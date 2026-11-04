// Package service 是 pay 模块的业务层：渠道配置、下单、退款与回调处理都在这里编排。
package service

import (
	"context"
	"strings"

	"eqray-go-starter/app/facades"
	baseEnums "eqray-go-starter/app/modules/base/enums"
	baseErrs "eqray-go-starter/app/modules/base/errs"
	"eqray-go-starter/app/modules/base/support"
	"eqray-go-starter/app/modules/base/types"
	payChannel "eqray-go-starter/app/modules/pay/channel"
	payErrs "eqray-go-starter/app/modules/pay/errs"
	payModel "eqray-go-starter/app/modules/pay/model"
	payRepository "eqray-go-starter/app/modules/pay/repository"
)

// CallbackPath 是异步通知固定路由前缀，完整地址 = APP_URL + 该前缀 + 渠道编码。
const CallbackPath = "/api/v1/pay/notify/"

// PayChannelSaveInput 是新增 / 编辑支付渠道的入参。
type PayChannelSaveInput struct {
	// Name 渠道名称
	Name string
	// Code 渠道编码，新增必填且唯一，编辑忽略
	Code string
	// Config config JSON 明文（secret 字段留空表示沿用原值）
	Config string
	// Status 状态：0 停用 / 1 启用
	Status int8
	// Sort 排序
	Sort int
	// Remark 备注
	Remark string
}

// PayChannelListQuery 是渠道列表的查询条件。
type PayChannelListQuery struct {
	// Keyword 名称模糊匹配
	Keyword string
	// Code 渠道编码精确匹配
	Code string
	// Status 状态过滤
	Status *baseEnums.CommonStatusEnum
	// PageNo / PageSize 分页参数
	PageNo   int
	PageSize int
}

// PayChannelService 负责支付渠道配置的增删改查与渠道客户端解析。
type PayChannelService struct {
	channels *payRepository.PayChannelRepository
}

// NewPayChannelService 构造支付渠道服务。
func NewPayChannelService() *PayChannelService {
	return &PayChannelService{channels: payRepository.NewPayChannelRepository()}
}

// Page 分页查询渠道配置（config 脱敏后返回）。
// 入参 q 为查询条件；返回脱敏后的渠道列表与总数。
func (s *PayChannelService) Page(ctx context.Context, q PayChannelListQuery) ([]payModel.PayChannelModel, int64, error) {
	list, total, err := s.channels.Paginate(ctx, payRepository.PayChannelFilter{
		Keyword: strings.TrimSpace(q.Keyword),
		Code:    strings.TrimSpace(q.Code),
		Status:  q.Status,
	}, q.PageNo, q.PageSize)
	if err != nil {
		return nil, 0, err
	}

	for i := range list {
		list[i].Config = payChannel.MaskConfig(list[i].Code, list[i].Config)
	}

	return list, total, nil
}

// Schema 返回指定渠道编码的动态表单字段定义与元信息，前端据此渲染表单。
// 入参 code 为渠道编码；返回元数据，编码不受支持返回业务错误码。
func (s *PayChannelService) Schema(code string) (payChannel.Meta, error) {
	meta, ok := payChannel.Metas[code]
	if !ok {
		return payChannel.Meta{}, baseErrs.New(payErrs.PayChannelUnsupported, code)
	}

	return meta, nil
}

// NotifyURL 返回系统生成的异步通知地址：APP_URL + 固定路由 + 渠道编码。
// 站内渠道（钱包 / 模拟支付）没有异步通知，直接返回空串。
// 入参 code 为渠道编码；返回完整地址，APP_URL 未配置返回业务错误码。
func (s *PayChannelService) NotifyURL(code string) (string, error) {
	if payChannel.IsInternal(code) {
		return "", nil
	}

	appURL := strings.TrimRight(strings.TrimSpace(facades.Config().GetString("http.url")), "/")
	if appURL == "" || appURL == "http://localhost" {
		return "", baseErrs.New(payErrs.PayAppUrlNotConfigured)
	}

	return appURL + CallbackPath + code, nil
}

// Detail 查询渠道详情（config 脱敏后返回）。
// 入参 id 为渠道 ID；返回渠道模型，不存在返回业务错误码。
func (s *PayChannelService) Detail(ctx context.Context, id types.ID) (*payModel.PayChannelModel, error) {
	m, err := s.channels.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if m == nil {
		return nil, baseErrs.New(payErrs.PayChannelNotExists, id)
	}

	m.Config = payChannel.MaskConfig(m.Code, m.Config)

	return m, nil
}

// Create 新增渠道：编码唯一、必须受支持、config 必填项校验，明文入库。
// 入参 in 为渠道入参；返回新渠道 ID。
func (s *PayChannelService) Create(ctx context.Context, in PayChannelSaveInput) (types.ID, error) {
	in.Name = strings.TrimSpace(in.Name)
	in.Code = strings.TrimSpace(in.Code)
	in.Config = strings.TrimSpace(in.Config)
	if in.Name == "" {
		return 0, baseErrs.WithMsg(baseErrs.BadRequest, "渠道名称不能为空")
	}

	meta, ok := payChannel.Metas[in.Code]
	if !ok {
		return 0, baseErrs.New(payErrs.PayChannelUnsupported, in.Code)
	}
	if err := payChannel.ValidateConfig(in.Code, in.Config); err != nil {
		return 0, err
	}
	// 离线自检：密钥长度、私钥格式、证书与序列号是否配套，尽量在保存时就暴露问题
	if err := payChannel.SelfCheck(in.Code, in.Config); err != nil {
		return 0, err
	}

	exists, err := s.channels.ExistsByCode(ctx, in.Code, 0)
	if err != nil {
		return 0, err
	}
	if exists {
		return 0, baseErrs.New(payErrs.PayChannelCodeExists, in.Code)
	}

	m := &payModel.PayChannelModel{
		Code:      in.Code,
		Name:      in.Name,
		Channel:   meta.Vendor,
		PayMethod: meta.PayMethod,
		Config:    in.Config,
		Status:    baseEnums.CommonStatusEnum(in.Status),
		Sort:      in.Sort,
		Remark:    in.Remark,
	}
	if err := s.channels.Create(ctx, m); err != nil {
		return 0, err
	}

	return m.ID, nil
}

// Update 编辑渠道：编码不可改；config 中 secret 字段留空表示沿用原值。
// 入参 id 为渠道 ID、in 为渠道入参；返回写库错误或业务错误码。
func (s *PayChannelService) Update(ctx context.Context, id types.ID, in PayChannelSaveInput) error {
	m, err := s.channels.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if m == nil {
		return baseErrs.New(payErrs.PayChannelNotExists, id)
	}

	in.Name = strings.TrimSpace(in.Name)
	in.Config = strings.TrimSpace(in.Config)
	if in.Name == "" {
		return baseErrs.WithMsg(baseErrs.BadRequest, "渠道名称不能为空")
	}
	// 先把留空的敏感字段补回库中原值，再校验与自检，避免「留空不修改」被误判为缺失
	merged := payChannel.MergeSecretFields(m.Code, in.Config, m.Config)
	if err := payChannel.ValidateConfig(m.Code, merged); err != nil {
		return err
	}
	if err := payChannel.SelfCheck(m.Code, merged); err != nil {
		return err
	}
	in.Config = merged

	m.Name = in.Name
	m.Config = in.Config
	m.Status = baseEnums.CommonStatusEnum(in.Status)
	m.Sort = in.Sort
	m.Remark = in.Remark

	return s.channels.Update(ctx, m)
}

// SetStatus 启停渠道；停用后 C 端不可选、下单拒绝，回调处理不受影响。
// 入参 id 为渠道 ID、enable 为目标状态；返回写库错误。
func (s *PayChannelService) SetStatus(ctx context.Context, id types.ID, enable bool) error {
	m, err := s.channels.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if m == nil {
		return baseErrs.New(payErrs.PayChannelNotExists, id)
	}

	status := baseEnums.CommonStatusDisabled
	if enable {
		status = baseEnums.CommonStatusEnabled
	}

	return s.channels.UpdateStatus(ctx, id, status, operatorName(ctx))
}

// Delete 物理删除渠道配置。
// 入参 id 为渠道 ID；返回写库错误。
func (s *PayChannelService) Delete(ctx context.Context, id types.ID) error {
	m, err := s.channels.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if m == nil {
		return baseErrs.New(payErrs.PayChannelNotExists, id)
	}

	return s.channels.Delete(ctx, id)
}

// EnabledChannels 返回全部启用渠道（config 置空不外泄），供 C 端选择支付方式。
// 入参 ctx 为请求上下文；返回渠道列表。
func (s *PayChannelService) EnabledChannels(ctx context.Context) ([]payModel.PayChannelModel, error) {
	list, err := s.channels.Enabled(ctx)
	if err != nil {
		return nil, err
	}

	for i := range list {
		list[i].Config = ""
	}

	return list, nil
}

// CheckChannels 对全部启用渠道做一次离线凭据自检，仅记录日志、不返回错误。
// 供启动时巡检使用：配置有问题时先告警，避免用户付款才发现。
// 入参 ctx 为请求上下文。
func (s *PayChannelService) CheckChannels(ctx context.Context) {
	list, err := s.channels.Enabled(ctx)
	if err != nil {
		return
	}

	for _, m := range list {
		if err := payChannel.SelfCheck(m.Code, m.Config); err != nil {
			facades.Log().Warningf("支付渠道【%s】配置自检未通过，请到后台核对：%v", m.Code, err)
		}
	}
}

// Resolve 按渠道编码解析出可用渠道客户端。
// 入参 code 为渠道编码、requireEnabled 为是否要求启用状态（回调处理不要求）；
// 返回渠道模型与客户端，渠道不存在 / 停用 / 配置不合法返回业务错误码。
func (s *PayChannelService) Resolve(ctx context.Context, code string, requireEnabled bool) (*payModel.PayChannelModel, payChannel.Channel, error) {
	var m *payModel.PayChannelModel
	var err error

	if requireEnabled {
		m, err = s.channels.FindEnabledByCode(ctx, code)
	} else {
		m, err = s.channels.FindByCode(ctx, code)
	}
	if err != nil {
		return nil, nil, err
	}
	if m == nil {
		return nil, nil, baseErrs.New(payErrs.PayChannelNotExists, code)
	}
	if requireEnabled && m.Status != baseEnums.CommonStatusEnabled {
		return nil, nil, baseErrs.New(payErrs.PayChannelDisabled, code)
	}

	ch, err := payChannel.Resolve(m.Code, m.Config)
	if err != nil {
		return nil, nil, err
	}

	return m, ch, nil
}

// operatorName 取当前操作人标识（后台账号 ID），未登录场景为空。
// 入参 ctx 为请求上下文；返回字符串形式的用户 ID。
func operatorName(ctx context.Context) string {
	if uid := support.UserID(ctx); uid != 0 {
		return types.ID(uid).String()
	}
	return ""
}
