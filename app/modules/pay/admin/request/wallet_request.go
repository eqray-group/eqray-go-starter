package request

import (
	"github.com/goravel/framework/contracts/http"

	"eqray-go-starter/app/modules/base/types"
)

// PayWalletAdjustRequest 是后台调整钱包余额的入参。
type PayWalletAdjustRequest struct {
	// UserID 会员 ID；用 types.ID 兼容前端传来的字符串形式雪花 ID
	UserID types.ID `form:"userId" json:"userId"`
	// Amount 调整金额（分）：正数充值、负数扣减
	Amount int64 `form:"amount" json:"amount"`
	// Remark 备注
	Remark string `form:"remark" json:"remark"`
}

// PayWalletFreezeRequest 是后台冻结 / 解冻钱包资金的入参。
type PayWalletFreezeRequest struct {
	// UserID 会员 ID；用 types.ID 兼容前端传来的字符串形式雪花 ID
	UserID types.ID `form:"userId" json:"userId"`
	// Amount 金额（分），必须大于 0
	Amount int64 `form:"amount" json:"amount"`
	// Remark 备注
	Remark string `form:"remark" json:"remark"`
}

// Rules 返回校验规则。
func (r *PayWalletFreezeRequest) Rules(ctx http.Context) map[string]any {
	return map[string]any{
		"userId": "required|integer|min:1",
		"amount": "required|integer|min:1",
		"remark": "max:255",
	}
}

// Messages 覆盖默认提示。
func (r *PayWalletFreezeRequest) Messages(ctx http.Context) map[string]string {
	return map[string]string{
		"userId.required": "请先选择会员",
		"userId.integer":  "会员 ID 不合法",
		"userId.min":      "会员 ID 不合法",
		"amount.required": "请输入金额",
		"amount.integer":  "金额必须是整数（单位：分）",
		"amount.min":      "金额必须大于 0",
		"remark.max":      "备注不能超过 255 个字符",
	}
}

// Rules 返回校验规则。
func (r *PayWalletAdjustRequest) Rules(ctx http.Context) map[string]any {
	return map[string]any{
		"userId": "required|integer|min:1",
		"amount": "required|integer",
		"remark": "max:255",
	}
}

// Messages 覆盖默认提示。
func (r *PayWalletAdjustRequest) Messages(ctx http.Context) map[string]string {
	return map[string]string{
		"userId.required": "请先选择会员",
		"userId.integer":  "会员 ID 不合法",
		"userId.min":      "会员 ID 不合法",
		"amount.required": "请输入调整金额",
		"amount.integer":  "调整金额必须是整数（单位：分）",
		"remark.max":      "备注不能超过 255 个字符",
	}
}
