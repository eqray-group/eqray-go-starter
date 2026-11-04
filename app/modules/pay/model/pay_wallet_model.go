package model

import (
	"eqray-go-starter/app/modules/base/types"
)

// 钱包相关表名。
const (
	PayWalletTable            = "pay_wallet"
	PayWalletTransactionTable = "pay_wallet_transaction"
)

// 钱包流水类型。
const (
	// PayWalletTypeRecharge 充值（后台调增 / 后续在线充值）
	PayWalletTypeRecharge int8 = 1
	// PayWalletTypeConsume 消费（余额支付）
	PayWalletTypeConsume int8 = 2
	// PayWalletTypeRefund 退款回钱包
	PayWalletTypeRefund int8 = 3
	// PayWalletTypeAdjust 后台调账（可正可负）
	PayWalletTypeAdjust int8 = 4
	// PayWalletTypeFreeze 冻结（可用余额减少、冻结金额增加）
	PayWalletTypeFreeze int8 = 5
	// PayWalletTypeUnfreeze 解冻（冻结金额减少、可用余额增加）
	PayWalletTypeUnfreeze int8 = 6
)

// 钱包状态。
const (
	// PayWalletStatusDisabled 冻结，不可支付
	PayWalletStatusDisabled int8 = 0
	// PayWalletStatusEnabled 正常
	PayWalletStatusEnabled int8 = 1
)

// PayWalletModel 是会员钱包账户，一个会员一行，余额单位「分」。
type PayWalletModel struct {
	// ID 主键
	ID types.ID `db:"id" json:"id"`
	// UserID 会员 ID
	UserID types.ID `db:"user_id" json:"userId"`
	// Balance 可用余额（分）
	Balance int64 `db:"balance" json:"balance"`
	// FrozenBalance 冻结金额（分），不计入可用余额；总资产 = Balance + FrozenBalance
	FrozenBalance int64 `db:"frozen_balance" json:"frozenBalance"`
	// TotalRecharge 累计充值（分）
	TotalRecharge int64 `db:"total_recharge" json:"totalRecharge"`
	// TotalExpense 累计消费（分）
	TotalExpense int64 `db:"total_expense" json:"totalExpense"`
	// Status 0 冻结 / 1 正常
	Status int8 `db:"status" json:"status"`
	// Creator 创建人
	Creator string `db:"creator" json:"creator"`
	// CreatedAt 创建时间
	CreatedAt types.DateTime `db:"created_at" json:"createdAt"`
	// Updater 更新人
	Updater string `db:"updater" json:"updater"`
	// UpdatedAt 更新时间
	UpdatedAt types.DateTime `db:"updated_at" json:"updatedAt"`
}

// TableName 返回表名。
func (m *PayWalletModel) TableName() string {
	return PayWalletTable
}

// PayWalletTransactionModel 是钱包流水，只追加不修改。
type PayWalletTransactionModel struct {
	// ID 主键
	ID types.ID `db:"id" json:"id"`
	// WalletID 钱包 ID
	WalletID types.ID `db:"wallet_id" json:"walletId"`
	// UserID 会员 ID
	UserID types.ID `db:"user_id" json:"userId"`
	// Type 类型：1 充值 / 2 消费 / 3 退款 / 4 调整
	Type int8 `db:"type" json:"type"`
	// Amount 变动金额（分），正数入账、负数出账
	Amount int64 `db:"amount" json:"amount"`
	// BalanceAfter 变动后余额（分）
	BalanceAfter int64 `db:"balance_after" json:"balanceAfter"`
	// BizType 业务类型，如 demo_order
	BizType string `db:"biz_type" json:"bizType"`
	// BizOrderNo 业务单号 / 支付单号
	BizOrderNo string `db:"biz_order_no" json:"bizOrderNo"`
	// Remark 备注
	Remark string `db:"remark" json:"remark"`
	// Creator 创建人
	Creator string `db:"creator" json:"creator"`
	// CreatedAt 创建时间
	CreatedAt types.DateTime `db:"created_at" json:"createdAt"`
}

// TableName 返回表名。
func (m *PayWalletTransactionModel) TableName() string {
	return PayWalletTransactionTable
}
