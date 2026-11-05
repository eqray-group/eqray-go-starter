package bootstrap

import (
	"context"
	"fmt"

	infraJob "eqray-go-starter/app/modules/infra/job"
	infraService "eqray-go-starter/app/modules/infra/service"
	payService "eqray-go-starter/app/modules/pay/service"
)

// registerJobs 注册全部可调度的任务处理器。
// 库里只存处理器标识，真实逻辑始终在代码里，改逻辑走发版，改频率 / 启停在后台即可。
// 入参无；无返回值。
func registerJobs() {
	infraJob.Register(infraJob.Definition{
		Key:         "infra:purge-api-access-log",
		Name:        "清理 API 访问日志",
		Description: "删除超过保留天数的 infra_api_access_log 记录（保留天数见 config/infra.go）",
		DefaultCron: "10 3 * * *",
		Handler: func(ctx context.Context) (string, error) {
			deleted, err := infraService.NewApiLogService().PurgeExpiredAccess(ctx)
			if err != nil {
				return "", err
			}

			return fmt.Sprintf("删除 %d 条访问日志", deleted), nil
		},
	})

	infraJob.Register(infraJob.Definition{
		Key:         "pay:close-expired-orders",
		Name:        "关闭超时支付订单",
		Description: "扫描超过有效期的未支付订单，先关渠道侧再关本地",
		DefaultCron: "*/5 * * * *",
		Handler: func(ctx context.Context) (string, error) {
			closed, err := payService.NewPayOrderService().CloseExpired(ctx)
			if err != nil {
				return "", err
			}

			return fmt.Sprintf("关闭 %d 笔超时订单", closed), nil
		},
	})

	infraJob.Register(infraJob.Definition{
		Key:         "pay:sync-refunding",
		Name:        "同步退款中订单",
		Description: "查单收敛处于「退款中」的退款单（微信退款异步，支付宝同步无此状态）",
		DefaultCron: "*/5 * * * *",
		Handler: func(ctx context.Context) (string, error) {
			settled, err := payService.NewPayRefundService().SyncRefunding(ctx)
			if err != nil {
				return "", err
			}

			return fmt.Sprintf("收敛 %d 笔退款单", settled), nil
		},
	})

	infraJob.Register(infraJob.Definition{
		Key:         "pay:purge-notify-logs",
		Name:        "清理支付回调日志",
		Description: "物理删除超过保留天数的 pay_notify_log 记录",
		DefaultCron: "40 3 * * *",
		Handler: func(ctx context.Context) (string, error) {
			deleted, err := payService.NewPayNotifyService().PurgeNotifyLogs(ctx)
			if err != nil {
				return "", err
			}

			return fmt.Sprintf("删除 %d 条回调日志", deleted), nil
		},
	})
}
