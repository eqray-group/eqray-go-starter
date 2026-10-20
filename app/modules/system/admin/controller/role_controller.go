package controller

import (
	"github.com/goravel/framework/contracts/http"

	baseEnums "eqray-go-starter/app/modules/base/enums"
	"eqray-go-starter/app/modules/base/param"
	"eqray-go-starter/app/modules/base/result"
	"eqray-go-starter/app/modules/system/admin/request"
	"eqray-go-starter/app/modules/system/service"
)

// RoleController 提供 system_role 的 CRUD 与角色授权接口。
type RoleController struct {
	service *service.RoleService
}

func NewRoleController() *RoleController {
	return &RoleController{service: service.NewRoleService()}
}

// Index GET /api/v1/system/roles
func (c *RoleController) Index(ctx http.Context) http.Response {
	page := param.Page{
		PageNo:   ctx.Request().QueryInt("pageNo", 1),
		PageSize: ctx.Request().QueryInt("pageSize", param.DefaultPageSize),
	}
	if err := page.Normalize(); err != nil {
		return result.Fail(ctx, err)
	}

	q := service.RoleListQuery{
		Name:     ctx.Request().Query("name"),
		Code:     ctx.Request().Query("code"),
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

// Options GET /api/v1/system/roles/options
// 用户分配角色时复用，返回全部启用角色。
func (c *RoleController) Options(ctx http.Context) http.Response {
	list, err := c.service.Options(ctx.Context())
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, list)
}

// Show GET /api/v1/system/roles/{id}
func (c *RoleController) Show(ctx http.Context) http.Response {
	id, err := param.PathID(ctx)
	if err != nil {
		return result.Fail(ctx, err)
	}

	role, err := c.service.Detail(ctx.Context(), id)
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, role)
}

// Store POST /api/v1/system/roles
func (c *RoleController) Store(ctx http.Context) http.Response {
	var req request.RoleCreateRequest
	if err := param.Bind(ctx, &req); err != nil {
		return result.Fail(ctx, err)
	}

	id, err := c.service.Create(ctx.Context(), service.RoleSaveInput{
		Name:   req.Name,
		Code:   req.Code,
		Sort:   req.Sort,
		Status: req.Status,
		Remark: req.Remark,
	})
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, map[string]any{"id": id})
}

// Update PUT /api/v1/system/roles/{id}
func (c *RoleController) Update(ctx http.Context) http.Response {
	id, err := param.PathID(ctx)
	if err != nil {
		return result.Fail(ctx, err)
	}

	var req request.RoleUpdateRequest
	if err := param.Bind(ctx, &req); err != nil {
		return result.Fail(ctx, err)
	}

	err = c.service.Update(ctx.Context(), id, service.RoleSaveInput{
		Name:   req.Name,
		Code:   req.Code,
		Sort:   req.Sort,
		Status: req.Status,
		Remark: req.Remark,
	})
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, nil)
}

// Destroy DELETE /api/v1/system/roles/{id}
func (c *RoleController) Destroy(ctx http.Context) http.Response {
	id, err := param.PathID(ctx)
	if err != nil {
		return result.Fail(ctx, err)
	}

	if err := c.service.Delete(ctx.Context(), id); err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, nil)
}

// Menus GET /api/v1/system/roles/{id}/menus
func (c *RoleController) Menus(ctx http.Context) http.Response {
	id, err := param.PathID(ctx)
	if err != nil {
		return result.Fail(ctx, err)
	}

	menuIDs, err := c.service.MenuIDs(ctx.Context(), id)
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, menuIDs)
}

// AssignMenus PUT /api/v1/system/roles/{id}/menus
func (c *RoleController) AssignMenus(ctx http.Context) http.Response {
	id, err := param.PathID(ctx)
	if err != nil {
		return result.Fail(ctx, err)
	}

	var req request.RoleAssignMenuRequest
	if err := param.Bind(ctx, &req); err != nil {
		return result.Fail(ctx, err)
	}

	if err := c.service.AssignMenus(ctx.Context(), id, req.MenuIDs); err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, nil)
}
