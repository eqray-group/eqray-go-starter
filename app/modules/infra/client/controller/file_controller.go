// Package controller 存放 infra 模块 C 端的控制器，只做取参、调 Service、出响应。
package controller

import (
	"github.com/goravel/framework/contracts/http"

	baseErrs "eqray-go-starter/app/modules/base/errs"
	"eqray-go-starter/app/modules/base/result"
	"eqray-go-starter/app/modules/infra/service"
)

// FileController 提供 C 端的文件上传接口，供头像、评价图等自助场景使用。
type FileController struct {
	service *service.FileService
}

func NewFileController() *FileController {
	return &FileController{service: service.NewFileService()}
}

// Upload POST /api/v1/client/infra/files
// 表单字段固定为 file，成功返回文件 id 与可直接访问的 url。
// 上传后的文件仍记在 infra_file，由后台文件管理统一查看与清理。
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
