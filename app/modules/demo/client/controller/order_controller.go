// Package controller 是演示业务模块的 C 端接入层：仅用于联调，上线可整体删除。
package controller

import (
	"github.com/goravel/framework/contracts/http"

	baseErrs "eqray-go-starter/app/modules/base/errs"
	"eqray-go-starter/app/modules/base/param"
	"eqray-go-starter/app/modules/base/result"
	"eqray-go-starter/app/modules/base/support"
	demoService "eqray-go-starter/app/modules/demo/service"
)

// DemoOrderController 提供演示订单的创建支付、列表与状态查询。
type DemoOrderController struct {
	service *demoService.DemoOrderService
}

// NewDemoOrderController 构造演示订单控制器。
func NewDemoOrderController() *DemoOrderController {
	return &DemoOrderController{service: demoService.NewDemoOrderService()}
}

// Store 创建演示订单并发起支付：金额服务端写死 1 分，渠道由前端指定。
// 入参：channelCode（请求体）；返回业务单号、支付单号与前端拉起参数。
func (c *DemoOrderController) Store(ctx http.Context) http.Response {
	channelCode := ctx.Request().Input("channelCode")
	if channelCode == "" {
		return result.Fail(ctx, baseErrs.WithMsg(baseErrs.BadRequest, "请选择支付方式"))
	}

	memberID := support.MemberID(ctx.Context())
	created, orderNo, err := c.service.CreateAndPay(ctx.Context(), memberID, ctx.Request().Ip(), channelCode)
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, map[string]any{
		"orderNo":         orderNo,
		"merchantOrderNo": created.Order.MerchantOrderNo,
		"payInfo":         created.PayInfo,
	})
}

// Index 分页查询本人的演示订单。
// 入参：pageNo、pageSize（查询串）；返回订单列表与总数。
func (c *DemoOrderController) Index(ctx http.Context) http.Response {
	page := param.Page{
		PageNo:   ctx.Request().QueryInt("pageNo", 1),
		PageSize: ctx.Request().QueryInt("pageSize", param.DefaultPageSize),
	}
	if err := page.Normalize(); err != nil {
		return result.Fail(ctx, err)
	}

	list, total, err := c.service.PageMine(ctx.Context(), support.MemberID(ctx.Context()), page.PageNo, page.PageSize)
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Page(ctx, list, total)
}

// Show 查询本人的演示订单状态（含支付回写结果）。
// 入参：路径 orderNo；返回订单详情。
func (c *DemoOrderController) Show(ctx http.Context) http.Response {
	m, err := c.service.Detail(ctx.Context(), ctx.Request().Route("orderNo"), support.MemberID(ctx.Context()))
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, m)
}
