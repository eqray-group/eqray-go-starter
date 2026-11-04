package model

import (
	baseModel "eqray-go-starter/app/modules/base/model"
	"eqray-go-starter/app/modules/base/types"
)

// DemoOrderTable 是演示业务订单表名。
const DemoOrderTable = "demo_order"

// 演示订单状态，与 demo_order.status 对齐。
const (
	DemoOrderStatusPending int8 = 0  // 待支付
	DemoOrderStatusPaid    int8 = 10 // 已支付
	DemoOrderStatusCancel  int8 = 20 // 已取消
)

// DemoOrderModel 对应 demo_order 表：仅用于联调，不承载真实业务。
// 它模拟「业务模块」的角色——自己下单、自己订阅支付成功事件，pay 模块对它一无所知。
type DemoOrderModel struct {
	baseModel.BaseModel
	// OrderNo 演示业务单号
	OrderNo string `db:"order_no" json:"orderNo"`
	// UserID 下单会员 ID
	UserID types.ID `db:"user_id" json:"userId"`
	// Subject 商品标题
	Subject string `db:"subject" json:"subject"`
	// Price 金额（分）
	Price int64 `db:"price" json:"price"`
	// Status 状态：0 待支付 / 10 已支付 / 20 已取消
	Status int8 `db:"status" json:"status"`
	// PayOrderNo 支付单号
	PayOrderNo string `db:"pay_order_no" json:"payOrderNo"`
	// PaidAt 支付完成时间
	PaidAt *types.DateTime `db:"paid_at" json:"paidAt"`
}

// TableName 返回表名。
func (m *DemoOrderModel) TableName() string {
	return DemoOrderTable
}
