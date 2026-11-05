package controller

import (
	"strconv"

	"github.com/goravel/framework/contracts/http"

	"eqray-go-starter/app/modules/base/param"
	"eqray-go-starter/app/modules/base/result"
	"eqray-go-starter/app/modules/base/support"
	"eqray-go-starter/app/modules/base/types"
	"eqray-go-starter/app/modules/infra/admin/request"
	infraService "eqray-go-starter/app/modules/infra/service"
)

// NoticeController 提供通知渠道、模板、发送与记录的接口。
type NoticeController struct {
	service *infraService.NoticeService
}

// NewNoticeController 构造通知控制器。
func NewNoticeController() *NoticeController {
	return &NoticeController{service: infraService.NewNoticeService()}
}

// Types 返回支持的渠道类型与配置字段，供前端动态渲染表单。
// 入参：无；返回类型列表。
func (c *NoticeController) Types(ctx http.Context) http.Response {
	return result.Success(ctx, c.service.Types())
}

// Channels 分页查询渠道。
// 入参：keyword、type、status、pageNo、pageSize；返回 list / total。
func (c *NoticeController) Channels(ctx http.Context) http.Response {
	query, err := c.listQuery(ctx)
	if err != nil {
		return result.Fail(ctx, err)
	}

	list, total, err := c.service.PageChannels(ctx.Context(), query)
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Page(ctx, list, total)
}

// StoreChannel 新增渠道。
// 入参：name、type、config、status、remark；返回新渠道 ID。
func (c *NoticeController) StoreChannel(ctx http.Context) http.Response {
	var req request.NoticeChannelRequest
	if err := param.Bind(ctx, &req); err != nil {
		return result.Fail(ctx, err)
	}

	id, err := c.service.CreateChannel(ctx.Context(), infraService.NoticeChannelInput{
		Name:   req.Name,
		Type:   req.Type,
		Config: req.Config,
		Status: req.Status,
		Remark: req.Remark,
	}, noticeOperator(ctx))
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, map[string]any{"id": id})
}

// UpdateChannel 编辑渠道。
// 入参：name、config、status、remark；返回 null。
func (c *NoticeController) UpdateChannel(ctx http.Context) http.Response {
	id, err := param.PathID(ctx)
	if err != nil {
		return result.Fail(ctx, err)
	}

	var req request.NoticeChannelRequest
	if err := param.Bind(ctx, &req); err != nil {
		return result.Fail(ctx, err)
	}

	if err := c.service.UpdateChannel(ctx.Context(), id, infraService.NoticeChannelInput{
		Name:   req.Name,
		Type:   req.Type,
		Config: req.Config,
		Status: req.Status,
		Remark: req.Remark,
	}, noticeOperator(ctx)); err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, nil)
}

// DestroyChannel 删除渠道。
// 入参：路径 id；返回 null。
func (c *NoticeController) DestroyChannel(ctx http.Context) http.Response {
	id, err := param.PathID(ctx)
	if err != nil {
		return result.Fail(ctx, err)
	}

	if err := c.service.DeleteChannel(ctx.Context(), id); err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, nil)
}

// TestChannel 用样例内容发一条测试消息。
// 入参：路径 id、target；返回本次发送记录。
func (c *NoticeController) TestChannel(ctx http.Context) http.Response {
	id, err := param.PathID(ctx)
	if err != nil {
		return result.Fail(ctx, err)
	}

	var req request.NoticeTestRequest
	if err := param.Bind(ctx, &req); err != nil {
		return result.Fail(ctx, err)
	}

	logModel, err := c.service.TestChannel(ctx.Context(), id, req.Target, noticeOperator(ctx))
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, logModel)
}

// Templates 分页查询模板。
func (c *NoticeController) Templates(ctx http.Context) http.Response {
	query, err := c.listQuery(ctx)
	if err != nil {
		return result.Fail(ctx, err)
	}

	list, total, err := c.service.PageTemplates(ctx.Context(), query)
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Page(ctx, list, total)
}

// StoreTemplate 新增模板。
func (c *NoticeController) StoreTemplate(ctx http.Context) http.Response {
	var req request.NoticeTemplateRequest
	if err := param.Bind(ctx, &req); err != nil {
		return result.Fail(ctx, err)
	}

	id, err := c.service.CreateTemplate(ctx.Context(), infraService.NoticeTemplateInput{
		Code:    req.Code,
		Name:    req.Name,
		Type:    req.Type,
		Title:   req.Title,
		Content: req.Content,
		Status:  req.Status,
		Remark:  req.Remark,
	}, noticeOperator(ctx))
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, map[string]any{"id": id})
}

// UpdateTemplate 编辑模板。
func (c *NoticeController) UpdateTemplate(ctx http.Context) http.Response {
	id, err := param.PathID(ctx)
	if err != nil {
		return result.Fail(ctx, err)
	}

	var req request.NoticeTemplateRequest
	if err := param.Bind(ctx, &req); err != nil {
		return result.Fail(ctx, err)
	}

	if err := c.service.UpdateTemplate(ctx.Context(), id, infraService.NoticeTemplateInput{
		Name:    req.Name,
		Type:    req.Type,
		Title:   req.Title,
		Content: req.Content,
		Status:  req.Status,
		Remark:  req.Remark,
	}, noticeOperator(ctx)); err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, nil)
}

// DestroyTemplate 删除模板。
func (c *NoticeController) DestroyTemplate(ctx http.Context) http.Response {
	id, err := param.PathID(ctx)
	if err != nil {
		return result.Fail(ctx, err)
	}

	if err := c.service.DeleteTemplate(ctx.Context(), id); err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, nil)
}

// Logs 分页查询发送记录。
func (c *NoticeController) Logs(ctx http.Context) http.Response {
	page := param.Page{
		PageNo:   ctx.Request().QueryInt("pageNo", 1),
		PageSize: ctx.Request().QueryInt("pageSize", param.DefaultPageSize),
	}
	if err := page.Normalize(); err != nil {
		return result.Fail(ctx, err)
	}

	list, total, err := c.service.PageLogs(ctx.Context(), infraService.NoticeLogQuery{
		Type:     ctx.Request().Query("type"),
		Status:   parseNoticeStatus(ctx.Request().Query("status")),
		PageNo:   page.PageNo,
		PageSize: page.PageSize,
	})
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Page(ctx, list, total)
}

// Send 通用发送：可给渠道 ID 或类型，配合模板 + 变量。
func (c *NoticeController) Send(ctx http.Context) http.Response {
	var req request.NoticeSendRequest
	if err := param.Bind(ctx, &req); err != nil {
		return result.Fail(ctx, err)
	}

	logModel, err := c.service.Send(ctx.Context(), infraService.NoticeSendInput{
		ChannelID:    req.ChannelID,
		Type:         req.Type,
		TemplateCode: req.TemplateCode,
		Target:       req.Target,
		Title:        req.Title,
		Content:      req.Content,
		Variables:    req.Variables,
		Operator:     noticeOperator(ctx),
	})
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, logModel)
}

// listQuery 解析渠道 / 模板列表的公共筛选参数。
// 入参 ctx 为请求上下文；返回查询条件与分页解析错误。
func (c *NoticeController) listQuery(ctx http.Context) (infraService.NoticeListQuery, error) {
	page := param.Page{
		PageNo:   ctx.Request().QueryInt("pageNo", 1),
		PageSize: ctx.Request().QueryInt("pageSize", param.DefaultPageSize),
	}
	if err := page.Normalize(); err != nil {
		return infraService.NoticeListQuery{}, err
	}

	return infraService.NoticeListQuery{
		Keyword:  ctx.Request().Query("keyword"),
		Type:     ctx.Request().Query("type"),
		Status:   parseNoticeStatus(ctx.Request().Query("status")),
		PageNo:   page.PageNo,
		PageSize: page.PageSize,
	}, nil
}

// noticeOperator 取当前登录账号，未登录记 system。
// 入参 ctx 为请求上下文；返回操作人标识。
func noticeOperator(ctx http.Context) string {
	if uid := support.UserID(ctx.Context()); uid != 0 {
		return types.ID(uid).String()
	}

	return "system"
}

// parseNoticeStatus 把查询串转成状态指针，空串或非法值返回 nil。
// 入参 value 为查询串；返回状态指针。
func parseNoticeStatus(value string) *int8 {
	if value == "" {
		return nil
	}

	n, err := strconv.Atoi(value)
	if err != nil {
		return nil
	}

	v := int8(n)

	return &v
}
