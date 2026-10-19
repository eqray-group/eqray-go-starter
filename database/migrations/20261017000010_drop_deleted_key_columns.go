package migrations

import (
	"eqray-go-starter/app/facades"
)

type M20261017000010DropDeletedKeyColumns struct{}

// Signature The unique signature for the migration.
func (r *M20261017000010DropDeletedKeyColumns) Signature() string {
	return "20261017000010_drop_deleted_key_columns"
}

// Up 去掉 system_role、system_user_role、system_role_menu 的 deleted_key 生成列，
// 唯一键直接建在业务键上。关联表改为硬删除，历史软删行先清理，避免占用新唯一键。
func (r *M20261017000010DropDeletedKeyColumns) Up() error {
	if facades.Schema().HasTable("system_user_role") && facades.Schema().HasColumn("system_user_role", "deleted_key") {
		if _, err := facades.DB().Table("system_user_role").WhereNotNull("deleted_at").Delete(); err != nil {
			return err
		}
		if err := facades.DB().Statement(`
ALTER TABLE system_user_role
  DROP INDEX uk_user_role,
  ADD UNIQUE KEY uk_user_role (user_id, role_id),
  DROP COLUMN deleted_key`); err != nil {
			return err
		}
	}

	if facades.Schema().HasTable("system_role_menu") && facades.Schema().HasColumn("system_role_menu", "deleted_key") {
		if _, err := facades.DB().Table("system_role_menu").WhereNotNull("deleted_at").Delete(); err != nil {
			return err
		}
		if err := facades.DB().Statement(`
ALTER TABLE system_role_menu
  DROP INDEX uk_role_menu,
  ADD UNIQUE KEY uk_role_menu (role_id, menu_id),
  DROP COLUMN deleted_key`); err != nil {
			return err
		}
	}

	if facades.Schema().HasTable("system_role") && facades.Schema().HasColumn("system_role", "deleted_key") {
		if err := facades.DB().Statement(`
ALTER TABLE system_role
  DROP INDEX uk_code,
  ADD UNIQUE KEY uk_code (code),
  DROP COLUMN deleted_key`); err != nil {
			return err
		}
	}

	return nil
}

// Down 恢复生成列与带 deleted_key 的唯一键。
func (r *M20261017000010DropDeletedKeyColumns) Down() error {
	if facades.Schema().HasTable("system_user_role") && !facades.Schema().HasColumn("system_user_role", "deleted_key") {
		if err := facades.DB().Statement(`
ALTER TABLE system_user_role
  ADD COLUMN deleted_key datetime GENERATED ALWAYS AS (IFNULL(deleted_at, '1970-01-01 00:00:00')) STORED COMMENT '软删除让位生成列',
  DROP INDEX uk_user_role,
  ADD UNIQUE KEY uk_user_role (user_id, role_id, deleted_key)`); err != nil {
			return err
		}
	}

	if facades.Schema().HasTable("system_role_menu") && !facades.Schema().HasColumn("system_role_menu", "deleted_key") {
		if err := facades.DB().Statement(`
ALTER TABLE system_role_menu
  ADD COLUMN deleted_key datetime GENERATED ALWAYS AS (IFNULL(deleted_at, '1970-01-01 00:00:00')) STORED COMMENT '软删除让位生成列',
  DROP INDEX uk_role_menu,
  ADD UNIQUE KEY uk_role_menu (role_id, menu_id, deleted_key)`); err != nil {
			return err
		}
	}

	if facades.Schema().HasTable("system_role") && !facades.Schema().HasColumn("system_role", "deleted_key") {
		if err := facades.DB().Statement(`
ALTER TABLE system_role
  ADD COLUMN deleted_key datetime GENERATED ALWAYS AS (IFNULL(deleted_at, '1970-01-01 00:00:00')) STORED COMMENT '软删除让位生成列',
  DROP INDEX uk_code,
  ADD UNIQUE KEY uk_code (code, deleted_key)`); err != nil {
			return err
		}
	}

	return nil
}
