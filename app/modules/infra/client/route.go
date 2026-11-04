// Package client 注册 infra 模块的 C 端接口。
// 这里只放会员本人会直接调用的通用能力（如文件上传），不含任何管理配置类接口。
package client

import (
	"github.com/goravel/framework/contracts/route"

	"eqray-go-starter/app/facades"
	baseMiddleware "eqray-go-starter/app/modules/base/middleware"
	"eqray-go-starter/app/modules/infra/client/controller"
)

// Prefix 是 C 端接口前缀。
const Prefix = "/api/v1/client/infra"

// Register 注册 C 端接口，由 infra.Register() 统一调用。
func Register() {
	fileController := controller.NewFileController()
	publicConfigController := controller.NewPublicConfigController()
	regionController := controller.NewRegionController()

	facades.Route().Prefix(Prefix).Group(func(r route.Router) {
		// 公开配置免登录：页面渲染前就要取到地图 Key，且不含任何敏感信息（服务端白名单过滤）
		r.Get("/public-configs", publicConfigController.Show)

		// 省市区数据只读、已落库：免登录查询，前端可长期缓存
		r.Get("/regions", regionController.Index)

		// 上传需要会员登录态；不挂操作日志，C 端行为不进后台操作审计
		memberAuth := r.Middleware(&baseMiddleware.MemberAuth{})
		memberAuth.Post("/files", fileController.Upload)
	})
}
