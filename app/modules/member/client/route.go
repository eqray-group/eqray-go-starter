// Package client 注册会员模块的 C 端接口（小程序 / H5 / App）。
// C 端访问的是「会员本人」的数据，只做登录态校验，不进入后台 RBAC 权限体系。
package client

import (
	"github.com/goravel/framework/contracts/route"

	"eqray-go-starter/app/facades"
	baseMiddleware "eqray-go-starter/app/modules/base/middleware"
	"eqray-go-starter/app/modules/member/client/controller"
)

// Prefix 是 C 端接口前缀。
const Prefix = "/api/v1/client/member"

// Register 注册 C 端接口，由 member.Register() 统一调用。
func Register() {
	authController := controller.NewMemberAuthController()

	facades.Route().Prefix(Prefix).Group(func(r route.Router) {
		// ---------- 免登录 ----------
		// 登录、续期与授权地址不需要令牌
		r.Get("/auth/social/{type}/authorize-url", authController.AuthorizeURL)
		r.Post("/auth/social/login", authController.SocialLogin)
		r.Post("/auth/sms-code", authController.SmsCode)
		r.Post("/auth/sms-login", authController.SmsLogin)
		r.Post("/auth/social-bind-login", authController.SocialBindLogin)
		r.Post("/auth/refresh", authController.Refresh)

		// ---------- 需要会员令牌 ----------
		// 与后台令牌作用域隔离，互不通用
		memberAuth := r.Middleware(&baseMiddleware.MemberAuth{})
		memberAuth.Get("/auth/me", authController.Me)
		memberAuth.Post("/auth/logout", authController.Logout)
		memberAuth.Post("/auth/social/bind", authController.Bind)
		memberAuth.Get("/auth/social/bindings", authController.Bindings)
		memberAuth.Delete("/auth/social/bindings/{id}", authController.Unbind)

		// 后续 C 端业务接口（我的资料、我的地址等）统一加在 memberAuth 下，
		// controller 放在 client/controller，request 放在 client/request，
		// 用 support.MemberID(ctx.Context()) 取当前会员，禁止信任前端传入的会员 ID。
	})
}
