package controller

import (
	"github.com/goravel/framework/contracts/http"

	baseEnums "eqray-go-starter/app/modules/base/enums"
	"eqray-go-starter/app/modules/base/param"
	"eqray-go-starter/app/modules/base/result"
	"eqray-go-starter/app/modules/member/request"
	"eqray-go-starter/app/modules/member/service"
)

// MemberSocialClientController 提供第三方登录渠道配置的增删改查。
type MemberSocialClientController struct {
	service *service.MemberSocialClientService
}

func NewMemberSocialClientController() *MemberSocialClientController {
	return &MemberSocialClientController{service: service.NewMemberSocialClientService()}
}

// Index GET /api/v1/member/social-clients
func (c *MemberSocialClientController) Index(ctx http.Context) http.Response {
	page := param.Page{
		PageNo:   ctx.Request().QueryInt("pageNo", 1),
		PageSize: ctx.Request().QueryInt("pageSize", param.DefaultPageSize),
	}
	if err := page.Normalize(); err != nil {
		return result.Fail(ctx, err)
	}

	q := service.MemberSocialClientListQuery{
		Keyword:  ctx.Request().Query("keyword"),
		PageNo:   page.PageNo,
		PageSize: page.PageSize,
	}
	if status := ctx.Request().QueryInt("status", -1); status >= 0 {
		v := baseEnums.CommonStatusEnum(status)
		q.Status = &v
	}

	list, total, err := c.service.Page(ctx.Context(), q)
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Page(ctx, list, total)
}

// Show GET /api/v1/member/social-clients/{id}
func (c *MemberSocialClientController) Show(ctx http.Context) http.Response {
	id, err := param.PathID(ctx)
	if err != nil {
		return result.Fail(ctx, err)
	}

	m, err := c.service.Detail(ctx.Context(), id)
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, m)
}

// Store POST /api/v1/member/social-clients
func (c *MemberSocialClientController) Store(ctx http.Context) http.Response {
	var req request.MemberSocialClientSaveRequest
	if err := param.Bind(ctx, &req); err != nil {
		return result.Fail(ctx, err)
	}

	id, err := c.service.Create(ctx.Context(), service.MemberSocialClientSaveInput{
		Name:         req.Name,
		SocialType:   req.SocialType,
		ClientID:     req.ClientID,
		ClientSecret: req.ClientSecret,
		AgentID:      req.AgentID,
		Status:       req.Status,
	})
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, map[string]any{"id": id})
}

// Update PUT /api/v1/member/social-clients/{id}
func (c *MemberSocialClientController) Update(ctx http.Context) http.Response {
	id, err := param.PathID(ctx)
	if err != nil {
		return result.Fail(ctx, err)
	}

	var req request.MemberSocialClientSaveRequest
	if err := param.Bind(ctx, &req); err != nil {
		return result.Fail(ctx, err)
	}

	err = c.service.Update(ctx.Context(), id, service.MemberSocialClientSaveInput{
		Name:         req.Name,
		SocialType:   req.SocialType,
		ClientID:     req.ClientID,
		ClientSecret: req.ClientSecret,
		AgentID:      req.AgentID,
		Status:       req.Status,
	})
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, nil)
}

// Destroy DELETE /api/v1/member/social-clients/{id}
func (c *MemberSocialClientController) Destroy(ctx http.Context) http.Response {
	id, err := param.PathID(ctx)
	if err != nil {
		return result.Fail(ctx, err)
	}

	if err := c.service.Delete(ctx.Context(), id); err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, nil)
}
