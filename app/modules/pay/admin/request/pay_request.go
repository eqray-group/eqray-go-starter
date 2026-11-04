// Package request 定义 pay 模块管理端接口的入参结构与校验规则。
package request

import (
	"github.com/goravel/framework/contracts/http"
)

// PayChannelSaveRequest 是新增 / 编辑支付渠道的入参。
type PayChannelSaveRequest struct {
	// Name 渠道名称
	Name string `form:"name" json:"name"`
	// Code 渠道编码（编辑时忽略，以路径 ID 为准）
	Code string `form:"code" json:"code"`
	// Config config JSON 字符串，敏感字段留空表示沿用原值
	Config string `form:"config" json:"config"`
	// Status 状态：0 停用 / 1 启用
	Status int8 `form:"status" json:"status"`
	// Sort 排序
	Sort int `form:"sort" json:"sort"`
	// Remark 备注
	Remark string `form:"remark" json:"remark"`
}

// Rules 返回校验规则。
func (r *PayChannelSaveRequest) Rules(ctx http.Context) map[string]any {
	return map[string]any{
		"name":   "required|max:64",
		"code":   "required_without:id|max:32",
		"config": "required|max:65535",
		"status": "required|integer|in:0,1",
		"sort":   "integer",
		"remark": "max:255",
	}
}

// Messages 覆盖默认提示。
func (r *PayChannelSaveRequest) Messages(ctx http.Context) map[string]string {
	return map[string]string{
		"name.required":         "渠道名称不能为空",
		"name.max":              "渠道名称不能超过 64 个字符",
		"code.required_without": "渠道编码不能为空",
		"code.max":              "渠道编码不能超过 32 个字符",
		"config.required":       "渠道配置不能为空",
		"config.max":            "渠道配置过长",
		"status.required":       "状态不能为空",
		"status.integer":        "状态取值不合法",
		"status.in":             "状态取值不合法",
		"remark.max":            "备注不能超过 255 个字符",
	}
}

// PayChannelStatusRequest 是渠道启停的入参。
type PayChannelStatusRequest struct {
	// Enable 是否启用
	Enable bool `form:"enable" json:"enable"`
}

// Rules 返回校验规则。
func (r *PayChannelStatusRequest) Rules(ctx http.Context) map[string]any {
	return map[string]any{}
}

// Messages 覆盖默认提示。
func (r *PayChannelStatusRequest) Messages(ctx http.Context) map[string]string {
	return map[string]string{}
}

// PayRefundCreateRequest 是发起退款的入参。
type PayRefundCreateRequest struct {
	// OrderID 支付订单 ID
	OrderID int64 `form:"orderId" json:"orderId"`
	// RefundPrice 退款金额（分）
	RefundPrice int64 `form:"refundPrice" json:"refundPrice"`
	// Reason 退款原因
	Reason string `form:"reason" json:"reason"`
}

// Rules 返回校验规则。
func (r *PayRefundCreateRequest) Rules(ctx http.Context) map[string]any {
	return map[string]any{
		"orderId":     "required",
		"refundPrice": "required|integer|min:1",
		"reason":      "required|max:255",
	}
}

// Messages 覆盖默认提示。
func (r *PayRefundCreateRequest) Messages(ctx http.Context) map[string]string {
	return map[string]string{
		"orderId.required":     "请指定支付订单",
		"refundPrice.required": "退款金额不能为空",
		"refundPrice.integer":  "退款金额必须为整数（分）",
		"refundPrice.min":      "退款金额必须大于 0",
		"reason.required":      "退款原因不能为空",
		"reason.max":           "退款原因不能超过 255 个字符",
	}
}
