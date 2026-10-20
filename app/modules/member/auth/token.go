// Package auth 提供 C 端会员的登录态，复用 base/auth 的令牌实现但使用独立的 Redis 作用域，
// 会员令牌与后台令牌互不通用。
package auth

import (
	baseAuth "eqray-go-starter/app/modules/base/auth"
	"eqray-go-starter/app/modules/base/types"
)

// tokenScope 是会员令牌的作用域，最终 Redis 键形如 auth:member:access:<hash>。
const tokenScope = "member"

// Issue 为会员签发一对令牌。
func Issue(memberID types.ID) (*baseAuth.TokenPair, error) {
	return baseAuth.IssueScoped(tokenScope, memberID)
}

// Resolve 用会员访问令牌换取会员 ID，无效令牌返回未登录错误。
func Resolve(accessToken string) (types.ID, error) {
	return baseAuth.ResolveScoped(tokenScope, accessToken)
}

// Refresh 用会员刷新令牌换一对新令牌。
func Refresh(refreshToken string) (*baseAuth.TokenPair, error) {
	return baseAuth.RefreshScoped(tokenScope, refreshToken)
}

// Revoke 注销会员令牌。
func Revoke(accessToken, refreshToken string) {
	baseAuth.RevokeScoped(tokenScope, accessToken, refreshToken)
}
