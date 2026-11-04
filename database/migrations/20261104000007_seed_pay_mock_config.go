package migrations

import (
	"time"

	"eqray-go-starter/app/facades"
	"eqray-go-starter/app/modules/base/support"
)

type M20261104000007SeedPayMockConfig struct{}

// Signature The unique signature for the migration.
func (r *M20261104000007SeedPayMockConfig) Signature() string {
	return "20261104000007_seed_pay_mock_config"
}

// Up 写入模拟支付开关（默认关闭）。生产环境不要开启，
// 开启后 C 端可对 mock 渠道订单调用「模拟支付成功」接口。
func (r *M20261104000007SeedPayMockConfig) Up() error {
	if !facades.Schema().HasTable("infra_config") {
		return nil
	}

	const configKey = "pay.mock.enabled"

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
		"category":   "支付",
		"name":       "模拟支付开关",
		"config_key": configKey,
		"value":      "0",
		"type":       1,
		"visible":    1,
		"remark":     "1 开启 / 0 关闭；仅联调环境开启，开启后 mock 渠道订单可一键模拟支付成功",
		"creator":    "system",
		"created_at": now,
		"updater":    "system",
		"updated_at": now,
	})

	return err
}

// Down 删除模拟支付开关。
func (r *M20261104000007SeedPayMockConfig) Down() error {
	if !facades.Schema().HasTable("infra_config") {
		return nil
	}

	_, err := facades.DB().Table("infra_config").Where("config_key", "pay.mock.enabled").Delete()

	return err
}
