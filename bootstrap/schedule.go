package bootstrap

import (
	"context"
	"time"

	"github.com/goravel/framework/contracts/schedule"

	"eqray-go-starter/app/facades"
	infraService "eqray-go-starter/app/modules/infra/service"
	payService "eqray-go-starter/app/modules/pay/service"
)

// Schedule 注册应用级定时任务，由 foundation.Setup().WithSchedule 装载，
// 随 app.Start() 的 schedule runner 一起启动。
func Schedule() []schedule.Event {
	return []schedule.Event{
		// 访问日志保留策略：每天 03:10 清理 infra_api_access_log 中超过保留天数的记录，
		// 保留天数见 config/infra.go（默认 7 天，可用 INFRA_API_LOG_RETENTION_DAYS 覆盖）。
		// OnOneServer 保证多实例部署时同一时刻只有一台机器执行。
		facades.Schedule().Call(purgeExpiredApiAccessLogs).
			DailyAt("03:10").
			Name("infra:purge-api-access-log").
			OnOneServer(),

		// 支付超时关单：每 5 分钟扫描超时未支付订单，先关渠道侧再关本地
		facades.Schedule().Call(closeExpiredPayOrders).
			EveryFiveMinutes().
			Name("pay:close-expired-orders").
			OnOneServer(),

		// 退款状态同步：每 5 分钟查单收敛「退款中」的微信退款单
		facades.Schedule().Call(syncRefundingPayOrders).
			EveryFiveMinutes().
			Name("pay:sync-refunding").
			OnOneServer(),

		// 支付回调日志清理：每天 03:40 删除超过保留天数的 pay_notify_log
		facades.Schedule().Call(purgePayNotifyLogs).
			DailyAt("03:40").
			Name("pay:purge-notify-logs").
			OnOneServer(),
	}
}

// purgeExpiredApiAccessLogs 执行一次访问日志清理，并把结果写入应用日志。
// 无入参与返回值：异常只记录日志，避免把错误抛给调度器打断后续任务。
func purgeExpiredApiAccessLogs() {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	deleted, err := infraService.NewApiLogService().PurgeExpiredAccess(ctx)
	if err != nil {
		facades.Log().Errorf("清理 API 访问日志失败：%v", err)
		return
	}

	if deleted > 0 {
		facades.Log().Infof("清理 API 访问日志完成，共删除 %d 条", deleted)
	}
}

// closeExpiredPayOrders 关闭超时未支付订单：先关渠道侧订单再关本地，渠道失败不阻断。
// 无入参与返回值：异常只记录日志，避免打断后续任务。
func closeExpiredPayOrders() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	closed, err := payService.NewPayOrderService().CloseExpired(ctx)
	if err != nil {
		facades.Log().Errorf("关闭超时支付订单失败：%v", err)
		return
	}

	if closed > 0 {
		facades.Log().Infof("关闭超时支付订单完成，共关闭 %d 单", closed)
	}
}

// syncRefundingPayOrders 查单收敛「退款中」的退款单（微信退款异步，支付宝同步无此状态）。
// 无入参与返回值：异常只记录日志。
func syncRefundingPayOrders() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	settled, err := payService.NewPayRefundService().SyncRefunding(ctx)
	if err != nil {
		facades.Log().Errorf("同步退款状态失败：%v", err)
		return
	}

	if settled > 0 {
		facades.Log().Infof("退款状态同步完成，共收敛 %d 单", settled)
	}
}

// purgePayNotifyLogs 物理清理过期支付回调日志（保留天数见 pay service 常量）。
// 无入参与返回值：异常只记录日志。
func purgePayNotifyLogs() {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	deleted, err := payService.NewPayNotifyService().PurgeNotifyLogs(ctx)
	if err != nil {
		facades.Log().Errorf("清理支付回调日志失败：%v", err)
		return
	}

	if deleted > 0 {
		facades.Log().Infof("清理支付回调日志完成，共删除 %d 条", deleted)
	}
}
