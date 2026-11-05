package bootstrap

import (
	"github.com/goravel/framework/contracts/console"

	"eqray-go-starter/app/console/commands"
)

// Commands 注册自定义 artisan 命令。
// 入参无；返回命令列表。
func Commands() []console.Command {
	return []console.Command{
		&commands.PermCheck{},
	}
}
