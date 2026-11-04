// Package demo 是演示业务模块：验证「下单 → 支付 → 事件回写」完整链路，联调后可整体删除。
// demo 只订阅 pay 的 events、调用 pay 的 service，pay 模块对 demo 一无所知。
package demo

import (
	"github.com/goravel/framework/contracts/route"

	"eqray-go-starter/app/facades"
	baseMiddleware "eqray-go-starter/app/modules/base/middleware"
	demoController "eqray-go-starter/app/modules/demo/client/controller"
)

// Prefix 是演示业务 C 端接口前缀。
const Prefix = "/api/v1/client/demo"

// Register 注册演示业务的路由（只挂 C 端，不做后台页面）。
func Register() {
	orderController := demoController.NewDemoOrderController()

	facades.Route().Prefix(Prefix).Group(func(r route.Router) {
		memberAuth := r.Middleware(&baseMiddleware.MemberAuth{})
		memberAuth.Post("/orders", orderController.Store)
		memberAuth.Get("/orders", orderController.Index)
		memberAuth.Get("/orders/{orderNo}", orderController.Show)
	})
}
