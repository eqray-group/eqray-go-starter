package migrations

import (
	"strconv"
	"time"

	"eqray-go-starter/app/facades"
	"eqray-go-starter/app/modules/base/support"
	"eqray-go-starter/app/modules/base/types"
	"eqray-go-starter/app/modules/system/model"
)

// M20261105000005SeedInfraNoticeMenu 新增「基础设施 / 消息通知」菜单与权限点。
type M20261105000005SeedInfraNoticeMenu struct{}

// Signature The unique signature for the migration.
func (r *M20261105000005SeedInfraNoticeMenu) Signature() string {
	return "20261105000005_seed_infra_notice_menu"
}

// Up 新增页面与四个权限点。
func (r *M20261105000005SeedInfraNoticeMenu) Up() error {
	if !facades.Schema().HasTable("system_menu") {
		return nil
	}

	now := time.Now()

	var infraDir model.SystemMenuModel
	if err := facades.DB().Table("system_menu").
		Select("id").
		Where("name", "基础设施").
		Where("type", 1).
		WhereNull("deleted_at").
		First(&infraDir); err != nil {
		return err
	}
	if infraDir.ID == 0 {
		return nil
	}

	pageID, err := r.ensureMenu(infraDir.ID, "消息通知", 2, 8, "notice", "/infra/notice", "ri:notification-3-line", "", now)
	if err != nil {
		return err
	}

	buttons := []struct {
		name       string
		permission string
		sort       int
	}{
		{"通知查询", "infra:notice:list", 1},
		{"通知配置", "infra:notice:update", 2},
		{"通知删除", "infra:notice:delete", 3},
		{"通知发送", "infra:notice:send", 4},
	}
	for _, button := range buttons {
		if _, err := r.ensureMenu(pageID, button.name, 3, button.sort, "", "", "", button.permission, now); err != nil {
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
func (r *M20261105000005SeedInfraNoticeMenu) ensureMenu(parentID types.ID, name string, menuType, sort int, path, component, icon, permission string, now time.Time) (types.ID, error) {
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
func (r *M20261105000005SeedInfraNoticeMenu) grantBuiltinRoles(now time.Time) error {
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

// Down 删除消息通知菜单。
func (r *M20261105000005SeedInfraNoticeMenu) Down() error {
	if !facades.Schema().HasTable("system_menu") {
		return nil
	}

	var page model.SystemMenuModel
	if err := facades.DB().Table("system_menu").
		Select("id").
		Where("name", "消息通知").
		Where("type", 2).
		WhereNull("deleted_at").
		First(&page); err != nil || page.ID == 0 {
		return err
	}

	buttons := make([]model.SystemMenuModel, 0)
	if err := facades.DB().Table("system_menu").Select("id").Where("parent_id", page.ID).Get(&buttons); err != nil {
		return err
	}
	for _, button := range buttons {
		if _, err := facades.DB().Table("system_menu").Where("id", button.ID).Delete(); err != nil {
			return err
		}
	}

	_, err := facades.DB().Table("system_menu").Where("id", page.ID).Delete()

	return err
}
