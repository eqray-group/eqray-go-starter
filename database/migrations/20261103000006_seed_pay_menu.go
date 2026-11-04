package migrations

import (
	"strconv"
	"time"

	"eqray-go-starter/app/facades"
	"eqray-go-starter/app/modules/base/support"
	"eqray-go-starter/app/modules/base/types"
	"eqray-go-starter/app/modules/system/model"
)

type M20261103000006SeedPayMenu struct{}

// Signature The unique signature for the migration.
func (r *M20261103000006SeedPayMenu) Signature() string {
	return "20261103000006_seed_pay_menu"
}

// Up 新增「支付管理」一级目录与支付渠道 / 支付订单 / 退款订单 / 回调日志四个页面及按钮权限点，
// 并授权给内置角色；随后前移 RBAC 缓存版本，让管理员立即拿到新权限。
func (r *M20261103000006SeedPayMenu) Up() error {
	if !facades.Schema().HasTable("system_menu") {
		return nil
	}

	now := time.Now()

	payID, err := r.ensureMenu(0, "支付管理", 1, 30, "/pay", "", "ri:bank-card-line", "", now)
	if err != nil {
		return err
	}

	pages := []struct {
		name      string
		path      string
		component string
		icon      string
		sort      int
		buttons   [][2]string
	}{
		{
			"支付渠道", "channel", "/pay/channel", "ri:bank-card-2-line", 1,
			[][2]string{
				{"渠道查询", "pay:channel:list"},
				{"渠道新增", "pay:channel:create"},
				{"渠道修改", "pay:channel:update"},
				{"渠道删除", "pay:channel:delete"},
			},
		},
		{
			"支付订单", "order", "/pay/order", "ri:file-list-3-line", 2,
			[][2]string{
				{"订单查询", "pay:order:list"},
				{"订单操作", "pay:order:update"},
			},
		},
		{
			"退款订单", "refund", "/pay/refund", "ri:refund-2-line", 3,
			[][2]string{
				{"退款查询", "pay:refund:list"},
				{"发起退款", "pay:refund:create"},
			},
		},
		{
			"回调日志", "notify-log", "/pay/notify-log", "ri:terminal-box-line", 4,
			[][2]string{
				{"日志查询", "pay:notify-log:list"},
			},
		},
	}

	for _, page := range pages {
		menuID, err := r.ensureMenu(payID, page.name, 2, page.sort, page.path, page.component, page.icon, "", now)
		if err != nil {
			return err
		}

		for index, button := range page.buttons {
			if _, err := r.ensureMenu(menuID, button[0], 3, index+1, "", "", "", button[1], now); err != nil {
				return err
			}
		}
	}

	if err := r.grantBuiltinRoles(now); err != nil {
		return err
	}

	// 新增权限点后前移权限缓存版本，避免管理员要等 30 分钟才能拿到新按钮权限
	_ = facades.Cache().Put("rbac:version", strconv.FormatInt(time.Now().UnixNano(), 10), 30*time.Minute)

	return nil
}

// ensureMenu 按「父级 + 类型 + 名称」查找菜单，不存在时插入。
func (r *M20261103000006SeedPayMenu) ensureMenu(parentID types.ID, name string, menuType, sort int, path, component, icon, permission string, now time.Time) (types.ID, error) {
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

// grantBuiltinRoles 给内置角色补齐全部菜单授权。
func (r *M20261103000006SeedPayMenu) grantBuiltinRoles(now time.Time) error {
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

// Down 删除「支付管理」目录及全部子菜单、按钮与授权。
func (r *M20261103000006SeedPayMenu) Down() error {
	if !facades.Schema().HasTable("system_menu") {
		return nil
	}

	dirs := make([]model.SystemMenuModel, 0)
	if err := facades.DB().Table("system_menu").
		Select("id").
		Where("name", "支付管理").
		Where("type", 1).
		WhereNull("deleted_at").
		Get(&dirs); err != nil {
		return err
	}

	for _, dir := range dirs {
		children := make([]model.SystemMenuModel, 0)
		if err := facades.DB().Table("system_menu").Select("id").Where("parent_id", dir.ID).Get(&children); err != nil {
			return err
		}
		for _, page := range children {
			buttons := make([]model.SystemMenuModel, 0)
			if err := facades.DB().Table("system_menu").Select("id").Where("parent_id", page.ID).Get(&buttons); err != nil {
				return err
			}
			for _, button := range buttons {
				if err := r.removeMenu(button.ID); err != nil {
					return err
				}
			}
			if err := r.removeMenu(page.ID); err != nil {
				return err
			}
		}
		if err := r.removeMenu(dir.ID); err != nil {
			return err
		}
	}

	return nil
}

// removeMenu 删除单个菜单及其角色授权。
func (r *M20261103000006SeedPayMenu) removeMenu(id types.ID) error {
	if facades.Schema().HasTable("system_role_menu") {
		if _, err := facades.DB().Table("system_role_menu").Where("menu_id", id).Delete(); err != nil {
			return err
		}
	}
	if _, err := facades.DB().Table("system_menu").Where("id", id).Delete(); err != nil {
		return err
	}

	return nil
}
