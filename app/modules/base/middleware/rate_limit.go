package middleware

import (
	"fmt"
	"strings"
	"time"

	"github.com/goravel/framework/contracts/http"

	"eqray-go-starter/app/facades"
	"eqray-go-starter/app/modules/base/support"
)

// RateLimit 是基于缓存（Redis）的固定窗口限流中间件。
// 计数主体优先取登录用户 ID，未登录时退化为客户端 IP，避免同一出口 IP 的多个用户互相影响。
type RateLimit struct {
	// Name 限流器名称，用于区分不同策略的计数键
	Name string
	// Limit 窗口内允许的请求数，<=0 时读取配置 security.rate_limit.per_minute
	Limit int
	// Window 统计窗口，默认 1 分钟
	Window time.Duration
}

// Signature 让不同名称的限流中间件拥有独立签名。
func (m *RateLimit) Signature() string {
	return "rate_limit_" + strings.NewReplacer(":", "_", "-", "_").Replace(m.Name)
}

// Handle 累加计数并判断是否超限；超限返回 429，不进入后续处理器。
func (m *RateLimit) Handle(ctx http.Context) {
	if !facades.Config().GetBool("security.rate_limit.enabled", true) {
		ctx.Request().Next()
		return
	}

	limit := m.Limit
	if limit <= 0 {
		limit = facades.Config().GetInt("security.rate_limit.per_minute", 240)
	}

	window := m.Window
	if window <= 0 {
		window = time.Minute
	}

	key := fmt.Sprintf("security:rate:%s:%s", m.Name, rateLimitSubject(ctx))
	count, err := facades.Cache().Increment(key, 1)
	if err != nil {
		// 缓存不可用时放行，避免限流组件本身把接口打挂
		ctx.Request().Next()
		return
	}
	if count == 1 {
		// 首次计数时补上过期时间，形成固定窗口
		_ = facades.Cache().Put(key, count, window)
	}

	if count > int64(limit) {
		ctx.Request().Abort(http.StatusTooManyRequests)
		return
	}

	ctx.Request().Next()
}

// rateLimitSubject 取限流计数主体：已登录用用户 ID，未登录用客户端 IP。
// 入参 ctx 为请求上下文；返回主体标识。
func rateLimitSubject(ctx http.Context) string {
	if uid := support.UserID(ctx.Context()); uid != 0 {
		return "u" + fmt.Sprintf("%d", uid)
	}

	return "ip" + ctx.Request().Ip()
}
