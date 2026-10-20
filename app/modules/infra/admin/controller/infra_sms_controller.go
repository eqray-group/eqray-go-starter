package controller

import (
	"github.com/goravel/framework/contracts/http"

	baseEnums "eqray-go-starter/app/modules/base/enums"
	"eqray-go-starter/app/modules/base/param"
	"eqray-go-starter/app/modules/base/result"
	"eqray-go-starter/app/modules/base/types"
	"eqray-go-starter/app/modules/infra/admin/request"
	"eqray-go-starter/app/modules/infra/service"
)

// InfraSmsChannelController 提供短信渠道配置的增删改查。
type InfraSmsChannelController struct {
	service *service.InfraSmsChannelService
}

func NewInfraSmsChannelController() *InfraSmsChannelController {
	return &InfraSmsChannelController{service: service.NewInfraSmsChannelService()}
}

// Index GET /api/v1/infra/sms-channels
func (c *InfraSmsChannelController) Index(ctx http.Context) http.Response {
	page := param.Page{
		PageNo:   ctx.Request().QueryInt("pageNo", 1),
		PageSize: ctx.Request().QueryInt("pageSize", param.DefaultPageSize),
	}
	if err := page.Normalize(); err != nil {
		return result.Fail(ctx, err)
	}

	q := service.InfraSmsChannelListQuery{
		Keyword:  ctx.Request().Query("keyword"),
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

// Options GET /api/v1/infra/sms-channels/options
func (c *InfraSmsChannelController) Options(ctx http.Context) http.Response {
	list, err := c.service.Options(ctx.Context())
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, list)
}

// Show GET /api/v1/infra/sms-channels/{id}
func (c *InfraSmsChannelController) Show(ctx http.Context) http.Response {
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

// Store POST /api/v1/infra/sms-channels
func (c *InfraSmsChannelController) Store(ctx http.Context) http.Response {
	var req request.InfraSmsChannelSaveRequest
	if err := param.Bind(ctx, &req); err != nil {
		return result.Fail(ctx, err)
	}

	id, err := c.service.Create(ctx.Context(), service.InfraSmsChannelSaveInput{
		Name:      req.Name,
		Code:      req.Code,
		APIKey:    req.APIKey,
		APISecret: req.APISecret,
		Signature: req.Signature,
		Extra:     req.Extra,
		Status:    req.Status,
		Remark:    req.Remark,
	})
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, map[string]any{"id": id})
}

// Update PUT /api/v1/infra/sms-channels/{id}
func (c *InfraSmsChannelController) Update(ctx http.Context) http.Response {
	id, err := param.PathID(ctx)
	if err != nil {
		return result.Fail(ctx, err)
	}

	var req request.InfraSmsChannelSaveRequest
	if err := param.Bind(ctx, &req); err != nil {
		return result.Fail(ctx, err)
	}

	err = c.service.Update(ctx.Context(), id, service.InfraSmsChannelSaveInput{
		Name:      req.Name,
		Code:      req.Code,
		APIKey:    req.APIKey,
		APISecret: req.APISecret,
		Signature: req.Signature,
		Extra:     req.Extra,
		Status:    req.Status,
		Remark:    req.Remark,
	})
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, nil)
}

// Destroy DELETE /api/v1/infra/sms-channels/{id}
func (c *InfraSmsChannelController) Destroy(ctx http.Context) http.Response {
	id, err := param.PathID(ctx)
	if err != nil {
		return result.Fail(ctx, err)
	}

	if err := c.service.Delete(ctx.Context(), id); err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, nil)
}

// InfraSmsTemplateController 提供短信模板的增删改查。
type InfraSmsTemplateController struct {
	service *service.InfraSmsTemplateService
}

func NewInfraSmsTemplateController() *InfraSmsTemplateController {
	return &InfraSmsTemplateController{service: service.NewInfraSmsTemplateService()}
}

// Index GET /api/v1/infra/sms-templates
func (c *InfraSmsTemplateController) Index(ctx http.Context) http.Response {
	page := param.Page{
		PageNo:   ctx.Request().QueryInt("pageNo", 1),
		PageSize: ctx.Request().QueryInt("pageSize", param.DefaultPageSize),
	}
	if err := page.Normalize(); err != nil {
		return result.Fail(ctx, err)
	}

	q := service.InfraSmsTemplateListQuery{
		Keyword:   ctx.Request().Query("keyword"),
		ChannelID: types.ID(ctx.Request().QueryInt("channelId", 0)),
		PageNo:    page.PageNo,
		PageSize:  page.PageSize,
	}
	if templateType := ctx.Request().QueryInt("type", 0); templateType > 0 {
		v := int8(templateType)
		q.Type = &v
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

// Show GET /api/v1/infra/sms-templates/{id}
func (c *InfraSmsTemplateController) Show(ctx http.Context) http.Response {
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

// Store POST /api/v1/infra/sms-templates
func (c *InfraSmsTemplateController) Store(ctx http.Context) http.Response {
	var req request.InfraSmsTemplateSaveRequest
	if err := param.Bind(ctx, &req); err != nil {
		return result.Fail(ctx, err)
	}

	id, err := c.service.Create(ctx.Context(), service.InfraSmsTemplateSaveInput{
		ChannelID:     req.ChannelID,
		Code:          req.Code,
		Name:          req.Name,
		Type:          req.Type,
		APITemplateID: req.APITemplateID,
		Content:       req.Content,
		Params:        req.Params,
		Status:        req.Status,
		Remark:        req.Remark,
	})
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, map[string]any{"id": id})
}

// Update PUT /api/v1/infra/sms-templates/{id}
func (c *InfraSmsTemplateController) Update(ctx http.Context) http.Response {
	id, err := param.PathID(ctx)
	if err != nil {
		return result.Fail(ctx, err)
	}

	var req request.InfraSmsTemplateSaveRequest
	if err := param.Bind(ctx, &req); err != nil {
		return result.Fail(ctx, err)
	}

	err = c.service.Update(ctx.Context(), id, service.InfraSmsTemplateSaveInput{
		ChannelID:     req.ChannelID,
		Code:          req.Code,
		Name:          req.Name,
		Type:          req.Type,
		APITemplateID: req.APITemplateID,
		Content:       req.Content,
		Params:        req.Params,
		Status:        req.Status,
		Remark:        req.Remark,
	})
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, nil)
}

// Destroy DELETE /api/v1/infra/sms-templates/{id}
func (c *InfraSmsTemplateController) Destroy(ctx http.Context) http.Response {
	id, err := param.PathID(ctx)
	if err != nil {
		return result.Fail(ctx, err)
	}

	if err := c.service.Delete(ctx.Context(), id); err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, nil)
}
