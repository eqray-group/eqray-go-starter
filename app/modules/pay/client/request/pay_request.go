// Package request 定义 pay 模块 C 端接口的入参结构。
// C 端请求体不含金额与会员 ID：金额由业务模块反查得出，会员身份一律取令牌。
package request

import (
	"github.com/goravel/framework/contracts/http"
)

// PayCreateRequest 是 C 端创建支付的入参。
type PayCreateRequest struct {
	// BizType 业务类型，如 demo_order
	BizType string `form:"bizType" json:"bizType"`
	// BizOrderNo 业务方订单号
	BizOrderNo string `form:"bizOrderNo" json:"bizOrderNo"`
	// ChannelCode 渠道编码，如 wx_mini
	ChannelCode string `form:"channelCode" json:"channelCode"`
	// ReturnURL 同步跳转地址（支付宝网页类支付用），可空
	ReturnURL string `form:"returnUrl" json:"returnUrl"`
}

// Rules 返回校验规则。
func (r *PayCreateRequest) Rules(ctx http.Context) map[string]any {
	return map[string]any{
		"bizType":     "required|max:32",
		"bizOrderNo":  "required|max:64",
		"channelCode": "required|max:32",
		"returnUrl":   "max:255",
	}
}

// Messages 覆盖默认提示。
func (r *PayCreateRequest) Messages(ctx http.Context) map[string]string {
	return map[string]string{
		"bizType.required":     "业务类型不能为空",
		"bizType.max":          "业务类型不能超过 32 个字符",
		"bizOrderNo.required":  "业务单号不能为空",
		"bizOrderNo.max":       "业务单号不能超过 64 个字符",
		"channelCode.required": "请选择支付方式",
		"channelCode.max":      "渠道编码不能超过 32 个字符",
		"returnUrl.max":        "跳转地址不能超过 255 个字符",
	}
}
