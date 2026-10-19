package controller

import (
	"github.com/goravel/framework/contracts/http"

	"eqray-go-starter/app/modules/base/result"
	"eqray-go-starter/app/modules/system/service"
)

// DashboardController 提供控制台首页的聚合统计。
type DashboardController struct {
	service *service.DashboardService
}

func NewDashboardController() *DashboardController {
	return &DashboardController{service: service.NewDashboardService()}
}

// Overview GET /api/v1/system/dashboard/overview
// 控制台是所有登录用户的默认首页，只挂登录态，不挂具体权限点。
func (c *DashboardController) Overview(ctx http.Context) http.Response {
	data, err := c.service.Overview(ctx.Context())
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, data)
}
