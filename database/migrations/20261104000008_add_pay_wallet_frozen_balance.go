package migrations

import (
	"eqray-go-starter/app/facades"
)

// M20261104000008AddPayWalletFrozenBalance 给钱包增加冻结金额。
// 语义：balance 为可用余额，frozen_balance 为被冻结部分，总资产 = 两者之和。
type M20261104000008AddPayWalletFrozenBalance struct{}

// Signature The unique signature for the migration.
func (r *M20261104000008AddPayWalletFrozenBalance) Signature() string {
	return "20261104000008_add_pay_wallet_frozen_balance"
}

// Up 新增 frozen_balance 字段。
func (r *M20261104000008AddPayWalletFrozenBalance) Up() error {
	if !facades.Schema().HasTable("pay_wallet") {
		return nil
	}
	if facades.Schema().HasColumn("pay_wallet", "frozen_balance") {
		return nil
	}

	return facades.DB().Statement(`
ALTER TABLE pay_wallet
  ADD COLUMN frozen_balance bigint NOT NULL DEFAULT 0 COMMENT '冻结金额（分）' AFTER balance`)
}

// Down 删除 frozen_balance 字段。
func (r *M20261104000008AddPayWalletFrozenBalance) Down() error {
	if !facades.Schema().HasTable("pay_wallet") {
		return nil
	}

	return facades.DB().Statement(`ALTER TABLE pay_wallet DROP COLUMN frozen_balance`)
}
