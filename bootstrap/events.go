package bootstrap

import (
	"github.com/goravel/framework/contracts/event"

	payevents "eqray-go-starter/app/events/pay"
	demoListener "eqray-go-starter/app/modules/demo/listener"
)

// Events 注册事件与监听器的映射，随 WithEvents 装载。
// 事件契约在 app/events 下按域划分，任何业务模块订阅只需在此追加监听器，
// 发布方（如 pay）不反向依赖业务模块。
func Events() map[event.Event][]event.Listener {
	return map[event.Event][]event.Listener{
		// 事件键使用零值（非指针）：框架按事件实例查映射，指针每次都是新地址会查不到
		payevents.OrderPaidEvent{}: {
			demoListener.NewOrderPaidListener(),
		},
	}
}
