package routes

import (
	"eqray-go-starter/app/modules/infra"
	"eqray-go-starter/app/modules/member"
	"eqray-go-starter/app/modules/system"
)

// Modules 聚合各业务模块的路由，在 bootstrap/app.go 的 WithRouting 中调用。
func Modules() {
	system.Register()
	infra.Register()
	member.Register()
}
