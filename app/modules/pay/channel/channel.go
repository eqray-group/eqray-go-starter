// Package channel 是支付渠道适配层：把微信 / 支付宝的 SDK 差异收敛到统一接口后面。
// 本包只依赖 gopay 与基础模型，不依赖 service / repository，方向为 service → channel。
package channel

import (
	"context"
	nethttp "net/http"
	"strings"
)

// 交易状态：各渠道返回值统一映射为以下四个语义。
const (
	// TradeStatePaid 已支付
	TradeStatePaid = "paid"
	// TradeStateUnpaid 未支付（含用户支付中）
	TradeStateUnpaid = "unpaid"
	// TradeStateClosed 已关闭（渠道侧关单 / 超时）
	TradeStateClosed = "closed"
	// TradeStateUnknown 未知（渠道返回异常状态）
	TradeStateUnknown = "unknown"
)

// 退款状态：同步调渠道退款 / 查询退款后的统一语义。
const (
	// RefundStateSuccess 退款成功
	RefundStateSuccess = "success"
	// RefundStatePending 退款处理中
	RefundStatePending = "pending"
	// RefundStateFailed 退款失败
	RefundStateFailed = "failed"
)

// CreateOrderRequest 是向渠道下单的统一入参，金额一律用「分」。
type CreateOrderRequest struct {
	// MerchantOrderNo 商户单号（pay_order.merchant_order_no），对渠道唯一
	MerchantOrderNo string
	// Subject 商品标题
	Subject string
	// Description 商品描述，可空
	Description string
	// Price 金额（分）
	Price int64
	// NotifyURL 异步通知地址，由 pay 模块按 APP_URL 生成
	NotifyURL string
	// ReturnURL 同步跳转地址（支付宝网页类支付用），可空
	ReturnURL string
	// OpenID 渠道侧付款人标识：微信 JSAPI / 小程序必传，其他方式忽略
	OpenID string
	// ClientIP 下单 IP，微信 H5 必传
	ClientIP string
}

// CreateOrderResult 是下单后给前端的统一结果，前端按 Kind 分支处理。
type CreateOrderResult struct {
	// Kind 拉起方式：qr（渲染二维码）/ redirect（跳转）/ jsapi（JSAPI/小程序调起参数）/ app（App 调起参数）
	Kind string
	// URL 二维码内容或跳转地址；qr / redirect 时有效
	URL string
	// Params JSAPI / 小程序 / App 的调起参数（结构体转 JSON 的 map），jsapi / app 时有效
	Params map[string]any
}

// QueryOrderResult 是渠道侧订单查询结果。
type QueryOrderResult struct {
	// TradeState 统一交易状态，取值见本包 TradeState* 常量
	TradeState string
	// ChannelOrderNo 渠道侧单号（微信 transaction_id / 支付宝 trade_no）
	ChannelOrderNo string
	// ChannelUser 渠道侧付款人标识（openid / buyer_id）
	ChannelUser string
	// Amount 实付金额（分），渠道未返回时为 0
	Amount int64
	// SuccessTime 支付完成时间（RFC3339），渠道未返回时为空
	SuccessTime string
	// Raw 渠道原始响应 JSON，落库到 pay_order.channel_extra
	Raw string
}

// RefundRequest 是发起退款的统一入参，金额一律用「分」。
type RefundRequest struct {
	// MerchantRefundNo 商户退款单号（pay_refund.merchant_refund_no），对渠道唯一
	MerchantRefundNo string
	// MerchantOrderNo 原支付商户单号
	MerchantOrderNo string
	// RefundPrice 本次退款金额（分）
	RefundPrice int64
	// TotalPrice 原订单总金额（分），微信要求传原订单金额
	TotalPrice int64
	// Reason 退款原因
	Reason string
}

// RefundResult 是渠道退款处理结果。
type RefundResult struct {
	// State 统一退款状态，取值见本包 RefundState* 常量
	State string
	// ChannelRefundNo 渠道退款单号
	ChannelRefundNo string
	// Raw 渠道原始响应 JSON
	Raw string
}

// NotifyResult 是异步通知解析（验签 + 解密）后的统一结果。
type NotifyResult struct {
	// MerchantOrderNo 商户单号
	MerchantOrderNo string
	// ChannelOrderNo 渠道侧单号
	ChannelOrderNo string
	// TradeState 统一交易状态，取值见本包 TradeState* 常量
	TradeState string
	// Amount 通知金额（分）
	Amount int64
	// ChannelUser 渠道侧付款人标识
	ChannelUser string
	// SuccessTime 支付完成时间（RFC3339），渠道未返回时为空
	SuccessTime string
	// RawRequest 原始请求报文（微信为密文报文、支付宝为表单），写 pay_notify_log.request_body
	RawRequest string
	// Decrypted 解密后的业务报文 JSON，写 pay_notify_log.decrypted_body
	Decrypted string
}

// NotifyAck 是渠道要求的应答报文。
type NotifyAck struct {
	// Body 应答报文内容
	Body string
	// ContentType 应答 Content-Type
	ContentType string
	// Success 是否为「处理成功」应答；false 时渠道会按各自策略重试
	Success bool
}

// IsOrderNotExist 判断渠道错误是否等价于「订单不存在 / 已关闭」。
// 连通性测试会用一个不存在的单号查单：能拿到这类错误说明签名与凭据都正确，
// 应当视为「接口连通」，而不是凭据错误。
// 入参 err 为渠道调用错误；返回是否为订单不存在类错误。
func IsOrderNotExist(err error) bool {
	if err == nil {
		return false
	}

	message := err.Error()
	for _, keyword := range []string{
		"ORDER_NOT_EXIST", "ORDERNOTEXIST", "ORDER_CLOSED", "ORDERNOTEXIST",
		"TRADE_NOT_EXIST", "ACQ.TRADE_NOT_EXIST",
	} {
		if strings.Contains(message, keyword) {
			return true
		}
	}

	return false
}

// Channel 统一支付渠道能力，屏蔽微信 / 支付宝差异。
// 实现必须可并发使用（内部 client 缓存由 registry 负责）。
type Channel interface {
	// Code 返回渠道编码，如 wx_native
	Code() string
	// CreateOrder 向渠道下单，返回前端拉起支付所需的统一结果
	CreateOrder(ctx context.Context, req CreateOrderRequest) (CreateOrderResult, error)
	// QueryOrder 按商户单号查询渠道侧订单状态
	QueryOrder(ctx context.Context, merchantOrderNo string) (QueryOrderResult, error)
	// CloseOrder 关闭渠道侧订单；渠道报「订单不存在」时应视为已关闭返回 nil
	CloseOrder(ctx context.Context, merchantOrderNo string) error
	// Refund 发起退款，返回同步可确认的结果（支付宝终态 / 微信受理中）
	Refund(ctx context.Context, req RefundRequest) (RefundResult, error)
	// QueryRefund 按商户退款单号查询退款状态；不支持查单的渠道返回错误，由调用方跳过
	QueryRefund(ctx context.Context, merchantRefundNo, merchantOrderNo string) (RefundResult, error)
	// ParseNotify 解析并验签异步通知；验签失败必须返回错误，不得静默通过
	ParseNotify(ctx context.Context, r *nethttp.Request) (NotifyResult, error)
	// NotifyAck 返回该渠道要求的成功应答
	NotifyAck() NotifyAck
	// NotifyFailAck 返回该渠道要求的失败应答，用于验签失败 / 处理异常时触发渠道重试
	NotifyFailAck(reason string) NotifyAck
}
