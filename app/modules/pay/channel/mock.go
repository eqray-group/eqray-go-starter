package channel

import (
	"context"
	nethttp "net/http"

	"eqray-go-starter/app/modules/base/errs"
	payErrs "eqray-go-starter/app/modules/pay/errs"
)

// MockChannel 是联调用的模拟支付渠道：不请求任何第三方，
// 下单后由 C 端调用「模拟支付成功」接口推进订单（见 PayOrderService.MockPay）。
type MockChannel struct {
	// code 渠道编码（mock）
	code string
}

// newMockChannel 构造模拟支付渠道，不需要任何凭据。
func newMockChannel(code, _ string) (Channel, error) {
	return &MockChannel{code: code}, nil
}

// Code 返回渠道编码。
func (c *MockChannel) Code() string {
	return c.code
}

// CreateOrder 返回 kind=mock，前端据此展示「模拟支付」按钮。
func (c *MockChannel) CreateOrder(context.Context, CreateOrderRequest) (CreateOrderResult, error) {
	return CreateOrderResult{Kind: KindMock, Params: map[string]any{"mock": true}}, nil
}

// QueryOrder 模拟渠道只能通过模拟支付接口推进，查单返回未支付。
func (c *MockChannel) QueryOrder(context.Context, string) (QueryOrderResult, error) {
	return QueryOrderResult{TradeState: TradeStateUnpaid}, nil
}

// CloseOrder 模拟渠道没有真实订单，直接返回成功。
func (c *MockChannel) CloseOrder(context.Context, string) error {
	return nil
}

// Refund 模拟退款直接成功。
func (c *MockChannel) Refund(context.Context, RefundRequest) (RefundResult, error) {
	return RefundResult{State: RefundStateSuccess}, nil
}

// QueryRefund 模拟退款直接返回成功。
func (c *MockChannel) QueryRefund(context.Context, string, string) (RefundResult, error) {
	return RefundResult{State: RefundStateSuccess}, nil
}

// ParseNotify 模拟渠道没有异步通知。
func (c *MockChannel) ParseNotify(context.Context, *nethttp.Request) (NotifyResult, error) {
	return NotifyResult{}, errs.New(payErrs.PayChannelUnsupported, c.code)
}

// NotifyAck 返回成功应答占位。
func (c *MockChannel) NotifyAck() NotifyAck {
	return NotifyAck{Body: "success", ContentType: "text/plain", Success: true}
}

// NotifyFailAck 返回失败应答占位。
func (c *MockChannel) NotifyFailAck(string) NotifyAck {
	return NotifyAck{Body: "success", ContentType: "text/plain", Success: false}
}
