package model

import (
	baseModel "eqray-go-starter/app/modules/base/model"
	"eqray-go-starter/app/modules/base/types"
)

// PayRefundTable 是支付退款单表名。
const PayRefundTable = "pay_refund"

// PayRefundStatus 是退款单状态，取值与 pay_refund.status 对齐。
type PayRefundStatus int8

const (
	PayRefundStatusPending   PayRefundStatus = 0  // 待退款：记录已创建，尚未调渠道
	PayRefundStatusRefunding PayRefundStatus = 10 // 退款中：渠道已受理，等待结果
	PayRefundStatusSuccess   PayRefundStatus = 20 // 已退款
	PayRefundStatusFailed    PayRefundStatus = 30 // 退款失败
)

// String 返回状态的中文描述，便于日志排查。
func (s PayRefundStatus) String() string {
	switch s {
	case PayRefundStatusPending:
		return "待退款"
	case PayRefundStatusRefunding:
		return "退款中"
	case PayRefundStatusSuccess:
		return "已退款"
	case PayRefundStatusFailed:
		return "退款失败"
	default:
		return "未知"
	}
}

// PayRefundModel 对应 pay_refund 表：一张支付订单可有多条退款记录。
type PayRefundModel struct {
	baseModel.BaseModel
	// MerchantRefundNo 退款单号，对渠道唯一
	MerchantRefundNo string `db:"merchant_refund_no" json:"merchantRefundNo"`
	// OrderID 支付订单 ID
	OrderID types.ID `db:"order_id" json:"orderId"`
	// MerchantOrderNo 支付单号快照
	MerchantOrderNo string `db:"merchant_order_no" json:"merchantOrderNo"`
	// ChannelID 支付渠道 ID
	ChannelID types.ID `db:"channel_id" json:"channelId"`
	// ChannelCode 渠道编码
	ChannelCode string `db:"channel_code" json:"channelCode"`
	// RefundPrice 退款金额（分）
	RefundPrice int64 `db:"refund_price" json:"refundPrice"`
	// Reason 退款原因
	Reason string `db:"reason" json:"reason"`
	// Status 退款状态
	Status PayRefundStatus `db:"status" json:"status"`
	// ChannelRefundNo 渠道退款单号
	ChannelRefundNo string `db:"channel_refund_no" json:"channelRefundNo"`
	// ChannelExtra 渠道返回的原始数据 JSON
	ChannelExtra string `db:"channel_extra" json:"-"`
	// ErrorMsg 失败原因
	ErrorMsg string `db:"error_msg" json:"errorMsg"`
	// SuccessTime 退款成功时间
	SuccessTime *types.DateTime `db:"success_time" json:"successTime"`
}

// TableName 返回表名。
func (m *PayRefundModel) TableName() string {
	return PayRefundTable
}
