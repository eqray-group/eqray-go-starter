// Package repository 是 pay 模块的数据访问层：只做查询与写库，不含业务分支。
package repository

import (
	"context"

	"eqray-go-starter/app/facades"
	baseEnums "eqray-go-starter/app/modules/base/enums"
	baseTypes "eqray-go-starter/app/modules/base/types"
	payModel "eqray-go-starter/app/modules/pay/model"
)

const payChannelTable = payModel.PayChannelTable

// PayChannelFilter 是支付渠道列表的筛选条件。
type PayChannelFilter struct {
	// Keyword 名称模糊匹配
	Keyword string
	// Code 渠道编码精确匹配
	Code string
	// Status 状态过滤，nil 表示全部
	Status *baseEnums.CommonStatusEnum
}

// PayChannelRepository 是 pay_channel 表的仓储。
// 渠道属配置数据：写操作走物理删除，避免 uk_code 被软删记录占用。
type PayChannelRepository struct{}

// NewPayChannelRepository 构造支付渠道仓储。
func NewPayChannelRepository() *PayChannelRepository {
	return &PayChannelRepository{}
}

// Create 插入渠道配置（幂等性由 uk_code 唯一键保证，冲突转业务错误码）。
// 入参 m 为待插入模型；返回冲突或写库错误。
func (r *PayChannelRepository) Create(ctx context.Context, m *payModel.PayChannelModel) error {
	m.FillCreate(ctx)

	_, err := facades.DB().WithContext(ctx).Table(payChannelTable).Insert(map[string]any{
		"id":         m.ID,
		"code":       m.Code,
		"name":       m.Name,
		"channel":    m.Channel,
		"pay_method": m.PayMethod,
		"config":     m.Config,
		"status":     m.Status,
		"sort":       m.Sort,
		"remark":     m.Remark,
		"creator":    m.Creator,
		"created_at": m.CreatedAt,
		"updater":    m.Updater,
		"updated_at": m.UpdatedAt,
	})

	return err
}

// Update 更新渠道配置字段（编码不可改，编码唯一由新增时校验）。
// 入参 m 为携带新值的模型；返回写库错误。
func (r *PayChannelRepository) Update(ctx context.Context, m *payModel.PayChannelModel) error {
	m.FillUpdate(ctx)

	_, err := facades.DB().WithContext(ctx).Table(payChannelTable).
		Where("id", m.ID).
		Update(map[string]any{
			"name":       m.Name,
			"config":     m.Config,
			"status":     m.Status,
			"sort":       m.Sort,
			"remark":     m.Remark,
			"updater":    m.Updater,
			"updated_at": m.UpdatedAt,
		})

	return err
}

// UpdateStatus 只更新启用状态，供启停接口使用。
// 入参 id 为渠道 ID、status 为目标状态、updater 为操作人；返回写库错误。
func (r *PayChannelRepository) UpdateStatus(ctx context.Context, id baseTypes.ID, status baseEnums.CommonStatusEnum, updater string) error {
	_, err := facades.DB().WithContext(ctx).Table(payChannelTable).
		Where("id", id).
		Update(map[string]any{
			"status":     status,
			"updater":    updater,
			"updated_at": baseTypes.Now(),
		})

	return err
}

// Delete 物理删除渠道配置；配置数据不软删，避免编码被占用。
// 入参 id 为渠道 ID；返回写库错误。
func (r *PayChannelRepository) Delete(ctx context.Context, id baseTypes.ID) error {
	_, err := facades.DB().WithContext(ctx).Table(payChannelTable).
		Where("id", id).
		Delete()

	return err
}

// FindByID 按主键查询渠道；未命中返回 (nil, nil)。
// 入参 id 为渠道 ID；返回模型或查询错误。
func (r *PayChannelRepository) FindByID(ctx context.Context, id baseTypes.ID) (*payModel.PayChannelModel, error) {
	var m payModel.PayChannelModel

	if err := facades.DB().WithContext(ctx).Table(payChannelTable).
		Where("id", id).
		First(&m); err != nil {
		return nil, err
	}
	if m.ID == 0 {
		return nil, nil
	}

	return &m, nil
}

// FindByCode 按渠道编码查询（回调处理与下单解析都用编码定位）。
// 入参 code 为渠道编码；返回模型，未命中 (nil, nil)。
func (r *PayChannelRepository) FindByCode(ctx context.Context, code string) (*payModel.PayChannelModel, error) {
	var m payModel.PayChannelModel

	if err := facades.DB().WithContext(ctx).Table(payChannelTable).
		Where("code", code).
		First(&m); err != nil {
		return nil, err
	}
	if m.ID == 0 {
		return nil, nil
	}

	return &m, nil
}

// FindEnabledByCode 查询启用中的渠道，供下单解析使用；停用渠道不允许收款。
// 入参 code 为渠道编码；返回模型，未命中或已停用 (nil, nil)。
func (r *PayChannelRepository) FindEnabledByCode(ctx context.Context, code string) (*payModel.PayChannelModel, error) {
	var m payModel.PayChannelModel

	if err := facades.DB().WithContext(ctx).Table(payChannelTable).
		Where("code", code).
		Where("status", baseEnums.CommonStatusEnabled).
		First(&m); err != nil {
		return nil, err
	}
	if m.ID == 0 {
		return nil, nil
	}

	return &m, nil
}

// Enabled 返回全部启用渠道，按 sort 升序，供 C 端选择支付方式。
// 入参 ctx 为请求上下文；返回启用渠道列表。
func (r *PayChannelRepository) Enabled(ctx context.Context) ([]payModel.PayChannelModel, error) {
	list := make([]payModel.PayChannelModel, 0)

	if err := facades.DB().WithContext(ctx).Table(payChannelTable).
		Where("status", baseEnums.CommonStatusEnabled).
		OrderBy("sort").
		OrderBy("id").
		Get(&list); err != nil {
		return nil, err
	}

	return list, nil
}

// ExistsByCode 判断渠道编码是否已被占用，excludeID 用于编辑时排除自身。
// 入参 code 为渠道编码、excludeID 为排除的渠道 ID；返回是否占用与查询错误。
func (r *PayChannelRepository) ExistsByCode(ctx context.Context, code string, excludeID baseTypes.ID) (bool, error) {
	query := facades.DB().WithContext(ctx).Table(payChannelTable).
		Where("code", code)
	if excludeID > 0 {
		query = query.Where("id <> ?", excludeID)
	}

	total, err := query.Count()

	return total > 0, err
}

// Paginate 分页查询渠道配置。
// 入参 filter 为筛选条件、pageNo/pageSize 为分页参数；返回列表与总数。
func (r *PayChannelRepository) Paginate(ctx context.Context, filter PayChannelFilter, pageNo, pageSize int) ([]payModel.PayChannelModel, int64, error) {
	list := make([]payModel.PayChannelModel, 0)
	var total int64

	query := facades.DB().WithContext(ctx).Table(payChannelTable)
	if filter.Keyword != "" {
		query = query.WhereLike("name", "%"+filter.Keyword+"%")
	}
	if filter.Code != "" {
		query = query.Where("code", filter.Code)
	}
	if filter.Status != nil {
		query = query.Where("status", *filter.Status)
	}

	if err := query.OrderBy("sort").OrderByDesc("id").Paginate(pageNo, pageSize, &list, &total); err != nil {
		return nil, 0, err
	}

	return list, total, nil
}
