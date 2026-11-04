package service

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"eqray-go-starter/app/modules/base/errs"
	"eqray-go-starter/app/modules/base/types"
)

// BizOrderInfo 是支付模块向业务模块反查订单得到的关键信息。
type BizOrderInfo struct {
	// Price 应付金额（分），金额只信业务侧，C 端请求不接受金额字段
	Price int64
	// Subject 商品标题
	Subject string
	// Description 商品描述，可空
	Description string
	// UserID 订单归属会员，C 端下单时必须与当前登录会员一致
	UserID types.ID
}

// BizOrderResolver 由业务模块注册：按业务单号反查订单信息。
// 入参 bizOrderNo 为业务单号、userID 为当前登录会员（管理端 / 服务间调用可能为 0）；
// 返回订单信息；订单不存在或归属不符返回错误。
type BizOrderResolver func(ctx context.Context, bizOrderNo string, userID types.ID) (*BizOrderInfo, error)

// bizResolvers 是业务类型 → 反查函数的注册表。
// pay 模块不 import 业务模块，业务模块启动时调用 RegisterBizOrder 登记反查能力。
var bizResolvers sync.Map

// RegisterBizOrder 注册某业务类型的订单反查函数，重复注册以后者为准。
// 入参 bizType 为业务类型（如 demo_order）、resolver 为反查函数。
func RegisterBizOrder(bizType string, resolver BizOrderResolver) {
	bizResolvers.Store(strings.TrimSpace(bizType), resolver)
}

// resolveBizOrder 反查业务订单；业务类型未注册或反查失败返回业务错误码。
// 入参 bizType/bizOrderNo 定位订单、userID 为当前会员；返回订单信息。
func resolveBizOrder(ctx context.Context, bizType, bizOrderNo string, userID types.ID) (*BizOrderInfo, error) {
	value, ok := bizResolvers.Load(strings.TrimSpace(bizType))
	if !ok {
		return nil, errs.WithMsg(errs.BadRequest, fmt.Sprintf("不支持的业务类型【%v】", bizType))
	}

	return value.(BizOrderResolver)(ctx, bizOrderNo, userID)
}
