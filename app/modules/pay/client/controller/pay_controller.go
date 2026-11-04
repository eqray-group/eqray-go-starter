// Package controller 是 pay 模块 C 端接入层：会员令牌鉴权，身份一律取令牌。
package controller

import (
	"github.com/goravel/framework/contracts/http"

	"eqray-go-starter/app/modules/base/param"
	"eqray-go-starter/app/modules/base/result"
	"eqray-go-starter/app/modules/base/support"
	"eqray-go-starter/app/modules/pay/client/request"
	payService "eqray-go-starter/app/modules/pay/service"
)

// PayController 提供 C 端支付能力：可用渠道、创建支付、轮询支付结果。
type PayController struct {
	orderService *payService.PayOrderService
}

// NewPayController 构造 C 端支付控制器。
func NewPayController() *PayController {
	return &PayController{orderService: payService.NewPayOrderService()}
}

// Channels 查询启用中的支付渠道，供前端选择支付方式。
// 入参：无；返回渠道列表（不含任何凭据信息）。
func (c *PayController) Channels(ctx http.Context) http.Response {
	list, err := payService.NewPayChannelService().EnabledChannels(ctx.Context())
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, list)
}

// Store 创建支付：金额与标题由业务模块按业务单号反查得出，请求体不接受金额。
// 入参：bizType、bizOrderNo、channelCode、returnUrl（可选）；返回支付单号与前端拉起参数。
func (c *PayController) Store(ctx http.Context) http.Response {
	var req request.PayCreateRequest
	if err := param.Bind(ctx, &req); err != nil {
		return result.Fail(ctx, err)
	}

	memberID := support.MemberID(ctx.Context())
	created, err := c.orderService.CreateOrderForClient(
		ctx.Context(),
		req.BizType, req.BizOrderNo, req.ChannelCode,
		memberID,
		ctx.Request().Ip(), req.ReturnURL,
	)
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, map[string]any{
		"merchantOrderNo": created.Order.MerchantOrderNo,
		"status":          created.Order.Status,
		"price":           created.Order.Price,
		"payInfo":         created.PayInfo,
	})
}

// Show 查询支付结果（前端轮询）：支付中订单会顺带向渠道主动查单，避免回调延迟卡住前端。
// 入参：路径 merchantOrderNo；返回订单状态与渠道侧信息。
func (c *PayController) Show(ctx http.Context) http.Response {
	merchantOrderNo := ctx.Request().Route("merchantOrderNo")

	m, err := c.orderService.SyncUserOrder(ctx.Context(), merchantOrderNo, support.MemberID(ctx.Context()))
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, m)
}
