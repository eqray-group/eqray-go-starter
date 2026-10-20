package service

import (
	"context"
	"strings"

	baseEnums "eqray-go-starter/app/modules/base/enums"
	baseErrs "eqray-go-starter/app/modules/base/errs"
	"eqray-go-starter/app/modules/base/types"
	memberErrs "eqray-go-starter/app/modules/member/errs"
	"eqray-go-starter/app/modules/member/model"
	"eqray-go-starter/app/modules/member/repository"
)

// MemberTagSaveInput 是新增 / 编辑会员标签的入参。
type MemberTagSaveInput struct {
	Name   string
	Status int8
	Remark string
}

// MemberTagListQuery 是标签列表的查询条件。
type MemberTagListQuery struct {
	Keyword  string
	Status   *baseEnums.CommonStatusEnum
	PageNo   int
	PageSize int
}

// MemberTagService 负责会员标签的增删改查。
type MemberTagService struct {
	tags *repository.MemberTagRepository
}

func NewMemberTagService() *MemberTagService {
	return &MemberTagService{tags: repository.NewMemberTagRepository()}
}

func (s *MemberTagService) Page(ctx context.Context, q MemberTagListQuery) ([]model.MemberTagModel, int64, error) {
	return s.tags.Paginate(ctx, repository.MemberTagFilter{
		Keyword: strings.TrimSpace(q.Keyword),
		Status:  q.Status,
	}, q.PageNo, q.PageSize)
}

// Options 返回启用中的标签，供会员表单多选。
func (s *MemberTagService) Options(ctx context.Context) ([]model.MemberTagModel, error) {
	return s.tags.Enabled(ctx)
}

func (s *MemberTagService) Detail(ctx context.Context, id types.ID) (*model.MemberTagModel, error) {
	m, err := s.tags.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if m == nil {
		return nil, baseErrs.New(memberErrs.MemberTagNotExists, id)
	}

	return m, nil
}

func (s *MemberTagService) Create(ctx context.Context, in MemberTagSaveInput) (types.ID, error) {
	if err := normalizeTagInput(&in); err != nil {
		return 0, err
	}

	m := &model.MemberTagModel{
		Name:   in.Name,
		Status: baseEnums.CommonStatusEnum(in.Status),
		Remark: in.Remark,
	}
	if err := s.tags.Create(ctx, m); err != nil {
		return 0, err
	}

	return m.ID, nil
}

func (s *MemberTagService) Update(ctx context.Context, id types.ID, in MemberTagSaveInput) error {
	if err := normalizeTagInput(&in); err != nil {
		return err
	}

	m, err := s.Detail(ctx, id)
	if err != nil {
		return err
	}

	m.Name = in.Name
	m.Status = baseEnums.CommonStatusEnum(in.Status)
	m.Remark = in.Remark

	return s.tags.Update(ctx, m)
}

func (s *MemberTagService) Delete(ctx context.Context, id types.ID) error {
	m, err := s.Detail(ctx, id)
	if err != nil {
		return err
	}

	return s.tags.Delete(ctx, m)
}

func normalizeTagInput(in *MemberTagSaveInput) error {
	in.Name = strings.TrimSpace(in.Name)
	in.Remark = strings.TrimSpace(in.Remark)

	if in.Name == "" {
		return baseErrs.WithMsg(baseErrs.BadRequest, "标签名不能为空")
	}
	if len([]rune(in.Name)) > 30 {
		return baseErrs.WithMsg(baseErrs.BadRequest, "标签名不能超过 30 个字符")
	}
	if !baseEnums.CommonStatusEnum(in.Status).IsValid() {
		return baseErrs.WithMsg(baseErrs.BadRequest, "标签状态取值不合法")
	}

	return nil
}
