package controller

import (
	"github.com/goravel/framework/contracts/http"

	"eqray-go-starter/app/modules/base/param"
	"eqray-go-starter/app/modules/base/result"
	payRepository "eqray-go-starter/app/modules/pay/repository"
)

// PayNotifyLogController 提供支付回调日志的分页查询，用于排查「为什么没到账」。
type PayNotifyLogController struct {
	logs *payRepository.PayNotifyLogRepository
}

// NewPayNotifyLogController 构造回调日志控制器。
func NewPayNotifyLogController() *PayNotifyLogController {
	return &PayNotifyLogController{logs: payRepository.NewPayNotifyLogRepository()}
}

// Index 分页查询回调日志，支持渠道、类型、单号与结果筛选。
// 入参：channelCode、type、merchantOrderNo、verifyStatus、handleStatus、pageNo、pageSize（查询串）；
// 返回日志列表与总数。
func (c *PayNotifyLogController) Index(ctx http.Context) http.Response {
	page := param.Page{
		PageNo:   ctx.Request().QueryInt("pageNo", 1),
		PageSize: ctx.Request().QueryInt("pageSize", param.DefaultPageSize),
	}
	if err := page.Normalize(); err != nil {
		return result.Fail(ctx, err)
	}

	filter := payRepository.PayNotifyLogFilter{
		ChannelCode:     ctx.Request().Query("channelCode"),
		Type:            ctx.Request().Query("type"),
		MerchantOrderNo: ctx.Request().Query("merchantOrderNo"),
	}
	if v := ctx.Request().QueryInt("verifyStatus", -1); v >= 0 {
		status := int8(v)
		filter.VerifyStatus = &status
	}
	if v := ctx.Request().QueryInt("handleStatus", -1); v >= 0 {
		status := int8(v)
		filter.HandleStatus = &status
	}

	list, total, err := c.logs.Paginate(ctx.Context(), filter, page.PageNo, page.PageSize)
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Page(ctx, list, total)
}
