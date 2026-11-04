package controller

import (
	"github.com/goravel/framework/contracts/http"

	"eqray-go-starter/app/modules/base/result"
	infraService "eqray-go-starter/app/modules/infra/service"
)

// RegionController 提供行政区划的同步入口。
// 数据一次性落库后，C 端只查库（见 infra/client/controller/region_controller.go），不再实时调高德。
type RegionController struct {
	service *infraService.RegionService
}

// NewRegionController 构造行政区划控制器。
func NewRegionController() *RegionController {
	return &RegionController{service: infraService.NewRegionService()}
}

// Sync 从高德拉取全国省市区并整表刷新，返回写入条数。
// 入参：无（后端读取参数配置 map.amap.serverKey）；返回 { count: 条数 }。
func (c *RegionController) Sync(ctx http.Context) http.Response {
	count, err := c.service.SyncFromAMap(ctx.Context())
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, map[string]any{"count": count})
}
