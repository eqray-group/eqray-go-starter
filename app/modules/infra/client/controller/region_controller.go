package controller

import (
	"github.com/goravel/framework/contracts/http"

	"eqray-go-starter/app/modules/base/result"
	infraService "eqray-go-starter/app/modules/infra/service"
)

// RegionController 提供省市区数据查询，供前端级联懒加载。
// 数据来自 infra_region 表（后台一次性同步），接口只读、免登录、可放心加缓存。
type RegionController struct {
	regions *infraService.RegionService
}

// NewRegionController 构造行政区划查询控制器。
func NewRegionController() *RegionController {
	return &RegionController{regions: infraService.NewRegionService()}
}

// Index 查询某级区划的下级列表。
// 入参：parentAdcode（查询串，空表示查省份，根节点为 100000）；
// 返回 [{ adcode, name, level, leaf }]，leaf 为 true 表示该节点没有下级。
func (c *RegionController) Index(ctx http.Context) http.Response {
	list, err := c.regions.Children(ctx.Context(), ctx.Request().Query("parentAdcode"))
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, list)
}
