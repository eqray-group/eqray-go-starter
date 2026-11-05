package migrations

import (
	"strconv"
	"time"

	"eqray-go-starter/app/facades"
	"eqray-go-starter/app/modules/base/support"
	"eqray-go-starter/app/modules/base/types"
	"eqray-go-starter/app/modules/system/model"
)

// M20261105000002SeedInfraJobMenu 新增「基础设施 / 定时任务」菜单与权限点，
// 并把原先硬编码在 bootstrap/schedule.go 的 4 个任务写入 infra_job（处理器标识与默认 cron）。
type M20261105000002SeedInfraJobMenu struct{}

// Signature The unique signature for the migration.
func (r *M20261105000002SeedInfraJobMenu) Signature() string {
	return "20261105000002_seed_infra_job_menu"
}

// Up 新增菜单、权限点与内置任务。
func (r *M20261105000002SeedInfraJobMenu) Up() error {
	now := time.Now()

	if err := r.seedJobs(now); err != nil {
		return err
	}

	if !facades.Schema().HasTable("system_menu") {
		return nil
	}

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

	jobPageID, err := r.ensureMenu(infraDir.ID, "定时任务", 2, 6, "job", "/infra/job", "ri:timer-line", "", now)
	if err != nil {
		return err
	}

	buttons := []struct {
		name       string
		permission string
		sort       int
	}{
		{"任务查询", "infra:job:list", 1},
		{"任务编辑", "infra:job:update", 2},
		{"任务删除", "infra:job:delete", 3},
		{"手动执行", "infra:job:run", 4},
	}
	for _, button := range buttons {
		if _, err := r.ensureMenu(jobPageID, button.name, 3, button.sort, "", "", "", button.permission, now); err != nil {
			return err
		}
	}

	if err := r.grantBuiltinRoles(now); err != nil {
		return err
	}

	_ = facades.Cache().Put("rbac:version", strconv.FormatInt(time.Now().UnixNano(), 10), 30*time.Minute)

	return nil
}

// seedJobs 写入内置定时任务；按 handler 判重，重复执行安全。
// next_run_at 留空，由调度器首次分发时补算，避免迁移时集中补跑。
// 入参 now 为当前时间；返回写库错误。
func (r *M20261105000002SeedInfraJobMenu) seedJobs(now time.Time) error {
	if !facades.Schema().HasTable("infra_job") {
		return nil
	}

	presets := []struct {
		name    string
		handler string
		cron    string
		remark  string
	}{
		{"清理 API 访问日志", "infra:purge-api-access-log", "10 3 * * *", "每天 03:10 删除超过保留天数的访问日志"},
		{"关闭超时支付订单", "pay:close-expired-orders", "*/5 * * * *", "每 5 分钟关闭超时未支付订单"},
		{"同步退款中订单", "pay:sync-refunding", "*/5 * * * *", "每 5 分钟查单收敛退款中的退款单"},
		{"清理支付回调日志", "pay:purge-notify-logs", "40 3 * * *", "每天 03:40 删除过期的支付回调日志"},
	}

	for _, preset := range presets {
		total, err := facades.DB().Table("infra_job").Where("handler", preset.handler).Count()
		if err != nil {
			return err
		}
		if total > 0 {
			continue
		}

		if _, err := facades.DB().Table("infra_job").Insert(map[string]any{
			"id":            support.NewID(),
			"name":          preset.name,
			"handler":       preset.handler,
			"cron":          preset.cron,
			"status":        1,
			"remark":        preset.remark,
			"next_run_at":   nil,
			"last_status":   1,
			"last_duration": 0,
			"last_output":   "",
			"creator":       "system",
			"created_at":    now,
			"updater":       "system",
			"updated_at":    now,
		}); err != nil {
			return err
		}
	}

	return nil
}

// ensureMenu 按「父级 + 类型 + 名称」查找菜单，不存在时插入。
func (r *M20261105000002SeedInfraJobMenu) ensureMenu(parentID types.ID, name string, menuType, sort int, path, component, icon, permission string, now time.Time) (types.ID, error) {
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
func (r *M20261105000002SeedInfraJobMenu) grantBuiltinRoles(now time.Time) error {
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

// Down 删除定时任务菜单与内置任务。
func (r *M20261105000002SeedInfraJobMenu) Down() error {
	if facades.Schema().HasTable("infra_job") {
		if _, err := facades.DB().Table("infra_job_log").Delete(); err != nil {
			return err
		}
		if _, err := facades.DB().Table("infra_job").Delete(); err != nil {
			return err
		}
	}

	if !facades.Schema().HasTable("system_menu") {
		return nil
	}

	var page model.SystemMenuModel
	if err := facades.DB().Table("system_menu").
		Select("id").
		Where("name", "定时任务").
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
