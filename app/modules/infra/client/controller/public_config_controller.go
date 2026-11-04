// Package controller 是 infra 模块 C 端接入层。
package controller

import (
	"strings"

	"github.com/goravel/framework/contracts/http"

	"eqray-go-starter/app/modules/base/result"
	infraService "eqray-go-starter/app/modules/infra/service"
)

// PublicConfigController 提供 C 端只读的公开参数（如高德地图 Key）。
// 只返回 service 层白名单里的键，避免把敏感参数（支付密钥、短信密钥）暴露出去。
type PublicConfigController struct {
	configs *infraService.ConfigService
}

// NewPublicConfigController 构造公开参数控制器。
func NewPublicConfigController() *PublicConfigController {
	return &PublicConfigController{configs: infraService.NewConfigService()}
}

// Show 批量读取公开参数，返回「键 → 值」映射。
// 入参：keys（查询串，英文逗号分隔，如 map.amap.key,map.amap.securityCode）；
// 返回仅包含白名单键且值非空的映射，未配置的键不会出现在结果里。
func (c *PublicConfigController) Show(ctx http.Context) http.Response {
	keys := strings.Split(ctx.Request().Query("keys"), ",")
	values := make(map[string]string, len(keys))

	for _, key := range keys {
		key = strings.TrimSpace(key)
		if key == "" || !infraService.IsPublicConfigKey(key) {
			continue
		}

		if value := c.configs.GetString(ctx.Context(), key, ""); value != "" {
			values[key] = value
		}
	}

	return result.Success(ctx, values)
}
