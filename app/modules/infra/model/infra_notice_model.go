package model

import (
	"eqray-go-starter/app/modules/base/types"
)

// 表名常量。
const (
	// InfraNoticeChannelTable 通知渠道表
	InfraNoticeChannelTable = "infra_notice_channel"
	// InfraNoticeTemplateTable 通知模板表
	InfraNoticeTemplateTable = "infra_notice_template"
	// InfraNoticeLogTable 通知发送记录表
	InfraNoticeLogTable = "infra_notice_log"
)

// 通知渠道类型。
const (
	// NoticeTypeEmail 邮件（SMTP）
	NoticeTypeEmail = "email"
	// NoticeTypeWeCom 企业微信群机器人
	NoticeTypeWeCom = "wecom"
	// NoticeTypeDingTalk 钉钉群机器人
	NoticeTypeDingTalk = "dingtalk"
)

// 通知发送结果。
const (
	// NoticeResultFailed 失败
	NoticeResultFailed int8 = 0
	// NoticeResultSuccess 成功
	NoticeResultSuccess int8 = 1
)

// InfraNoticeChannelModel 对应 infra_notice_channel：一条记录是一个可发送的渠道实例。
type InfraNoticeChannelModel struct {
	// ID 主键
	ID types.ID `db:"id" json:"id"`
	// Name 渠道名称
	Name string `db:"name" json:"name"`
	// Type 渠道类型：email / wecom / dingtalk
	Type string `db:"type" json:"type"`
	// Config 渠道配置 JSON 文本
	Config string `db:"config" json:"config"`
	// Status 0 停用 / 1 启用
	Status int8 `db:"status" json:"status"`
	// Remark 备注
	Remark string `db:"remark" json:"remark"`
	// Creator 创建人
	Creator string `db:"creator" json:"creator"`
	// CreatedAt 创建时间
	CreatedAt types.DateTime `db:"created_at" json:"createdAt"`
	// Updater 更新人
	Updater string `db:"updater" json:"updater"`
	// UpdatedAt 更新时间
	UpdatedAt types.DateTime `db:"updated_at" json:"updatedAt"`
}

// TableName 返回表名。
func (m *InfraNoticeChannelModel) TableName() string {
	return InfraNoticeChannelTable
}

// InfraNoticeTemplateModel 对应 infra_notice_template：业务按 code 取模板渲染。
type InfraNoticeTemplateModel struct {
	// ID 主键
	ID types.ID `db:"id" json:"id"`
	// Code 模板编码
	Code string `db:"code" json:"code"`
	// Name 模板名称
	Name string `db:"name" json:"name"`
	// Type 适用渠道类型
	Type string `db:"type" json:"type"`
	// Title 标题
	Title string `db:"title" json:"title"`
	// Content 内容，支持 {{变量}}
	Content string `db:"content" json:"content"`
	// Status 0 停用 / 1 启用
	Status int8 `db:"status" json:"status"`
	// Remark 备注
	Remark string `db:"remark" json:"remark"`
	// Creator 创建人
	Creator string `db:"creator" json:"creator"`
	// CreatedAt 创建时间
	CreatedAt types.DateTime `db:"created_at" json:"createdAt"`
	// Updater 更新人
	Updater string `db:"updater" json:"updater"`
	// UpdatedAt 更新时间
	UpdatedAt types.DateTime `db:"updated_at" json:"updatedAt"`
}

// TableName 返回表名。
func (m *InfraNoticeTemplateModel) TableName() string {
	return InfraNoticeTemplateTable
}

// InfraNoticeLogModel 对应 infra_notice_log：一次发送一行，只追加。
type InfraNoticeLogModel struct {
	// ID 主键
	ID types.ID `db:"id" json:"id"`
	// ChannelID 渠道 ID
	ChannelID types.ID `db:"channel_id" json:"channelId"`
	// ChannelName 渠道名称（冗余）
	ChannelName string `db:"channel_name" json:"channelName"`
	// Type 渠道类型
	Type string `db:"type" json:"type"`
	// TemplateCode 模板编码
	TemplateCode string `db:"template_code" json:"templateCode"`
	// Target 接收目标
	Target string `db:"target" json:"target"`
	// Title 标题
	Title string `db:"title" json:"title"`
	// Content 实际发送内容
	Content string `db:"content" json:"content"`
	// Status 0 失败 / 1 成功
	Status int8 `db:"status" json:"status"`
	// Error 失败原因
	Error string `db:"error" json:"error"`
	// Duration 耗时（毫秒）
	Duration int `db:"duration" json:"duration"`
	// Creator 操作人
	Creator string `db:"creator" json:"creator"`
	// CreatedAt 创建时间
	CreatedAt types.DateTime `db:"created_at" json:"createdAt"`
}

// TableName 返回表名。
func (m *InfraNoticeLogModel) TableName() string {
	return InfraNoticeLogTable
}
