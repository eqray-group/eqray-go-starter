package controller

import (
	"github.com/goravel/framework/contracts/http"

	baseEnums "eqray-go-starter/app/modules/base/enums"
	"eqray-go-starter/app/modules/base/param"
	"eqray-go-starter/app/modules/base/result"
	"eqray-go-starter/app/modules/system/service"
)

// LogController 提供日志查询接口；登录日志数据归 base 的认证组件，操作日志归 system，均只读。
type LogController struct {
	service        *service.LogService
	operateLogImpl *service.OperateLogService
}

func NewLogController() *LogController {
	return &LogController{
		service:        service.NewLogService(),
		operateLogImpl: service.NewOperateLogService(),
	}
}

// Login GET /api/v1/system/logs/login
func (c *LogController) Login(ctx http.Context) http.Response {
	page := param.Page{
		PageNo:   ctx.Request().QueryInt("pageNo", 1),
		PageSize: ctx.Request().QueryInt("pageSize", param.DefaultPageSize),
	}
	if err := page.Normalize(); err != nil {
		return result.Fail(ctx, err)
	}

	q := service.LoginLogListQuery{
		Username: ctx.Request().Query("username"),
		PageNo:   page.PageNo,
		PageSize: page.PageSize,
	}
	if v := ctx.Request().QueryInt("type", -1); v > 0 {
		t := baseEnums.LoginTypeEnum(v)
		q.Type = &t
	}
	if v := ctx.Request().QueryInt("result", -1); v >= 0 {
		r := baseEnums.LoginResultEnum(v)
		q.Result = &r
	}

	list, total, err := c.service.LoginLogs(ctx.Context(), q)
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Page(ctx, list, total)
}

// Operate GET /api/v1/system/logs/operate
func (c *LogController) Operate(ctx http.Context) http.Response {
	page := param.Page{
		PageNo:   ctx.Request().QueryInt("pageNo", 1),
		PageSize: ctx.Request().QueryInt("pageSize", param.DefaultPageSize),
	}
	if err := page.Normalize(); err != nil {
		return result.Fail(ctx, err)
	}

	q := service.OperateLogListQuery{
		Type:     ctx.Request().Query("type"),
		SubType:  ctx.Request().Query("subType"),
		Keyword:  ctx.Request().Query("keyword"),
		PageNo:   page.PageNo,
		PageSize: page.PageSize,
	}
	if v := ctx.Request().QueryInt("success", -1); v >= 0 {
		success := int8(v)
		q.Success = &success
	}

	list, total, err := c.operateLogImpl.Page(ctx.Context(), q)
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Page(ctx, list, total)
}
