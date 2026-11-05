package repository

import (
	"context"
	"strings"
	"time"

	"github.com/goravel/framework/contracts/database/db"

	"eqray-go-starter/app/facades"
	"eqray-go-starter/app/modules/base/types"
	"eqray-go-starter/app/modules/infra/model"
)

// NoticeChannelFilter 是渠道列表筛选条件。
type NoticeChannelFilter struct {
	// Keyword 名称模糊匹配
	Keyword string
	// Type 渠道类型，空表示全部
	Type string
	// Status 状态，nil 表示全部
	Status *int8
}

// NoticeTemplateFilter 是模板列表筛选条件。
type NoticeTemplateFilter struct {
	// Keyword 编码 / 名称模糊匹配
	Keyword string
	// Type 渠道类型，空表示全部
	Type string
	// Status 状态，nil 表示全部
	Status *int8
}

// NoticeLogFilter 是发送记录筛选条件。
type NoticeLogFilter struct {
	// Type 渠道类型，空表示全部
	Type string
	// Status 结果，nil 表示全部
	Status *int8
}

// InfraNoticeRepository 是通知渠道 / 模板 / 记录三张表的仓储。
type InfraNoticeRepository struct{}

// NewInfraNoticeRepository 构造通知仓储。
func NewInfraNoticeRepository() *InfraNoticeRepository {
	return &InfraNoticeRepository{}
}

// PaginateChannels 分页查询渠道。
// 入参 filter 为筛选条件、pageNo/pageSize 为分页；返回列表与总数。
func (r *InfraNoticeRepository) PaginateChannels(ctx context.Context, filter NoticeChannelFilter, pageNo, pageSize int) ([]model.InfraNoticeChannelModel, int64, error) {
	apply := func(query db.Query) db.Query {
		if keyword := strings.TrimSpace(filter.Keyword); keyword != "" {
			query = query.Where("name", "like", "%"+keyword+"%")
		}
		if filter.Type != "" {
			query = query.Where("type", filter.Type)
		}
		if filter.Status != nil {
			query = query.Where("status", *filter.Status)
		}

		return query
	}

	total, err := apply(facades.DB().WithContext(ctx).Table(model.InfraNoticeChannelTable)).Count()
	if err != nil {
		return nil, 0, err
	}

	list := make([]model.InfraNoticeChannelModel, 0)
	if total == 0 {
		return list, 0, nil
	}

	if err := apply(facades.DB().WithContext(ctx).Table(model.InfraNoticeChannelTable)).
		OrderBy("id desc").
		Offset(uint64((pageNo - 1) * pageSize)).
		Limit(uint64(pageSize)).
		Get(&list); err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

// FindChannelByID 按主键查渠道；未命中返回 (nil, nil)。
func (r *InfraNoticeRepository) FindChannelByID(ctx context.Context, id types.ID) (*model.InfraNoticeChannelModel, error) {
	m := &model.InfraNoticeChannelModel{}
	if err := facades.DB().WithContext(ctx).Table(model.InfraNoticeChannelTable).Where("id", id).First(m); err != nil {
		return nil, nil
	}

	return m, nil
}

// FirstEnabledChannel 取指定类型下第一个启用渠道；没有返回 (nil, nil)。
func (r *InfraNoticeRepository) FirstEnabledChannel(ctx context.Context, noticeType string) (*model.InfraNoticeChannelModel, error) {
	m := &model.InfraNoticeChannelModel{}
	if err := facades.DB().WithContext(ctx).Table(model.InfraNoticeChannelTable).
		Where("type", noticeType).
		Where("status", 1).
		OrderBy("id").
		First(m); err != nil {
		return nil, nil
	}

	return m, nil
}

// InsertChannel 新增渠道。
func (r *InfraNoticeRepository) InsertChannel(ctx context.Context, m *model.InfraNoticeChannelModel) error {
	_, err := facades.DB().WithContext(ctx).Table(model.InfraNoticeChannelTable).Insert(map[string]any{
		"id":         m.ID,
		"name":       m.Name,
		"type":       m.Type,
		"config":     m.Config,
		"status":     m.Status,
		"remark":     m.Remark,
		"creator":    m.Creator,
		"created_at": m.CreatedAt,
		"updater":    m.Updater,
		"updated_at": m.UpdatedAt,
	})

	return err
}

// UpdateChannel 更新渠道字段。
func (r *InfraNoticeRepository) UpdateChannel(ctx context.Context, id types.ID, data map[string]any, updater string) error {
	data["updater"] = updater
	data["updated_at"] = types.Now()

	_, err := facades.DB().WithContext(ctx).Table(model.InfraNoticeChannelTable).Where("id", id).Update(data)

	return err
}

// DeleteChannel 删除渠道。
func (r *InfraNoticeRepository) DeleteChannel(ctx context.Context, id types.ID) error {
	_, err := facades.DB().WithContext(ctx).Table(model.InfraNoticeChannelTable).Where("id", id).Delete()

	return err
}

// PaginateTemplates 分页查询模板。
func (r *InfraNoticeRepository) PaginateTemplates(ctx context.Context, filter NoticeTemplateFilter, pageNo, pageSize int) ([]model.InfraNoticeTemplateModel, int64, error) {
	apply := func(query db.Query) db.Query {
		if keyword := strings.TrimSpace(filter.Keyword); keyword != "" {
			query = query.Where("CONCAT(code, ' ', name) LIKE ?", "%"+keyword+"%")
		}
		if filter.Type != "" {
			query = query.Where("type", filter.Type)
		}
		if filter.Status != nil {
			query = query.Where("status", *filter.Status)
		}

		return query
	}

	total, err := apply(facades.DB().WithContext(ctx).Table(model.InfraNoticeTemplateTable)).Count()
	if err != nil {
		return nil, 0, err
	}

	list := make([]model.InfraNoticeTemplateModel, 0)
	if total == 0 {
		return list, 0, nil
	}

	if err := apply(facades.DB().WithContext(ctx).Table(model.InfraNoticeTemplateTable)).
		OrderBy("id desc").
		Offset(uint64((pageNo - 1) * pageSize)).
		Limit(uint64(pageSize)).
		Get(&list); err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

// FindTemplateByID 按主键查模板。
func (r *InfraNoticeRepository) FindTemplateByID(ctx context.Context, id types.ID) (*model.InfraNoticeTemplateModel, error) {
	m := &model.InfraNoticeTemplateModel{}
	if err := facades.DB().WithContext(ctx).Table(model.InfraNoticeTemplateTable).Where("id", id).First(m); err != nil {
		return nil, nil
	}

	return m, nil
}

// FindTemplateByCode 按编码查模板。
func (r *InfraNoticeRepository) FindTemplateByCode(ctx context.Context, code string) (*model.InfraNoticeTemplateModel, error) {
	m := &model.InfraNoticeTemplateModel{}
	if err := facades.DB().WithContext(ctx).Table(model.InfraNoticeTemplateTable).Where("code", code).First(m); err != nil {
		return nil, nil
	}

	return m, nil
}

// InsertTemplate 新增模板。
func (r *InfraNoticeRepository) InsertTemplate(ctx context.Context, m *model.InfraNoticeTemplateModel) error {
	_, err := facades.DB().WithContext(ctx).Table(model.InfraNoticeTemplateTable).Insert(map[string]any{
		"id":         m.ID,
		"code":       m.Code,
		"name":       m.Name,
		"type":       m.Type,
		"title":      m.Title,
		"content":    m.Content,
		"status":     m.Status,
		"remark":     m.Remark,
		"creator":    m.Creator,
		"created_at": m.CreatedAt,
		"updater":    m.Updater,
		"updated_at": m.UpdatedAt,
	})

	return err
}

// UpdateTemplate 更新模板字段。
func (r *InfraNoticeRepository) UpdateTemplate(ctx context.Context, id types.ID, data map[string]any, updater string) error {
	data["updater"] = updater
	data["updated_at"] = types.Now()

	_, err := facades.DB().WithContext(ctx).Table(model.InfraNoticeTemplateTable).Where("id", id).Update(data)

	return err
}

// DeleteTemplate 删除模板。
func (r *InfraNoticeRepository) DeleteTemplate(ctx context.Context, id types.ID) error {
	_, err := facades.DB().WithContext(ctx).Table(model.InfraNoticeTemplateTable).Where("id", id).Delete()

	return err
}

// InsertLog 写发送记录。
func (r *InfraNoticeRepository) InsertLog(ctx context.Context, m *model.InfraNoticeLogModel) error {
	_, err := facades.DB().WithContext(ctx).Table(model.InfraNoticeLogTable).Insert(map[string]any{
		"id":            m.ID,
		"channel_id":    m.ChannelID,
		"channel_name":  m.ChannelName,
		"type":          m.Type,
		"template_code": m.TemplateCode,
		"target":        m.Target,
		"title":         m.Title,
		"content":       m.Content,
		"status":        m.Status,
		"error":         m.Error,
		"duration":      m.Duration,
		"creator":       m.Creator,
		"created_at":    m.CreatedAt,
	})

	return err
}

// PaginateLogs 分页查询发送记录，按时间倒序。
func (r *InfraNoticeRepository) PaginateLogs(ctx context.Context, filter NoticeLogFilter, pageNo, pageSize int) ([]model.InfraNoticeLogModel, int64, error) {
	apply := func(query db.Query) db.Query {
		if filter.Type != "" {
			query = query.Where("type", filter.Type)
		}
		if filter.Status != nil {
			query = query.Where("status", *filter.Status)
		}

		return query
	}

	total, err := apply(facades.DB().WithContext(ctx).Table(model.InfraNoticeLogTable)).Count()
	if err != nil {
		return nil, 0, err
	}

	list := make([]model.InfraNoticeLogModel, 0)
	if total == 0 {
		return list, 0, nil
	}

	if err := apply(facades.DB().WithContext(ctx).Table(model.InfraNoticeLogTable)).
		OrderBy("created_at desc").
		Offset(uint64((pageNo - 1) * pageSize)).
		Limit(uint64(pageSize)).
		Get(&list); err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

// PurgeLogsBefore 清理早于指定时间的发送记录。
func (r *InfraNoticeRepository) PurgeLogsBefore(ctx context.Context, before time.Time) (int64, error) {
	res, err := facades.DB().WithContext(ctx).Table(model.InfraNoticeLogTable).
		Where("created_at", "<", before).
		Delete()
	if err != nil {
		return 0, err
	}

	return res.RowsAffected, nil
}
