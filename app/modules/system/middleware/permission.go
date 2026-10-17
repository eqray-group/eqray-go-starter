// Package middleware 存放 system 模块的中间件（依赖本模块表，故不放在 base）。
package middleware

import (
	"strings"

	"github.com/goravel/framework/contracts/http"

	"eqray-go-starter/app/modules/base/support"
	"eqray-go-starter/app/modules/base/types"
	"eqray-go-starter/app/modules/system/service"
)

// Permission 按路由声明的权限标识做接口级 RBAC 校验（docs 9.3）。
// 超级管理员（角色 code = SUPER_ADMIN）直接放行。
type Permission struct {
	Code string
}

// Signature 让每个权限标识对应独立的中间件签名。
func (m *Permission) Signature() string {
	return "permission_" + strings.NewReplacer(":", "_", "-", "_").Replace(m.Code)
}

// Handle 校验当前用户是否拥有路由要求的权限。
func (m *Permission) Handle(ctx http.Context) {
	userID := support.UserID(ctx.Context())
	if userID == 0 {
		ctx.Request().Abort(http.StatusUnauthorized)
		return
	}

	permission := service.NewPermissionService()

	super, err := permission.IsSuperAdmin(ctx.Context(), types.ID(userID))
	if err != nil {
		ctx.Request().Abort(http.StatusInternalServerError)
		return
	}
	if super {
		ctx.Request().Next()
		return
	}

	codes, err := permission.Permissions(ctx.Context(), types.ID(userID))
	if err != nil {
		ctx.Request().Abort(http.StatusInternalServerError)
		return
	}

	for _, code := range codes {
		if code == m.Code {
			ctx.Request().Next()
			return
		}
	}

	ctx.Request().Abort(http.StatusForbidden)
}
