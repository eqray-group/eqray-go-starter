package routes

import (
	"eqray-go-starter/app/modules/demo"
	"eqray-go-starter/app/modules/infra"
	"eqray-go-starter/app/modules/member"
	"eqray-go-starter/app/modules/pay"
	"eqray-go-starter/app/modules/system"
)

// Modules 聚合各业务模块的路由，在 bootstrap/app.go 的 WithRouting 中调用。
func Modules() {
	system.Register()
	infra.Register()
	member.Register()
	pay.Register()
	// demo 仅用于支付链路联调，上线可整体删除
	demo.Register()
}
