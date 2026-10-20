// Package middleware 提供会员模块（C 端）的请求中间件。
package middleware

import (
	"github.com/goravel/framework/contracts/http"

	baseAuth "eqray-go-starter/app/modules/base/auth"
	baseErrs "eqray-go-starter/app/modules/base/errs"
	"eqray-go-starter/app/modules/base/result"
	"eqray-go-starter/app/modules/base/support"
	memberAuth "eqray-go-starter/app/modules/member/auth"
)

// MemberAuth 校验会员访问令牌，通过后把会员 ID 写入上下文。
// 与后台 TokenAuth 使用不同的令牌作用域，两者不能互相冒充。
type MemberAuth struct{}

// Signature 是中间件的唯一标识。
func (m *MemberAuth) Signature() string {
	return "member_auth"
}

// Handle 校验会员访问令牌，未通过时直接中断请求。
func (m *MemberAuth) Handle(ctx http.Context) {
	token := baseAuth.BearerToken(ctx.Request().Header("Authorization"))
	memberID, err := memberAuth.Resolve(token)
	if err != nil {
		result.MarkFail(ctx, baseErrs.Unauthorized)
		ctx.Request().Abort(http.StatusUnauthorized)
		return
	}

	ctx.WithContext(support.WithMemberID(ctx.Context(), int64(memberID)))
	ctx.Request().Next()
}
