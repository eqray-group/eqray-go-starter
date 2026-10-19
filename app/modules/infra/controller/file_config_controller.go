package controller

import (
	"github.com/goravel/framework/contracts/http"

	"eqray-go-starter/app/modules/base/param"
	"eqray-go-starter/app/modules/base/result"
	"eqray-go-starter/app/modules/infra/request"
	"eqray-go-starter/app/modules/infra/service"
)

// FileConfigController 提供文件配置的增删改查接口。
type FileConfigController struct {
	service *service.FileConfigService
}

func NewFileConfigController() *FileConfigController {
	return &FileConfigController{service: service.NewFileConfigService()}
}

// Index GET /api/v1/infra/file-configs
// 文件配置是低频数据，直接返回全量列表。
func (c *FileConfigController) Index(ctx http.Context) http.Response {
	list, err := c.service.List(ctx.Context())
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, map[string]any{"list": list})
}

// Store POST /api/v1/infra/file-configs
func (c *FileConfigController) Store(ctx http.Context) http.Response {
	var req request.FileConfigRequest
	if err := param.Bind(ctx, &req); err != nil {
		return result.Fail(ctx, err)
	}

	err := c.service.Create(ctx.Context(), service.FileConfigInput{
		Name:    req.Name,
		Storage: req.Storage,
		Master:  req.Master,
		Config:  req.Config,
	})
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, nil)
}

// Update PUT /api/v1/infra/file-configs/{id}
func (c *FileConfigController) Update(ctx http.Context) http.Response {
	id, err := param.PathID(ctx)
	if err != nil {
		return result.Fail(ctx, err)
	}

	var req request.FileConfigRequest
	if err := param.Bind(ctx, &req); err != nil {
		return result.Fail(ctx, err)
	}

	err = c.service.Update(ctx.Context(), id, service.FileConfigInput{
		Name:    req.Name,
		Storage: req.Storage,
		Master:  req.Master,
		Config:  req.Config,
	})
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, nil)
}

// SetMaster PUT /api/v1/infra/file-configs/{id}/master
// 把指定配置设为主配置，新上传文件将落到该配置的目录。
func (c *FileConfigController) SetMaster(ctx http.Context) http.Response {
	id, err := param.PathID(ctx)
	if err != nil {
		return result.Fail(ctx, err)
	}

	if err := c.service.SetMaster(ctx.Context(), id); err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, nil)
}

// Destroy DELETE /api/v1/infra/file-configs/{id}
func (c *FileConfigController) Destroy(ctx http.Context) http.Response {
	id, err := param.PathID(ctx)
	if err != nil {
		return result.Fail(ctx, err)
	}

	if err := c.service.Delete(ctx.Context(), id); err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, nil)
}
