package migrations

import (
	"time"

	"eqray-go-starter/app/facades"
	"eqray-go-starter/app/modules/base/support"
)

type M20261017000003CreateSystemRoleTable struct{}

// Signature The unique signature for the migration.
func (r *M20261017000003CreateSystemRoleTable) Signature() string {
	return "20261017000003_create_system_role_table"
}

// Up 建 system_role 表并写入内置角色 SUPER_ADMIN（docs 6.3）。
// 角色编码允许「删除后复用」，因此唯一键带上 deleted_key 生成列。
func (r *M20261017000003CreateSystemRoleTable) Up() error {
	if facades.Schema().HasTable("system_role") {
		return nil
	}

	if err := facades.DB().Statement(`
CREATE TABLE system_role (
  id           bigint       NOT NULL COMMENT '主键',
  name         varchar(64)  NOT NULL COMMENT '角色名称',
  code         varchar(64)  NOT NULL COMMENT '角色标识，唯一，如 SUPER_ADMIN',
  sort         int          NOT NULL DEFAULT 0 COMMENT '排序',
  status       tinyint      NOT NULL DEFAULT 1 COMMENT '状态：0 禁用 / 1 启用',
  type         tinyint      NOT NULL DEFAULT 2 COMMENT '类型：1 内置角色 / 2 自定义角色',
  remark       varchar(500) NOT NULL DEFAULT '' COMMENT '备注',
  creator      varchar(64)  NOT NULL DEFAULT '' COMMENT '创建人',
  created_at   datetime     NULL COMMENT '创建时间',
  updater      varchar(64)  NOT NULL DEFAULT '' COMMENT '更新人',
  updated_at   datetime     NULL COMMENT '更新时间',
  deleted_at   datetime     NULL COMMENT '删除时间',
  deleted_key  datetime     GENERATED ALWAYS AS (IFNULL(deleted_at, '1970-01-01 00:00:00')) STORED COMMENT '软删除让位生成列',
  PRIMARY KEY (id),
  UNIQUE KEY uk_code (code, deleted_key)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='角色表'`); err != nil {
		return err
	}

	return r.seedAdminRole()
}

// seedAdminRole 写入内置超管角色，权限校验时按 code = SUPER_ADMIN 直接放行（docs 9.3）。
func (r *M20261017000003CreateSystemRoleTable) seedAdminRole() error {
	now := time.Now()

	_, err := facades.DB().Table("system_role").Insert(map[string]any{
		"id":         support.NewID(),
		"name":       "超级管理员",
		"code":       "SUPER_ADMIN",
		"sort":       0,
		"status":     1,
		"type":       1,
		"remark":     "内置角色，拥有全部权限",
		"creator":    "system",
		"created_at": now,
		"updater":    "system",
		"updated_at": now,
	})

	return err
}

// Down Reverse the migrations.
func (r *M20261017000003CreateSystemRoleTable) Down() error {
	return facades.Schema().DropIfExists("system_role")
}
