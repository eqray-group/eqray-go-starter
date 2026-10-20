package repository

import (
	"context"

	"eqray-go-starter/app/facades"
	baseEnums "eqray-go-starter/app/modules/base/enums"
	baseErrs "eqray-go-starter/app/modules/base/errs"
	"eqray-go-starter/app/modules/base/types"
	memberErrs "eqray-go-starter/app/modules/member/errs"
	"eqray-go-starter/app/modules/member/model"
)

const tagTable = model.MemberTagTable

// MemberTagFilter 是标签列表的筛选条件。
type MemberTagFilter struct {
	Keyword string
	Status  *baseEnums.CommonStatusEnum
}

// MemberTagRepository 是 member_tag 表的仓储。
type MemberTagRepository struct{}

func NewMemberTagRepository() *MemberTagRepository {
	return &MemberTagRepository{}
}

func (r *MemberTagRepository) Create(ctx context.Context, m *model.MemberTagModel) error {
	m.FillCreate(ctx)

	_, err := facades.DB().WithContext(ctx).Table(tagTable).Insert(map[string]any{
		"id":         m.ID,
		"name":       m.Name,
		"status":     m.Status,
		"remark":     m.Remark,
		"creator":    m.Creator,
		"created_at": m.CreatedAt,
		"updater":    m.Updater,
		"updated_at": m.UpdatedAt,
	})

	return err
}

func (r *MemberTagRepository) Update(ctx context.Context, m *model.MemberTagModel) error {
	m.FillUpdate(ctx)

	_, err := facades.DB().WithContext(ctx).Table(tagTable).
		Where("id", m.ID).
		WhereNull("deleted_at").
		Update(map[string]any{
			"name":       m.Name,
			"status":     m.Status,
			"remark":     m.Remark,
			"updater":    m.Updater,
			"updated_at": m.UpdatedAt,
		})

	return err
}

func (r *MemberTagRepository) Delete(ctx context.Context, m *model.MemberTagModel) error {
	m.FillUpdate(ctx)

	res, err := facades.DB().WithContext(ctx).Table(tagTable).
		Where("id", m.ID).
		WhereNull("deleted_at").
		Update(map[string]any{
			"deleted_at": types.Now(),
			"updater":    m.Updater,
			"updated_at": m.UpdatedAt,
		})
	if err != nil {
		return err
	}
	if res.RowsAffected == 0 {
		return baseErrs.New(memberErrs.MemberTagNotExists, m.ID)
	}

	return nil
}

func (r *MemberTagRepository) FindByID(ctx context.Context, id types.ID) (*model.MemberTagModel, error) {
	var m model.MemberTagModel

	if err := facades.DB().WithContext(ctx).Table(tagTable).
		Where("id", id).
		WhereNull("deleted_at").
		First(&m); err != nil {
		return nil, err
	}
	if m.ID == 0 {
		return nil, nil
	}

	return &m, nil
}

func (r *MemberTagRepository) Enabled(ctx context.Context) ([]model.MemberTagModel, error) {
	list := make([]model.MemberTagModel, 0)

	if err := facades.DB().WithContext(ctx).Table(tagTable).
		Where("status", baseEnums.CommonStatusEnabled).
		WhereNull("deleted_at").
		OrderBy("id").
		Get(&list); err != nil {
		return nil, err
	}

	return list, nil
}

func (r *MemberTagRepository) Paginate(ctx context.Context, filter MemberTagFilter, pageNo, pageSize int) ([]model.MemberTagModel, int64, error) {
	list := make([]model.MemberTagModel, 0)
	var total int64

	query := facades.DB().WithContext(ctx).Table(tagTable).WhereNull("deleted_at")
	if filter.Keyword != "" {
		query = query.WhereLike("name", "%"+filter.Keyword+"%")
	}
	if filter.Status != nil {
		query = query.Where("status", *filter.Status)
	}

	if err := query.OrderByDesc("id").Paginate(pageNo, pageSize, &list, &total); err != nil {
		return nil, 0, err
	}

	return list, total, nil
}
