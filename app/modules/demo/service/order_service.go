// Package service 是演示业务模块的业务层：下单并调起支付、反查订单、回写支付结果。
// demo 模块演示「业务方接入支付」的标准姿势：调 pay 模块下单 + 订阅 pay.order.paid 事件。
package service

import (
	"context"
	"strconv"
	"time"

	"eqray-go-starter/app/facades"
	baseErrs "eqray-go-starter/app/modules/base/errs"
	"eqray-go-starter/app/modules/base/support"
	"eqray-go-starter/app/modules/base/types"
	demoModel "eqray-go-starter/app/modules/demo/model"
	demoRepository "eqray-go-starter/app/modules/demo/repository"
	payService "eqray-go-starter/app/modules/pay/service"
)

// DemoPrice 是演示商品金额（分）：写死 1 分钱，联调用，不信任前端传值。
const DemoPrice = 1

// DemoSubject 是演示商品标题。
const DemoSubject = "演示商品（1 分钱联调）"

// DemoBizType 是演示业务在 pay 模块登记的业务类型。
const DemoBizType = "demo_order"

// DemoOrderService 负责演示订单的下单与支付结果回写。
type DemoOrderService struct {
	orders   *demoRepository.DemoOrderRepository
	payOrder *payService.PayOrderService
}

// NewDemoOrderService 构造演示订单服务。
func NewDemoOrderService() *DemoOrderService {
	return &DemoOrderService{
		orders:   demoRepository.NewDemoOrderRepository(),
		payOrder: payService.NewPayOrderService(),
	}
}

// CreateAndPay 创建演示订单并直接发起支付。
// 入参 ctx 为请求上下文、userID 为当前会员、clientIP 为下单 IP、channelCode 为支付渠道；
// 返回业务单号与支付拉起参数。
func (s *DemoOrderService) CreateAndPay(ctx context.Context, userID types.ID, clientIP, channelCode string) (*payService.CreateOrderResult, string, error) {
	orderNo := "D" + strconv.FormatInt(int64(support.NewID()), 10)

	if err := s.orders.Insert(ctx, &demoModel.DemoOrderModel{
		OrderNo: orderNo,
		UserID:  userID,
		Subject: DemoSubject,
		Price:   DemoPrice,
		Status:  demoModel.DemoOrderStatusPending,
	}); err != nil {
		return nil, "", err
	}

	created, err := s.payOrder.CreateOrder(ctx, payService.CreateOrderInput{
		BizType:     DemoBizType,
		BizOrderNo:  orderNo,
		ChannelCode: channelCode,
		Subject:     DemoSubject,
		Price:       DemoPrice,
		UserID:      userID,
		ClientIP:    clientIP,
	})
	if err != nil {
		return nil, "", err
	}

	if err := s.orders.MarkPayOrderNo(ctx, orderNo, created.Order.MerchantOrderNo); err != nil {
		facades.Log().Warningf("演示订单回写支付单号失败：%v", err)
	}

	return created, orderNo, nil
}

// PageMine 分页查询当前会员的演示订单。
// 入参 ctx 为请求上下文、userID 为会员 ID、pageNo/pageSize 为分页参数；返回列表与总数。
func (s *DemoOrderService) PageMine(ctx context.Context, userID types.ID, pageNo, pageSize int) ([]demoModel.DemoOrderModel, int64, error) {
	return s.orders.PageMine(ctx, userID, pageNo, pageSize)
}

// Detail 查询本人的演示订单详情，校验归属防止越权。
// 入参 ctx 为请求上下文、orderNo 为业务单号、userID 为会员 ID；返回订单模型。
func (s *DemoOrderService) Detail(ctx context.Context, orderNo string, userID types.ID) (*demoModel.DemoOrderModel, error) {
	m, err := s.orders.FindByOrderNo(ctx, orderNo)
	if err != nil {
		return nil, err
	}
	if m == nil || m.UserID != userID {
		return nil, baseErrs.WithMsg(baseErrs.BadRequest, "演示订单不存在")
	}

	return m, nil
}

// ResolveOrder 实现业务订单反查：pay 模块 C 端下单时按业务单号取金额与标题。
// 金额只信库里的值，归属不符直接拒绝，防止替他人付款。
// 入参 ctx 为请求上下文、bizOrderNo 为业务单号、userID 为当前会员；返回订单信息。
func (s *DemoOrderService) ResolveOrder(ctx context.Context, bizOrderNo string, userID types.ID) (*payService.BizOrderInfo, error) {
	m, err := s.orders.FindByOrderNo(ctx, bizOrderNo)
	if err != nil {
		return nil, err
	}
	if m == nil {
		return nil, baseErrs.WithMsg(baseErrs.BadRequest, "演示订单不存在")
	}
	if m.Status != demoModel.DemoOrderStatusPending {
		return nil, baseErrs.WithMsg(baseErrs.BadRequest, "演示订单当前状态不允许支付")
	}
	if userID > 0 && m.UserID != userID {
		return nil, baseErrs.WithMsg(baseErrs.BadRequest, "无权支付该演示订单")
	}

	return &payService.BizOrderInfo{
		Price:   m.Price,
		Subject: m.Subject,
		UserID:  m.UserID,
	}, nil
}

// OnOrderPaid 处理支付成功事件：按业务单号回写状态（where status=0，天然幂等）。
// 入参 ctx 为请求上下文、orderNo 为业务单号、payOrderNo 为支付单号、paidAtStr 为支付时间字符串；
// 处理失败只记日志，不影响回调应答。
func (s *DemoOrderService) OnOrderPaid(ctx context.Context, orderNo, payOrderNo, paidAtStr string) {
	paidAt := types.DateTime(time.Now())
	if t, err := time.Parse(time.RFC3339, paidAtStr); err == nil {
		paidAt = types.DateTime(t)
	}

	if err := s.orders.MarkPaid(ctx, orderNo, payOrderNo, paidAt); err != nil {
		facades.Log().Errorf("演示订单回写支付状态失败（%s）：%v", orderNo, err)
	}
}

// init 在包加载时把演示业务类型注册到 pay 模块的反查表，
// pay 模块对 demo 一无所知，方向为 demo → pay 单向依赖。
func init() {
	svc := NewDemoOrderService()
	payService.RegisterBizOrder(DemoBizType, svc.ResolveOrder)
}
