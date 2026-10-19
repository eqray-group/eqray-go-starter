package controller

import (
	"github.com/goravel/framework/contracts/http"

	"eqray-go-starter/app/modules/base/param"
	"eqray-go-starter/app/modules/base/result"
	"eqray-go-starter/app/modules/infra/request"
	"eqray-go-starter/app/modules/infra/service"
)

// ConfigController 提供参数配置的增删改查接口。
type ConfigController struct {
	service *service.ConfigService
}

func NewConfigController() *ConfigController {
	return &ConfigController{service: service.NewConfigService()}
}

// Index GET /api/v1/infra/configs
func (c *ConfigController) Index(ctx http.Context) http.Response {
	page := param.Page{
		PageNo:   ctx.Request().QueryInt("pageNo", 1),
		PageSize: ctx.Request().QueryInt("pageSize", param.DefaultPageSize),
	}
	if err := page.Normalize(); err != nil {
		return result.Fail(ctx, err)
	}

	list, total, err := c.service.Page(ctx.Context(), service.ConfigListQuery{
		Category: ctx.Request().Query("category"),
		Keyword:  ctx.Request().Query("keyword"),
		PageNo:   page.PageNo,
		PageSize: page.PageSize,
	})
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Page(ctx, list, total)
}

// Store POST /api/v1/infra/configs
func (c *ConfigController) Store(ctx http.Context) http.Response {
	var req request.ConfigCreateRequest
	if err := param.Bind(ctx, &req); err != nil {
		return result.Fail(ctx, err)
	}

	err := c.service.Create(ctx.Context(), service.ConfigCreateInput{
		Category:  req.Category,
		Name:      req.Name,
		ConfigKey: req.ConfigKey,
		Value:     req.Value,
		Type:      req.Type,
		Visible:   req.Visible,
		Remark:    req.Remark,
	})
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, nil)
}

// Update PUT /api/v1/infra/configs/{id}
func (c *ConfigController) Update(ctx http.Context) http.Response {
	id, err := param.PathID(ctx)
	if err != nil {
		return result.Fail(ctx, err)
	}

	var req request.ConfigUpdateRequest
	if err := param.Bind(ctx, &req); err != nil {
		return result.Fail(ctx, err)
	}

	err = c.service.Update(ctx.Context(), id, service.ConfigUpdateInput{
		Category: req.Category,
		Name:     req.Name,
		Value:    req.Value,
		Visible:  req.Visible,
		Remark:   req.Remark,
	})
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, nil)
}

// Destroy DELETE /api/v1/infra/configs/{id}
func (c *ConfigController) Destroy(ctx http.Context) http.Response {
	id, err := param.PathID(ctx)
	if err != nil {
		return result.Fail(ctx, err)
	}

	if err := c.service.Delete(ctx.Context(), id); err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, nil)
}
