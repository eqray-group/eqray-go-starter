package model

import (
	baseModel "eqray-go-starter/app/modules/base/model"
)

// PayNotifyLogTable 是支付回调日志表名。
const PayNotifyLogTable = "pay_notify_log"

// 回调类型，用于区分支付通知与退款通知。
const (
	PayNotifyTypePay    = "pay"
	PayNotifyTypeRefund = "refund"
)

// 回调处理结果，用于后台列表筛选。
const (
	PayNotifyVerifyFailed int8 = 0 // 验签失败
	PayNotifyVerifyPassed int8 = 1 // 验签通过
	PayNotifyHandleFailed int8 = 0 // 业务处理失败
	PayNotifyHandleOK     int8 = 1 // 业务处理成功
)

// PayNotifyLogModel 对应 pay_notify_log 表：只追加，不更新。
type PayNotifyLogModel struct {
	baseModel.BaseModel
	// ChannelCode 渠道编码
	ChannelCode string `db:"channel_code" json:"channelCode"`
	// Type 类型：pay / refund
	Type string `db:"type" json:"type"`
	// MerchantOrderNo 支付单号，可能为空（报文异常时解析不出）
	MerchantOrderNo string `db:"merchant_order_no" json:"merchantOrderNo"`
	// RequestHeaders 关键请求头，已脱敏
	RequestHeaders string `db:"request_headers" json:"requestHeaders"`
	// RequestBody 原始报文
	RequestBody string `db:"request_body" json:"requestBody"`
	// DecryptedBody 解密后的业务报文
	DecryptedBody string `db:"decrypted_body" json:"decryptedBody"`
	// VerifyStatus 验签结果
	VerifyStatus int8 `db:"verify_status" json:"verifyStatus"`
	// HandleStatus 业务处理结果
	HandleStatus int8 `db:"handle_status" json:"handleStatus"`
	// ErrorMsg 失败原因
	ErrorMsg string `db:"error_msg" json:"errorMsg"`
}

// TableName 返回表名。
func (m *PayNotifyLogModel) TableName() string {
	return PayNotifyLogTable
}
