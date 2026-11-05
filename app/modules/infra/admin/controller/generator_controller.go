package controller

import (
	"github.com/goravel/framework/contracts/http"

	"eqray-go-starter/app/modules/base/param"
	"eqray-go-starter/app/modules/base/result"
	"eqray-go-starter/app/modules/infra/admin/request"
	infraService "eqray-go-starter/app/modules/infra/service"
)

// GeneratorController 提供代码生成：读表结构、预览与落盘。
type GeneratorController struct {
	service *infraService.GeneratorService
}

// NewGeneratorController 构造生成器控制器。
func NewGeneratorController() *GeneratorController {
	return &GeneratorController{service: infraService.NewGeneratorService()}
}

// Tables 列出可生成的业务表。
// 入参无；返回表摘要列表。
func (c *GeneratorController) Tables(ctx http.Context) http.Response {
	list, err := c.service.ListTables(ctx.Context())
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, list)
}

// Columns 读取指定表的字段定义。
// 入参：路径 table；返回字段列表。
func (c *GeneratorController) Columns(ctx http.Context) http.Response {
	list, err := c.service.DescribeTable(ctx.Context(), ctx.Request().Route("table"))
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, list)
}

// Preview 生成代码内容但不落盘。
// 入参：table、module、pascal；返回文件列表。
func (c *GeneratorController) Preview(ctx http.Context) http.Response {
	var req request.GeneratorRequest
	if err := param.Bind(ctx, &req); err != nil {
		return result.Fail(ctx, err)
	}

	files, err := c.service.Preview(ctx.Context(), infraService.GeneratorInput{
		Table:  req.Table,
		Module: req.Module,
		Pascal: req.Pascal,
	})
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, files)
}

// Generate 生成并写入文件（已存在且未开启覆盖时跳过）。
// 入参：table、module、pascal、overwrite；返回写入 / 跳过清单。
func (c *GeneratorController) Generate(ctx http.Context) http.Response {
	var req request.GeneratorRequest
	if err := param.Bind(ctx, &req); err != nil {
		return result.Fail(ctx, err)
	}

	res, err := c.service.Generate(ctx.Context(), infraService.GeneratorInput{
		Table:     req.Table,
		Module:    req.Module,
		Pascal:    req.Pascal,
		Overwrite: req.Overwrite,
	})
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, res)
}
