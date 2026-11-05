package bootstrap

import (
	"context"

	contractsFoundation "github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/contracts/foundation/configuration"
	"github.com/goravel/framework/foundation"

	baseMiddleware "eqray-go-starter/app/modules/base/middleware"
	"eqray-go-starter/app/modules/base/types"
	infraMiddleware "eqray-go-starter/app/modules/infra/middleware"
	memberService "eqray-go-starter/app/modules/member/service"
	payChannel "eqray-go-starter/app/modules/pay/channel"
	payService "eqray-go-starter/app/modules/pay/service"
	systemService "eqray-go-starter/app/modules/system/service"
	"eqray-go-starter/config"
	"eqray-go-starter/routes"
)

func Boot() contractsFoundation.Application {
	// 定时任务处理器必须先注册，调度器每分钟按 infra_job 表里的 handler 查找
	registerJobs()

	return foundation.Setup().
		WithMigrations(Migrations).
		WithRouting(func() {
			// 权限数据归 system，中间件在 base；启动时注入实现，保持模块单向依赖
			baseMiddleware.RegisterPermissionChecker(
				func(ctx context.Context, userID types.ID, code string) (bool, error) {
					return systemService.NewPermissionService().Allows(ctx, userID, code)
				},
			)

			// 操作日志中间件在 base，写库实现归 system，启动时注入
			baseMiddleware.RegisterOperateLogRecorder(
				func(ctx context.Context, in baseMiddleware.OperateLogInput) error {
					return systemService.NewOperateLogService().Record(ctx, systemService.OperateLogInput{
						UserID:        in.UserID,
						UserType:      in.UserType,
						Type:          in.Type,
						SubType:       in.SubType,
						Action:        in.Action,
						Success:       in.Success,
						Extra:         in.Extra,
						RequestMethod: in.RequestMethod,
						RequestURL:    in.RequestURL,
						RequestIP:     in.RequestIP,
						UserAgent:     in.UserAgent,
						Duration:      in.Duration,
					})
				},
			)

			// 支付模块的 openid 反查能力归 member，启动时注入，保持 pay 不依赖 member。
			// 小程序与公众号 openid 不通用，两个社交类型各注册一次。
			socialService := memberService.NewMemberSocialService()
			for _, socialType := range []string{
				payChannel.SocialTypeWechatMini,
				payChannel.SocialTypeWechatMp,
			} {
				payService.RegisterOpenIDResolver(socialType,
					func(ctx context.Context, userID types.ID) (string, error) {
						return socialService.FindOpenIDByMember(ctx, userID, socialType)
					})
			}

			routes.Web()
			routes.Grpc()
			routes.Modules()
		}).
		WithProviders(Providers).
		WithCommands(Commands).
		WithSchedule(Schedule).
		WithEvents(Events).
		WithConfig(config.Boot).
		// 全局中间件：接口限流在前，日志在后，被限流的请求也能留下访问记录
		WithMiddleware(func(handler configuration.Middleware) {
			handler.Append(
				&baseMiddleware.RateLimit{Name: "api"},
				&infraMiddleware.RequestLog{},
			)
		}).
		Create()
}
