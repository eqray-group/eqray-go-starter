package migrations

import (
	"time"

	"eqray-go-starter/app/facades"
	"eqray-go-starter/app/modules/base/support"
)

type M20261104000003SeedAmapServerKey struct{}

// Signature The unique signature for the migration.
func (r *M20261104000003SeedAmapServerKey) Signature() string {
	return "20261104000003_seed_amap_server_key"
}

// Up 写入高德「Web 服务」Key 参数，供后端一次性拉取省市区数据入库。
// 该 Key 与前端 JS Key 不是同一种类型，也不在任何公开接口中返回。
func (r *M20261104000003SeedAmapServerKey) Up() error {
	if !facades.Schema().HasTable("infra_config") {
		return nil
	}

	const configKey = "map.amap.serverKey"

	total, err := facades.DB().Table("infra_config").
		Where("config_key", configKey).
		Count()
	if err != nil {
		return err
	}
	if total > 0 {
		return nil
	}

	now := time.Now()
	_, err = facades.DB().Table("infra_config").Insert(map[string]any{
		"id":         support.NewID(),
		"category":   "地图",
		"name":       "高德 Web 服务 Key",
		"config_key": configKey,
		"value":      "",
		"type":       1,
		"visible":    0,
		"remark":     "仅后端使用：同步省市区数据（行政区划），不会暴露给前端；与 JS API Key 不是同一种类型",
		"creator":    "system",
		"created_at": now,
		"updater":    "system",
		"updated_at": now,
	})

	return err
}

// Down 删除高德 Web 服务 Key 参数。
func (r *M20261104000003SeedAmapServerKey) Down() error {
	if !facades.Schema().HasTable("infra_config") {
		return nil
	}

	_, err := facades.DB().Table("infra_config").
		Where("config_key", "map.amap.serverKey").
		Delete()

	return err
}
