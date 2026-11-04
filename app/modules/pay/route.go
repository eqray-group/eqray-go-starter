// Package pay 是支付模块：收单、退款、回调与事件派发。
// 模块边界：pay 只做支付，不 import 任何业务模块；业务侧通过 events 接入支付结果。
package pay

import (
	"context"
	"time"

	payAdmin "eqray-go-starter/app/modules/pay/admin"
	payCallback "eqray-go-starter/app/modules/pay/callback"
	payClient "eqray-go-starter/app/modules/pay/client"
	payService "eqray-go-starter/app/modules/pay/service"
)

// Register 注册 pay 模块全部路由：管理端、C 端与免鉴权回调，并触发一次启动自检。
func Register() {
	payAdmin.Register()
	payClient.Register()
	payCallback.Register()

	go checkChannelsOnBoot()
}

// checkChannelsOnBoot 启动后异步对启用渠道做一次凭据自检，失败只告警不阻断启动。
// 延迟 3 秒执行，避免与迁移、数据库连接初始化抢时序；表不存在等错误静默跳过。
func checkChannelsOnBoot() {
	time.Sleep(3 * time.Second)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	payService.NewPayChannelService().CheckChannels(ctx)
}
