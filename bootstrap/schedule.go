package bootstrap

import (
	"context"
	"time"

	"github.com/goravel/framework/contracts/schedule"

	"eqray-go-starter/app/facades"
	infraService "eqray-go-starter/app/modules/infra/service"
)

// Schedule 注册应用级定时任务，由 foundation.Setup().WithSchedule 装载。
// 这里只挂一个「每分钟分发器」：具体任务与其 cron 存在 infra_job 表，
// 运营可在「基础设施 → 定时任务」里启停与改频率，无需改代码重启。
func Schedule() []schedule.Event {
	return []schedule.Event{
		// OnOneServer 保证多实例部署时同一分钟只有一台机器执行
		facades.Schedule().Call(dispatchDueJobs).
			EveryMinute().
			Name("infra:job-dispatcher").
			OnOneServer(),
	}
}

// dispatchDueJobs 每分钟执行一次到点的任务，并顺带清理过期执行日志。
// 无入参与返回值：内部自行兜底超时与异常，不把错误抛给调度器。
func dispatchDueJobs() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	infraService.NewJobService().DispatchDue(ctx)
}
