package controller

import (
	"github.com/goravel/framework/contracts/http"

	baseEnums "eqray-go-starter/app/modules/base/enums"
	"eqray-go-starter/app/modules/base/param"
	"eqray-go-starter/app/modules/base/result"
	"eqray-go-starter/app/modules/member/admin/request"
	"eqray-go-starter/app/modules/member/service"
)

// MemberLevelController 提供会员等级的增删改查接口。
type MemberLevelController struct {
	service *service.MemberLevelService
}

func NewMemberLevelController() *MemberLevelController {
	return &MemberLevelController{service: service.NewMemberLevelService()}
}

// Index GET /api/v1/member/levels
func (c *MemberLevelController) Index(ctx http.Context) http.Response {
	page := param.Page{
		PageNo:   ctx.Request().QueryInt("pageNo", 1),
		PageSize: ctx.Request().QueryInt("pageSize", param.DefaultPageSize),
	}
	if err := page.Normalize(); err != nil {
		return result.Fail(ctx, err)
	}

	q := service.MemberLevelListQuery{
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

// Options GET /api/v1/member/levels/options
func (c *MemberLevelController) Options(ctx http.Context) http.Response {
	list, err := c.service.Options(ctx.Context())
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, list)
}

// Show GET /api/v1/member/levels/{id}
func (c *MemberLevelController) Show(ctx http.Context) http.Response {
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

// Store POST /api/v1/member/levels
func (c *MemberLevelController) Store(ctx http.Context) http.Response {
	var req request.MemberLevelSaveRequest
	if err := param.Bind(ctx, &req); err != nil {
		return result.Fail(ctx, err)
	}

	id, err := c.service.Create(ctx.Context(), service.MemberLevelSaveInput{
		Level:           req.Level,
		Name:            req.Name,
		Icon:            req.Icon,
		Experience:      req.Experience,
		DiscountPercent: req.DiscountPercent,
		Description:     req.Description,
		Status:          req.Status,
		Sort:            req.Sort,
	})
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, map[string]any{"id": id})
}

// Update PUT /api/v1/member/levels/{id}
func (c *MemberLevelController) Update(ctx http.Context) http.Response {
	id, err := param.PathID(ctx)
	if err != nil {
		return result.Fail(ctx, err)
	}

	var req request.MemberLevelSaveRequest
	if err := param.Bind(ctx, &req); err != nil {
		return result.Fail(ctx, err)
	}

	err = c.service.Update(ctx.Context(), id, service.MemberLevelSaveInput{
		Level:           req.Level,
		Name:            req.Name,
		Icon:            req.Icon,
		Experience:      req.Experience,
		DiscountPercent: req.DiscountPercent,
		Description:     req.Description,
		Status:          req.Status,
		Sort:            req.Sort,
	})
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, nil)
}

// Destroy DELETE /api/v1/member/levels/{id}
func (c *MemberLevelController) Destroy(ctx http.Context) http.Response {
	id, err := param.PathID(ctx)
	if err != nil {
		return result.Fail(ctx, err)
	}

	if err := c.service.Delete(ctx.Context(), id); err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, nil)
}
