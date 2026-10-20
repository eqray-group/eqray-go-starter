package member

import (
	"github.com/goravel/framework/contracts/route"

	"eqray-go-starter/app/facades"
	baseMiddleware "eqray-go-starter/app/modules/base/middleware"
	"eqray-go-starter/app/modules/member/controller"
	memberMiddleware "eqray-go-starter/app/modules/member/middleware"
)

// Register 注册 member 模块的路由，由 routes.Modules() 统一调用。
// 路由分两组：/auth/* 是 C 端会员登录态，其余是后台管理接口。
func Register() {
	memberController := controller.NewMemberController()
	authController := controller.NewMemberAuthController()
	levelController := controller.NewMemberLevelController()
	tagController := controller.NewMemberTagController()
	groupController := controller.NewMemberGroupController()
	socialClientController := controller.NewMemberSocialClientController()

	facades.Route().Prefix("/api/v1/member").Group(func(r route.Router) {
		// ---------- C 端认证 ----------
		// 登录、续期与授权地址不需要令牌
		r.Get("/auth/social/{type}/authorize-url", authController.AuthorizeURL)
		r.Post("/auth/social/login", authController.SocialLogin)
		r.Post("/auth/sms-code", authController.SmsCode)
		r.Post("/auth/sms-login", authController.SmsLogin)
		r.Post("/auth/social-bind-login", authController.SocialBindLogin)
		r.Post("/auth/refresh", authController.Refresh)

		// 需要会员令牌：与后台令牌作用域隔离，互不通用
		memberAuth := r.Middleware(&memberMiddleware.MemberAuth{})
		memberAuth.Get("/auth/me", authController.Me)
		memberAuth.Post("/auth/logout", authController.Logout)
		memberAuth.Post("/auth/social/bind", authController.Bind)
		memberAuth.Get("/auth/social/bindings", authController.Bindings)
		memberAuth.Delete("/auth/social/bindings/{id}", authController.Unbind)

		// ---------- 管理端 ----------
		auth := r.Middleware(&baseMiddleware.TokenAuth{})

		// 权限中间件在 base，member 只声明自己的权限点
		perm := func(code string) route.Router {
			return auth.Middleware(&baseMiddleware.Permission{Code: code})
		}

		// 写操作留痕，中间件注册在 Permission 之内
		operate := func(typ, sub string) *baseMiddleware.OperateLog {
			return &baseMiddleware.OperateLog{Type: typ, SubType: sub}
		}

		// ---------- 会员 ----------
		perm("member:user:list").Get("/users", memberController.Index)
		perm("member:user:list").Get("/users/{id}", memberController.Show)
		perm("member:user:create").Middleware(operate("会员管理", "新增会员")).
			Post("/users", memberController.Store)
		perm("member:user:update").Middleware(operate("会员管理", "编辑会员")).
			Put("/users/{id}", memberController.Update)
		perm("member:user:update").Middleware(operate("会员管理", "变更会员状态")).
			Put("/users/{id}/status", memberController.UpdateStatus)
		perm("member:user:delete").Middleware(operate("会员管理", "删除会员")).
			Delete("/users/{id}", memberController.Destroy)

		// 积分：列表随会员权限，调整单独授权
		perm("member:user:list").Get("/users/{id}/points", memberController.PointRecords)
		perm("member:user:points").Middleware(operate("会员管理", "调整积分")).
			Post("/users/{id}/point", memberController.AdjustPoint)

		// 等级调整与等级记录
		perm("member:user:list").Get("/users/{id}/level-records", memberController.LevelRecords)
		perm("member:user:level").Middleware(operate("会员管理", "调整等级")).
			Put("/users/{id}/level", memberController.AdjustLevel)

		// 签到记录、收货地址与第三方绑定：后台只读
		perm("member:user:list").Get("/users/{id}/sign-records", memberController.SignRecords)
		perm("member:user:list").Get("/users/{id}/addresses", memberController.Addresses)
		perm("member:user:list").Get("/users/{id}/social-bindings", memberController.SocialBindings)

		// ---------- 等级 ----------
		// 下拉供会员表单使用，因此只要求会员列表权限
		perm("member:user:list").Get("/levels/options", levelController.Options)
		perm("member:level:list").Get("/levels", levelController.Index)
		perm("member:level:list").Get("/levels/{id}", levelController.Show)
		perm("member:level:create").Middleware(operate("会员管理", "新增等级")).
			Post("/levels", levelController.Store)
		perm("member:level:update").Middleware(operate("会员管理", "编辑等级")).
			Put("/levels/{id}", levelController.Update)
		perm("member:level:delete").Middleware(operate("会员管理", "删除等级")).
			Delete("/levels/{id}", levelController.Destroy)

		// ---------- 标签 ----------
		perm("member:user:list").Get("/tags/options", tagController.Options)
		perm("member:tag:list").Get("/tags", tagController.Index)
		perm("member:tag:create").Middleware(operate("会员管理", "新增标签")).
			Post("/tags", tagController.Store)
		perm("member:tag:update").Middleware(operate("会员管理", "编辑标签")).
			Put("/tags/{id}", tagController.Update)
		perm("member:tag:delete").Middleware(operate("会员管理", "删除标签")).
			Delete("/tags/{id}", tagController.Destroy)

		// ---------- 分组 ----------
		perm("member:user:list").Get("/groups/options", groupController.Options)
		perm("member:group:list").Get("/groups", groupController.Index)
		perm("member:group:create").Middleware(operate("会员管理", "新增分组")).
			Post("/groups", groupController.Store)
		perm("member:group:update").Middleware(operate("会员管理", "编辑分组")).
			Put("/groups/{id}", groupController.Update)
		perm("member:group:delete").Middleware(operate("会员管理", "删除分组")).
			Delete("/groups/{id}", groupController.Destroy)

		// ---------- 第三方登录渠道配置 ----------
		perm("member:social-client:list").Get("/social-clients", socialClientController.Index)
		perm("member:social-client:list").Get("/social-clients/{id}", socialClientController.Show)
		perm("member:social-client:create").Middleware(operate("会员管理", "新增社交客户端")).
			Post("/social-clients", socialClientController.Store)
		perm("member:social-client:update").Middleware(operate("会员管理", "编辑社交客户端")).
			Put("/social-clients/{id}", socialClientController.Update)
		perm("member:social-client:delete").Middleware(operate("会员管理", "删除社交客户端")).
			Delete("/social-clients/{id}", socialClientController.Destroy)
	})
}
