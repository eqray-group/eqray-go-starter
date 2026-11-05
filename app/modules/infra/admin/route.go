// Package admin 注册 infra 模块的后台管理接口。
package admin

import (
	"github.com/goravel/framework/contracts/route"

	"eqray-go-starter/app/facades"
	baseMiddleware "eqray-go-starter/app/modules/base/middleware"
	"eqray-go-starter/app/modules/infra/admin/controller"
)

// Prefix 是后台管理端接口前缀。
const Prefix = "/api/v1/infra"

// Register 注册后台管理接口，由 infra.Register() 统一调用（docs 4）。
func Register() {
	fileController := controller.NewFileController()
	fileConfigController := controller.NewFileConfigController()
	configController := controller.NewConfigController()
	regionController := controller.NewRegionController()
	apiLogController := controller.NewApiLogController()
	smsChannelController := controller.NewInfraSmsChannelController()
	smsTemplateController := controller.NewInfraSmsTemplateController()
	jobController := controller.NewJobController()
	monitorController := controller.NewMonitorController()
	noticeController := controller.NewNoticeController()
	generatorController := controller.NewGeneratorController()

	facades.Route().Prefix(Prefix).Group(func(r route.Router) {
		auth := r.Middleware(&baseMiddleware.TokenAuth{})

		// 权限中间件已下沉到 base，infra 按自身权限点校验，不再依赖 system
		perm := func(code string) route.Router {
			return auth.Middleware(&baseMiddleware.Permission{Code: code})
		}

		// operate 给写操作声明留痕语义，中间件注册在 Permission 之内（docs 9.5）
		operate := func(typ, sub string) *baseMiddleware.OperateLog {
			return &baseMiddleware.OperateLog{Type: typ, SubType: sub}
		}

		// 上传只要求登录态：头像等自助场景不应被具体权限点挡住，但上传动作需要留痕
		auth.Middleware(operate("文件管理", "上传文件")).Post("/files", fileController.Upload)

		// 文件列表与删除；C 端上传见 client/route.go
		perm("infra:file:list").Get("/files", fileController.Index)
		perm("infra:file:delete").Middleware(operate("文件管理", "删除文件")).
			Delete("/files/{id}", fileController.Destroy)

		perm("infra:file-config:list").Get("/file-configs", fileConfigController.Index)
		perm("infra:file-config:create").Middleware(operate("文件配置", "新增配置")).
			Post("/file-configs", fileConfigController.Store)
		perm("infra:file-config:update").Middleware(operate("文件配置", "编辑配置")).
			Put("/file-configs/{id}", fileConfigController.Update)
		perm("infra:file-config:update").Middleware(operate("文件配置", "设为主配置")).
			Put("/file-configs/{id}/master", fileConfigController.SetMaster)
		perm("infra:file-config:delete").Middleware(operate("文件配置", "删除配置")).
			Delete("/file-configs/{id}", fileConfigController.Destroy)

		perm("infra:config:list").Get("/configs", configController.Index)
		perm("infra:config:create").Middleware(operate("参数配置", "新增参数")).
			Post("/configs", configController.Store)
		perm("infra:config:update").Middleware(operate("参数配置", "编辑参数")).
			Put("/configs/{id}", configController.Update)
		perm("infra:config:delete").Middleware(operate("参数配置", "删除参数")).
			Delete("/configs/{id}", configController.Destroy)
		// 省市县：列表只读，同步按钮一次性从高德灌数据，之后 C 端级联只查库
		perm("infra:region:list").Get("/regions", regionController.Index)
		perm("infra:region:sync").Middleware(operate("基础设施", "同步行政区划")).
			Post("/regions/sync", regionController.Sync)

		// 系统监控：只读聚合接口，供后台看进程 / 依赖 / 请求指标
		perm("infra:monitor:list").Get("/monitor/overview", monitorController.Overview)

		// 代码生成：读表结构 → 预览 → 落盘（写文件单独授权）
		perm("infra:generator:list").Get("/generator/tables", generatorController.Tables)
		perm("infra:generator:list").Get("/generator/tables/{table}/columns", generatorController.Columns)
		perm("infra:generator:list").Post("/generator/preview", generatorController.Preview)
		perm("infra:generator:write").Middleware(operate("基础设施", "生成代码")).
			Post("/generator/generate", generatorController.Generate)

		// 消息通知：渠道与模板可配置，发送与测试单独授权
		perm("infra:notice:list").Get("/notice/types", noticeController.Types)
		perm("infra:notice:list").Get("/notice-channels", noticeController.Channels)
		perm("infra:notice:list").Get("/notice-templates", noticeController.Templates)
		perm("infra:notice:list").Get("/notice-logs", noticeController.Logs)
		perm("infra:notice:update").Middleware(operate("消息通知", "新增通知渠道")).
			Post("/notice-channels", noticeController.StoreChannel)
		perm("infra:notice:update").Middleware(operate("消息通知", "编辑通知渠道")).
			Put("/notice-channels/{id}", noticeController.UpdateChannel)
		perm("infra:notice:delete").Middleware(operate("消息通知", "删除通知渠道")).
			Delete("/notice-channels/{id}", noticeController.DestroyChannel)
		perm("infra:notice:update").Middleware(operate("消息通知", "新增通知模板")).
			Post("/notice-templates", noticeController.StoreTemplate)
		perm("infra:notice:update").Middleware(operate("消息通知", "编辑通知模板")).
			Put("/notice-templates/{id}", noticeController.UpdateTemplate)
		perm("infra:notice:delete").Middleware(operate("消息通知", "删除通知模板")).
			Delete("/notice-templates/{id}", noticeController.DestroyTemplate)
		perm("infra:notice:send").Middleware(operate("消息通知", "测试发送")).
			Post("/notice-channels/{id}/test", noticeController.TestChannel)
		perm("infra:notice:send").Middleware(operate("消息通知", "发送通知")).
			Post("/notices/send", noticeController.Send)

		// 定时任务：逻辑在代码里注册，这里只管启停、频率、手动执行与日志
		perm("infra:job:list").Get("/jobs", jobController.Index)
		perm("infra:job:list").Get("/jobs/handlers", jobController.Handlers)
		perm("infra:job:list").Get("/jobs/logs", jobController.Logs)
		perm("infra:job:update").Middleware(operate("基础设施", "新增定时任务")).
			Post("/jobs", jobController.Store)
		perm("infra:job:update").Middleware(operate("基础设施", "编辑定时任务")).
			Put("/jobs/{id}", jobController.Update)
		perm("infra:job:update").Middleware(operate("基础设施", "启停定时任务")).
			Put("/jobs/{id}/status", jobController.Toggle)
		perm("infra:job:run").Middleware(operate("基础设施", "手动执行任务")).
			Post("/jobs/{id}/run", jobController.Run)
		perm("infra:job:delete").Middleware(operate("基础设施", "删除定时任务")).
			Delete("/jobs/{id}", jobController.Destroy)

		perm("infra:api-log:list").Get("/api-logs/access", apiLogController.Access)
		perm("infra:api-log:list").Get("/api-logs/error", apiLogController.Error)
		perm("infra:api-log:update").Middleware(operate("API 日志", "标记处理")).
			Put("/api-logs/error/{id}/process", apiLogController.Process)

		// 短信渠道与模板：配置在 infra，member 等模块通过 InfraSmsService 复用发送能力
		perm("infra:sms-channel:list").Get("/sms-channels", smsChannelController.Index)
		perm("infra:sms-template:list").Get("/sms-channels/options", smsChannelController.Options)
		perm("infra:sms-channel:list").Get("/sms-channels/{id}", smsChannelController.Show)
		perm("infra:sms-channel:create").Middleware(operate("短信管理", "新增渠道")).
			Post("/sms-channels", smsChannelController.Store)
		perm("infra:sms-channel:update").Middleware(operate("短信管理", "编辑渠道")).
			Put("/sms-channels/{id}", smsChannelController.Update)
		perm("infra:sms-channel:delete").Middleware(operate("短信管理", "删除渠道")).
			Delete("/sms-channels/{id}", smsChannelController.Destroy)

		perm("infra:sms-template:list").Get("/sms-templates", smsTemplateController.Index)
		perm("infra:sms-template:list").Get("/sms-templates/{id}", smsTemplateController.Show)
		perm("infra:sms-template:create").Middleware(operate("短信管理", "新增模板")).
			Post("/sms-templates", smsTemplateController.Store)
		perm("infra:sms-template:update").Middleware(operate("短信管理", "编辑模板")).
			Put("/sms-templates/{id}", smsTemplateController.Update)
		perm("infra:sms-template:delete").Middleware(operate("短信管理", "删除模板")).
			Delete("/sms-templates/{id}", smsTemplateController.Destroy)
	})
}
