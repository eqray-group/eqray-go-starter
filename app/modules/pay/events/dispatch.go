package events

import (
	"encoding/json"

	"github.com/goravel/framework/contracts/event"

	"eqray-go-starter/app/facades"
)

// dispatch 统一派发事件：载荷序列化为 JSON 字符串经 Arg 传递，
// 监听器侧反序列化还原，避免依赖框架队列的 struct 序列化差异。
// 入参 key 为事件键实例、payload 为业务载荷；无监听器或派发失败只记日志。
func dispatch(key event.Event, payload any) {
	data, err := json.Marshal(payload)
	if err != nil {
		facades.Log().Errorf("pay 事件载荷序列化失败：%v", err)
		return
	}

	err = facades.Event().Job(key, []event.Arg{
		{Type: "string", Value: string(data)},
	}).Dispatch()
	if err != nil {
		// 没有任何业务模块订阅时框架返回「未绑定」错误，属正常情况
		facades.Log().Warningf("pay 事件派发未完成：%v", err)
	}
}
