package service

import (
	"context"
	"strings"
	"time"

	"github.com/robfig/cron/v3"

	"eqray-go-starter/app/facades"
	baseErrs "eqray-go-starter/app/modules/base/errs"
	"eqray-go-starter/app/modules/base/support"
	"eqray-go-starter/app/modules/base/types"
	infraJob "eqray-go-starter/app/modules/infra/job"
	"eqray-go-starter/app/modules/infra/model"
	"eqray-go-starter/app/modules/infra/repository"
)

const (
	// jobRunTimeout 是单次任务执行的超时时间，避免异常任务长期占用调度器
	jobRunTimeout = 5 * time.Minute
	// jobLogKeepDays 是执行日志保留天数
	jobLogKeepDays = 30
	// jobOutputLimit 是写入库的输出摘要字符上限
	jobOutputLimit = 4000
)

// JobHandlerItem 是后台「新增任务」可选的处理器。
type JobHandlerItem struct {
	// Key 处理器标识
	Key string `json:"key"`
	// Name 默认名称
	Name string `json:"name"`
	// Description 说明
	Description string `json:"description"`
	// DefaultCron 默认 cron 表达式
	DefaultCron string `json:"defaultCron"`
}

// JobListQuery 是任务列表查询条件。
type JobListQuery struct {
	// Keyword 名称 / 处理器标识
	Keyword string
	// Status 状态过滤，nil 表示全部
	Status *int8
	// PageNo / PageSize 分页
	PageNo   int
	PageSize int
}

// JobSaveInput 是新增 / 编辑任务的入参。
type JobSaveInput struct {
	// Name 任务名称
	Name string
	// Handler 处理器标识
	Handler string
	// Cron cron 表达式（5 段）
	Cron string
	// Status 状态：0 停用 / 1 启用
	Status int8
	// Remark 备注
	Remark string
}

// JobLogQuery 是执行日志查询条件。
type JobLogQuery struct {
	// JobID 任务 ID，0 表示全部
	JobID types.ID
	// Status 结果过滤，nil 表示全部
	Status *int8
	// PageNo / PageSize 分页
	PageNo   int
	PageSize int
}

// JobService 负责定时任务的配置、执行与调度分发。
// 任务逻辑不在库里，只存处理器标识；启停与频率改库即生效（下一分钟调度周期）。
type JobService struct {
	jobs *repository.InfraJobRepository
}

// NewJobService 构造定时任务服务。
func NewJobService() *JobService {
	return &JobService{jobs: repository.NewInfraJobRepository()}
}

// Handlers 返回代码里注册的全部处理器，供后台下拉选择。
// 入参无；返回处理器列表。
func (s *JobService) Handlers() []JobHandlerItem {
	defs := infraJob.All()
	items := make([]JobHandlerItem, 0, len(defs))
	for _, def := range defs {
		items = append(items, JobHandlerItem{
			Key:         def.Key,
			Name:        def.Name,
			Description: def.Description,
			DefaultCron: def.DefaultCron,
		})
	}

	return items
}

// Page 分页查询任务。
// 入参 q 为查询条件；返回任务列表与总数。
func (s *JobService) Page(ctx context.Context, q JobListQuery) ([]model.InfraJobModel, int64, error) {
	pageNo, pageSize := normalizePage(q.PageNo, q.PageSize)

	return s.jobs.PaginateJobs(ctx, repository.JobFilter{
		Keyword: q.Keyword,
		Status:  q.Status,
	}, pageNo, pageSize)
}

// Logs 分页查询执行日志。
// 入参 q 为查询条件；返回日志列表与总数。
func (s *JobService) Logs(ctx context.Context, q JobLogQuery) ([]model.InfraJobLogModel, int64, error) {
	pageNo, pageSize := normalizePage(q.PageNo, q.PageSize)

	return s.jobs.PaginateLogs(ctx, repository.JobLogFilter{
		JobID:  q.JobID,
		Status: q.Status,
	}, pageNo, pageSize)
}

// Create 新增任务：校验处理器存在、cron 合法、处理器未被占用。
// 入参 in 为任务配置、operator 为操作人；返回新任务 ID。
func (s *JobService) Create(ctx context.Context, in JobSaveInput, operator string) (types.ID, error) {
	if _, ok := infraJob.Get(in.Handler); !ok {
		return 0, baseErrs.WithMsg(baseErrs.BadRequest, "处理器不存在："+in.Handler)
	}

	existing, err := s.jobs.FindJobByHandler(ctx, in.Handler)
	if err != nil {
		return 0, err
	}
	if existing != nil {
		return 0, baseErrs.WithMsg(baseErrs.BadRequest, "该处理器已有任务，请直接编辑频率")
	}

	nextRunAt, err := computeNextRunAt(in.Cron, in.Status)
	if err != nil {
		return 0, err
	}

	now := types.Now()
	id := support.NewID()
	m := &model.InfraJobModel{
		ID:         id,
		Name:       strings.TrimSpace(in.Name),
		Handler:    in.Handler,
		Cron:       strings.TrimSpace(in.Cron),
		Status:     in.Status,
		Remark:     strings.TrimSpace(in.Remark),
		NextRunAt:  nextRunAt,
		LastStatus: model.InfraJobResultSuccess,
		Creator:    operator,
		CreatedAt:  now,
		Updater:    operator,
		UpdatedAt:  now,
	}
	if err := s.jobs.InsertJob(ctx, m); err != nil {
		return 0, err
	}

	return id, nil
}

// Update 编辑任务：可改名称、cron、状态与备注，处理器标识创建后不可改。
// 入参 id 为任务 ID、in 为任务配置、operator 为操作人；返回业务错误或写库错误。
func (s *JobService) Update(ctx context.Context, id types.ID, in JobSaveInput, operator string) error {
	m, err := s.findJob(ctx, id)
	if err != nil {
		return err
	}

	nextRunAt, err := computeNextRunAt(in.Cron, in.Status)
	if err != nil {
		return err
	}

	return s.jobs.UpdateJobFields(ctx, m.ID, map[string]any{
		"name":        strings.TrimSpace(in.Name),
		"cron":        strings.TrimSpace(in.Cron),
		"status":      in.Status,
		"remark":      strings.TrimSpace(in.Remark),
		"next_run_at": nextRunAt,
	}, operator)
}

// Toggle 启用 / 停用任务，重新计算下次执行时间。
// 入参 id 为任务 ID、status 为目标状态、operator 为操作人；返回业务错误或写库错误。
func (s *JobService) Toggle(ctx context.Context, id types.ID, status int8, operator string) error {
	m, err := s.findJob(ctx, id)
	if err != nil {
		return err
	}
	if status != model.InfraJobStatusEnabled && status != model.InfraJobStatusDisabled {
		return baseErrs.WithMsg(baseErrs.BadRequest, "状态取值不合法")
	}

	nextRunAt, err := computeNextRunAt(m.Cron, status)
	if err != nil {
		return err
	}

	return s.jobs.UpdateJobFields(ctx, m.ID, map[string]any{
		"status":      status,
		"next_run_at": nextRunAt,
	}, operator)
}

// Delete 删除任务及其日志。
// 入参 id 为任务 ID；返回业务错误或写库错误。
func (s *JobService) Delete(ctx context.Context, id types.ID) error {
	m, err := s.findJob(ctx, id)
	if err != nil {
		return err
	}

	return s.jobs.DeleteJob(ctx, m.ID)
}

// Run 手动执行一次任务（同步），无论成败都会写一条 manual 日志。
// 入参 id 为任务 ID、operator 为操作人；返回本次执行日志。
func (s *JobService) Run(ctx context.Context, id types.ID, operator string) (*model.InfraJobLogModel, error) {
	m, err := s.findJob(ctx, id)
	if err != nil {
		return nil, err
	}

	return s.runJob(ctx, m, model.InfraJobTriggerManual, operator), nil
}

// DispatchDue 由调度器每分钟调用：执行到点的启用任务，并顺带清理过期日志。
// 入参 ctx 为调度上下文；无返回值（失败只记日志，不打断后续任务）。
func (s *JobService) DispatchDue(ctx context.Context) {
	if !facades.Schema().HasTable(model.InfraJobTable) {
		return
	}

	now := time.Now()
	list, err := s.jobs.ListEnabled(ctx)
	if err != nil {
		facades.Log().Errorf("[定时任务] 查询待执行任务失败：%v", err)
		return
	}

	for index := range list {
		item := &list[index]
		if item.NextRunAt != nil && item.NextRunAt.Time().After(now) {
			continue
		}

		// 从未计算过下次执行时间（历史数据）时，只补算不执行，避免集中补跑
		if item.NextRunAt == nil {
			if next, err := nextRunAt(item.Cron, now); err == nil {
				_ = s.jobs.UpdateJobFields(ctx, item.ID, map[string]any{"next_run_at": next}, "system")
			}
			continue
		}

		s.runJob(ctx, item, model.InfraJobTriggerCron, "system")
	}

	// 每天 04:00 清理超过保留期的执行日志（调度器每分钟触发，命中一次即可）
	if now.Hour() == 4 && now.Minute() == 0 {
		if deleted, err := s.jobs.PurgeLogsBefore(ctx, now.AddDate(0, 0, -jobLogKeepDays)); err == nil && deleted > 0 {
			facades.Log().Infof("[定时任务] 清理执行日志 %d 条", deleted)
		}
	}
}

// runJob 执行任务并落库：写日志、更新任务的上次结果与下次执行时间。
// 入参 m 为任务、trigger 为触发方式、operator 为操作人；返回日志模型。
func (s *JobService) runJob(ctx context.Context, m *model.InfraJobModel, trigger, operator string) *model.InfraJobLogModel {
	startedAt := time.Now()

	output, runErr := s.invoke(ctx, m.Handler, trigger)

	finishedAt := time.Now()
	duration := int(finishedAt.Sub(startedAt).Milliseconds())
	status := model.InfraJobResultSuccess
	errText := ""
	if runErr != nil {
		status = model.InfraJobResultFailed
		errText = runErr.Error()
	}

	logModel := &model.InfraJobLogModel{
		ID:         support.NewID(),
		JobID:      m.ID,
		JobName:    m.Name,
		Handler:    m.Handler,
		Trigger:    trigger,
		Status:     status,
		StartedAt:  types.DateTime(startedAt),
		FinishedAt: types.DateTime(finishedAt),
		Duration:   duration,
		Output:     truncateOutput(output),
		Error:      errText,
		CreatedAt:  types.DateTime(finishedAt),
	}
	if err := s.jobs.InsertLog(ctx, logModel); err != nil {
		facades.Log().Errorf("[定时任务] 写执行日志失败：%v", err)
	}

	next, _ := nextRunAt(m.Cron, finishedAt)
	_ = s.jobs.UpdateJobFields(ctx, m.ID, map[string]any{
		"last_run_at":   finishedAt,
		"last_status":   status,
		"last_duration": duration,
		"last_output":   truncateOutput(output),
		"next_run_at":   next,
	}, operator)

	if runErr != nil {
		facades.Log().Errorf("[定时任务] %s 执行失败：%v", m.Name, runErr)
	} else if output != "" {
		facades.Log().Infof("[定时任务] %s 执行完成：%s", m.Name, output)
	}

	return logModel
}

// invoke 按处理器标识取注册逻辑并执行，带超时兜底。
// 入参 ctx 为父上下文、handler 为处理器标识、trigger 为触发方式；返回摘要与错误。
func (s *JobService) invoke(ctx context.Context, handler, trigger string) (string, error) {
	def, ok := infraJob.Get(handler)
	if !ok {
		return "", baseErrs.WithMsg(baseErrs.BadRequest, "处理器未注册："+handler)
	}

	runCtx, cancel := context.WithTimeout(ctx, jobRunTimeout)
	defer cancel()

	type result struct {
		output string
		err    error
	}
	done := make(chan result, 1)
	go func() {
		defer func() {
			// 处理器 panic 不能把调度器带走
			if r := recover(); r != nil {
				done <- result{err: baseErrs.WithMsg(baseErrs.BadRequest, "任务 panic")}
			}
		}()

		output, err := def.Handler(runCtx)
		done <- result{output: output, err: err}
	}()

	select {
	case r := <-done:
		return r.output, r.err
	case <-runCtx.Done():
		return "", baseErrs.WithMsg(baseErrs.BadRequest, "任务执行超时（"+jobRunTimeout.String()+"）")
	}
}

// findJob 按 ID 查任务，不存在返回业务错误。
// 入参 id 为任务 ID；返回任务模型。
func (s *JobService) findJob(ctx context.Context, id types.ID) (*model.InfraJobModel, error) {
	m, err := s.jobs.FindJobByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if m == nil {
		return nil, baseErrs.WithMsg(baseErrs.BadRequest, "任务不存在")
	}

	return m, nil
}

// computeNextRunAt 按 cron 与状态计算下次执行时间；停用时返回 nil。
// 入参 expr 为 cron 表达式、status 为状态；返回时间指针或 cron 解析错误。
func computeNextRunAt(expr string, status int8) (*types.DateTime, error) {
	if status != model.InfraJobStatusEnabled {
		return nil, nil
	}

	next, err := nextRunAt(expr, time.Now())
	if err != nil {
		return nil, err
	}

	return &next, nil
}

// nextRunAt 计算 cron 表达式在 after 之后的下一次执行时间。
// 入参 expr 为 cron 表达式、after 为基准时间；返回该时间点的 DateTime 或解析错误。
func nextRunAt(expr string, after time.Time) (types.DateTime, error) {
	expr = strings.TrimSpace(expr)
	if expr == "" {
		return types.DateTime{}, baseErrs.WithMsg(baseErrs.BadRequest, "cron 表达式不能为空")
	}

	schedule, err := cron.ParseStandard(expr)
	if err != nil {
		return types.DateTime{}, baseErrs.WithMsg(baseErrs.BadRequest, "cron 表达式不合法："+err.Error())
	}

	next := schedule.Next(after)
	if next.IsZero() {
		return types.DateTime{}, baseErrs.WithMsg(baseErrs.BadRequest, "cron 表达式无法计算下次执行时间")
	}

	return types.DateTime(next), nil
}

// truncateOutput 截断执行输出，避免超长文本占满日志表。
// 入参 value 为原始输出；返回截断后的字符串。
func truncateOutput(value string) string {
	runes := []rune(value)
	if len(runes) <= jobOutputLimit {
		return value
	}

	return string(runes[:jobOutputLimit]) + "…"
}

// normalizePage 规范化分页参数：页码最小 1、页大小默认 20 且不超过 200。
// 入参 pageNo/pageSize 为原始分页；返回可用的分页值。
func normalizePage(pageNo, pageSize int) (int, int) {
	if pageNo <= 0 {
		pageNo = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 200 {
		pageSize = 200
	}

	return pageNo, pageSize
}
