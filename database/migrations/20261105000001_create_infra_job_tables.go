package migrations

import (
	"eqray-go-starter/app/facades"
)

// M20261105000001CreateInfraJobTables 建定时任务与执行日志表。
// 任务本身只存「处理器标识 + cron 表达式 + 启停状态」，具体逻辑由代码注册，避免在库里存代码。
type M20261105000001CreateInfraJobTables struct{}

// Signature The unique signature for the migration.
func (r *M20261105000001CreateInfraJobTables) Signature() string {
	return "20261105000001_create_infra_job_tables"
}

// Up 建 infra_job / infra_job_log 表。
func (r *M20261105000001CreateInfraJobTables) Up() error {
	if !facades.Schema().HasTable("infra_job") {
		if err := facades.DB().Statement(`
CREATE TABLE infra_job (
  id            bigint       NOT NULL COMMENT '主键',
  name          varchar(64)  NOT NULL COMMENT '任务名称',
  handler       varchar(64)  NOT NULL COMMENT '处理器标识（代码注册）',
  cron          varchar(64)  NOT NULL COMMENT 'cron 表达式（5 段）',
  status        tinyint      NOT NULL DEFAULT 1 COMMENT '状态：0 停用 / 1 启用',
  remark        varchar(255) NOT NULL DEFAULT '' COMMENT '备注',
  next_run_at   datetime     DEFAULT NULL COMMENT '下次执行时间',
  last_run_at   datetime     DEFAULT NULL COMMENT '上次执行时间',
  last_status   tinyint      NOT NULL DEFAULT 0 COMMENT '上次结果：0 失败 / 1 成功',
  last_duration int          NOT NULL DEFAULT 0 COMMENT '上次耗时（毫秒）',
  last_output   text         COMMENT '上次输出摘要',
  creator       varchar(64)  DEFAULT NULL COMMENT '创建人',
  created_at    datetime     DEFAULT NULL COMMENT '创建时间',
  updater       varchar(64)  DEFAULT NULL COMMENT '更新人',
  updated_at    datetime     DEFAULT NULL COMMENT '更新时间',
  PRIMARY KEY (id),
  UNIQUE KEY uk_handler (handler),
  KEY idx_status_next (status, next_run_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='定时任务'`); err != nil {
			return err
		}
	}

	if !facades.Schema().HasTable("infra_job_log") {
		if err := facades.DB().Statement(`
CREATE TABLE infra_job_log (
  id          bigint      NOT NULL COMMENT '主键',
  job_id      bigint      NOT NULL COMMENT '任务 ID',
  job_name    varchar(64) NOT NULL COMMENT '任务名称（冗余，删任务后仍可查）',
  handler     varchar(64) NOT NULL COMMENT '处理器标识',
  trigger     varchar(16) NOT NULL DEFAULT 'cron' COMMENT '触发方式：cron / manual',
  status      tinyint     NOT NULL DEFAULT 1 COMMENT '结果：0 失败 / 1 成功',
  started_at  datetime    DEFAULT NULL COMMENT '开始时间',
  finished_at datetime    DEFAULT NULL COMMENT '结束时间',
  duration    int         NOT NULL DEFAULT 0 COMMENT '耗时（毫秒）',
  output      text        COMMENT '执行输出',
  error       text        COMMENT '失败原因',
  created_at  datetime    DEFAULT NULL COMMENT '创建时间',
  PRIMARY KEY (id),
  KEY idx_job_time (job_id, started_at),
  KEY idx_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='定时任务执行日志'`); err != nil {
			return err
		}
	}

	return nil
}

// Down 删除定时任务相关表。
func (r *M20261105000001CreateInfraJobTables) Down() error {
	if err := facades.Schema().DropIfExists("infra_job_log"); err != nil {
		return err
	}

	return facades.Schema().DropIfExists("infra_job")
}
