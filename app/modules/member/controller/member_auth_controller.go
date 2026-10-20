package controller

import (
	"github.com/goravel/framework/contracts/http"

	baseAuth "eqray-go-starter/app/modules/base/auth"
	baseErrs "eqray-go-starter/app/modules/base/errs"
	"eqray-go-starter/app/modules/base/param"
	"eqray-go-starter/app/modules/base/result"
	"eqray-go-starter/app/modules/base/support"
	"eqray-go-starter/app/modules/base/types"
	"eqray-go-starter/app/modules/member/request"
	"eqray-go-starter/app/modules/member/service"
)

// MemberAuthController 提供 C 端会员的第三方登录、手机号验证码登录与登录态接口。
type MemberAuthController struct {
	service *service.MemberSocialService
	auth    *service.MemberAuthService
}

func NewMemberAuthController() *MemberAuthController {
	return &MemberAuthController{
		service: service.NewMemberSocialService(),
		auth:    service.NewMemberAuthService(),
	}
}

// SmsCode POST /api/v1/member/auth/sms-code
func (c *MemberAuthController) SmsCode(ctx http.Context) http.Response {
	var req request.MemberSmsCodeRequest
	if err := param.Bind(ctx, &req); err != nil {
		return result.Fail(ctx, err)
	}

	if err := c.auth.SendSmsCode(ctx.Context(), req.Mobile, req.Scene); err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, nil)
}

// SmsLogin POST /api/v1/member/auth/sms-login
func (c *MemberAuthController) SmsLogin(ctx http.Context) http.Response {
	var req request.MemberSmsLoginRequest
	if err := param.Bind(ctx, &req); err != nil {
		return result.Fail(ctx, err)
	}

	res, err := c.auth.SmsLogin(ctx.Context(), service.SmsLoginInput{
		Mobile:    req.Mobile,
		Code:      req.Code,
		IP:        ctx.Request().Ip(),
		UserAgent: ctx.Request().Header("User-Agent"),
	})
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, res)
}

// SocialBindLogin POST /api/v1/member/auth/social-bind-login
func (c *MemberAuthController) SocialBindLogin(ctx http.Context) http.Response {
	var req request.MemberSocialBindLoginRequest
	if err := param.Bind(ctx, &req); err != nil {
		return result.Fail(ctx, err)
	}

	res, err := c.auth.SocialBindLogin(ctx.Context(), service.SocialBindLoginInput{
		Mobile:      req.Mobile,
		Code:        req.Code,
		SocialType:  req.SocialType,
		SocialCode:  req.SocialCode,
		SocialState: req.SocialState,
		IP:          ctx.Request().Ip(),
		UserAgent:   ctx.Request().Header("User-Agent"),
	})
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, res)
}

// AuthorizeURL GET /api/v1/member/auth/social/{type}/authorize-url
// 公众号 / 网站扫码返回跳转地址；小程序返回空串，由前端 wx.login 取 code。
func (c *MemberAuthController) AuthorizeURL(ctx http.Context) http.Response {
	url, err := c.service.AuthorizeURL(
		ctx.Context(),
		ctx.Request().Route("type"),
		ctx.Request().Query("redirectUri"),
		ctx.Request().Query("state"),
	)
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, map[string]any{"url": url})
}

// SocialLogin POST /api/v1/member/auth/social/login
func (c *MemberAuthController) SocialLogin(ctx http.Context) http.Response {
	var req request.MemberSocialLoginRequest
	if err := param.Bind(ctx, &req); err != nil {
		return result.Fail(ctx, err)
	}

	res, err := c.service.Login(ctx.Context(), service.SocialLoginInput{
		Type:      req.Type,
		Code:      req.Code,
		IP:        ctx.Request().Ip(),
		UserAgent: ctx.Request().Header("User-Agent"),
	})
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, res)
}

// Refresh POST /api/v1/member/auth/refresh
func (c *MemberAuthController) Refresh(ctx http.Context) http.Response {
	var req request.MemberRefreshRequest
	if err := param.Bind(ctx, &req); err != nil {
		return result.Fail(ctx, err)
	}

	tokens, err := c.service.Refresh(req.RefreshToken)
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, tokens)
}

// Logout POST /api/v1/member/auth/logout
func (c *MemberAuthController) Logout(ctx http.Context) http.Response {
	var req request.MemberLogoutRequest
	if err := param.Bind(ctx, &req); err != nil {
		return result.Fail(ctx, err)
	}

	c.service.Logout(
		baseAuth.BearerToken(ctx.Request().Header("Authorization")),
		req.RefreshToken,
	)

	return result.Success(ctx, nil)
}

// Me GET /api/v1/member/auth/me
func (c *MemberAuthController) Me(ctx http.Context) http.Response {
	memberID := support.MemberID(ctx.Context())
	if memberID == 0 {
		return result.Fail(ctx, baseErrs.New(baseErrs.Unauthorized))
	}

	member, err := c.service.Me(ctx.Context(), types.ID(memberID))
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, member)
}

// Bind POST /api/v1/member/auth/social/bind
func (c *MemberAuthController) Bind(ctx http.Context) http.Response {
	memberID := support.MemberID(ctx.Context())
	if memberID == 0 {
		return result.Fail(ctx, baseErrs.New(baseErrs.Unauthorized))
	}

	var req request.MemberSocialBindRequest
	if err := param.Bind(ctx, &req); err != nil {
		return result.Fail(ctx, err)
	}

	err := c.service.Bind(ctx.Context(), service.SocialBindInput{
		MemberID: types.ID(memberID),
		Type:     req.Type,
		Code:     req.Code,
	})
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, nil)
}

// Bindings GET /api/v1/member/auth/social/bindings
func (c *MemberAuthController) Bindings(ctx http.Context) http.Response {
	memberID := support.MemberID(ctx.Context())
	if memberID == 0 {
		return result.Fail(ctx, baseErrs.New(baseErrs.Unauthorized))
	}

	list, err := c.service.Bindings(ctx.Context(), types.ID(memberID))
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, list)
}

// Unbind DELETE /api/v1/member/auth/social/bindings/{id}
func (c *MemberAuthController) Unbind(ctx http.Context) http.Response {
	memberID := support.MemberID(ctx.Context())
	if memberID == 0 {
		return result.Fail(ctx, baseErrs.New(baseErrs.Unauthorized))
	}

	id, err := param.PathID(ctx)
	if err != nil {
		return result.Fail(ctx, err)
	}

	if err := c.service.Unbind(ctx.Context(), types.ID(memberID), id); err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, nil)
}
