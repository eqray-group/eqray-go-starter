package system

import (
	"github.com/goravel/framework/contracts/route"

	"eqray-go-starter/app/facades"
	baseMiddleware "eqray-go-starter/app/modules/base/middleware"
	"eqray-go-starter/app/modules/system/controller"
	systemMiddleware "eqray-go-starter/app/modules/system/middleware"
)

// Register 注册 system 模块的路由，由 routes.Modules() 统一调用（docs 4）。
func Register() {
	authController := controller.NewAuthController()
	userController := controller.NewUserController()
	menuController := controller.NewMenuController()
	roleController := controller.NewRoleController()

	facades.Route().Prefix("/api/v1/system").Group(func(r route.Router) {
		r.Get("/auth/captcha", authController.Captcha)
		r.Post("/auth/login", authController.Login)
		// 刷新与登出不挂鉴权：访问令牌过期后仍要能续期，登出也允许用已失效的令牌调用
		r.Post("/auth/refresh", authController.Refresh)
		r.Post("/auth/logout", authController.Logout)

		auth := r.Middleware(&baseMiddleware.TokenAuth{})

		// 按权限标识分组注册，权限标识与 system_menu.permission 一一对应（docs 7.2、9.3）
		perm := func(code string) route.Router {
			return auth.Middleware(&systemMiddleware.Permission{Code: code})
		}

		auth.Get("/profile", authController.Profile)
		// 动态路由按当前登录者计算，不挂具体权限点，避免没有菜单权限时连导航都拿不到
		auth.Get("/menus/routes", menuController.Routes)

		// 用户
		userList := perm("system:user:list")
		userList.Get("/users", userController.Index)
		userList.Get("/users/{id}", userController.Show)

		perm("system:user:create").Post("/users", userController.Store)

		userUpdate := perm("system:user:update")
		userUpdate.Put("/users/{id}", userController.Update)
		userUpdate.Put("/users/{id}/password", userController.ResetPassword)

		perm("system:user:delete").Delete("/users/{id}", userController.Destroy)

		userRole := perm("system:user:assign-role")
		userRole.Get("/users/{id}/roles", userController.Roles)
		userRole.Put("/users/{id}/roles", userController.AssignRoles)

		// 菜单：/menus/simple 注册在 /menus/{id} 之前
		menuRead := perm("system:menu:list")
		menuRead.Get("/menus", menuController.Index)
		menuRead.Get("/menus/simple", menuController.Simple)
		menuRead.Get("/menus/{id}", menuController.Show)

		perm("system:menu:create").Post("/menus", menuController.Store)
		perm("system:menu:update").Put("/menus/{id}", menuController.Update)
		perm("system:menu:delete").Delete("/menus/{id}", menuController.Destroy)

		// 角色：/roles/options 注册在 /roles/{id} 之前
		roleRead := perm("system:role:list")
		roleRead.Get("/roles", roleController.Index)
		roleRead.Get("/roles/options", roleController.Options)
		roleRead.Get("/roles/{id}", roleController.Show)

		perm("system:role:create").Post("/roles", roleController.Store)
		perm("system:role:update").Put("/roles/{id}", roleController.Update)
		perm("system:role:delete").Delete("/roles/{id}", roleController.Destroy)

		roleMenu := perm("system:role:assign-menu")
		roleMenu.Get("/roles/{id}/menus", roleController.Menus)
		roleMenu.Put("/roles/{id}/menus", roleController.AssignMenus)
	})
}
