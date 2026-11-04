package migrations

import (
	"strconv"
	"time"

	"eqray-go-starter/app/facades"
	"eqray-go-starter/app/modules/base/support"
	"eqray-go-starter/app/modules/base/types"
	"eqray-go-starter/app/modules/system/model"
)

// M20261104000006SeedPayWalletMenu 在「支付管理」下新增「钱包」页面与按钮权限点。
type M20261104000006SeedPayWalletMenu struct{}

// Signature The unique signature for the migration.
func (r *M20261104000006SeedPayWalletMenu) Signature() string {
	return "20261104000006_seed_pay_wallet_menu"
}

// Up 新增钱包菜单（列表 / 调账权限），并授权给内置角色。
func (r *M20261104000006SeedPayWalletMenu) Up() error {
	if !facades.Schema().HasTable("system_menu") {
		return nil
	}

	now := time.Now()

	var payDir model.SystemMenuModel
	if err := facades.DB().Table("system_menu").
		Select("id").
		Where("name", "支付管理").
		Where("type", 1).
		WhereNull("deleted_at").
		First(&payDir); err != nil {
		return err
	}
	if payDir.ID == 0 {
		// 支付菜单未初始化时跳过，等支付菜单迁移执行完再手动跑一次即可
		return nil
	}

	walletID, err := r.ensureMenu(payDir.ID, "钱包", 2, 5, "wallet", "/pay/wallet", "ri:wallet-3-line", "", now)
	if err != nil {
		return err
	}

	buttons := [][2]string{
		{"钱包查询", "pay:wallet:list"},
		{"余额调整", "pay:wallet:adjust"},
	}
	for index, button := range buttons {
		if _, err := r.ensureMenu(walletID, button[0], 3, index+1, "", "", "", button[1], now); err != nil {
			return err
		}
	}

	if err := r.grantBuiltinRoles(now); err != nil {
		return err
	}

	_ = facades.Cache().Put("rbac:version", strconv.FormatInt(time.Now().UnixNano(), 10), 30*time.Minute)

	return nil
}

// ensureMenu 按「父级 + 类型 + 名称」查找菜单，不存在时插入。
func (r *M20261104000006SeedPayWalletMenu) ensureMenu(parentID types.ID, name string, menuType, sort int, path, component, icon, permission string, now time.Time) (types.ID, error) {
	var existing model.SystemMenuModel
	if err := facades.DB().Table("system_menu").
		Select("id").
		Where("parent_id", parentID).
		Where("type", menuType).
		Where("name", name).
		WhereNull("deleted_at").
		First(&existing); err != nil {
		return 0, err
	}
	if existing.ID > 0 {
		return existing.ID, nil
	}

	id := support.NewID()
	_, err := facades.DB().Table("system_menu").Insert(map[string]any{
		"id":             id,
		"name":           name,
		"permission":     permission,
		"type":           menuType,
		"sort":           sort,
		"parent_id":      parentID,
		"path":           path,
		"icon":           icon,
		"component":      component,
		"component_name": "",
		"status":         1,
		"visible":        1,
		"keep_alive":     1,
		"always_show":    0,
		"creator":        "system",
		"created_at":     now,
		"updater":        "system",
		"updated_at":     now,
	})

	return id, err
}

// grantBuiltinRoles 给内置角色补齐缺失的菜单授权。
func (r *M20261104000006SeedPayWalletMenu) grantBuiltinRoles(now time.Time) error {
	if !facades.Schema().HasTable("system_role") || !facades.Schema().HasTable("system_role_menu") {
		return nil
	}

	roles := make([]model.SystemRoleModel, 0)
	if err := facades.DB().Table("system_role").
		Select("id").
		Where("type", 1).
		WhereNull("deleted_at").
		Get(&roles); err != nil {
		return err
	}

	menus := make([]model.SystemMenuModel, 0)
	if err := facades.DB().Table("system_menu").
		Select("id").
		WhereNull("deleted_at").
		Get(&menus); err != nil {
		return err
	}

	for _, role := range roles {
		granted := make(map[types.ID]bool, len(menus))

		rows := make([]model.SystemRoleMenuModel, 0)
		if err := facades.DB().Table("system_role_menu").
			Select("menu_id").
			Where("role_id", role.ID).
			WhereNull("deleted_at").
			Get(&rows); err != nil {
			return err
		}
		for _, row := range rows {
			granted[row.MenuID] = true
		}

		for _, menu := range menus {
			if granted[menu.ID] {
				continue
			}

			if _, err := facades.DB().Table("system_role_menu").Insert(map[string]any{
				"id":         support.NewID(),
				"role_id":    role.ID,
				"menu_id":    menu.ID,
				"creator":    "system",
				"created_at": now,
				"updater":    "system",
				"updated_at": now,
			}); err != nil {
				return err
			}
		}
	}

	return nil
}

// Down 删除钱包菜单及其角色授权。
func (r *M20261104000006SeedPayWalletMenu) Down() error {
	if !facades.Schema().HasTable("system_menu") {
		return nil
	}

	var wallet model.SystemMenuModel
	if err := facades.DB().Table("system_menu").
		Select("id").
		Where("name", "钱包").
		Where("type", 2).
		WhereNull("deleted_at").
		First(&wallet); err != nil || wallet.ID == 0 {
		return err
	}

	buttons := make([]model.SystemMenuModel, 0)
	if err := facades.DB().Table("system_menu").Select("id").Where("parent_id", wallet.ID).Get(&buttons); err != nil {
		return err
	}
	for _, button := range buttons {
		if _, err := facades.DB().Table("system_menu").Where("id", button.ID).Delete(); err != nil {
			return err
		}
	}

	_, err := facades.DB().Table("system_menu").Where("id", wallet.ID).Delete()

	return err
}
