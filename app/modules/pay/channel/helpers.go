package channel

import (
	"encoding/json"
)

// parseJSON 把 JSON 字符串解析为 map，非法输入返回空 map（不报错，交由校验流程处理）。
// 入参 raw 为 JSON 字符串；返回键值 map。
func parseJSON(raw string) map[string]any {
	plain := make(map[string]any)
	if raw == "" {
		return plain
	}
	if err := json.Unmarshal([]byte(raw), &plain); err != nil {
		return make(map[string]any)
	}

	return plain
}
