package model

import (
	baseModel "eqray-go-starter/app/modules/base/model"
	"eqray-go-starter/app/modules/base/types"
)

// PayOrderTable 是支付订单表名。
const PayOrderTable = "pay_order"

// PayOrderStatus 是支付订单状态，取值与 pay_order.status 对齐。
type PayOrderStatus int8

const (
	PayOrderStatusPending PayOrderStatus = 0  // 待支付：订单已创建，尚未向渠道下单
	PayOrderStatusPaying  PayOrderStatus = 10 // 支付中：已向渠道下单，等待用户付款
	PayOrderStatusPaid    PayOrderStatus = 20 // 已支付：渠道确认收款
	PayOrderStatusClosed  PayOrderStatus = 30 // 已关闭：超时或人工关闭
)

// String 返回状态的中文描述，便于日志排查。
func (s PayOrderStatus) String() string {
	switch s {
	case PayOrderStatusPending:
		return "待支付"
	case PayOrderStatusPaying:
		return "支付中"
	case PayOrderStatusPaid:
		return "已支付"
	case PayOrderStatusClosed:
		return "已关闭"
	default:
		return "未知"
	}
}

// PayOrderRefundStatus 是订单整体的退款进度。
type PayOrderRefundStatus int8

const (
	PayOrderRefundStatusNone    PayOrderRefundStatus = 0  // 未退款
	PayOrderRefundStatusPartial PayOrderRefundStatus = 10 // 部分退款
	PayOrderRefundStatusAll     PayOrderRefundStatus = 20 // 全额退款
)

// PayOrderModel 对应 pay_order 表：一次「业务单 + 渠道」的收款记录。
type PayOrderModel struct {
	baseModel.BaseModel
	// MerchantOrderNo 支付模块生成的商户单号，对渠道唯一
	MerchantOrderNo string `db:"merchant_order_no" json:"merchantOrderNo"`
	// BizType 业务类型，如 demo_order
	BizType string `db:"biz_type" json:"bizType"`
	// BizOrderNo 业务方订单号
	BizOrderNo string `db:"biz_order_no" json:"bizOrderNo"`
	// ChannelID 支付渠道 ID
	ChannelID types.ID `db:"channel_id" json:"channelId"`
	// ChannelCode 渠道编码快照，渠道被删也能追溯
	ChannelCode string `db:"channel_code" json:"channelCode"`
	// Subject 商品标题
	Subject string `db:"subject" json:"subject"`
	// Description 商品描述
	Description string `db:"description" json:"description"`
	// Price 金额（分）
	Price int64 `db:"price" json:"price"`
	// Currency 币种，默认 CNY
	Currency string `db:"currency" json:"currency"`
	// Status 订单状态
	Status PayOrderStatus `db:"status" json:"status"`
	// RefundStatus 退款进度
	RefundStatus PayOrderRefundStatus `db:"refund_status" json:"refundStatus"`
	// UserID 下单会员 ID
	UserID types.ID `db:"user_id" json:"userId"`
	// ClientIP 下单 IP
	ClientIP string `db:"client_ip" json:"clientIp"`
	// ChannelOrderNo 渠道侧单号（微信 transaction_id / 支付宝 trade_no）
	ChannelOrderNo string `db:"channel_order_no" json:"channelOrderNo"`
	// ChannelUser 渠道侧付款人标识（openid / buyer_id）
	ChannelUser string `db:"channel_user" json:"channelUser"`
	// ChannelExtra 渠道返回的原始数据 JSON
	ChannelExtra string `db:"channel_extra" json:"-"`
	// NotifyURL 下单时实际传给渠道的通知地址
	NotifyURL string `db:"notify_url" json:"notifyUrl"`
	// ExpireTime 支付超时时间
	ExpireTime types.DateTime `db:"expire_time" json:"expireTime"`
	// SuccessTime 支付成功时间
	SuccessTime *types.DateTime `db:"success_time" json:"successTime"`
}

// TableName 返回表名。
func (m *PayOrderModel) TableName() string {
	return PayOrderTable
}
