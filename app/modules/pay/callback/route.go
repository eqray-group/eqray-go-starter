// Package callback 注册渠道异步回调路由：免鉴权、不参与 RBAC、不写操作日志。
// 只做验签 + 入日志 + 幂等处理，应答格式由各渠道适配器决定。
package callback

import (
	"github.com/goravel/framework/contracts/route"

	"eqray-go-starter/app/facades"
	payController "eqray-go-starter/app/modules/pay/callback/controller"
	payService "eqray-go-starter/app/modules/pay/service"
)

// Prefix 是异步回调路由前缀，完整地址 = APP_URL + Prefix + 渠道编码。
const Prefix = "/api/v1/pay/notify"

// Register 注册异步回调路由，由 pay.Register() 统一调用。
func Register() {
	// 与 payService.CallbackPath 保持一致，改路由必须同步修改
	if payService.CallbackPath != Prefix+"/" {
		facades.Log().Errorf("回调路由前缀不一致：route=%s service=%s", Prefix+"/", payService.CallbackPath)
	}

	notifyController := payController.NewNotifyController()

	facades.Route().Prefix(Prefix).Group(func(r route.Router) {
		// 不挂 TokenAuth / MemberAuth：渠道服务器不会携带本系统令牌
		r.Post("/{code}", notifyController.Pay)
	})
}
