package controller

import (
	"github.com/goravel/framework/contracts/http"

	"eqray-go-starter/app/modules/base/param"
	"eqray-go-starter/app/modules/base/result"
	"eqray-go-starter/app/modules/base/support"
	"eqray-go-starter/app/modules/base/types"
	"eqray-go-starter/app/modules/system/admin/request"
	"eqray-go-starter/app/modules/system/service"
)

// MenuController 提供 system_menu 的树形管理接口。
type MenuController struct {
	service *service.MenuService
}

func NewMenuController() *MenuController {
	return &MenuController{service: service.NewMenuService()}
}

// Index GET /api/v1/system/menus
func (c *MenuController) Index(ctx http.Context) http.Response {
	tree, err := c.service.Tree(ctx.Context())
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, tree)
}

// Simple GET /api/v1/system/menus/simple
// 菜单管理选上级、角色授权勾选菜单都复用它，返回整棵树的精简形态。
func (c *MenuController) Simple(ctx http.Context) http.Response {
	tree, err := c.service.Tree(ctx.Context())
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, tree)
}

// Routes GET /api/v1/system/menus/routes
// 后端权限模式：按当前登录者可见的菜单生成前端动态路由。
func (c *MenuController) Routes(ctx http.Context) http.Response {
	routes, err := c.service.Routes(ctx.Context(), types.ID(support.UserID(ctx.Context())))
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, routes)
}

// Show GET /api/v1/system/menus/{id}
func (c *MenuController) Show(ctx http.Context) http.Response {
	id, err := param.PathID(ctx)
	if err != nil {
		return result.Fail(ctx, err)
	}

	menu, err := c.service.Detail(ctx.Context(), id)
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, menu)
}

// Store POST /api/v1/system/menus
func (c *MenuController) Store(ctx http.Context) http.Response {
	var req request.MenuCreateRequest
	if err := param.Bind(ctx, &req); err != nil {
		return result.Fail(ctx, err)
	}

	id, err := c.service.Create(ctx.Context(), service.MenuSaveInput{
		Name:          req.Name,
		Permission:    req.Permission,
		Type:          req.Type,
		Sort:          req.Sort,
		ParentID:      req.ParentID,
		Path:          req.Path,
		Icon:          req.Icon,
		Component:     req.Component,
		ComponentName: req.ComponentName,
		Status:        req.Status,
		Visible:       req.Visible,
		KeepAlive:     req.KeepAlive,
		AlwaysShow:    req.AlwaysShow,
	})
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, map[string]any{"id": id})
}

// Update PUT /api/v1/system/menus/{id}
func (c *MenuController) Update(ctx http.Context) http.Response {
	id, err := param.PathID(ctx)
	if err != nil {
		return result.Fail(ctx, err)
	}

	var req request.MenuUpdateRequest
	if err := param.Bind(ctx, &req); err != nil {
		return result.Fail(ctx, err)
	}

	err = c.service.Update(ctx.Context(), id, service.MenuSaveInput{
		Name:          req.Name,
		Permission:    req.Permission,
		Type:          req.Type,
		Sort:          req.Sort,
		ParentID:      req.ParentID,
		Path:          req.Path,
		Icon:          req.Icon,
		Component:     req.Component,
		ComponentName: req.ComponentName,
		Status:        req.Status,
		Visible:       req.Visible,
		KeepAlive:     req.KeepAlive,
		AlwaysShow:    req.AlwaysShow,
	})
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, nil)
}

// Destroy DELETE /api/v1/system/menus/{id}
func (c *MenuController) Destroy(ctx http.Context) http.Response {
	id, err := param.PathID(ctx)
	if err != nil {
		return result.Fail(ctx, err)
	}

	if err := c.service.Delete(ctx.Context(), id); err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, nil)
}
