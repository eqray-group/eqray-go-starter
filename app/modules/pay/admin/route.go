// Package admin 注册 pay 模块的后台管理接口（令牌 + 权限点 + 操作留痕）。
package admin

import (
	"github.com/goravel/framework/contracts/route"

	"eqray-go-starter/app/facades"
	baseMiddleware "eqray-go-starter/app/modules/base/middleware"
	"eqray-go-starter/app/modules/pay/admin/controller"
)

// Prefix 是后台管理端接口前缀。
const Prefix = "/api/v1/pay"

// Register 注册后台管理接口，由 pay.Register() 统一调用。
func Register() {
	channelController := controller.NewPayChannelController()
	orderController := controller.NewPayOrderController()
	refundController := controller.NewPayRefundController()
	notifyLogController := controller.NewPayNotifyLogController()

	facades.Route().Prefix(Prefix).Group(func(r route.Router) {
		auth := r.Middleware(&baseMiddleware.TokenAuth{})

		perm := func(code string) route.Router {
			return auth.Middleware(&baseMiddleware.Permission{Code: code})
		}

		operate := func(typ, sub string) *baseMiddleware.OperateLog {
			return &baseMiddleware.OperateLog{Type: typ, SubType: sub}
		}

		// 渠道配置：静态路径（schema / notify-urls）必须注册在 /{id} 之前，避免被路径参数吞掉
		perm("pay:channel:list").Get("/channels", channelController.Index)
		perm("pay:channel:list").Get("/channels/schema", channelController.Schema)
		perm("pay:channel:list").Get("/channels/notify-urls", channelController.NotifyURLs)
		perm("pay:channel:list").Get("/channels/{id}", channelController.Show)
		perm("pay:channel:create").Middleware(operate("支付管理", "新增渠道")).
			Post("/channels", channelController.Store)
		perm("pay:channel:update").Middleware(operate("支付管理", "编辑渠道")).
			Put("/channels/{id}", channelController.Update)
		perm("pay:channel:update").Middleware(operate("支付管理", "启停渠道")).
			Put("/channels/{id}/status", channelController.SetStatus)
		perm("pay:channel:update").Middleware(operate("支付管理", "渠道测通")).
			Post("/channels/{id}/test", channelController.Test)
		perm("pay:channel:delete").Middleware(operate("支付管理", "删除渠道")).
			Delete("/channels/{id}", channelController.Destroy)

		// 支付订单
		perm("pay:order:list").Get("/orders", orderController.Index)
		perm("pay:order:list").Get("/orders/{id}", orderController.Show)
		perm("pay:order:update").Middleware(operate("支付管理", "同步订单")).
			Post("/orders/{id}/sync", orderController.Sync)
		perm("pay:order:update").Middleware(operate("支付管理", "关闭订单")).
			Post("/orders/{id}/close", orderController.Close)

		// 退款
		perm("pay:refund:list").Get("/refunds", refundController.Index)
		perm("pay:refund:list").Get("/refunds/{id}", refundController.Show)
		perm("pay:refund:create").Middleware(operate("支付管理", "发起退款")).
			Post("/refunds", refundController.Store)

		// 回调日志
		perm("pay:notify-log:list").Get("/notify-logs", notifyLogController.Index)
	})
}
