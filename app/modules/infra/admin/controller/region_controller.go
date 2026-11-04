package controller

import (
	"errors"

	"github.com/goravel/framework/contracts/http"

	baseErrs "eqray-go-starter/app/modules/base/errs"
	"eqray-go-starter/app/modules/base/param"
	"eqray-go-starter/app/modules/base/result"
	infraService "eqray-go-starter/app/modules/infra/service"
)

// asBusinessError 把非业务异常转成带真实原因的 400：后台运维操作需要看到失败细节，
// 而不是统一的「系统异常」。业务异常原样返回。
// 入参 err 为服务层错误、prefix 为提示前缀；返回可直接交给 result.Fail 的错误。
func asBusinessError(err error, prefix string) error {
	if err == nil {
		return nil
	}

	var businessErr *baseErrs.BusinessError
	if errors.As(err, &businessErr) {
		return err
	}

	return baseErrs.WithMsg(baseErrs.BadRequest, prefix+"："+err.Error())
}

// RegionController 提供行政区划的同步入口。
// 数据一次性落库后，C 端只查库（见 infra/client/controller/region_controller.go），不再实时调高德。
type RegionController struct {
	service *infraService.RegionService
}

// NewRegionController 构造行政区划控制器。
func NewRegionController() *RegionController {
	return &RegionController{service: infraService.NewRegionService()}
}

// Index 分页查询省市区数据，带各层级条数与最后同步时间。
// 入参：keyword（名称或编码）、level（province/city/district）、parentAdcode、pageNo、pageSize；
// 返回 { list, total, stats }。
func (c *RegionController) Index(ctx http.Context) http.Response {
	page := param.Page{
		PageNo:   ctx.Request().QueryInt("pageNo", 1),
		PageSize: ctx.Request().QueryInt("pageSize", param.DefaultPageSize),
	}
	if err := page.Normalize(); err != nil {
		return result.Fail(ctx, err)
	}

	res, err := c.service.Page(ctx.Context(), infraService.RegionListQuery{
		Keyword:      ctx.Request().Query("keyword"),
		Level:        ctx.Request().Query("level"),
		ParentAdcode: ctx.Request().Query("parentAdcode"),
		PageNo:       page.PageNo,
		PageSize:     page.PageSize,
	})
	if err != nil {
		return result.Fail(ctx, asBusinessError(err, "查询省市县失败"))
	}

	return result.Success(ctx, res)
}

// Sync 从高德拉取全国省市区并整表刷新，返回写入条数。
// 入参：无（后端读取参数配置 map.amap.serverKey）；返回 { count: 条数 }。
func (c *RegionController) Sync(ctx http.Context) http.Response {
	count, err := c.service.SyncFromAMap(ctx.Context())
	if err != nil {
		return result.Fail(ctx, asBusinessError(err, "同步行政区划失败"))
	}

	return result.Success(ctx, map[string]any{"count": count})
}
