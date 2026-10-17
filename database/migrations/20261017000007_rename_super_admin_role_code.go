package migrations

import (
	"eqray-go-starter/app/facades"
	"eqray-go-starter/app/modules/base/types"
)

type M20261017000007RenameSuperAdminRoleCode struct{}

// Signature The unique signature for the migration.
func (r *M20261017000007RenameSuperAdminRoleCode) Signature() string {
	return "20261017000007_rename_super_admin_role_code"
}

// Up 把存量库里的内置超管角色标识从 admin 改为 SUPER_ADMIN，
// 与代码中的放行判断保持一致；新建库由 0003 的初始化数据直接写入新标识。
func (r *M20261017000007RenameSuperAdminRoleCode) Up() error {
	if !facades.Schema().HasTable("system_role") {
		return nil
	}

	// 已经是新标识（例如新建库）时直接跳过，避免撞唯一键
	count, err := facades.DB().Table("system_role").
		Where("code", "SUPER_ADMIN").
		WhereNull("deleted_at").
		Count()
	if err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	_, err = facades.DB().Table("system_role").
		Where("code", "admin").
		Where("type", 1).
		Update(map[string]any{
			"code":       "SUPER_ADMIN",
			"updater":    "system",
			"updated_at": types.Now(),
		})

	return err
}

// Down 回滚为 admin。
func (r *M20261017000007RenameSuperAdminRoleCode) Down() error {
	if !facades.Schema().HasTable("system_role") {
		return nil
	}

	_, err := facades.DB().Table("system_role").
		Where("code", "SUPER_ADMIN").
		Where("type", 1).
		Update(map[string]any{
			"code":       "admin",
			"updater":    "system",
			"updated_at": types.Now(),
		})

	return err
}
