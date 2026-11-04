package controller

import (
	"github.com/goravel/framework/contracts/http"

	"eqray-go-starter/app/modules/base/param"
	"eqray-go-starter/app/modules/base/result"
	baseTypes "eqray-go-starter/app/modules/base/types"
	"eqray-go-starter/app/modules/pay/admin/request"
	payModel "eqray-go-starter/app/modules/pay/model"
	payService "eqray-go-starter/app/modules/pay/service"
)

// PayRefundController 提供退款单的分页查询、详情与发起退款。
type PayRefundController struct {
	service *payService.PayRefundService
}

// NewPayRefundController 构造退款控制器。
func NewPayRefundController() *PayRefundController {
	return &PayRefundController{service: payService.NewPayRefundService()}
}

// Index 分页查询退款单。
// 入参：merchantRefundNo、merchantOrderNo、status、pageNo、pageSize（查询串）；返回退款单列表与总数。
func (c *PayRefundController) Index(ctx http.Context) http.Response {
	page := param.Page{
		PageNo:   ctx.Request().QueryInt("pageNo", 1),
		PageSize: ctx.Request().QueryInt("pageSize", param.DefaultPageSize),
	}
	if err := page.Normalize(); err != nil {
		return result.Fail(ctx, err)
	}

	q := payService.PayRefundListQuery{
		MerchantRefundNo: ctx.Request().Query("merchantRefundNo"),
		MerchantOrderNo:  ctx.Request().Query("merchantOrderNo"),
		PageNo:           page.PageNo,
		PageSize:         page.PageSize,
	}
	if status := ctx.Request().QueryInt("status", -1); status >= 0 {
		v := payModel.PayRefundStatus(status)
		q.Status = &v
	}

	list, total, err := c.service.Page(ctx.Context(), q)
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Page(ctx, list, total)
}

// Show 查询退款单详情。
// 入参：路径 id；返回退款单详情，不存在返回业务错误码。
func (c *PayRefundController) Show(ctx http.Context) http.Response {
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

// Store 发起退款：无人工审核，提交后直接调渠道；支付宝同步返回终态，微信受理中由定时任务收敛。
// 入参：orderId、refundPrice（分）、reason；返回退款单 ID。
func (c *PayRefundController) Store(ctx http.Context) http.Response {
	var req request.PayRefundCreateRequest
	if err := param.Bind(ctx, &req); err != nil {
		return result.Fail(ctx, err)
	}

	id, err := c.service.Create(ctx.Context(), payService.RefundCreateInput{
		OrderID:     baseTypes.ID(req.OrderID),
		RefundPrice: req.RefundPrice,
		Reason:      req.Reason,
	})
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, map[string]any{"id": id})
}
