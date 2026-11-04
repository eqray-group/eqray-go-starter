package model

import (
	baseEnums "eqray-go-starter/app/modules/base/enums"
	baseModel "eqray-go-starter/app/modules/base/model"
)

// PayChannelTable 是支付渠道配置表名。
const PayChannelTable = "pay_channel"

// PayChannelModel 对应 pay_channel 表：一行是「一种支付方式」，不是「一个商户」。
// 同一商户号要同时支持扫码和 H5，就配两行 code 不同的记录。
type PayChannelModel struct {
	baseModel.BaseModel
	// Code 渠道编码，如 wx_native / ali_pc
	Code string `db:"code" json:"code"`
	// Name 渠道名称，仅后台展示
	Name string `db:"name" json:"name"`
	// Channel 支付厂商：wx / alipay
	Channel string `db:"channel" json:"channel"`
	// PayMethod 支付方式：mini/jsapi/app/native/h5/page/wap/qr
	PayMethod string `db:"pay_method" json:"payMethod"`
	// Config 渠道参数 JSON 明文，含商户私钥等敏感信息；
	// 返回给前端前必须由 service 层脱敏（channel.MaskConfig），敏感字段只回掩码
	Config string `db:"config" json:"config"`
	// Status 状态：0 停用 / 1 启用
	Status baseEnums.CommonStatusEnum `db:"status" json:"status"`
	// Sort 排序，越小越靠前
	Sort int `db:"sort" json:"sort"`
	// Remark 备注
	Remark string `db:"remark" json:"remark"`
}

// TableName 返回表名。
func (m *PayChannelModel) TableName() string {
	return PayChannelTable
}
