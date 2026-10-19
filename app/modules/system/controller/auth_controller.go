// Package controller 存放 system 模块的控制器，只做取参、调 Service、出响应。
package controller

import (
	"github.com/goravel/framework/contracts/http"

	baseAuth "eqray-go-starter/app/modules/base/auth"
	baseErrs "eqray-go-starter/app/modules/base/errs"
	"eqray-go-starter/app/modules/base/param"
	"eqray-go-starter/app/modules/base/result"
	"eqray-go-starter/app/modules/base/support"
	"eqray-go-starter/app/modules/base/types"
	"eqray-go-starter/app/modules/system/request"
	"eqray-go-starter/app/modules/system/service"
)

// AuthController 提供后台认证接口。
type AuthController struct {
	service *service.AuthService
	users   *service.UserService
}

func NewAuthController() *AuthController {
	return &AuthController{
		service: service.NewAuthService(),
		users:   service.NewUserService(),
	}
}

// Captcha GET /api/v1/system/auth/captcha
func (c *AuthController) Captcha(ctx http.Context) http.Response {
	id, image, err := baseAuth.GenerateCaptcha()
	if err != nil {
		return result.Fail(ctx, baseErrs.WithMsg(baseErrs.InternalServerError, "验证码生成失败"))
	}

	return result.Success(ctx, map[string]any{
		"captchaId": id,
		"image":     image,
	})
}

// Profile GET /api/v1/system/profile
func (c *AuthController) Profile(ctx http.Context) http.Response {
	userID := support.UserID(ctx.Context())
	if userID == 0 {
		return result.Fail(ctx, baseErrs.New(baseErrs.Unauthorized))
	}

	data, err := c.service.Profile(ctx.Context(), types.ID(userID))
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, data)
}

// UpdateProfile PUT /api/v1/system/profile
func (c *AuthController) UpdateProfile(ctx http.Context) http.Response {
	userID := support.UserID(ctx.Context())
	if userID == 0 {
		return result.Fail(ctx, baseErrs.New(baseErrs.Unauthorized))
	}

	var req request.ProfileUpdateRequest
	if err := param.Bind(ctx, &req); err != nil {
		return result.Fail(ctx, err)
	}

	err := c.users.UpdateProfile(ctx.Context(), types.ID(userID), service.ProfileUpdateInput{
		Nickname: req.Nickname,
		Email:    req.Email,
		Mobile:   req.Mobile,
		Sex:      req.Sex,
		Avatar:   req.Avatar,
	})
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, nil)
}

// UpdatePassword PUT /api/v1/system/profile/password
func (c *AuthController) UpdatePassword(ctx http.Context) http.Response {
	userID := support.UserID(ctx.Context())
	if userID == 0 {
		return result.Fail(ctx, baseErrs.New(baseErrs.Unauthorized))
	}

	var req request.ProfilePasswordRequest
	if err := param.Bind(ctx, &req); err != nil {
		return result.Fail(ctx, err)
	}

	if err := c.users.ChangePassword(ctx.Context(), types.ID(userID), req.OldPassword, req.NewPassword); err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, nil)
}

// Login POST /api/v1/system/auth/login
func (c *AuthController) Login(ctx http.Context) http.Response {
	var req request.LoginRequest
	if err := param.Bind(ctx, &req); err != nil {
		return result.Fail(ctx, err)
	}

	res, err := c.service.Login(ctx.Context(), service.LoginInput{
		IP:          ctx.Request().Ip(),
		UserAgent:   ctx.Request().Header("User-Agent"),
		Username:    req.Username,
		Password:    req.Password,
		CaptchaID:   req.CaptchaID,
		CaptchaCode: req.CaptchaCode,
	})
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, res)
}

// Refresh POST /api/v1/system/auth/refresh
func (c *AuthController) Refresh(ctx http.Context) http.Response {
	var req request.RefreshRequest
	if err := param.Bind(ctx, &req); err != nil {
		return result.Fail(ctx, err)
	}

	tokens, err := c.service.Refresh(req.RefreshToken)
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, tokens)
}

// Logout POST /api/v1/system/auth/logout
func (c *AuthController) Logout(ctx http.Context) http.Response {
	c.service.Logout(
		ctx.Context(),
		baseAuth.BearerToken(ctx.Request().Header("Authorization")),
		ctx.Request().Input("refreshToken"),
		ctx.Request().Ip(),
		ctx.Request().Header("User-Agent"),
	)

	return result.Success(ctx, nil)
}
