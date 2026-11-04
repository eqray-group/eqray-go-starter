package migrations

import (
	"eqray-go-starter/app/facades"
)

// M20261104000003CreatePayWalletTable 建会员钱包账户表。
// 一个会员一行，余额单位「分」；充值 / 消费 / 退款 / 后台调账都会更新这里并写流水。
// balance 为可用余额，frozen_balance 为冻结金额，总资产 = 两者之和。
type M20261104000003CreatePayWalletTable struct{}

// Signature The unique signature for the migration.
func (r *M20261104000003CreatePayWalletTable) Signature() string {
	return "20261104000003_create_pay_wallet_table"
}

// Up 建 pay_wallet 表。
func (r *M20261104000003CreatePayWalletTable) Up() error {
	if facades.Schema().HasTable("pay_wallet") {
		return nil
	}

	return facades.DB().Statement(`
CREATE TABLE pay_wallet (
  id             bigint      NOT NULL COMMENT '主键',
  user_id        bigint      NOT NULL COMMENT '会员 ID',
  balance        bigint      NOT NULL DEFAULT 0 COMMENT '可用余额（分）',
  frozen_balance bigint      NOT NULL DEFAULT 0 COMMENT '冻结金额（分）',
  total_recharge bigint      NOT NULL DEFAULT 0 COMMENT '累计充值（分）',
  total_expense  bigint      NOT NULL DEFAULT 0 COMMENT '累计消费（分）',
  status         tinyint     NOT NULL DEFAULT 1 COMMENT '状态：0 冻结 / 1 正常',
  creator        varchar(64) DEFAULT NULL COMMENT '创建人',
  created_at     datetime    DEFAULT NULL COMMENT '创建时间',
  updater        varchar(64) DEFAULT NULL COMMENT '更新人',
  updated_at     datetime    DEFAULT NULL COMMENT '更新时间',
  PRIMARY KEY (id),
  UNIQUE KEY uk_user (user_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='会员钱包账户'`)
}

// Down 删除钱包账户表。
func (r *M20261104000003CreatePayWalletTable) Down() error {
	return facades.Schema().DropIfExists("pay_wallet")
}
