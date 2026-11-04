package controller

import (
	"github.com/goravel/framework/contracts/http"

	"eqray-go-starter/app/modules/base/param"
	"eqray-go-starter/app/modules/base/result"
	"eqray-go-starter/app/modules/base/types"
	payModel "eqray-go-starter/app/modules/pay/model"
	payService "eqray-go-starter/app/modules/pay/service"
)

// PayOrderController 提供支付订单的分页查询、详情、主动同步与关闭。
type PayOrderController struct {
	service *payService.PayOrderService
}

// NewPayOrderController 构造支付订单控制器。
func NewPayOrderController() *PayOrderController {
	return &PayOrderController{service: payService.NewPayOrderService()}
}

// Index 分页查询支付订单，支持单号、业务单号、渠道、状态与创建时间区间筛选。
// 入参：merchantOrderNo、bizType、bizOrderNo、channelCode、status、beginTime、endTime、
// pageNo、pageSize（查询串）；返回订单列表与总数。
func (c *PayOrderController) Index(ctx http.Context) http.Response {
	page := param.Page{
		PageNo:   ctx.Request().QueryInt("pageNo", 1),
		PageSize: ctx.Request().QueryInt("pageSize", param.DefaultPageSize),
	}
	if err := page.Normalize(); err != nil {
		return result.Fail(ctx, err)
	}

	q := payService.PayOrderListQuery{
		MerchantOrderNo: ctx.Request().Query("merchantOrderNo"),
		BizType:         ctx.Request().Query("bizType"),
		BizOrderNo:      ctx.Request().Query("bizOrderNo"),
		ChannelCode:     ctx.Request().Query("channelCode"),
		BeginTime:       ctx.Request().Query("beginTime"),
		EndTime:         ctx.Request().Query("endTime"),
		PageNo:          page.PageNo,
		PageSize:        page.PageSize,
	}
	if status := ctx.Request().QueryInt("status", -1); status >= 0 {
		v := payModel.PayOrderStatus(status)
		q.Status = &v
	}

	list, total, err := c.service.Page(ctx.Context(), q)
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Page(ctx, list, total)
}

// Show 查询支付订单详情。
// 入参：路径 id；返回订单详情，不存在返回业务错误码。
func (c *PayOrderController) Show(ctx http.Context) http.Response {
	id, err := param.PathID(ctx)
	if err != nil {
		return result.Fail(ctx, err)
	}

	m, err := c.service.Detail(ctx.Context(), id)
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, m)
}

// Sync 主动向渠道查单并同步本地状态，用于「渠道已收款但回调未达」的兜底处理。
// 入参：路径 id；返回同步后的订单。
func (c *PayOrderController) Sync(ctx http.Context) http.Response {
	id, err := param.PathID(ctx)
	if err != nil {
		return result.Fail(ctx, err)
	}

	m, err := c.service.Sync(ctx.Context(), id)
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, m)
}

// Close 关闭支付订单：先关渠道侧订单再关本地，仅待支付 / 支付中可关。
// 入参：路径 id；返回 null。
func (c *PayOrderController) Close(ctx http.Context) http.Response {
	id, err := param.PathID(ctx)
	if err != nil {
		return result.Fail(ctx, err)
	}

	if err := c.service.Close(ctx.Context(), types.ID(id)); err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, nil)
}
