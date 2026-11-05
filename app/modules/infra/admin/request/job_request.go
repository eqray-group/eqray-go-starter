package request

import (
	"github.com/goravel/framework/contracts/http"
)

// JobSaveRequest 是新增 / 编辑定时任务的入参。
type JobSaveRequest struct {
	// Name 任务名称
	Name string `form:"name" json:"name"`
	// Handler 处理器标识（编辑时忽略，创建后不可改）
	Handler string `form:"handler" json:"handler"`
	// Cron cron 表达式（5 段）
	Cron string `form:"cron" json:"cron"`
	// Status 状态：0 停用 / 1 启用
	Status int8 `form:"status" json:"status"`
	// Remark 备注
	Remark string `form:"remark" json:"remark"`
}

// Rules 返回校验规则。
func (r *JobSaveRequest) Rules(ctx http.Context) map[string]any {
	return map[string]any{
		"name":    "required|max:64",
		"handler": "max:64",
		"cron":    "required|max:64",
		"status":  "required|integer|in:0,1",
		"remark":  "max:255",
	}
}

// Messages 覆盖默认提示。
func (r *JobSaveRequest) Messages(ctx http.Context) map[string]string {
	return map[string]string{
		"name.required": "请输入任务名称",
		"name.max":      "任务名称不能超过 64 个字符",
		"handler.max":   "处理器标识不合法",
		"cron.required": "请填写 cron 表达式",
		"cron.max":      "cron 表达式不能超过 64 个字符",
		"status.in":     "状态取值不合法",
		"remark.max":    "备注不能超过 255 个字符",
	}
}

// JobStatusRequest 是启停任务的入参。
type JobStatusRequest struct {
	// Status 目标状态：0 停用 / 1 启用
	Status int8 `form:"status" json:"status"`
}

// Rules 返回校验规则。
func (r *JobStatusRequest) Rules(ctx http.Context) map[string]any {
	return map[string]any{
		"status": "required|integer|in:0,1",
	}
}

// Messages 覆盖默认提示。
func (r *JobStatusRequest) Messages(ctx http.Context) map[string]string {
	return map[string]string{
		"status.in": "状态取值不合法",
	}
}
