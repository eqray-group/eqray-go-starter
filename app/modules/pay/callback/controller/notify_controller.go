// Package controller 处理渠道免鉴权异步回调。
package controller

import (
	"github.com/goravel/framework/contracts/http"

	payService "eqray-go-starter/app/modules/pay/service"
)

// NotifyController 接收微信 / 支付宝的支付结果异步通知。
type NotifyController struct {
	notifyService *payService.PayNotifyService
}

// NewNotifyController 构造回调控制器。
func NewNotifyController() *NotifyController {
	return &NotifyController{notifyService: payService.NewPayNotifyService()}
}

// Pay 处理支付结果通知：验签、金额校验、幂等落单、派发事件、按渠道格式应答。
// 入参：路径 code 为渠道编码；返回渠道要求的成功 / 失败应答（JSON 或纯文本）。
func (c *NotifyController) Pay(ctx http.Context) http.Response {
	ack := c.notifyService.HandlePayNotify(ctx.Context(), ctx.Request().Route("code"), ctx.Request().Origin())

	resp := ctx.Response().String(200, ack.Body)
	if ack.ContentType != "" {
		resp = ctx.Response().Header("Content-Type", ack.ContentType).String(200, ack.Body)
	}

	return resp
}
