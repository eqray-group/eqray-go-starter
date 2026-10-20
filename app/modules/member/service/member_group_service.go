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

// MemberGroupSaveInput 是新增 / 编辑会员分组的入参。
type MemberGroupSaveInput struct {
	Name   string
	Status int8
	Remark string
}

// MemberGroupListQuery 是分组列表的查询条件。
type MemberGroupListQuery struct {
	Keyword  string
	Status   *baseEnums.CommonStatusEnum
	PageNo   int
	PageSize int
}

// MemberGroupService 负责会员分组的增删改查。
type MemberGroupService struct {
	groups *repository.MemberGroupRepository
}

func NewMemberGroupService() *MemberGroupService {
	return &MemberGroupService{groups: repository.NewMemberGroupRepository()}
}

func (s *MemberGroupService) Page(ctx context.Context, q MemberGroupListQuery) ([]model.MemberGroupModel, int64, error) {
	return s.groups.Paginate(ctx, repository.MemberGroupFilter{
		Keyword: strings.TrimSpace(q.Keyword),
		Status:  q.Status,
	}, q.PageNo, q.PageSize)
}

// Options 返回启用中的分组，供会员表单下拉。
func (s *MemberGroupService) Options(ctx context.Context) ([]model.MemberGroupModel, error) {
	return s.groups.Enabled(ctx)
}

func (s *MemberGroupService) Detail(ctx context.Context, id types.ID) (*model.MemberGroupModel, error) {
	m, err := s.groups.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if m == nil {
		return nil, baseErrs.New(memberErrs.MemberGroupNotExists, id)
	}

	return m, nil
}

func (s *MemberGroupService) Create(ctx context.Context, in MemberGroupSaveInput) (types.ID, error) {
	if err := normalizeGroupInput(&in); err != nil {
		return 0, err
	}

	m := &model.MemberGroupModel{
		Name:   in.Name,
		Status: baseEnums.CommonStatusEnum(in.Status),
		Remark: in.Remark,
	}
	if err := s.groups.Create(ctx, m); err != nil {
		return 0, err
	}

	return m.ID, nil
}

func (s *MemberGroupService) Update(ctx context.Context, id types.ID, in MemberGroupSaveInput) error {
	if err := normalizeGroupInput(&in); err != nil {
		return err
	}

	m, err := s.Detail(ctx, id)
	if err != nil {
		return err
	}

	m.Name = in.Name
	m.Status = baseEnums.CommonStatusEnum(in.Status)
	m.Remark = in.Remark

	return s.groups.Update(ctx, m)
}

func (s *MemberGroupService) Delete(ctx context.Context, id types.ID) error {
	m, err := s.Detail(ctx, id)
	if err != nil {
		return err
	}

	return s.groups.Delete(ctx, m)
}

func normalizeGroupInput(in *MemberGroupSaveInput) error {
	in.Name = strings.TrimSpace(in.Name)
	in.Remark = strings.TrimSpace(in.Remark)

	if in.Name == "" {
		return baseErrs.WithMsg(baseErrs.BadRequest, "分组名不能为空")
	}
	if len([]rune(in.Name)) > 30 {
		return baseErrs.WithMsg(baseErrs.BadRequest, "分组名不能超过 30 个字符")
	}
	if !baseEnums.CommonStatusEnum(in.Status).IsValid() {
		return baseErrs.WithMsg(baseErrs.BadRequest, "分组状态取值不合法")
	}

	return nil
}
