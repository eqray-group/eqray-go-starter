// Package controller 存放 infra 模块的控制器，只做取参、调 Service、出响应。
package controller

import (
	"github.com/goravel/framework/contracts/http"

	baseErrs "eqray-go-starter/app/modules/base/errs"
	"eqray-go-starter/app/modules/base/param"
	"eqray-go-starter/app/modules/base/result"
	"eqray-go-starter/app/modules/infra/service"
)

// FileController 提供后台的文件上传、列表与删除接口；C 端上传见 client/controller。
type FileController struct {
	service *service.FileService
}

func NewFileController() *FileController {
	return &FileController{service: service.NewFileService()}
}

// Upload POST /api/v1/infra/files
// 后台自助上传（头像等），只要求登录态、不校验具体权限点；写操作由路由上的 OperateLog 留痕。
func (c *FileController) Upload(ctx http.Context) http.Response {
	file, err := ctx.Request().File("file")
	if err != nil || file == nil {
		return result.Fail(ctx, baseErrs.WithMsg(baseErrs.BadRequest, "请选择要上传的文件"))
	}

	m, err := c.service.Upload(ctx.Context(), file)
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, map[string]any{
		"id":   m.ID,
		"name": m.Name,
		"url":  m.Url,
		"type": m.Type,
		"size": m.Size,
	})
}

// Index GET /api/v1/infra/files
// 文件管理页的列表，按上传时间倒序。
func (c *FileController) Index(ctx http.Context) http.Response {
	page := param.Page{
		PageNo:   ctx.Request().QueryInt("pageNo", 1),
		PageSize: ctx.Request().QueryInt("pageSize", param.DefaultPageSize),
	}
	if err := page.Normalize(); err != nil {
		return result.Fail(ctx, err)
	}

	list, total, err := c.service.Page(ctx.Context(), service.FileListQuery{
		Keyword:  ctx.Request().Query("keyword"),
		PageNo:   page.PageNo,
		PageSize: page.PageSize,
	})
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Page(ctx, list, total)
}

// Destroy DELETE /api/v1/infra/files/{id}
func (c *FileController) Destroy(ctx http.Context) http.Response {
	id, err := param.PathID(ctx)
	if err != nil {
		return result.Fail(ctx, err)
	}

	if err := c.service.Delete(ctx.Context(), id); err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, nil)
}
