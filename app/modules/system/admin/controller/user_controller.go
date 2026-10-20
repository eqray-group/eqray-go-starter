package controller

import (
	"github.com/goravel/framework/contracts/http"

	baseEnums "eqray-go-starter/app/modules/base/enums"
	"eqray-go-starter/app/modules/base/param"
	"eqray-go-starter/app/modules/base/result"
	"eqray-go-starter/app/modules/system/admin/request"
	"eqray-go-starter/app/modules/system/service"
)

// UserController 提供 system_users 的 CRUD 接口。
type UserController struct {
	service *service.UserService
}

func NewUserController() *UserController {
	return &UserController{service: service.NewUserService()}
}

// Index GET /api/v1/system/users
func (c *UserController) Index(ctx http.Context) http.Response {
	page := param.Page{
		PageNo:   ctx.Request().QueryInt("pageNo", 1),
		PageSize: ctx.Request().QueryInt("pageSize", param.DefaultPageSize),
	}
	if err := page.Normalize(); err != nil {
		return result.Fail(ctx, err)
	}

	q := service.UserListQuery{
		Username: ctx.Request().Query("username"),
		Nickname: ctx.Request().Query("nickname"),
		Mobile:   ctx.Request().Query("mobile"),
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

// Show GET /api/v1/system/users/{id}
func (c *UserController) Show(ctx http.Context) http.Response {
	id, err := param.PathID(ctx)
	if err != nil {
		return result.Fail(ctx, err)
	}

	user, err := c.service.Detail(ctx.Context(), id)
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, user)
}

// Store POST /api/v1/system/users
func (c *UserController) Store(ctx http.Context) http.Response {
	var req request.UserCreateRequest
	if err := param.Bind(ctx, &req); err != nil {
		return result.Fail(ctx, err)
	}

	id, err := c.service.Create(ctx.Context(), service.UserSaveInput{
		Username: req.Username,
		Password: req.Password,
		Nickname: req.Nickname,
		Remark:   req.Remark,
		Email:    req.Email,
		Mobile:   req.Mobile,
		Sex:      req.Sex,
		Avatar:   req.Avatar,
		Status:   req.Status,
	})
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, map[string]any{"id": id})
}

// Update PUT /api/v1/system/users/{id}
func (c *UserController) Update(ctx http.Context) http.Response {
	id, err := param.PathID(ctx)
	if err != nil {
		return result.Fail(ctx, err)
	}

	var req request.UserUpdateRequest
	if err := param.Bind(ctx, &req); err != nil {
		return result.Fail(ctx, err)
	}

	err = c.service.Update(ctx.Context(), id, service.UserSaveInput{
		Username: req.Username,
		Nickname: req.Nickname,
		Remark:   req.Remark,
		Email:    req.Email,
		Mobile:   req.Mobile,
		Sex:      req.Sex,
		Avatar:   req.Avatar,
		Status:   req.Status,
	})
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, nil)
}

// Destroy DELETE /api/v1/system/users/{id}
func (c *UserController) Destroy(ctx http.Context) http.Response {
	id, err := param.PathID(ctx)
	if err != nil {
		return result.Fail(ctx, err)
	}

	if err := c.service.Delete(ctx.Context(), id); err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, nil)
}

// ResetPassword PUT /api/v1/system/users/{id}/password
func (c *UserController) ResetPassword(ctx http.Context) http.Response {
	id, err := param.PathID(ctx)
	if err != nil {
		return result.Fail(ctx, err)
	}

	var req request.UserResetPasswordRequest
	if err := param.Bind(ctx, &req); err != nil {
		return result.Fail(ctx, err)
	}

	if err := c.service.ResetPassword(ctx.Context(), id, req.Password); err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, nil)
}

// Roles GET /api/v1/system/users/{id}/roles
func (c *UserController) Roles(ctx http.Context) http.Response {
	id, err := param.PathID(ctx)
	if err != nil {
		return result.Fail(ctx, err)
	}

	roleIDs, err := c.service.RoleIDs(ctx.Context(), id)
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, roleIDs)
}

// AssignRoles PUT /api/v1/system/users/{id}/roles
func (c *UserController) AssignRoles(ctx http.Context) http.Response {
	id, err := param.PathID(ctx)
	if err != nil {
		return result.Fail(ctx, err)
	}

	var req request.UserAssignRoleRequest
	if err := param.Bind(ctx, &req); err != nil {
		return result.Fail(ctx, err)
	}

	if err := c.service.AssignRoles(ctx.Context(), id, req.RoleIDs); err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, nil)
}
