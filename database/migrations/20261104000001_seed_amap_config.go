package migrations

import (
	"time"

	"eqray-go-starter/app/facades"
	"eqray-go-starter/app/modules/base/support"
)

type M20261104000001SeedAmapConfig struct{}

// Signature The unique signature for the migration.
func (r *M20261104000001SeedAmapConfig) Signature() string {
	return "20261104000001_seed_amap_config"
}

// Up 写入高德地图的内置参数（JS Key / 安全密钥 / Web 服务 Key），值留空由管理员在「参数配置」里填写。
// 前端通过 GET /api/v1/client/infra/public-configs 免登录读取前两个键；
// serverKey 仅后端使用（同步省市区），visible=0，不会出现在公开接口里。
func (r *M20261104000001SeedAmapConfig) Up() error {
	if !facades.Schema().HasTable("infra_config") {
		return nil
	}

	now := time.Now()
	presets := []struct {
		name    string
		key     string
		remark  string
		visible int
	}{
		{
			name:    "高德地图 Key",
			key:     "map.amap.key",
			remark:  "高德开放平台「Web 端（JS API）」Key，用于地图选点",
			visible: 1,
		},
		{
			name:    "高德地图安全密钥",
			key:     "map.amap.securityCode",
			remark:  "2021 年 12 月后申请的 Key 必须配套填写安全密钥（securityJsCode），老 Key 可留空",
			visible: 1,
		},
		{
			name:    "高德 Web 服务 Key",
			key:     "map.amap.serverKey",
			remark:  "仅后端使用：同步省市区数据（行政区划），不会暴露给前端；与 JS API Key 不是同一种类型",
			visible: 0,
		},
	}

	for _, preset := range presets {
		total, err := facades.DB().Table("infra_config").
			Where("config_key", preset.key).
			Count()
		if err != nil {
			return err
		}
		if total > 0 {
			continue
		}

		if _, err := facades.DB().Table("infra_config").Insert(map[string]any{
			"id":         support.NewID(),
			"category":   "地图",
			"name":       preset.name,
			"config_key": preset.key,
			"value":      "",
			"type":       1,
			"visible":    preset.visible,
			"remark":     preset.remark,
			"creator":    "system",
			"created_at": now,
			"updater":    "system",
			"updated_at": now,
		}); err != nil {
			return err
		}
	}

	return nil
}

// Down 删除高德地图内置参数。
func (r *M20261104000001SeedAmapConfig) Down() error {
	if !facades.Schema().HasTable("infra_config") {
		return nil
	}

	_, err := facades.DB().Table("infra_config").
		WhereIn("config_key", []any{"map.amap.key", "map.amap.securityCode", "map.amap.serverKey"}).
		Delete()

	return err
}
