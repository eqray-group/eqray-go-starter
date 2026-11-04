package migrations

import (
	"eqray-go-starter/app/facades"
)

// M20261103000001CreatePayChannelTable 建 pay_channel 支付渠道配置表。
// 一行对应一种支付方式（如 wx_native），凭据以 JSON 明文存在 config 列。
type M20261103000001CreatePayChannelTable struct{}

// Signature The unique signature for the migration.
func (r *M20261103000001CreatePayChannelTable) Signature() string {
	return "20261103000001_create_pay_channel_table"
}

// Up 建 pay_channel 表。
// 渠道属于配置数据，删除走物理删除，否则 uk_code 会被软删记录长期占用。
func (r *M20261103000001CreatePayChannelTable) Up() error {
	if facades.Schema().HasTable("pay_channel") {
		return nil
	}

	return facades.DB().Statement(`
CREATE TABLE pay_channel (
  id         bigint       NOT NULL COMMENT '主键',
  code       varchar(32)  NOT NULL COMMENT '渠道编码：wx_native / wx_mini / ali_pc 等',
  name       varchar(64)  NOT NULL COMMENT '渠道名称，仅后台展示',
  channel    varchar(16)  NOT NULL COMMENT '支付厂商：wx / alipay',
  pay_method varchar(16)  NOT NULL COMMENT '支付方式：mini/jsapi/app/native/h5/page/wap/qr',
  config     mediumtext   NOT NULL COMMENT '渠道参数 JSON，明文存储（含商户私钥 / APIv3 密钥）',
  status     tinyint      NOT NULL DEFAULT 1 COMMENT '状态：0 停用 / 1 启用',
  sort       int          NOT NULL DEFAULT 0 COMMENT '排序，越小越靠前',
  remark     varchar(255) NOT NULL DEFAULT '' COMMENT '备注',
  creator    varchar(64)  NOT NULL DEFAULT '' COMMENT '创建人',
  created_at datetime     NULL COMMENT '创建时间',
  updater    varchar(64)  NOT NULL DEFAULT '' COMMENT '更新人',
  updated_at datetime     NULL COMMENT '更新时间',
  PRIMARY KEY (id),
  UNIQUE KEY uk_code (code)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='支付渠道配置表'`)
}

// Down Reverse the migrations.
func (r *M20261103000001CreatePayChannelTable) Down() error {
	return facades.Schema().DropIfExists("pay_channel")
}
