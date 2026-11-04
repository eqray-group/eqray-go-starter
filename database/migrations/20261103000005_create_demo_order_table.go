package migrations

import (
	"eqray-go-starter/app/facades"
)

// M20261103000005CreateDemoOrderTable 建 demo_order 演示业务订单表。
// 仅用于联调验证「下单 → 支付 → 事件回写」整条链路，不承载真实业务，上线可整体删除。
type M20261103000005CreateDemoOrderTable struct{}

// Signature The unique signature for the migration.
func (r *M20261103000005CreateDemoOrderTable) Signature() string {
	return "20261103000005_create_demo_order_table"
}

// Up 建 demo_order 表。
func (r *M20261103000005CreateDemoOrderTable) Up() error {
	if facades.Schema().HasTable("demo_order") {
		return nil
	}

	return facades.DB().Statement(`
CREATE TABLE demo_order (
  id           bigint       NOT NULL COMMENT '主键',
  order_no     varchar(64)  NOT NULL COMMENT '演示业务单号',
  user_id      bigint       NOT NULL COMMENT '下单会员 ID',
  subject      varchar(128) NOT NULL COMMENT '商品标题',
  price        bigint       NOT NULL COMMENT '金额（分）',
  status       tinyint      NOT NULL DEFAULT 0 COMMENT '状态：0 待支付 / 10 已支付 / 20 已取消',
  pay_order_no varchar(64)  NOT NULL DEFAULT '' COMMENT '支付单号',
  paid_at      datetime     NULL COMMENT '支付完成时间',
  creator      varchar(64)  NOT NULL DEFAULT '' COMMENT '创建人',
  created_at   datetime     NULL COMMENT '创建时间',
  updater      varchar(64)  NOT NULL DEFAULT '' COMMENT '更新人',
  updated_at   datetime     NULL COMMENT '更新时间',
  PRIMARY KEY (id),
  UNIQUE KEY uk_order_no (order_no),
  KEY idx_user (user_id, created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='演示业务订单表（仅联调，上线可删）'`)
}

// Down Reverse the migrations.
func (r *M20261103000005CreateDemoOrderTable) Down() error {
	return facades.Schema().DropIfExists("demo_order")
}
