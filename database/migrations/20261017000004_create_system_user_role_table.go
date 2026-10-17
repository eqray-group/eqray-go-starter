package migrations

import (
	"time"

	"eqray-go-starter/app/facades"
	"eqray-go-starter/app/modules/base/support"
	"eqray-go-starter/app/modules/system/model"
)

type M20261017000004CreateSystemUserRoleTable struct{}

// Signature The unique signature for the migration.
func (r *M20261017000004CreateSystemUserRoleTable) Signature() string {
	return "20261017000004_create_system_user_role_table"
}

// Up 建 system_user_role 关联表（docs 6.3）。
// 关联表走软删除 + deleted_key，保证「解绑后重新绑定同一角色」不会撞唯一键（docs 6.8(4)）。
func (r *M20261017000004CreateSystemUserRoleTable) Up() error {
	if facades.Schema().HasTable("system_user_role") {
		return nil
	}

	if err := facades.DB().Statement(`
CREATE TABLE system_user_role (
  id           bigint   NOT NULL COMMENT '主键',
  user_id      bigint   NOT NULL COMMENT '用户 ID',
  role_id      bigint   NOT NULL COMMENT '角色 ID',
  creator      varchar(64) NOT NULL DEFAULT '' COMMENT '创建人',
  created_at   datetime NULL COMMENT '创建时间',
  updater      varchar(64) NOT NULL DEFAULT '' COMMENT '更新人',
  updated_at   datetime NULL COMMENT '更新时间',
  deleted_at   datetime NULL COMMENT '删除时间',
  deleted_key  datetime GENERATED ALWAYS AS (IFNULL(deleted_at, '1970-01-01 00:00:00')) STORED COMMENT '软删除让位生成列',
  PRIMARY KEY (id),
  UNIQUE KEY uk_user_role (user_id, role_id, deleted_key)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户角色关联表'`); err != nil {
		return err
	}

	return r.seedAdminUserRole()
}

// seedAdminUserRole 给初始管理员绑定内置角色，保证开箱即用。
func (r *M20261017000004CreateSystemUserRoleTable) seedAdminUserRole() error {
	var user model.SystemUserModel
	var role model.SystemRoleModel

	if err := facades.DB().Table("system_users").
		Select("id").
		Where("username", "admin").
		WhereNull("deleted_at").
		First(&user); err != nil {
		return err
	}
	if err := facades.DB().Table("system_role").
		Select("id").
		Where("code", "SUPER_ADMIN").
		WhereNull("deleted_at").
		First(&role); err != nil {
		return err
	}
	if user.ID == 0 || role.ID == 0 {
		return nil
	}

	now := time.Now()

	_, err := facades.DB().Table("system_user_role").Insert(map[string]any{
		"id":         support.NewID(),
		"user_id":    user.ID,
		"role_id":    role.ID,
		"creator":    "system",
		"created_at": now,
		"updater":    "system",
		"updated_at": now,
	})

	return err
}

// Down Reverse the migrations.
func (r *M20261017000004CreateSystemUserRoleTable) Down() error {
	return facades.Schema().DropIfExists("system_user_role")
}
