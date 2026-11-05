package migrations

import (
	"eqray-go-starter/app/facades"
)

// M20261105000004CreateInfraNoticeTables 建消息通知三张表：渠道、模板、发送记录。
// 渠道配置以 JSON 文本存放（不同渠道字段差异大，避免为每种渠道加列）。
type M20261105000004CreateInfraNoticeTables struct{}

// Signature The unique signature for the migration.
func (r *M20261105000004CreateInfraNoticeTables) Signature() string {
	return "20261105000004_create_infra_notice_tables"
}

// Up 建表。
func (r *M20261105000004CreateInfraNoticeTables) Up() error {
	if !facades.Schema().HasTable("infra_notice_channel") {
		if err := facades.DB().Statement(`
CREATE TABLE infra_notice_channel (
  id         bigint       NOT NULL COMMENT '主键',
  name       varchar(64)  NOT NULL COMMENT '渠道名称',
  type       varchar(32)  NOT NULL COMMENT '渠道类型：email / wecom / dingtalk',
  config     text         COMMENT '渠道配置（JSON 文本，字段随类型变化）',
  status     tinyint      NOT NULL DEFAULT 1 COMMENT '状态：0 停用 / 1 启用',
  remark     varchar(255) NOT NULL DEFAULT '' COMMENT '备注',
  creator    varchar(64)  DEFAULT NULL COMMENT '创建人',
  created_at datetime     DEFAULT NULL COMMENT '创建时间',
  updater    varchar(64)  DEFAULT NULL COMMENT '更新人',
  updated_at datetime     DEFAULT NULL COMMENT '更新时间',
  PRIMARY KEY (id),
  KEY idx_type_status (type, status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='消息通知渠道'`); err != nil {
			return err
		}
	}

	if !facades.Schema().HasTable("infra_notice_template") {
		if err := facades.DB().Statement(`
CREATE TABLE infra_notice_template (
  id           bigint       NOT NULL COMMENT '主键',
  code         varchar(64)  NOT NULL COMMENT '模板编码（业务唯一标识）',
  name         varchar(64)  NOT NULL COMMENT '模板名称',
  type         varchar(32)  NOT NULL COMMENT '适用渠道类型：email / wecom / dingtalk',
  title        varchar(128) NOT NULL DEFAULT '' COMMENT '标题（邮件主题 / 机器人标题）',
  content      text         COMMENT '模板内容，支持 {{变量}} 占位',
  status       tinyint      NOT NULL DEFAULT 1 COMMENT '状态：0 停用 / 1 启用',
  remark       varchar(255) NOT NULL DEFAULT '' COMMENT '备注',
  creator      varchar(64)  DEFAULT NULL COMMENT '创建人',
  created_at   datetime     DEFAULT NULL COMMENT '创建时间',
  updater      varchar(64)  DEFAULT NULL COMMENT '更新人',
  updated_at   datetime     DEFAULT NULL COMMENT '更新时间',
  PRIMARY KEY (id),
  UNIQUE KEY uk_code (code)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='消息通知模板'`); err != nil {
			return err
		}
	}

	if !facades.Schema().HasTable("infra_notice_log") {
		if err := facades.DB().Statement(`
CREATE TABLE infra_notice_log (
  id           bigint       NOT NULL COMMENT '主键',
  channel_id   bigint       NOT NULL DEFAULT 0 COMMENT '渠道 ID',
  channel_name varchar(64)  NOT NULL DEFAULT '' COMMENT '渠道名称（冗余）',
  type         varchar(32)  NOT NULL DEFAULT '' COMMENT '渠道类型',
  template_code varchar(64) NOT NULL DEFAULT '' COMMENT '模板编码，手动发送时为空',
  target       varchar(255) NOT NULL DEFAULT '' COMMENT '接收目标（邮箱 / 群机器人无目标）',
  title        varchar(128) NOT NULL DEFAULT '' COMMENT '标题',
  content      text         COMMENT '实际发送内容',
  status       tinyint      NOT NULL DEFAULT 1 COMMENT '结果：0 失败 / 1 成功',
  error        text         COMMENT '失败原因',
  duration     int          NOT NULL DEFAULT 0 COMMENT '耗时（毫秒）',
  creator      varchar(64)  DEFAULT NULL COMMENT '操作人',
  created_at   datetime     DEFAULT NULL COMMENT '创建时间',
  PRIMARY KEY (id),
  KEY idx_created (created_at),
  KEY idx_type_status (type, status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='消息通知发送记录'`); err != nil {
			return err
		}
	}

	return nil
}

// Down 删除三张表。
func (r *M20261105000004CreateInfraNoticeTables) Down() error {
	for _, table := range []string{"infra_notice_log", "infra_notice_template", "infra_notice_channel"} {
		if err := facades.Schema().DropIfExists(table); err != nil {
			return err
		}
	}

	return nil
}
