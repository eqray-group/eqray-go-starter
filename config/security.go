package config

import (
	"eqray-go-starter/app/facades"
)

func init() {
	config := facades.Config()
	config.Add("security", map[string]any{
		// 接口限流：按「登录用户 ID 或客户端 IP」做每分钟固定窗口计数
		"rate_limit": map[string]any{
			// 是否启用全局接口限流
			"enabled": config.Env("SECURITY_RATE_LIMIT_ENABLED", true),
			// 单个主体每分钟允许的请求数
			"per_minute": config.Env("SECURITY_RATE_LIMIT_PER_MINUTE", 240),
			// 登录 / 验证码等敏感接口每分钟允许的请求数
			"login_per_minute": config.Env("SECURITY_RATE_LIMIT_LOGIN_PER_MINUTE", 10),
		},
		// 登录安全：连续失败达阈值后锁定一段时间
		"login": map[string]any{
			// 允许的连续失败次数
			"max_fail": config.Env("SECURITY_LOGIN_MAX_FAIL", 5),
			// 统计窗口与锁定时长（分钟）
			"lock_minutes": config.Env("SECURITY_LOGIN_LOCK_MINUTES", 15),
		},
	})
}
