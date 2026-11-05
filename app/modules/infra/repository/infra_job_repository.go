package repository

import (
	"context"
	"strings"
	"time"

	"github.com/goravel/framework/contracts/database/db"

	"eqray-go-starter/app/facades"
	baseTypes "eqray-go-starter/app/modules/base/types"
	"eqray-go-starter/app/modules/infra/model"
)

// JobFilter 是定时任务列表的筛选条件。
type JobFilter struct {
	// Keyword 名称 / 处理器标识模糊匹配
	Keyword string
	// Status 状态过滤，nil 表示全部
	Status *int8
}

// JobLogFilter 是执行日志的筛选条件。
type JobLogFilter struct {
	// JobID 任务 ID，0 表示全部
	JobID baseTypes.ID
	// Status 结果过滤，nil 表示全部
	Status *int8
}

// InfraJobRepository 是 infra_job / infra_job_log 表的仓储。
type InfraJobRepository struct{}

// NewInfraJobRepository 构造定时任务仓储。
func NewInfraJobRepository() *InfraJobRepository {
	return &InfraJobRepository{}
}

// applyJobFilter 给任务查询套上筛选条件；列表与计数复用，避免条件不一致。
// 入参 query 为查询构造器、filter 为筛选条件；返回加了条件的查询。
func (r *InfraJobRepository) applyJobFilter(query db.Query, filter JobFilter) db.Query {
	if keyword := strings.TrimSpace(filter.Keyword); keyword != "" {
		// 名称与处理器标识拼一起匹配，避免 OR 与其他 AND 条件混用时需要括号
		query = query.Where("CONCAT(name, ' ', handler) LIKE ?", "%"+keyword+"%")
	}
	if filter.Status != nil {
		query = query.Where("status", *filter.Status)
	}

	return query
}

// PaginateJobs 分页查询任务，启用优先、同状态按 ID 升序。
// 入参 filter 为筛选条件、pageNo/pageSize 为分页；返回任务列表与总数。
func (r *InfraJobRepository) PaginateJobs(ctx context.Context, filter JobFilter, pageNo, pageSize int) ([]model.InfraJobModel, int64, error) {
	total, err := r.applyJobFilter(facades.DB().WithContext(ctx).Table(model.InfraJobTable), filter).Count()
	if err != nil {
		return nil, 0, err
	}

	list := make([]model.InfraJobModel, 0)
	if total == 0 {
		return list, 0, nil
	}

	if err := r.applyJobFilter(facades.DB().WithContext(ctx).Table(model.InfraJobTable), filter).
		OrderBy("status desc").
		OrderBy("id").
		Offset(uint64((pageNo - 1) * pageSize)).
		Limit(uint64(pageSize)).
		Get(&list); err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

// FindJobByID 按主键查询任务；未命中返回 (nil, nil)。
// 入参 id 为任务 ID；返回任务或空。
func (r *InfraJobRepository) FindJobByID(ctx context.Context, id baseTypes.ID) (*model.InfraJobModel, error) {
	m := &model.InfraJobModel{}
	if err := facades.DB().WithContext(ctx).Table(model.InfraJobTable).Where("id", id).First(m); err != nil {
		return nil, nil
	}

	return m, nil
}

// FindJobByHandler 按处理器标识查询任务；未命中返回 (nil, nil)。
// 入参 handler 为处理器标识；返回任务或空。
func (r *InfraJobRepository) FindJobByHandler(ctx context.Context, handler string) (*model.InfraJobModel, error) {
	m := &model.InfraJobModel{}
	if err := facades.DB().WithContext(ctx).Table(model.InfraJobTable).Where("handler", handler).First(m); err != nil {
		return nil, nil
	}

	return m, nil
}

// InsertJob 写入一条任务。
// 入参 m 为任务模型；返回写库错误。
func (r *InfraJobRepository) InsertJob(ctx context.Context, m *model.InfraJobModel) error {
	_, err := facades.DB().WithContext(ctx).Table(model.InfraJobTable).Insert(map[string]any{
		"id":            m.ID,
		"name":          m.Name,
		"handler":       m.Handler,
		"cron":          m.Cron,
		"status":        m.Status,
		"remark":        m.Remark,
		"next_run_at":   m.NextRunAt,
		"last_status":   m.LastStatus,
		"last_duration": 0,
		"last_output":   "",
		"creator":       m.Creator,
		"created_at":    m.CreatedAt,
		"updater":       m.Updater,
		"updated_at":    m.UpdatedAt,
	})

	return err
}

// UpdateJobFields 按 ID 更新指定字段（自动补 updater / updated_at）。
// 入参 id 为任务 ID、data 为列名 → 值、updater 为操作人；返回写库错误。
func (r *InfraJobRepository) UpdateJobFields(ctx context.Context, id baseTypes.ID, data map[string]any, updater string) error {
	if data == nil {
		data = make(map[string]any)
	}
	data["updater"] = updater
	data["updated_at"] = baseTypes.Now()

	_, err := facades.DB().WithContext(ctx).Table(model.InfraJobTable).Where("id", id).Update(data)

	return err
}

// DeleteJob 删除任务及其执行日志（日志按任务级联清掉，避免孤儿数据）。
// 入参 id 为任务 ID；返回写库错误。
func (r *InfraJobRepository) DeleteJob(ctx context.Context, id baseTypes.ID) error {
	return facades.DB().WithContext(ctx).Transaction(func(tx db.Tx) error {
		if _, err := tx.Table(model.InfraJobLogTable).Where("job_id", id).Delete(); err != nil {
			return err
		}

		_, err := tx.Table(model.InfraJobTable).Where("id", id).Delete()

		return err
	})
}

// ListEnabled 查询全部启用任务，供调度器逐分钟匹配。
// 入参 ctx 为请求上下文；返回启用任务列表。
func (r *InfraJobRepository) ListEnabled(ctx context.Context) ([]model.InfraJobModel, error) {
	list := make([]model.InfraJobModel, 0)
	if err := facades.DB().WithContext(ctx).Table(model.InfraJobTable).
		Where("status", model.InfraJobStatusEnabled).
		Get(&list); err != nil {
		return nil, err
	}

	return list, nil
}

// InsertLog 写入一条执行日志。
// 入参 m 为日志模型；返回写库错误。
func (r *InfraJobRepository) InsertLog(ctx context.Context, m *model.InfraJobLogModel) error {
	_, err := facades.DB().WithContext(ctx).Table(model.InfraJobLogTable).Insert(map[string]any{
		"id":          m.ID,
		"job_id":      m.JobID,
		"job_name":    m.JobName,
		"handler":     m.Handler,
		"trigger":     m.Trigger,
		"status":      m.Status,
		"started_at":  m.StartedAt,
		"finished_at": m.FinishedAt,
		"duration":    m.Duration,
		"output":      m.Output,
		"error":       m.Error,
		"created_at":  m.CreatedAt,
	})

	return err
}

// PaginateLogs 分页查询执行日志，按开始时间倒序。
// 入参 filter 为筛选条件、pageNo/pageSize 为分页；返回日志列表与总数。
func (r *InfraJobRepository) PaginateLogs(ctx context.Context, filter JobLogFilter, pageNo, pageSize int) ([]model.InfraJobLogModel, int64, error) {
	apply := func(query db.Query) db.Query {
		if filter.JobID > 0 {
			query = query.Where("job_id", filter.JobID)
		}
		if filter.Status != nil {
			query = query.Where("status", *filter.Status)
		}

		return query
	}

	total, err := apply(facades.DB().WithContext(ctx).Table(model.InfraJobLogTable)).Count()
	if err != nil {
		return nil, 0, err
	}

	list := make([]model.InfraJobLogModel, 0)
	if total == 0 {
		return list, 0, nil
	}

	if err := apply(facades.DB().WithContext(ctx).Table(model.InfraJobLogTable)).
		OrderBy("started_at desc").
		Offset(uint64((pageNo - 1) * pageSize)).
		Limit(uint64(pageSize)).
		Get(&list); err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

// PurgeLogsBefore 清理早于指定时间的执行日志（默认保留 30 天）。
// 入参 before 为保留边界；返回删除条数。
func (r *InfraJobRepository) PurgeLogsBefore(ctx context.Context, before time.Time) (int64, error) {
	res, err := facades.DB().WithContext(ctx).Table(model.InfraJobLogTable).
		Where("started_at", "<", before).
		Delete()
	if err != nil {
		return 0, err
	}

	return res.RowsAffected, nil
}
