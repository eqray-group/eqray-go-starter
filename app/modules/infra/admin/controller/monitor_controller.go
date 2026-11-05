package controller

import (
	"github.com/goravel/framework/contracts/http"

	"eqray-go-starter/app/modules/base/result"
	infraService "eqray-go-starter/app/modules/infra/service"
)

// MonitorController 提供系统监控数据（进程、内存、依赖探测、请求统计与最近错误）。
type MonitorController struct {
	service *infraService.MonitorService
}

// NewMonitorController 构造监控控制器。
func NewMonitorController() *MonitorController {
	return &MonitorController{service: infraService.NewMonitorService()}
}

// Overview 采集一次监控数据。
// 入参：无；返回进程 / 内存 / DB / 缓存 / 24h 请求 / 最近错误。
func (c *MonitorController) Overview(ctx http.Context) http.Response {
	return result.Success(ctx, c.service.Overview(ctx.Context()))
}
