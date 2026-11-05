// Package payevents 定义支付域的跨模块事件契约：事件键、业务载荷与派发入口。
// 放在 app/events 下，任何模块都只依赖这里的事件类型，不反向依赖 pay 的 service 层。
// 本包只放契约，不放业务逻辑；新增事件按域分文件，避免公共包变成杂物间。
package payevents

import (
	"github.com/goravel/framework/contracts/event"
)

// OrderPaidEvent 是「支付成功」事件的键类型。
// Goravel 事件用「事件实例」作 map key 查找监听器，而事件 struct 是空结构体：
// 必须用**值类型**参与注册与派发，零值恒等才能命中注册表（指针每次都是新地址，查不到）。
// 业务载荷通过 OrderPaidPayload 经 args 传递，见 DispatchOrderPaid。
type OrderPaidEvent struct{}

// Handle 透传事件载荷给监听器；框架约定 Event 需实现 Handle。
// 入参 args 为派发时传入的载荷，返回值即监听器收到的参数，原样透传。
func (e OrderPaidEvent) Handle(args []event.Arg) ([]event.Arg, error) {
	return args, nil
}

// OrderPaidPayload 是支付成功事件的业务载荷。
type OrderPaidPayload struct {
	// BizType 业务类型，如 demo_order
	BizType string `json:"bizType"`
	// BizOrderNo 业务方订单号
	BizOrderNo string `json:"bizOrderNo"`
	// OrderNo 支付单号（pay_order.merchant_order_no）
	OrderNo string `json:"orderNo"`
	// ChannelCode 支付渠道编码
	ChannelCode string `json:"channelCode"`
	// Price 实付金额（分）
	Price int64 `json:"price"`
	// PaidAt 支付完成时间（RFC3339 字符串）
	PaidAt string `json:"paidAt"`
}

// RefundSucceededEvent 是「退款成功」事件的键类型，用法与 OrderPaidEvent 一致。
type RefundSucceededEvent struct{}

// Handle 透传事件载荷给监听器；框架约定 Event 需实现 Handle。
func (e RefundSucceededEvent) Handle(args []event.Arg) ([]event.Arg, error) {
	return args, nil
}

// RefundSucceededPayload 是退款成功事件的业务载荷。
type RefundSucceededPayload struct {
	// BizType 业务类型
	BizType string `json:"bizType"`
	// BizOrderNo 业务方订单号
	BizOrderNo string `json:"bizOrderNo"`
	// OrderNo 支付单号
	OrderNo string `json:"orderNo"`
	// RefundNo 退款单号（pay_refund.merchant_refund_no）
	RefundNo string `json:"refundNo"`
	// ChannelCode 支付渠道编码
	ChannelCode string `json:"channelCode"`
	// RefundPrice 退款金额（分）
	RefundPrice int64 `json:"refundPrice"`
	// Reason 退款原因
	Reason string `json:"reason"`
	// SucceededAt 退款成功时间（RFC3339 字符串）
	SucceededAt string `json:"succeededAt"`
}

// DispatchOrderPaid 派发支付成功事件；载荷封装为单个 Arg 传递。
// 入参 payload 为业务载荷；无监听器时框架返回「未绑定」错误，仅记日志不打断调用方。
func DispatchOrderPaid(payload *OrderPaidPayload) {
	dispatch(OrderPaidEvent{}, payload)
}

// DispatchRefundSucceeded 派发退款成功事件，用法与 DispatchOrderPaid 一致。
func DispatchRefundSucceeded(payload *RefundSucceededPayload) {
	dispatch(RefundSucceededEvent{}, payload)
}
