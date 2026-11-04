package channel

import (
	"context"
	nethttp "net/http"
	"time"

	"eqray-go-starter/app/modules/base/errs"
	payErrs "eqray-go-starter/app/modules/pay/errs"
)

// WalletChannel 是「余额支付」渠道，不产生第三方请求。
// 真正的扣款、流水与订单流转由 service 层在事务里完成（见 PayOrderService.payByWallet），
// 这里只提供渠道接口所需的空实现，让渠道列表 / 收银台可以把它当成普通渠道对待。
type WalletChannel struct {
	// code 渠道编码（wallet）
	code string
}

// newWalletChannel 构造余额支付渠道，不需要任何凭据。
func newWalletChannel(code, _ string) (Channel, error) {
	return &WalletChannel{code: code}, nil
}

// Code 返回渠道编码。
func (c *WalletChannel) Code() string {
	return c.code
}

// CreateOrder 余额支付不经过渠道下单，返回 kind=wallet 由前端提示「支付成功」。
func (c *WalletChannel) CreateOrder(context.Context, CreateOrderRequest) (CreateOrderResult, error) {
	return CreateOrderResult{Kind: KindWallet}, nil
}

// QueryOrder 余额支付是同步完成的，查询时按「已支付」返回，避免轮询方误判。
func (c *WalletChannel) QueryOrder(_ context.Context, merchantOrderNo string) (QueryOrderResult, error) {
	return QueryOrderResult{
		TradeState:     TradeStatePaid,
		ChannelOrderNo: walletOrderNo(merchantOrderNo),
		SuccessTime:    time.Now().Format(time.RFC3339),
	}, nil
}

// CloseOrder 余额支付没有渠道侧订单，直接返回成功。
func (c *WalletChannel) CloseOrder(context.Context, string) error {
	return nil
}

// Refund 余额退款同步到账，返回成功。
func (c *WalletChannel) Refund(context.Context, RefundRequest) (RefundResult, error) {
	return RefundResult{State: RefundStateSuccess}, nil
}

// QueryRefund 余额退款同步完成，查询直接返回成功。
func (c *WalletChannel) QueryRefund(context.Context, string, string) (RefundResult, error) {
	return RefundResult{State: RefundStateSuccess}, nil
}

// ParseNotify 余额支付没有异步通知，收到通知说明回调地址配错，直接报错。
func (c *WalletChannel) ParseNotify(context.Context, *nethttp.Request) (NotifyResult, error) {
	return NotifyResult{}, errs.New(payErrs.PayChannelUnsupported, c.code)
}

// NotifyAck 余额支付没有通知，返回成功应答占位。
func (c *WalletChannel) NotifyAck() NotifyAck {
	return NotifyAck{Body: "success", ContentType: "text/plain", Success: true}
}

// NotifyFailAck 同上，返回失败应答占位。
func (c *WalletChannel) NotifyFailAck(string) NotifyAck {
	return NotifyAck{Body: "success", ContentType: "text/plain", Success: false}
}

// walletOrderNo 生成余额支付的虚拟渠道单号，便于对账时一眼识别。
func walletOrderNo(merchantOrderNo string) string {
	return "WALLET" + merchantOrderNo
}
