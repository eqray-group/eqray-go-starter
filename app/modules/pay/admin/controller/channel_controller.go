// Package controller 是 pay 模块管理端接入层：取参、绑定校验、调 service、统一响应。
package controller

import (
	"github.com/goravel/framework/contracts/http"

	baseEnums "eqray-go-starter/app/modules/base/enums"
	"eqray-go-starter/app/modules/base/param"
	"eqray-go-starter/app/modules/base/result"
	"eqray-go-starter/app/modules/base/types"
	"eqray-go-starter/app/modules/pay/admin/request"
	payChannel "eqray-go-starter/app/modules/pay/channel"
	payService "eqray-go-starter/app/modules/pay/service"
)

// PayChannelController 提供支付渠道配置的增删改查、启停、schema 与回调地址查询。
type PayChannelController struct {
	service *payService.PayChannelService
}

// NewPayChannelController 构造支付渠道控制器。
func NewPayChannelController() *PayChannelController {
	return &PayChannelController{service: payService.NewPayChannelService()}
}

// Index 分页查询支付渠道，config 脱敏后返回。
// 入参：keyword、code、status、pageNo、pageSize（查询串）；返回渠道列表与总数。
func (c *PayChannelController) Index(ctx http.Context) http.Response {
	page := param.Page{
		PageNo:   ctx.Request().QueryInt("pageNo", 1),
		PageSize: ctx.Request().QueryInt("pageSize", param.DefaultPageSize),
	}
	if err := page.Normalize(); err != nil {
		return result.Fail(ctx, err)
	}

	q := payService.PayChannelListQuery{
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

// Schema 按渠道编码返回动态表单字段定义，前端据此渲染渠道表单，不硬编码映射。
// 入参：查询串 code；返回渠道元数据与字段定义。
func (c *PayChannelController) Schema(ctx http.Context) http.Response {
	code := ctx.Request().Query("code")

	meta, err := c.service.Schema(code)
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, map[string]any{
		"code":      code,
		"vendor":    meta.Vendor,
		"payMethod": meta.PayMethod,
		"fields":    meta.Fields,
		"codes":     payChannel.SupportedCodes(),
	})
}

// NotifyURLs 按渠道编码返回系统生成的异步通知地址与当前 APP_URL，供后台页面只读展示。
// 入参：查询串 code；返回 appUrl 与 payNotifyUrl。
func (c *PayChannelController) NotifyURLs(ctx http.Context) http.Response {
	code := ctx.Request().Query("code")

	notifyURL, err := c.service.NotifyURL(code)
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, map[string]any{
		"payNotifyUrl": notifyURL,
	})
}

// Show 查询渠道详情（config 脱敏）。
// 入参：路径 id；返回渠道详情，不存在返回业务错误码。
func (c *PayChannelController) Show(ctx http.Context) http.Response {
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

// Store 新增支付渠道：编码唯一且受支持，config 明文入库。
// 入参：name、code、config、status、sort、remark；返回新渠道 ID。
func (c *PayChannelController) Store(ctx http.Context) http.Response {
	var req request.PayChannelSaveRequest
	if err := param.Bind(ctx, &req); err != nil {
		return result.Fail(ctx, err)
	}

	id, err := c.service.Create(ctx.Context(), payService.PayChannelSaveInput{
		Name:   req.Name,
		Code:   req.Code,
		Config: req.Config,
		Status: req.Status,
		Sort:   req.Sort,
		Remark: req.Remark,
	})
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, map[string]any{"id": id})
}

// Update 编辑支付渠道；config 中敏感字段留空表示沿用原值。
// 入参：路径 id + 渠道字段；返回 null。
func (c *PayChannelController) Update(ctx http.Context) http.Response {
	id, err := param.PathID(ctx)
	if err != nil {
		return result.Fail(ctx, err)
	}

	var req request.PayChannelSaveRequest
	if err := param.Bind(ctx, &req); err != nil {
		return result.Fail(ctx, err)
	}

	if err := c.service.Update(ctx.Context(), id, payService.PayChannelSaveInput{
		Name:   req.Name,
		Config: req.Config,
		Status: req.Status,
		Sort:   req.Sort,
		Remark: req.Remark,
	}); err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, nil)
}

// SetStatus 启用 / 停用渠道。
// 入参：路径 id + enable（布尔）；返回 null。
func (c *PayChannelController) SetStatus(ctx http.Context) http.Response {
	id, err := param.PathID(ctx)
	if err != nil {
		return result.Fail(ctx, err)
	}

	var req request.PayChannelStatusRequest
	if err := param.Bind(ctx, &req); err != nil {
		return result.Fail(ctx, err)
	}

	if err := c.service.SetStatus(ctx.Context(), id, req.Enable); err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, nil)
}

// Destroy 物理删除支付渠道。
// 入参：路径 id；返回 null。
func (c *PayChannelController) Destroy(ctx http.Context) http.Response {
	id, err := param.PathID(ctx)
	if err != nil {
		return result.Fail(ctx, err)
	}

	if err := c.service.Delete(ctx.Context(), id); err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, nil)
}

// Test 渠道连通性测试：用当前配置发起一次查单（单号不存在），凭据错误与订单不存在可区分。
// 入参：路径 id；返回测通结果说明。
func (c *PayChannelController) Test(ctx http.Context) http.Response {
	id, err := param.PathID(ctx)
	if err != nil {
		return result.Fail(ctx, err)
	}

	m, err := c.service.Detail(ctx.Context(), id)
	if err != nil {
		return result.Fail(ctx, err)
	}

	// Detail 返回的是脱敏 config，测通必须用原值：直接按 code 走 Resolve（读库原值）
	_, ch, err := c.service.Resolve(ctx.Context(), m.Code, false)
	if err != nil {
		return result.Fail(ctx, err)
	}

	qr, err := ch.QueryOrder(ctx.Context(), "pay_connectivity_test_"+types.ID(id).String())
	if err != nil {
		// 用不存在的单号查单，渠道必然回「订单不存在」：能拿到该错误说明验签与凭据都对
		if payChannel.IsOrderNotExist(err) {
			return result.Success(ctx, map[string]any{
				"ok":      true,
				"message": "凭据有效，接口连通正常（测试单号不存在）",
			})
		}

		return result.Fail(ctx, err)
	}

	return result.Success(ctx, map[string]any{
		"ok":      true,
		"message": "凭据有效，接口连通正常（测试单号状态：" + qr.TradeState + "）",
	})
}
