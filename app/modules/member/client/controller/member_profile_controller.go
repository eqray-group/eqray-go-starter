package controller

import (
	"github.com/goravel/framework/contracts/http"

	baseErrs "eqray-go-starter/app/modules/base/errs"
	"eqray-go-starter/app/modules/base/param"
	"eqray-go-starter/app/modules/base/result"
	"eqray-go-starter/app/modules/base/support"
	"eqray-go-starter/app/modules/member/client/request"
	"eqray-go-starter/app/modules/member/service"
)

// MemberProfileController 提供 C 端会员的「我的资料」查询与修改。
type MemberProfileController struct {
	service *service.MemberProfileService
}

func NewMemberProfileController() *MemberProfileController {
	return &MemberProfileController{service: service.NewMemberProfileService()}
}

// Show GET /api/v1/client/member/profile
func (c *MemberProfileController) Show(ctx http.Context) http.Response {
	profile, err := c.service.Profile(ctx.Context(), support.MemberID(ctx.Context()))
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, profile)
}

// Update PUT /api/v1/client/member/profile
func (c *MemberProfileController) Update(ctx http.Context) http.Response {
	var req request.MemberProfileUpdateRequest
	if err := param.Bind(ctx, &req); err != nil {
		return result.Fail(ctx, err)
	}

	birthday, err := req.BirthdayTime()
	if err != nil {
		return result.Fail(ctx, baseErrs.WithMsg(baseErrs.BadRequest, "生日格式应为 2006-01-02"))
	}

	if err := c.service.Update(ctx.Context(), support.MemberID(ctx.Context()), service.MemberProfileUpdateInput{
		Nickname: req.Nickname,
		Avatar:   req.Avatar,
		Name:     req.Name,
		Sex:      req.Sex,
		Birthday: birthday,
		AreaID:   req.AreaID,
	}); err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, nil)
}
