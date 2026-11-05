package request

import (
	"github.com/goravel/framework/contracts/http"

	"eqray-go-starter/app/modules/base/types"
)

// NoticeChannelRequest 是通知渠道新增 / 编辑入参。
type NoticeChannelRequest struct {
	// Name 渠道名称
	Name string `form:"name" json:"name"`
	// Type 渠道类型：email / wecom / dingtalk
	Type string `form:"type" json:"type"`
	// Config 渠道配置（随类型变化）
	Config map[string]string `form:"config" json:"config"`
	// Status 0 停用 / 1 启用
	Status int8 `form:"status" json:"status"`
	// Remark 备注
	Remark string `form:"remark" json:"remark"`
}

// Rules 返回校验规则；配置字段由服务层按类型校验。
func (r *NoticeChannelRequest) Rules(ctx http.Context) map[string]any {
	return map[string]any{
		"name":   "required|max:64",
		"type":   "required|in:email,wecom,dingtalk",
		"status": "required|integer|in:0,1",
		"remark": "max:255",
	}
}

// Messages 覆盖默认提示。
func (r *NoticeChannelRequest) Messages(ctx http.Context) map[string]string {
	return map[string]string{
		"name.required": "请输入渠道名称",
		"name.max":      "渠道名称不能超过 64 个字符",
		"type.in":       "不支持的渠道类型",
		"status.in":     "状态取值不合法",
		"remark.max":    "备注不能超过 255 个字符",
	}
}

// NoticeTemplateRequest 是通知模板新增 / 编辑入参。
type NoticeTemplateRequest struct {
	// Code 模板编码（编辑时忽略）
	Code string `form:"code" json:"code"`
	// Name 模板名称
	Name string `form:"name" json:"name"`
	// Type 适用渠道类型
	Type string `form:"type" json:"type"`
	// Title 标题
	Title string `form:"title" json:"title"`
	// Content 内容，支持 {{变量}}
	Content string `form:"content" json:"content"`
	// Status 0 停用 / 1 启用
	Status int8 `form:"status" json:"status"`
	// Remark 备注
	Remark string `form:"remark" json:"remark"`
}

// Rules 返回校验规则。
func (r *NoticeTemplateRequest) Rules(ctx http.Context) map[string]any {
	return map[string]any{
		"code":    "max:64",
		"name":    "required|max:64",
		"type":    "required|in:email,wecom,dingtalk",
		"title":   "max:128",
		"content": "required|max:20000",
		"status":  "required|integer|in:0,1",
		"remark":  "max:255",
	}
}

// Messages 覆盖默认提示。
func (r *NoticeTemplateRequest) Messages(ctx http.Context) map[string]string {
	return map[string]string{
		"code.max":         "模板编码不能超过 64 个字符",
		"name.required":    "请输入模板名称",
		"type.in":          "不支持的渠道类型",
		"title.max":        "标题不能超过 128 个字符",
		"content.required": "请填写模板内容",
		"content.max":      "模板内容过长（上限 20000 字符）",
		"status.in":        "状态取值不合法",
		"remark.max":       "备注不能超过 255 个字符",
	}
}

// NoticeTestRequest 是渠道测试发送的入参。
type NoticeTestRequest struct {
	// Target 接收目标（邮件地址）
	Target string `form:"target" json:"target"`
}

// Rules 返回校验规则。
func (r *NoticeTestRequest) Rules(ctx http.Context) map[string]any {
	return map[string]any{
		"target": "max:255",
	}
}

// NoticeSendRequest 是通用发送入参：指定渠道或类型 + 模板 + 变量。
type NoticeSendRequest struct {
	// ChannelID 渠道 ID，0 表示按 Type 取第一个启用渠道
	ChannelID types.ID `form:"channelId" json:"channelId"`
	// Type 渠道类型
	Type string `form:"type" json:"type"`
	// TemplateCode 模板编码
	TemplateCode string `form:"templateCode" json:"templateCode"`
	// Target 接收目标
	Target string `form:"target" json:"target"`
	// Title 标题（覆盖模板）
	Title string `form:"title" json:"title"`
	// Content 正文（覆盖模板）
	Content string `form:"content" json:"content"`
	// Variables 模板变量
	Variables map[string]string `form:"variables" json:"variables"`
}

// Rules 返回校验规则。
func (r *NoticeSendRequest) Rules(ctx http.Context) map[string]any {
	return map[string]any{
		"type":         "max:32",
		"templateCode": "max:64",
		"target":       "max:255",
		"title":        "max:128",
	}
}
