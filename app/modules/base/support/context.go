package support

import "context"

type userIDKey struct{}

// WithUserID 把当前登录者写入上下文，供 Model 基类自动填充 creator / updater。
func WithUserID(ctx context.Context, id int64) context.Context {
	return context.WithValue(ctx, userIDKey{}, id)
}

// UserID 从上下文取当前登录者，未登录或未注入时返回 0。
func UserID(ctx context.Context) int64 {
	if ctx == nil {
		return 0
	}
	if v, ok := ctx.Value(userIDKey{}).(int64); ok {
		return v
	}
	return 0
}
