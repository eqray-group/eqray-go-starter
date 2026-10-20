package controller

import (
	"github.com/goravel/framework/contracts/http"

	baseEnums "eqray-go-starter/app/modules/base/enums"
	"eqray-go-starter/app/modules/base/param"
	"eqray-go-starter/app/modules/base/result"
	"eqray-go-starter/app/modules/system/admin/request"
	"eqray-go-starter/app/modules/system/service"
)

// DictController 提供字典类型与字典数据的管理接口。
type DictController struct {
	service *service.DictService
}

func NewDictController() *DictController {
	return &DictController{service: service.NewDictService()}
}

// Types GET /api/v1/system/dict-types
func (c *DictController) Types(ctx http.Context) http.Response {
	page := param.Page{
		PageNo:   ctx.Request().QueryInt("pageNo", 1),
		PageSize: ctx.Request().QueryInt("pageSize", param.DefaultPageSize),
	}
	if err := page.Normalize(); err != nil {
		return result.Fail(ctx, err)
	}

	q := service.DictTypeListQuery{
		Name:     ctx.Request().Query("name"),
		Type:     ctx.Request().Query("type"),
		PageNo:   page.PageNo,
		PageSize: page.PageSize,
	}
	if status := ctx.Request().QueryInt("status", -1); status >= 0 {
		v := baseEnums.CommonStatusEnum(status)
		q.Status = &v
	}

	list, total, err := c.service.PageTypes(ctx.Context(), q)
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Page(ctx, list, total)
}

// TypeOptions GET /api/v1/system/dict-types/options
// 字典数据弹窗选择所属类型时使用，返回全部启用类型。
func (c *DictController) TypeOptions(ctx http.Context) http.Response {
	list, err := c.service.TypeOptions(ctx.Context())
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, list)
}

// TypeShow GET /api/v1/system/dict-types/{id}
func (c *DictController) TypeShow(ctx http.Context) http.Response {
	id, err := param.PathID(ctx)
	if err != nil {
		return result.Fail(ctx, err)
	}

	m, err := c.service.TypeDetail(ctx.Context(), id)
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, m)
}

// TypeStore POST /api/v1/system/dict-types
func (c *DictController) TypeStore(ctx http.Context) http.Response {
	var req request.DictTypeCreateRequest
	if err := param.Bind(ctx, &req); err != nil {
		return result.Fail(ctx, err)
	}

	id, err := c.service.CreateType(ctx.Context(), service.DictTypeSaveInput{
		Name:   req.Name,
		Type:   req.Type,
		Status: req.Status,
		Remark: req.Remark,
	})
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, map[string]any{"id": id})
}

// TypeUpdate PUT /api/v1/system/dict-types/{id}
func (c *DictController) TypeUpdate(ctx http.Context) http.Response {
	id, err := param.PathID(ctx)
	if err != nil {
		return result.Fail(ctx, err)
	}

	var req request.DictTypeUpdateRequest
	if err := param.Bind(ctx, &req); err != nil {
		return result.Fail(ctx, err)
	}

	err = c.service.UpdateType(ctx.Context(), id, service.DictTypeSaveInput{
		Name:   req.Name,
		Type:   req.Type,
		Status: req.Status,
		Remark: req.Remark,
	})
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, nil)
}

// TypeDestroy DELETE /api/v1/system/dict-types/{id}
func (c *DictController) TypeDestroy(ctx http.Context) http.Response {
	id, err := param.PathID(ctx)
	if err != nil {
		return result.Fail(ctx, err)
	}

	if err := c.service.DeleteType(ctx.Context(), id); err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, nil)
}

// Data GET /api/v1/system/dict-data
func (c *DictController) Data(ctx http.Context) http.Response {
	page := param.Page{
		PageNo:   ctx.Request().QueryInt("pageNo", 1),
		PageSize: ctx.Request().QueryInt("pageSize", param.DefaultPageSize),
	}
	if err := page.Normalize(); err != nil {
		return result.Fail(ctx, err)
	}

	q := service.DictDataListQuery{
		DictType: ctx.Request().Query("dictType"),
		Label:    ctx.Request().Query("label"),
		PageNo:   page.PageNo,
		PageSize: page.PageSize,
	}
	if status := ctx.Request().QueryInt("status", -1); status >= 0 {
		v := baseEnums.CommonStatusEnum(status)
		q.Status = &v
	}

	list, total, err := c.service.PageData(ctx.Context(), q)
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Page(ctx, list, total)
}

// DataShow GET /api/v1/system/dict-data/{id}
func (c *DictController) DataShow(ctx http.Context) http.Response {
	id, err := param.PathID(ctx)
	if err != nil {
		return result.Fail(ctx, err)
	}

	m, err := c.service.DataDetail(ctx.Context(), id)
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, m)
}

// DataStore POST /api/v1/system/dict-data
func (c *DictController) DataStore(ctx http.Context) http.Response {
	var req request.DictDataCreateRequest
	if err := param.Bind(ctx, &req); err != nil {
		return result.Fail(ctx, err)
	}

	id, err := c.service.CreateData(ctx.Context(), service.DictDataSaveInput{
		Sort:      req.Sort,
		Label:     req.Label,
		Value:     req.Value,
		DictType:  req.DictType,
		Status:    req.Status,
		ColorType: req.ColorType,
		CSSClass:  req.CSSClass,
		Remark:    req.Remark,
	})
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, map[string]any{"id": id})
}

// DataUpdate PUT /api/v1/system/dict-data/{id}
func (c *DictController) DataUpdate(ctx http.Context) http.Response {
	id, err := param.PathID(ctx)
	if err != nil {
		return result.Fail(ctx, err)
	}

	var req request.DictDataUpdateRequest
	if err := param.Bind(ctx, &req); err != nil {
		return result.Fail(ctx, err)
	}

	err = c.service.UpdateData(ctx.Context(), id, service.DictDataSaveInput{
		Sort:      req.Sort,
		Label:     req.Label,
		Value:     req.Value,
		DictType:  req.DictType,
		Status:    req.Status,
		ColorType: req.ColorType,
		CSSClass:  req.CSSClass,
		Remark:    req.Remark,
	})
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, nil)
}

// DataDestroy DELETE /api/v1/system/dict-data/{id}
func (c *DictController) DataDestroy(ctx http.Context) http.Response {
	id, err := param.PathID(ctx)
	if err != nil {
		return result.Fail(ctx, err)
	}

	if err := c.service.DeleteData(ctx.Context(), id); err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, nil)
}

// Items GET /api/v1/system/dicts/{type}
// 前端下拉与标签渲染统一走这里，只需登录态，不挂具体权限点。
func (c *DictController) Items(ctx http.Context) http.Response {
	items, err := c.service.Items(ctx.Context(), ctx.Request().Route("type"))
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, items)
}
