package service

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"eqray-go-starter/app/facades"
	baseErrs "eqray-go-starter/app/modules/base/errs"
	"eqray-go-starter/app/modules/base/support"
	"eqray-go-starter/app/modules/base/types"
	"eqray-go-starter/app/modules/infra/model"
	"eqray-go-starter/app/modules/infra/repository"
)

// noticeSendTimeout 是单次发送的超时时间，避免第三方接口卡住请求。
const noticeSendTimeout = 20 * time.Second

// NoticeListQuery 是渠道 / 模板列表的查询条件。
type NoticeListQuery struct {
	// Keyword 名称或编码模糊匹配
	Keyword string
	// Type 渠道类型
	Type string
	// Status 状态，nil 表示全部
	Status *int8
	// PageNo / PageSize 分页
	PageNo   int
	PageSize int
}

// NoticeLogQuery 是发送记录的查询条件。
type NoticeLogQuery struct {
	// Type 渠道类型
	Type string
	// Status 结果，nil 表示全部
	Status *int8
	// PageNo / PageSize 分页
	PageNo   int
	PageSize int
}

// NoticeChannelInput 是渠道新增 / 编辑入参。
type NoticeChannelInput struct {
	// Name 渠道名称
	Name string
	// Type 渠道类型
	Type string
	// Config 渠道配置
	Config map[string]string
	// Status 0 停用 / 1 启用
	Status int8
	// Remark 备注
	Remark string
}

// NoticeTemplateInput 是模板新增 / 编辑入参。
type NoticeTemplateInput struct {
	// Code 模板编码
	Code string
	// Name 模板名称
	Name string
	// Type 适用渠道类型
	Type string
	// Title 标题
	Title string
	// Content 内容，支持 {{变量}}
	Content string
	// Status 0 停用 / 1 启用
	Status int8
	// Remark 备注
	Remark string
}

// NoticeSendInput 是发送入参：可指定渠道 + 模板，也可直接给标题正文。
type NoticeSendInput struct {
	// ChannelID 渠道 ID，0 表示按 Type 取第一个启用渠道
	ChannelID types.ID
	// Type 渠道类型，ChannelID 为 0 时使用
	Type string
	// TemplateCode 模板编码，填写后按模板渲染
	TemplateCode string
	// Target 接收目标（邮件地址；机器人渠道可空）
	Target string
	// Title 标题，覆盖模板标题
	Title string
	// Content 正文，覆盖模板正文
	Content string
	// Variables 模板变量
	Variables map[string]string
	// Operator 操作人
	Operator string
}

// NoticeService 负责通知渠道 / 模板的配置与发送。
// 发送不依赖业务模块：业务只给「模板编码 + 变量 + 目标」，渠道与格式由后台配置决定。
type NoticeService struct {
	notices *repository.InfraNoticeRepository
}

// NewNoticeService 构造通知服务。
func NewNoticeService() *NoticeService {
	return &NoticeService{notices: repository.NewInfraNoticeRepository()}
}

// Types 返回支持的渠道类型与字段定义。
// 入参无；返回类型列表。
func (s *NoticeService) Types() []NoticeTypeSchema {
	return NoticeTypes()
}

// PageChannels 分页查询渠道。
func (s *NoticeService) PageChannels(ctx context.Context, q NoticeListQuery) ([]model.InfraNoticeChannelModel, int64, error) {
	pageNo, pageSize := normalizePage(q.PageNo, q.PageSize)

	return s.notices.PaginateChannels(ctx, repository.NoticeChannelFilter{
		Keyword: q.Keyword,
		Type:    q.Type,
		Status:  q.Status,
	}, pageNo, pageSize)
}

// CreateChannel 新增渠道：校验类型与必填配置。
func (s *NoticeService) CreateChannel(ctx context.Context, in NoticeChannelInput, operator string) (types.ID, error) {
	if _, ok := noticeTypeByName(in.Type); !ok {
		return 0, baseErrs.WithMsg(baseErrs.BadRequest, "不支持的渠道类型")
	}
	if err := validateNoticeConfig(in.Type, in.Config); err != nil {
		return 0, baseErrs.WithMsg(baseErrs.BadRequest, err.Error())
	}

	configJSON, err := marshalNoticeConfig(in.Config)
	if err != nil {
		return 0, baseErrs.WithMsg(baseErrs.BadRequest, "渠道配置序列化失败")
	}

	now := types.Now()
	id := support.NewID()
	m := &model.InfraNoticeChannelModel{
		ID:        id,
		Name:      strings.TrimSpace(in.Name),
		Type:      in.Type,
		Config:    configJSON,
		Status:    in.Status,
		Remark:    strings.TrimSpace(in.Remark),
		Creator:   operator,
		CreatedAt: now,
		Updater:   operator,
		UpdatedAt: now,
	}
	if err := s.notices.InsertChannel(ctx, m); err != nil {
		return 0, err
	}

	return id, nil
}

// UpdateChannel 编辑渠道（类型不可改，配置整体覆盖）。
func (s *NoticeService) UpdateChannel(ctx context.Context, id types.ID, in NoticeChannelInput, operator string) error {
	m, err := s.findChannel(ctx, id)
	if err != nil {
		return err
	}

	cfg := in.Config
	// 密码类字段前端留空表示「不修改」，此处沿用原值
	oldCfg := unmarshalNoticeConfig(m.Config)
	schema, _ := noticeTypeByName(m.Type)
	for _, field := range schema.Fields {
		if field.Type != "password" {
			continue
		}
		if strings.TrimSpace(cfg[field.Key]) == "" && strings.TrimSpace(oldCfg[field.Key]) != "" {
			cfg[field.Key] = oldCfg[field.Key]
		}
	}

	if err := validateNoticeConfig(m.Type, cfg); err != nil {
		return baseErrs.WithMsg(baseErrs.BadRequest, err.Error())
	}

	configJSON, err := marshalNoticeConfig(cfg)
	if err != nil {
		return baseErrs.WithMsg(baseErrs.BadRequest, "渠道配置序列化失败")
	}

	return s.notices.UpdateChannel(ctx, m.ID, map[string]any{
		"name":   strings.TrimSpace(in.Name),
		"config": configJSON,
		"status": in.Status,
		"remark": strings.TrimSpace(in.Remark),
	}, operator)
}

// DeleteChannel 删除渠道。
func (s *NoticeService) DeleteChannel(ctx context.Context, id types.ID) error {
	if _, err := s.findChannel(ctx, id); err != nil {
		return err
	}

	return s.notices.DeleteChannel(ctx, id)
}

// PageTemplates 分页查询模板。
func (s *NoticeService) PageTemplates(ctx context.Context, q NoticeListQuery) ([]model.InfraNoticeTemplateModel, int64, error) {
	pageNo, pageSize := normalizePage(q.PageNo, q.PageSize)

	return s.notices.PaginateTemplates(ctx, repository.NoticeTemplateFilter{
		Keyword: q.Keyword,
		Type:    q.Type,
		Status:  q.Status,
	}, pageNo, pageSize)
}

// CreateTemplate 新增模板：编码唯一、类型合法。
func (s *NoticeService) CreateTemplate(ctx context.Context, in NoticeTemplateInput, operator string) (types.ID, error) {
	if _, ok := noticeTypeByName(in.Type); !ok {
		return 0, baseErrs.WithMsg(baseErrs.BadRequest, "不支持的渠道类型")
	}
	if strings.TrimSpace(in.Code) == "" {
		return 0, baseErrs.WithMsg(baseErrs.BadRequest, "模板编码不能为空")
	}

	existing, err := s.notices.FindTemplateByCode(ctx, strings.TrimSpace(in.Code))
	if err != nil {
		return 0, err
	}
	if existing != nil {
		return 0, baseErrs.WithMsg(baseErrs.BadRequest, "模板编码已存在")
	}

	now := types.Now()
	id := support.NewID()
	m := &model.InfraNoticeTemplateModel{
		ID:        id,
		Code:      strings.TrimSpace(in.Code),
		Name:      strings.TrimSpace(in.Name),
		Type:      in.Type,
		Title:     strings.TrimSpace(in.Title),
		Content:   in.Content,
		Status:    in.Status,
		Remark:    strings.TrimSpace(in.Remark),
		Creator:   operator,
		CreatedAt: now,
		Updater:   operator,
		UpdatedAt: now,
	}
	if err := s.notices.InsertTemplate(ctx, m); err != nil {
		return 0, err
	}

	return id, nil
}

// UpdateTemplate 编辑模板（编码不可改）。
func (s *NoticeService) UpdateTemplate(ctx context.Context, id types.ID, in NoticeTemplateInput, operator string) error {
	m, err := s.findTemplate(ctx, id)
	if err != nil {
		return err
	}
	if _, ok := noticeTypeByName(in.Type); !ok {
		return baseErrs.WithMsg(baseErrs.BadRequest, "不支持的渠道类型")
	}

	return s.notices.UpdateTemplate(ctx, m.ID, map[string]any{
		"name":    strings.TrimSpace(in.Name),
		"type":    in.Type,
		"title":   strings.TrimSpace(in.Title),
		"content": in.Content,
		"status":  in.Status,
		"remark":  strings.TrimSpace(in.Remark),
	}, operator)
}

// DeleteTemplate 删除模板。
func (s *NoticeService) DeleteTemplate(ctx context.Context, id types.ID) error {
	if _, err := s.findTemplate(ctx, id); err != nil {
		return err
	}

	return s.notices.DeleteTemplate(ctx, id)
}

// PageLogs 分页查询发送记录。
func (s *NoticeService) PageLogs(ctx context.Context, q NoticeLogQuery) ([]model.InfraNoticeLogModel, int64, error) {
	pageNo, pageSize := normalizePage(q.PageNo, q.PageSize)

	return s.notices.PaginateLogs(ctx, repository.NoticeLogFilter{
		Type:   q.Type,
		Status: q.Status,
	}, pageNo, pageSize)
}

// TestChannel 用固定样例内容发一条测试消息，验证配置是否可用。
// 入参 id 为渠道 ID、target 为接收目标、operator 为操作人；返回本次发送记录。
func (s *NoticeService) TestChannel(ctx context.Context, id types.ID, target, operator string) (*model.InfraNoticeLogModel, error) {
	m, err := s.findChannel(ctx, id)
	if err != nil {
		return nil, err
	}

	content := "这是一条测试消息，收到即表示渠道配置可用。\n发送时间：" + time.Now().Format("2006-01-02 15:04:05")

	return s.dispatch(ctx, m, "", target, "测试通知", content, operator), nil
}

// Send 按渠道 + 模板发送通知；模板变量用 {{key}} 占位。
// 入参 in 为发送参数；返回发送记录（失败时记录里 error 有值）。
func (s *NoticeService) Send(ctx context.Context, in NoticeSendInput) (*model.InfraNoticeLogModel, error) {
	title := strings.TrimSpace(in.Title)
	content := in.Content
	templateCode := strings.TrimSpace(in.TemplateCode)
	noticeType := strings.TrimSpace(in.Type)

	if templateCode != "" {
		tpl, err := s.notices.FindTemplateByCode(ctx, templateCode)
		if err != nil {
			return nil, err
		}
		if tpl == nil {
			return nil, baseErrs.WithMsg(baseErrs.BadRequest, "模板不存在："+templateCode)
		}
		if tpl.Status != 1 {
			return nil, baseErrs.WithMsg(baseErrs.BadRequest, "模板已停用："+templateCode)
		}
		if title == "" {
			title = tpl.Title
		}
		if strings.TrimSpace(content) == "" {
			content = tpl.Content
		}
		if noticeType == "" {
			noticeType = tpl.Type
		}
	}

	title = renderNoticeTemplate(title, in.Variables)
	content = renderNoticeTemplate(content, in.Variables)
	if strings.TrimSpace(content) == "" {
		return nil, baseErrs.WithMsg(baseErrs.BadRequest, "通知内容不能为空")
	}

	channel, err := s.resolveChannel(ctx, in.ChannelID, noticeType)
	if err != nil {
		return nil, err
	}

	return s.dispatch(ctx, channel, templateCode, strings.TrimSpace(in.Target), title, content, in.Operator), nil
}

// resolveChannel 决定用哪个渠道：优先指定 ID，其次按类型取第一个启用渠道。
// 入参 id 为渠道 ID、noticeType 为类型；返回渠道模型或业务错误。
func (s *NoticeService) resolveChannel(ctx context.Context, id types.ID, noticeType string) (*model.InfraNoticeChannelModel, error) {
	if id > 0 {
		return s.findChannel(ctx, id)
	}
	if noticeType == "" {
		return nil, baseErrs.WithMsg(baseErrs.BadRequest, "请指定渠道或渠道类型")
	}

	channel, err := s.notices.FirstEnabledChannel(ctx, noticeType)
	if err != nil {
		return nil, err
	}
	if channel == nil {
		return nil, baseErrs.WithMsg(baseErrs.BadRequest, "没有启用中的渠道，请先在「消息通知 → 渠道」配置")
	}

	return channel, nil
}

// dispatch 执行发送并写记录：无论成败都落一条 infra_notice_log。
// 入参 channel 为渠道、templateCode 为模板编码、target/title/content 为内容、operator 为操作人；返回日志。
func (s *NoticeService) dispatch(ctx context.Context, channel *model.InfraNoticeChannelModel, templateCode, target, title, content, operator string) *model.InfraNoticeLogModel {
	startedAt := time.Now()

	cfg := unmarshalNoticeConfig(channel.Config)
	sendCtx, cancel := context.WithTimeout(ctx, noticeSendTimeout)
	err := sendNotice(sendCtx, channel.Type, cfg, target, title, content)
	cancel()

	duration := int(time.Since(startedAt).Milliseconds())
	status := model.NoticeResultSuccess
	errText := ""
	if err != nil {
		status = model.NoticeResultFailed
		errText = err.Error()
		facades.Log().Errorf("[消息通知] %s 发送失败（%s）：%v", channel.Name, channel.Type, err)
	}

	logModel := &model.InfraNoticeLogModel{
		ID:           support.NewID(),
		ChannelID:    channel.ID,
		ChannelName:  channel.Name,
		Type:         channel.Type,
		TemplateCode: templateCode,
		Target:       target,
		Title:        title,
		Content:      truncateNoticeContent(content),
		Status:       status,
		Error:        errText,
		Duration:     duration,
		Creator:      operator,
		CreatedAt:    types.DateTime(time.Now()),
	}
	if insertErr := s.notices.InsertLog(ctx, logModel); insertErr != nil {
		facades.Log().Errorf("[消息通知] 写发送记录失败：%v", insertErr)
	}

	return logModel
}

// findChannel 按 ID 查渠道，不存在返回业务错误。
func (s *NoticeService) findChannel(ctx context.Context, id types.ID) (*model.InfraNoticeChannelModel, error) {
	m, err := s.notices.FindChannelByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if m == nil {
		return nil, baseErrs.WithMsg(baseErrs.BadRequest, "渠道不存在")
	}

	return m, nil
}

// findTemplate 按 ID 查模板，不存在返回业务错误。
func (s *NoticeService) findTemplate(ctx context.Context, id types.ID) (*model.InfraNoticeTemplateModel, error) {
	m, err := s.notices.FindTemplateByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if m == nil {
		return nil, baseErrs.WithMsg(baseErrs.BadRequest, "模板不存在")
	}

	return m, nil
}

// marshalNoticeConfig 把配置 map 序列化为 JSON 文本。
// 入参 cfg 为配置；返回 JSON 字符串与错误。
func marshalNoticeConfig(cfg map[string]string) (string, error) {
	if cfg == nil {
		return "{}", nil
	}

	data, err := json.Marshal(cfg)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

// unmarshalNoticeConfig 把 JSON 文本还原成配置 map，解析失败返回空 map。
// 入参 raw 为 JSON 文本；返回配置。
func unmarshalNoticeConfig(raw string) map[string]string {
	cfg := make(map[string]string)
	if strings.TrimSpace(raw) == "" {
		return cfg
	}

	// 兼容历史数据里数字 / 布尔类型，统一转成字符串
	generic := make(map[string]any)
	if err := json.Unmarshal([]byte(raw), &generic); err != nil {
		return cfg
	}
	for key, value := range generic {
		if value == nil {
			continue
		}
		if text, ok := value.(string); ok {
			cfg[key] = text
			continue
		}
		cfg[key] = strings.TrimSpace(strings.Trim(strings.TrimSpace(jsonValue(value)), `"`))
	}

	return cfg
}

// jsonValue 把任意值序列化成 JSON 片段，供 unmarshalNoticeConfig 转字符串。
// 入参 value 为任意值；返回 JSON 文本（失败返回空串）。
func jsonValue(value any) string {
	data, err := json.Marshal(value)
	if err != nil {
		return ""
	}

	return string(data)
}

// renderNoticeTemplate 替换模板里的 {{key}} 占位符；未提供的变量原样保留，便于发现漏传。
// 入参 text 为模板文本、vars 为变量表；返回渲染结果。
func renderNoticeTemplate(text string, vars map[string]string) string {
	text = strings.TrimSpace(text)
	if text == "" || len(vars) == 0 {
		return text
	}

	for key, value := range vars {
		text = strings.ReplaceAll(text, "{{"+key+"}}", value)
		text = strings.ReplaceAll(text, "{{ "+key+" }}", value)
	}

	return text
}

// truncateNoticeContent 截断发送内容，避免超长文本占满记录表。
// 入参 value 为原始内容；返回截断后的内容（上限 4000 字符）。
func truncateNoticeContent(value string) string {
	runes := []rune(value)
	if len(runes) <= 4000 {
		return value
	}

	return string(runes[:4000]) + "…"
}
