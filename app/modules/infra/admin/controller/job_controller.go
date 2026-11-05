package controller

import (
	"strconv"

	"github.com/goravel/framework/contracts/http"

	"eqray-go-starter/app/modules/base/param"
	"eqray-go-starter/app/modules/base/result"
	"eqray-go-starter/app/modules/base/support"
	"eqray-go-starter/app/modules/base/types"
	"eqray-go-starter/app/modules/infra/admin/request"
	infraService "eqray-go-starter/app/modules/infra/service"
)

// JobController 提供定时任务的配置、手动执行与日志查询。
// 任务逻辑由代码注册，这里只改「跑不跑、多久跑一次」。
type JobController struct {
	service *infraService.JobService
}

// NewJobController 构造定时任务控制器。
func NewJobController() *JobController {
	return &JobController{service: infraService.NewJobService()}
}

// Index 分页查询任务。
// 入参：keyword、status、pageNo、pageSize；返回 list / total。
func (c *JobController) Index(ctx http.Context) http.Response {
	page := param.Page{
		PageNo:   ctx.Request().QueryInt("pageNo", 1),
		PageSize: ctx.Request().QueryInt("pageSize", param.DefaultPageSize),
	}
	if err := page.Normalize(); err != nil {
		return result.Fail(ctx, err)
	}

	status := parseInt8Ptr(ctx.Request().Query("status"))
	list, total, err := c.service.Page(ctx.Context(), infraService.JobListQuery{
		Keyword:  ctx.Request().Query("keyword"),
		Status:   status,
		PageNo:   page.PageNo,
		PageSize: page.PageSize,
	})
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Page(ctx, list, total)
}

// Handlers 返回代码里注册的处理器清单，供新增任务时选择。
// 入参：无；返回处理器列表。
func (c *JobController) Handlers(ctx http.Context) http.Response {
	return result.Success(ctx, c.service.Handlers())
}

// Store 新增任务。
// 入参：name、handler、cron、status、remark；返回新任务 ID。
func (c *JobController) Store(ctx http.Context) http.Response {
	var req request.JobSaveRequest
	if err := param.Bind(ctx, &req); err != nil {
		return result.Fail(ctx, err)
	}

	id, err := c.service.Create(ctx.Context(), infraService.JobSaveInput{
		Name:    req.Name,
		Handler: req.Handler,
		Cron:    req.Cron,
		Status:  req.Status,
		Remark:  req.Remark,
	}, operatorOf(ctx))
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, map[string]any{"id": id})
}

// Update 编辑任务（处理器标识不可改）。
// 入参：name、cron、status、remark；返回 null。
func (c *JobController) Update(ctx http.Context) http.Response {
	id, err := param.PathID(ctx)
	if err != nil {
		return result.Fail(ctx, err)
	}

	var req request.JobSaveRequest
	if err := param.Bind(ctx, &req); err != nil {
		return result.Fail(ctx, err)
	}

	if err := c.service.Update(ctx.Context(), id, infraService.JobSaveInput{
		Name:   req.Name,
		Cron:   req.Cron,
		Status: req.Status,
		Remark: req.Remark,
	}, operatorOf(ctx)); err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, nil)
}

// Toggle 启用 / 停用任务。
// 入参：路径 id、status；返回 null。
func (c *JobController) Toggle(ctx http.Context) http.Response {
	id, err := param.PathID(ctx)
	if err != nil {
		return result.Fail(ctx, err)
	}

	var req request.JobStatusRequest
	if err := param.Bind(ctx, &req); err != nil {
		return result.Fail(ctx, err)
	}

	if err := c.service.Toggle(ctx.Context(), id, req.Status, operatorOf(ctx)); err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, nil)
}

// Destroy 删除任务及其执行日志。
// 入参：路径 id；返回 null。
func (c *JobController) Destroy(ctx http.Context) http.Response {
	id, err := param.PathID(ctx)
	if err != nil {
		return result.Fail(ctx, err)
	}

	if err := c.service.Delete(ctx.Context(), id); err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, nil)
}

// Run 手动执行一次任务（同步），返回本次执行日志。
// 入参：路径 id；返回 { log }。
func (c *JobController) Run(ctx http.Context) http.Response {
	id, err := param.PathID(ctx)
	if err != nil {
		return result.Fail(ctx, err)
	}

	logModel, err := c.service.Run(ctx.Context(), id, operatorOf(ctx))
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, logModel)
}

// Logs 分页查询执行日志。
// 入参：jobId、status、pageNo、pageSize；返回 list / total。
func (c *JobController) Logs(ctx http.Context) http.Response {
	page := param.Page{
		PageNo:   ctx.Request().QueryInt("pageNo", 1),
		PageSize: ctx.Request().QueryInt("pageSize", param.DefaultPageSize),
	}
	if err := page.Normalize(); err != nil {
		return result.Fail(ctx, err)
	}

	list, total, err := c.service.Logs(ctx.Context(), infraService.JobLogQuery{
		JobID:    types.ID(ctx.Request().QueryInt("jobId", 0)),
		Status:   parseInt8Ptr(ctx.Request().Query("status")),
		PageNo:   page.PageNo,
		PageSize: page.PageSize,
	})
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Page(ctx, list, total)
}

// operatorOf 取当前登录账号作为操作人，未登录（调度器调用）记 system。
// 入参 ctx 为请求上下文；返回操作人标识。
func operatorOf(ctx http.Context) string {
	if uid := support.UserID(ctx.Context()); uid != 0 {
		return types.ID(uid).String()
	}

	return "system"
}

// parseInt8Ptr 把查询串转成 int8 指针，空串或非法值返回 nil（表示不过滤）。
// 入参 value 为查询串；返回状态指针。
func parseInt8Ptr(value string) *int8 {
	if value == "" {
		return nil
	}

	n, err := strconv.Atoi(value)
	if err != nil {
		return nil
	}

	v := int8(n)

	return &v
}
