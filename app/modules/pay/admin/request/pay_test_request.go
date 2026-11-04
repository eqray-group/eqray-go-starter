package request

import (
	"github.com/goravel/framework/contracts/http"

	"eqray-go-starter/app/modules/base/types"
)

// PayTestOrderRequest 是后台支付联调下单的入参。
type PayTestOrderRequest struct {
	// UserID 付款会员 ID；用 types.ID 兼容前端传来的字符串形式雪花 ID
	UserID types.ID `form:"userId" json:"userId"`
	// ChannelCode 渠道编码
	ChannelCode string `form:"channelCode" json:"channelCode"`
}

// Rules 返回校验规则。
func (r *PayTestOrderRequest) Rules(ctx http.Context) map[string]any {
	return map[string]any{
		"userId":      "required|integer|min:1",
		"channelCode": "required|max:32",
	}
}

// Messages 覆盖默认提示。
func (r *PayTestOrderRequest) Messages(ctx http.Context) map[string]string {
	return map[string]string{
		"userId.required":      "请输入付款会员 ID",
		"userId.integer":       "会员 ID 不合法",
		"userId.min":           "会员 ID 不合法",
		"channelCode.required": "请选择支付渠道",
		"channelCode.max":      "渠道编码不合法",
	}
}

// PayTestBatchRequest 是批量联调下单的入参。
type PayTestBatchRequest struct {
	// UserID 付款会员 ID；用 types.ID 兼容前端传来的字符串形式雪花 ID
	UserID types.ID `form:"userId" json:"userId"`
	// ChannelCodes 渠道编码列表，为空表示全部启用渠道
	ChannelCodes []string `form:"channelCodes" json:"channelCodes"`
}

// Rules 返回校验规则。
func (r *PayTestBatchRequest) Rules(ctx http.Context) map[string]any {
	return map[string]any{
		"userId": "required|integer|min:1",
	}
}

// Messages 覆盖默认提示。
func (r *PayTestBatchRequest) Messages(ctx http.Context) map[string]string {
	return map[string]string{
		"userId.required": "请输入付款会员 ID",
		"userId.integer":  "会员 ID 不合法",
		"userId.min":      "会员 ID 不合法",
	}
}
