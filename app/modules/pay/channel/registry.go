package channel

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"eqray-go-starter/app/modules/base/errs"
	payErrs "eqray-go-starter/app/modules/pay/errs"
)

// clientEntry 是进程内缓存的渠道客户端条目，key 含 config 哈希，配置变更自动换新 key。
type clientEntry struct {
	channel Channel
}

// clientCache 进程内 client 缓存：解析 RSA 私钥成本不低，避免每次下单都重建。
// key = 渠道编码 + config 哈希，渠道配置修改后自然使用新条目，无需手动失效。
var clientCache sync.Map

// Resolve 由渠道编码与 config 明文获取渠道客户端（带进程内缓存）。
// 入参 code 为渠道编码、configJSON 为 pay_channel.config 明文；
// 编码不受支持或配置不合法返回业务错误码。
func Resolve(code, configJSON string) (Channel, error) {
	meta, ok := Metas[code]
	if !ok {
		return nil, errs.New(payErrs.PayChannelUnsupported, code)
	}

	sum := sha256.Sum256([]byte(configJSON))
	key := code + "|" + hex.EncodeToString(sum[:])
	if cached, ok := clientCache.Load(key); ok {
		return cached.(clientEntry).channel, nil
	}

	ch, err := meta.Build(code, configJSON)
	if err != nil {
		return nil, err
	}
	clientCache.Store(key, clientEntry{channel: ch})

	return ch, nil
}

// MaskConfig 按 schema 对 config JSON 做响应层脱敏：secret 字段只回「前 4 + **** + 后 4」。
// 入参 code 为渠道编码、configJSON 为库中明文；返回脱敏后的 JSON 字符串（解析失败原样返回空对象）。
func MaskConfig(code, configJSON string) string {
	meta, ok := Metas[code]
	if !ok {
		return "{}"
	}

	masked := make(map[string]any)
	plain := parseJSON(configJSON)
	for _, field := range meta.Fields {
		value, _ := plain[field.Key]
		if field.Secret {
			value = maskSecret(fmt.Sprintf("%v", value))
		}
		masked[field.Key] = value
	}

	data, err := json.Marshal(masked)
	if err != nil {
		return "{}"
	}
	return string(data)
}

// MergeSecretFields 把表单提交的 config 与库中原值合并：secret 字段留空表示不修改。
// 入参 code 为渠道编码、submitted 为前端提交的 config JSON、current 为库中明文；
// 返回合并后的完整 config JSON。
func MergeSecretFields(code, submitted, current string) string {
	meta, ok := Metas[code]
	if !ok {
		return submitted
	}

	merged := parseJSON(submitted)
	plain := parseJSON(current)
	for _, field := range meta.Fields {
		if !field.Secret {
			continue
		}
		value, _ := merged[field.Key]
		if str, isStr := value.(string); isStr && strings.TrimSpace(str) == "" {
			if old, hasOld := plain[field.Key]; hasOld {
				merged[field.Key] = old
			}
		}
	}

	data, err := json.Marshal(merged)
	if err != nil {
		return submitted
	}
	return string(data)
}

// ValidateConfig 按 schema 校验必填字段，缺项返回业务错误码。
// 入参 code 为渠道编码、configJSON 为提交的 config 明文。
func ValidateConfig(code, configJSON string) error {
	meta, ok := Metas[code]
	if !ok {
		return errs.New(payErrs.PayChannelUnsupported, code)
	}

	plain := parseJSON(configJSON)
	missing := make([]string, 0)
	for _, field := range meta.Fields {
		if !field.Required {
			continue
		}

		value, exists := plain[field.Key]
		if !exists || value == nil {
			missing = append(missing, field.Label)
			continue
		}
		// 开关 / 数字类型只要求有值：false 与 0 都是合法取值，不能按空字符串判断
		if field.Type == "switch" {
			if _, ok := value.(bool); !ok {
				missing = append(missing, field.Label)
			}
			continue
		}
		if str, _ := value.(string); strings.TrimSpace(str) == "" {
			missing = append(missing, field.Label)
		}
	}
	if len(missing) > 0 {
		return errs.New(payErrs.PayChannelConfigBad, "缺少必填项："+strings.Join(missing, "、"))
	}

	return nil
}

// maskSecret 脱敏单个敏感值：保留前 4 与后 4 个字符，其余用 **** 代替。
// 入参 value 为原值；长度不足 8 时整体替换为 ****。
func maskSecret(value string) string {
	runes := []rune(strings.TrimSpace(value))
	if len(runes) == 0 {
		return ""
	}
	if len(runes) <= 8 {
		return "****"
	}

	return string(runes[:4]) + "****" + string(runes[len(runes)-4:])
}
