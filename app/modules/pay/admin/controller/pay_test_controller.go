package controller

import (
	"github.com/goravel/framework/contracts/http"

	"eqray-go-starter/app/modules/base/param"
	"eqray-go-starter/app/modules/base/result"
	"eqray-go-starter/app/modules/pay/admin/request"
	payService "eqray-go-starter/app/modules/pay/service"
)

// PayTestController 提供后台支付联调能力：创建 1 分钱测试单、查看状态、主动查单、模拟支付成功。
// 仅供联调使用，测试单的业务类型固定为 pay_test，不与任何业务表关联。
type PayTestController struct {
	orders *payService.PayOrderService
}

// NewPayTestController 构造支付联调控制器。
func NewPayTestController() *PayTestController {
	return &PayTestController{orders: payService.NewPayOrderService()}
}

// Store 创建联调测试单并发起支付（金额固定 1 分，钱包渠道会真实扣款）。
// 入参：userId、channelCode；返回支付单号、状态与前端拉起参数。
func (c *PayTestController) Store(ctx http.Context) http.Response {
	var req request.PayTestOrderRequest
	if err := param.Bind(ctx, &req); err != nil {
		return result.Fail(ctx, err)
	}

	created, err := c.orders.CreateTestOrder(ctx.Context(), payService.CreateTestOrderInput{
		UserID:      req.UserID,
		ChannelCode: req.ChannelCode,
	})
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, map[string]any{
		"id":              created.Order.ID,
		"merchantOrderNo": created.Order.MerchantOrderNo,
		"channelCode":     created.Order.ChannelCode,
		"price":           created.Order.Price,
		"status":          created.Order.Status,
		"payInfo":         created.PayInfo,
	})
}

// Batch 批量创建联调测试单：不传 channelCodes 时覆盖全部启用渠道。
// 单个渠道失败（如钱包余额不足）只体现在该行 Error 上，不影响其他渠道。
// 入参：userId、channelCodes（可选）；返回每个渠道的结果列表。
func (c *PayTestController) Batch(ctx http.Context) http.Response {
	var req request.PayTestBatchRequest
	if err := param.Bind(ctx, &req); err != nil {
		return result.Fail(ctx, err)
	}

	outcomes, err := c.orders.CreateTestOrders(ctx.Context(), req.UserID, req.ChannelCodes)
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, outcomes)
}

// Show 按支付单号查询联调单当前状态（前端轮询用）。
// 入参：路径 merchantOrderNo；返回订单详情。
func (c *PayTestController) Show(ctx http.Context) http.Response {
	m, err := c.orders.AdminOrder(ctx.Context(), ctx.Request().Route("merchantOrderNo"))
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, m)
}

// Sync 主动向渠道查单并同步状态，用于「用户已付款但回调未达」的兜底。
// 入参：路径 merchantOrderNo；返回同步后的订单。
func (c *PayTestController) Sync(ctx http.Context) http.Response {
	m, err := c.orders.SyncByMerchantOrderNo(ctx.Context(), ctx.Request().Route("merchantOrderNo"))
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, m)
}

// MockPay 对模拟渠道订单执行「模拟支付成功」。
// 入参：路径 merchantOrderNo；返回最新订单。
func (c *PayTestController) MockPay(ctx http.Context) http.Response {
	m, err := c.orders.MockPayByMerchantOrderNo(ctx.Context(), ctx.Request().Route("merchantOrderNo"))
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, m)
}
