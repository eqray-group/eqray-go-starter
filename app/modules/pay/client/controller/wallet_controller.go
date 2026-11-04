package controller

import (
	"github.com/goravel/framework/contracts/http"

	"eqray-go-starter/app/modules/base/param"
	"eqray-go-starter/app/modules/base/result"
	"eqray-go-starter/app/modules/base/support"
	payService "eqray-go-starter/app/modules/pay/service"
)

// PayWalletController 提供 C 端钱包查询与模拟支付。
type PayWalletController struct {
	wallets *payService.PayWalletService
	orders  *payService.PayOrderService
}

// NewPayWalletController 构造 C 端钱包控制器。
func NewPayWalletController() *PayWalletController {
	return &PayWalletController{
		wallets: payService.NewPayWalletService(),
		orders:  payService.NewPayOrderService(),
	}
}

// Show 查询本人钱包余额（首次访问自动建档，余额为 0）。
// 入参：无（会员身份取令牌）；返回余额与累计充值 / 消费。
func (c *PayWalletController) Show(ctx http.Context) http.Response {
	wallet, err := c.wallets.Detail(ctx.Context(), support.MemberID(ctx.Context()))
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, wallet)
}

// Records 分页查询本人钱包流水。
// 入参：pageNo、pageSize；返回流水列表与总数。
func (c *PayWalletController) Records(ctx http.Context) http.Response {
	page := param.Page{
		PageNo:   ctx.Request().QueryInt("pageNo", 1),
		PageSize: ctx.Request().QueryInt("pageSize", param.DefaultPageSize),
	}
	if err := page.Normalize(); err != nil {
		return result.Fail(ctx, err)
	}

	list, total, err := c.wallets.Transactions(
		ctx.Context(), support.MemberID(ctx.Context()), page.PageNo, page.PageSize,
	)
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Page(ctx, list, total)
}

// MockPay 模拟支付成功：仅 mock 渠道订单可用，且需参数配置 pay.mock.enabled=1。
// 入参：路径 merchantOrderNo；返回最新订单状态。
func (c *PayWalletController) MockPay(ctx http.Context) http.Response {
	merchantOrderNo := ctx.Request().Route("merchantOrderNo")

	m, err := c.orders.MockPay(ctx.Context(), merchantOrderNo, support.MemberID(ctx.Context()))
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, m)
}
