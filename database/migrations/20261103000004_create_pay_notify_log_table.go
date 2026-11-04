package migrations

import (
	"eqray-go-starter/app/facades"
)

// M20261103000004CreatePayNotifyLogTable 建 pay_notify_log 支付回调日志表。
// 渠道每次通知都留档原始报文与处理结果，排查「为什么没到账」时主要看这张表。
type M20261103000004CreatePayNotifyLogTable struct{}

// Signature The unique signature for the migration.
func (r *M20261103000004CreatePayNotifyLogTable) Signature() string {
	return "20261103000004_create_pay_notify_log_table"
}

// Up 建 pay_notify_log 表。
// 不建软删与更新字段：日志只追加，过期由定时任务物理清理。
func (r *M20261103000004CreatePayNotifyLogTable) Up() error {
	if facades.Schema().HasTable("pay_notify_log") {
		return nil
	}

	return facades.DB().Statement(`
CREATE TABLE pay_notify_log (
  id                bigint       NOT NULL COMMENT '主键',
  channel_code      varchar(32)  NOT NULL COMMENT '渠道编码',
  type              varchar(16)  NOT NULL COMMENT '类型：pay / refund',
  merchant_order_no varchar(64)  NOT NULL DEFAULT '' COMMENT '支付单号',
  request_headers   mediumtext   NOT NULL COMMENT '关键请求头，已脱敏',
  request_body      mediumtext   NOT NULL COMMENT '原始报文',
  decrypted_body    mediumtext   NOT NULL COMMENT '解密后的业务报文',
  verify_status     tinyint      NOT NULL DEFAULT 0 COMMENT '验签结果：0 失败 / 1 通过',
  handle_status     tinyint      NOT NULL DEFAULT 0 COMMENT '处理结果：0 失败 / 1 成功',
  error_msg         varchar(500) NOT NULL DEFAULT '' COMMENT '失败原因',
  created_at        datetime     NULL COMMENT '创建时间',
  PRIMARY KEY (id),
  KEY idx_order (merchant_order_no),
  KEY idx_created (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='支付回调日志表'`)
}

// Down Reverse the migrations.
func (r *M20261103000004CreatePayNotifyLogTable) Down() error {
	return facades.Schema().DropIfExists("pay_notify_log")
}
