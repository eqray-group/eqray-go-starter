package migrations

import (
	"eqray-go-starter/app/facades"
)

// M20261103000003CreatePayRefundTable 建 pay_refund 退款单表。
// 一张支付订单可以有多条退款记录（部分退款场景），累计金额不得超过订单金额。
type M20261103000003CreatePayRefundTable struct{}

// Signature The unique signature for the migration.
func (r *M20261103000003CreatePayRefundTable) Signature() string {
	return "20261103000003_create_pay_refund_table"
}

// Up 建 pay_refund 表。
func (r *M20261103000003CreatePayRefundTable) Up() error {
	if facades.Schema().HasTable("pay_refund") {
		return nil
	}

	return facades.DB().Statement(`
CREATE TABLE pay_refund (
  id                 bigint       NOT NULL COMMENT '主键',
  merchant_refund_no varchar(64)  NOT NULL COMMENT '退款单号',
  order_id           bigint       NOT NULL COMMENT '支付订单 ID',
  merchant_order_no  varchar(64)  NOT NULL COMMENT '支付单号快照',
  channel_id         bigint       NOT NULL COMMENT '支付渠道 ID',
  channel_code       varchar(32)  NOT NULL COMMENT '渠道编码',
  refund_price       bigint       NOT NULL COMMENT '退款金额（分）',
  reason             varchar(255) NOT NULL COMMENT '退款原因',
  status             tinyint      NOT NULL DEFAULT 0 COMMENT '状态：0 待退款 / 10 退款中 / 20 已退款 / 30 退款失败',
  channel_refund_no  varchar(64)  NOT NULL DEFAULT '' COMMENT '渠道退款单号',
  channel_extra      mediumtext   NOT NULL COMMENT '渠道返回的原始数据 JSON',
  error_msg          varchar(500) NOT NULL DEFAULT '' COMMENT '失败原因',
  success_time       datetime     NULL COMMENT '退款成功时间',
  creator            varchar(64)  NOT NULL DEFAULT '' COMMENT '创建人',
  created_at         datetime     NULL COMMENT '创建时间',
  updater            varchar(64)  NOT NULL DEFAULT '' COMMENT '更新人',
  updated_at         datetime     NULL COMMENT '更新时间',
  PRIMARY KEY (id),
  UNIQUE KEY uk_merchant_refund_no (merchant_refund_no),
  KEY idx_order (order_id),
  KEY idx_status (status, updated_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='支付退款单表'`)
}

// Down Reverse the migrations.
func (r *M20261103000003CreatePayRefundTable) Down() error {
	return facades.Schema().DropIfExists("pay_refund")
}
