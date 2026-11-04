package migrations

import (
	"eqray-go-starter/app/facades"
)

// M20261104000005CreatePayWalletTransactionTable 建钱包流水表。
// 只追加不修改：金额带符号（正入账、负出账），并记录变动后的余额便于对账。
type M20261104000005CreatePayWalletTransactionTable struct{}

// Signature The unique signature for the migration.
func (r *M20261104000005CreatePayWalletTransactionTable) Signature() string {
	return "20261104000005_create_pay_wallet_transaction_table"
}

// Up 建 pay_wallet_transaction 表。
func (r *M20261104000005CreatePayWalletTransactionTable) Up() error {
	if facades.Schema().HasTable("pay_wallet_transaction") {
		return nil
	}

	return facades.DB().Statement(`
CREATE TABLE pay_wallet_transaction (
  id            bigint       NOT NULL COMMENT '主键',
  wallet_id     bigint       NOT NULL COMMENT '钱包 ID',
  user_id       bigint       NOT NULL COMMENT '会员 ID',
  type          tinyint      NOT NULL COMMENT '类型：1 充值 / 2 消费 / 3 退款 / 4 调整',
  amount        bigint       NOT NULL COMMENT '变动金额（分），正数入账、负数出账',
  balance_after bigint       NOT NULL COMMENT '变动后余额（分）',
  biz_type      varchar(32)  NOT NULL DEFAULT '' COMMENT '业务类型，如 demo_order',
  biz_order_no  varchar(64)  NOT NULL DEFAULT '' COMMENT '业务单号 / 支付单号',
  remark        varchar(255) NOT NULL DEFAULT '' COMMENT '备注',
  creator       varchar(64)  DEFAULT NULL COMMENT '创建人',
  created_at    datetime     DEFAULT NULL COMMENT '创建时间',
  PRIMARY KEY (id),
  KEY idx_wallet (wallet_id, id),
  KEY idx_user (user_id, id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='钱包流水'`)
}

// Down 删除钱包流水表。
func (r *M20261104000005CreatePayWalletTransactionTable) Down() error {
	return facades.Schema().DropIfExists("pay_wallet_transaction")
}
