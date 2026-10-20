package controller

import (
	"github.com/goravel/framework/contracts/http"

	"eqray-go-starter/app/modules/base/param"
	"eqray-go-starter/app/modules/base/result"
	"eqray-go-starter/app/modules/base/support"
	"eqray-go-starter/app/modules/member/client/request"
	"eqray-go-starter/app/modules/member/service"
)

// MemberAddressController 提供 C 端会员自助维护收货地址的接口。
// 所有操作的会员身份都取自令牌，地址归属由 Service 用 user_id 过滤。
type MemberAddressController struct {
	service *service.MemberAddressService
}

func NewMemberAddressController() *MemberAddressController {
	return &MemberAddressController{service: service.NewMemberAddressService()}
}

// Index GET /api/v1/client/member/addresses
func (c *MemberAddressController) Index(ctx http.Context) http.Response {
	list, err := c.service.List(ctx.Context(), support.MemberID(ctx.Context()))
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, list)
}

// Store POST /api/v1/client/member/addresses
func (c *MemberAddressController) Store(ctx http.Context) http.Response {
	var req request.MemberAddressRequest
	if err := param.Bind(ctx, &req); err != nil {
		return result.Fail(ctx, err)
	}

	id, err := c.service.Create(ctx.Context(), support.MemberID(ctx.Context()), service.MemberAddressInput{
		Name:          req.Name,
		Mobile:        req.Mobile,
		AreaID:        req.AreaID,
		DetailAddress: req.DetailAddress,
		DefaultStatus: req.DefaultStatus,
	})
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, map[string]any{"id": id})
}

// Update PUT /api/v1/client/member/addresses/{id}
func (c *MemberAddressController) Update(ctx http.Context) http.Response {
	id, err := param.PathID(ctx)
	if err != nil {
		return result.Fail(ctx, err)
	}

	var req request.MemberAddressRequest
	if err := param.Bind(ctx, &req); err != nil {
		return result.Fail(ctx, err)
	}

	if err := c.service.Update(ctx.Context(), support.MemberID(ctx.Context()), id, service.MemberAddressInput{
		Name:          req.Name,
		Mobile:        req.Mobile,
		AreaID:        req.AreaID,
		DetailAddress: req.DetailAddress,
	}); err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, nil)
}

// Destroy DELETE /api/v1/client/member/addresses/{id}
func (c *MemberAddressController) Destroy(ctx http.Context) http.Response {
	id, err := param.PathID(ctx)
	if err != nil {
		return result.Fail(ctx, err)
	}

	if err := c.service.Delete(ctx.Context(), support.MemberID(ctx.Context()), id); err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, nil)
}

// SetDefault PUT /api/v1/client/member/addresses/{id}/default
func (c *MemberAddressController) SetDefault(ctx http.Context) http.Response {
	id, err := param.PathID(ctx)
	if err != nil {
		return result.Fail(ctx, err)
	}

	if err := c.service.SetDefault(ctx.Context(), support.MemberID(ctx.Context()), id); err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, nil)
}
