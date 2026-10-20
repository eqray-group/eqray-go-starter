package controller

import (
	"github.com/goravel/framework/contracts/http"

	baseEnums "eqray-go-starter/app/modules/base/enums"
	"eqray-go-starter/app/modules/base/param"
	"eqray-go-starter/app/modules/base/result"
	"eqray-go-starter/app/modules/member/admin/request"
	"eqray-go-starter/app/modules/member/service"
)

// MemberSignConfigController 提供后台签到奖励档位的维护接口。
type MemberSignConfigController struct {
	service *service.MemberSignConfigService
}

func NewMemberSignConfigController() *MemberSignConfigController {
	return &MemberSignConfigController{service: service.NewMemberSignConfigService()}
}

// Index GET /api/v1/member/sign-configs
// 支持 day（签到天数）与 status（状态）两个可选筛选条件。
func (c *MemberSignConfigController) Index(ctx http.Context) http.Response {
	page, err := param.FromQuery(ctx)
	if err != nil {
		return result.Fail(ctx, err)
	}

	q := service.MemberSignConfigListQuery{
		PageNo:   page.PageNo,
		PageSize: page.PageSize,
	}
	if day := ctx.Request().QueryInt("day", 0); day > 0 {
		q.Day = &day
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

// Store POST /api/v1/member/sign-configs
func (c *MemberSignConfigController) Store(ctx http.Context) http.Response {
	var req request.MemberSignConfigRequest
	if err := param.Bind(ctx, &req); err != nil {
		return result.Fail(ctx, err)
	}

	if err := c.service.Create(ctx.Context(), service.MemberSignConfigInput{
		Day:        req.Day,
		Point:      req.Point,
		Experience: req.Experience,
		Status:     req.Status,
	}); err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, nil)
}

// Update PUT /api/v1/member/sign-configs/{id}
func (c *MemberSignConfigController) Update(ctx http.Context) http.Response {
	id, err := param.PathID(ctx)
	if err != nil {
		return result.Fail(ctx, err)
	}

	var req request.MemberSignConfigRequest
	if err := param.Bind(ctx, &req); err != nil {
		return result.Fail(ctx, err)
	}

	if err := c.service.Update(ctx.Context(), id, service.MemberSignConfigInput{
		Day:        req.Day,
		Point:      req.Point,
		Experience: req.Experience,
		Status:     req.Status,
	}); err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, nil)
}

// Destroy DELETE /api/v1/member/sign-configs/{id}
func (c *MemberSignConfigController) Destroy(ctx http.Context) http.Response {
	id, err := param.PathID(ctx)
	if err != nil {
		return result.Fail(ctx, err)
	}

	if err := c.service.Delete(ctx.Context(), id); err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, nil)
}
