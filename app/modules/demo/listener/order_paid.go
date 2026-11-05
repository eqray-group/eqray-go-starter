// Package listener 订阅支付域事件（契约见 app/events/pay）。
// 监听器必须幂等、不得 panic，重活异步化由业务侧自行安排。
package listener

import (
	"context"
	"encoding/json"
	"time"

	"github.com/goravel/framework/contracts/event"

	"eqray-go-starter/app/events/pay"
	"eqray-go-starter/app/facades"
	demoService "eqray-go-starter/app/modules/demo/service"
)

// OrderPaidListener 订阅 pay.order.paid 事件，回写演示订单状态。
type OrderPaidListener struct {
	service *demoService.DemoOrderService
}

// NewOrderPaidListener 构造支付成功监听器。
func NewOrderPaidListener() *OrderPaidListener {
	return &OrderPaidListener{service: demoService.NewDemoOrderService()}
}

// Signature 返回监听器唯一标识。
func (l *OrderPaidListener) Signature() string {
	return "demo:order-paid"
}

// Queue 声明同步执行：回写状态很轻，异步反而引入延迟与失败排查成本。
// 入参 args 为监听器收到的载荷；返回空队列（Enable=false 即同步）。
func (l *OrderPaidListener) Queue(args ...any) event.Queue {
	return event.Queue{}
}

// Handle 处理支付成功事件：解析载荷，只处理 demo_order 业务类型，回写演示订单。
// 入参 args 为事件载荷（JSON 字符串）；返回错误会记入框架日志。
func (l *OrderPaidListener) Handle(args ...any) error {
	if len(args) == 0 {
		return nil
	}

	raw, ok := args[0].(string)
	if !ok {
		return nil
	}

	var payload payevents.OrderPaidPayload
	if err := json.Unmarshal([]byte(raw), &payload); err != nil {
		facades.Log().Errorf("演示订单监听器解析事件失败：%v", err)
		return nil
	}
	if payload.BizType != demoService.DemoBizType {
		return nil
	}

	// 用独立超时的后台上下文：事件派发时的请求上下文可能已被取消
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	l.service.OnOrderPaid(ctx, payload.BizOrderNo, payload.OrderNo, payload.PaidAt)

	return nil
}
