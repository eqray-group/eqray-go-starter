// Package client 注册 pay 模块的 C 端接口（小程序 / H5 / App）。
// 只认会员令牌，请求体不含金额与会员 ID。
package client

import (
	"github.com/goravel/framework/contracts/route"

	"eqray-go-starter/app/facades"
	baseMiddleware "eqray-go-starter/app/modules/base/middleware"
	payController "eqray-go-starter/app/modules/pay/client/controller"
)

// Prefix 是 C 端接口前缀。
const Prefix = "/api/v1/client/pay"

// Register 注册 C 端接口，由 pay.Register() 统一调用。
func Register() {
	payController := payController.NewPayController()

	facades.Route().Prefix(Prefix).Group(func(r route.Router) {
		// 可用渠道：登录后即可查看支付方式列表
		memberAuth := r.Middleware(&baseMiddleware.MemberAuth{})
		memberAuth.Get("/channels", payController.Channels)
		memberAuth.Post("/orders", payController.Store)
		memberAuth.Get("/orders/{merchantOrderNo}", payController.Show)
	})
}
