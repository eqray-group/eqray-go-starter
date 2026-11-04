package migrations

import (
	"eqray-go-starter/app/facades"
)

// M20261103000002CreatePayOrderTable 建 pay_order 支付订单表。
// 一个「业务单 + 渠道」只允许一条记录，靠 uk_biz_order 兜底重复下单。
type M20261103000002CreatePayOrderTable struct{}

// Signature The unique signature for the migration.
func (r *M20261103000002CreatePayOrderTable) Signature() string {
	return "20261103000002_create_pay_order_table"
}

// Up 建 pay_order 表。
// 金额统一用「分」（bigint），全链路不出现浮点数。
func (r *M20261103000002CreatePayOrderTable) Up() error {
	if facades.Schema().HasTable("pay_order") {
		return nil
	}

	return facades.DB().Statement(`
CREATE TABLE pay_order (
  id                bigint       NOT NULL COMMENT '主键',
  merchant_order_no varchar(64)  NOT NULL COMMENT '支付模块生成的商户单号',
  biz_type          varchar(32)  NOT NULL COMMENT '业务类型，如 demo_order',
  biz_order_no      varchar(64)  NOT NULL COMMENT '业务方订单号',
  channel_id        bigint       NOT NULL COMMENT '支付渠道 ID',
  channel_code      varchar(32)  NOT NULL COMMENT '渠道编码快照，渠道被删也能追溯',
  subject           varchar(128) NOT NULL COMMENT '商品标题',
  description       varchar(255) NOT NULL DEFAULT '' COMMENT '商品描述',
  price             bigint       NOT NULL COMMENT '金额（分）',
  currency          varchar(8)   NOT NULL DEFAULT 'CNY' COMMENT '币种',
  status            tinyint      NOT NULL DEFAULT 0 COMMENT '状态：0 待支付 / 10 支付中 / 20 已支付 / 30 已关闭',
  refund_status     tinyint      NOT NULL DEFAULT 0 COMMENT '退款状态：0 未退款 / 10 部分退款 / 20 全额退款',
  user_id           bigint       NOT NULL DEFAULT 0 COMMENT '下单会员 ID',
  client_ip         varchar(64)  NOT NULL DEFAULT '' COMMENT '下单 IP',
  channel_order_no  varchar(64)  NOT NULL DEFAULT '' COMMENT '渠道侧单号',
  channel_user      varchar(64)  NOT NULL DEFAULT '' COMMENT '渠道侧付款人标识',
  channel_extra     mediumtext   NOT NULL COMMENT '渠道返回的原始数据 JSON',
  notify_url        varchar(255) NOT NULL COMMENT '下单时实际传给渠道的通知地址',
  expire_time       datetime     NOT NULL COMMENT '支付超时时间',
  success_time      datetime     NULL COMMENT '支付成功时间',
  creator           varchar(64)  NOT NULL DEFAULT '' COMMENT '创建人',
  created_at        datetime     NULL COMMENT '创建时间',
  updater           varchar(64)  NOT NULL DEFAULT '' COMMENT '更新人',
  updated_at        datetime     NULL COMMENT '更新时间',
  PRIMARY KEY (id),
  UNIQUE KEY uk_merchant_order_no (merchant_order_no),
  UNIQUE KEY uk_biz_order (biz_type, biz_order_no, channel_code),
  KEY idx_user (user_id, created_at),
  KEY idx_status_expire (status, expire_time)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='支付订单表'`)
}

// Down Reverse the migrations.
func (r *M20261103000002CreatePayOrderTable) Down() error {
	return facades.Schema().DropIfExists("pay_order")
}
