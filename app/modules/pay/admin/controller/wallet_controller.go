package controller

import (
	"github.com/goravel/framework/contracts/http"

	baseErrs "eqray-go-starter/app/modules/base/errs"
	"eqray-go-starter/app/modules/base/param"
	"eqray-go-starter/app/modules/base/result"
	"eqray-go-starter/app/modules/base/types"
	"eqray-go-starter/app/modules/pay/admin/request"
	payService "eqray-go-starter/app/modules/pay/service"
)

// PayWalletController 提供钱包账户查询、流水查询与余额调整。
type PayWalletController struct {
	service *payService.PayWalletService
}

// NewPayWalletController 构造钱包控制器。
func NewPayWalletController() *PayWalletController {
	return &PayWalletController{service: payService.NewPayWalletService()}
}

// Index 分页查询钱包账户。
// 入参：userId（可选）、pageNo、pageSize；返回钱包列表与总数。
func (c *PayWalletController) Index(ctx http.Context) http.Response {
	page := param.Page{
		PageNo:   ctx.Request().QueryInt("pageNo", 1),
		PageSize: ctx.Request().QueryInt("pageSize", param.DefaultPageSize),
	}
	if err := page.Normalize(); err != nil {
		return result.Fail(ctx, err)
	}

	list, total, err := c.service.Page(ctx.Context(), payService.PayWalletListQuery{
		UserID:   types.ID(ctx.Request().QueryInt("userId", 0)),
		PageNo:   page.PageNo,
		PageSize: page.PageSize,
	})
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Page(ctx, list, total)
}

// Records 分页查询某会员的钱包流水。
// 入参：userId（必填）、pageNo、pageSize；返回流水列表与总数。
func (c *PayWalletController) Records(ctx http.Context) http.Response {
	userID := types.ID(ctx.Request().QueryInt("userId", 0))
	if userID == 0 {
		return result.Fail(ctx, baseErrs.WithMsg(baseErrs.BadRequest, "请先选择会员"))
	}

	page := param.Page{
		PageNo:   ctx.Request().QueryInt("pageNo", 1),
		PageSize: ctx.Request().QueryInt("pageSize", param.DefaultPageSize),
	}
	if err := page.Normalize(); err != nil {
		return result.Fail(ctx, err)
	}

	list, total, err := c.service.Transactions(ctx.Context(), userID, page.PageNo, page.PageSize)
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Page(ctx, list, total)
}

// Adjust 调整会员钱包余额：正数充值、负数扣减，写流水并留操作日志。
// 入参：userId、amount（分）、remark；返回 null。
func (c *PayWalletController) Adjust(ctx http.Context) http.Response {
	var req request.PayWalletAdjustRequest
	if err := param.Bind(ctx, &req); err != nil {
		return result.Fail(ctx, err)
	}

	if err := c.service.Adjust(ctx.Context(), payService.WalletAdjustInput{
		UserID: req.UserID,
		Amount: req.Amount,
		Remark: req.Remark,
	}); err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, nil)
}

// Freeze 冻结会员钱包资金：可用余额转入冻结金额，不可消费但仍是会员资产。
// 入参：userId、amount（分）、remark；返回 null。
func (c *PayWalletController) Freeze(ctx http.Context) http.Response {
	var req request.PayWalletFreezeRequest
	if err := param.Bind(ctx, &req); err != nil {
		return result.Fail(ctx, err)
	}

	if err := c.service.Freeze(ctx.Context(), payService.WalletFreezeInput{
		UserID: req.UserID,
		Amount: req.Amount,
		Remark: req.Remark,
	}); err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, nil)
}

// Unfreeze 解冻会员钱包资金：冻结金额转回可用余额。
// 入参：userId、amount（分）、remark；返回 null。
func (c *PayWalletController) Unfreeze(ctx http.Context) http.Response {
	var req request.PayWalletFreezeRequest
	if err := param.Bind(ctx, &req); err != nil {
		return result.Fail(ctx, err)
	}

	if err := c.service.Unfreeze(ctx.Context(), payService.WalletFreezeInput{
		UserID: req.UserID,
		Amount: req.Amount,
		Remark: req.Remark,
	}); err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, nil)
}
