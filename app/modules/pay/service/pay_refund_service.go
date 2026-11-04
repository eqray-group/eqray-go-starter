package service

import (
	"context"
	"strconv"
	"time"

	baseErrs "eqray-go-starter/app/modules/base/errs"
	"eqray-go-starter/app/modules/base/support"
	"eqray-go-starter/app/modules/base/types"
	payChannel "eqray-go-starter/app/modules/pay/channel"
	payErrs "eqray-go-starter/app/modules/pay/errs"
	payEvents "eqray-go-starter/app/modules/pay/events"
	payModel "eqray-go-starter/app/modules/pay/model"
	payRepository "eqray-go-starter/app/modules/pay/repository"
)

// RefundCreateInput 是发起退款的入参（后台直接发起，无人工审核）。
type RefundCreateInput struct {
	// OrderID 支付订单 ID
	OrderID types.ID
	// RefundPrice 退款金额（分）
	RefundPrice int64
	// Reason 退款原因
	Reason string
}

// PayRefundListQuery 是退款单列表的查询条件。
type PayRefundListQuery struct {
	// MerchantRefundNo 退款单号
	MerchantRefundNo string
	// MerchantOrderNo 支付单号
	MerchantOrderNo string
	// Status 退款状态过滤
	Status *payModel.PayRefundStatus
	// PageNo / PageSize 分页参数
	PageNo   int
	PageSize int
}

// PayRefundService 负责退款单的创建、查询与状态收敛。
type PayRefundService struct {
	refunds  *payRepository.PayRefundRepository
	orders   *payRepository.PayOrderRepository
	channels *PayChannelService
}

// NewPayRefundService 构造退款服务。
func NewPayRefundService() *PayRefundService {
	return &PayRefundService{
		refunds:  payRepository.NewPayRefundRepository(),
		orders:   payRepository.NewPayOrderRepository(),
		channels: NewPayChannelService(),
	}
}

// Create 发起退款：校验订单已支付、金额不超可退余额，创建退款单后直接调渠道。
// 支付宝同步返回终态；微信返回「受理中」，由 SyncRefunding 定时查单收敛。
// 入参 in 为退款入参；返回退款单 ID。
func (s *PayRefundService) Create(ctx context.Context, in RefundCreateInput) (types.ID, error) {
	if in.RefundPrice <= 0 {
		return 0, baseErrs.WithMsg(baseErrs.BadRequest, "退款金额必须大于 0")
	}

	order, err := s.orders.FindByID(ctx, in.OrderID)
	if err != nil {
		return 0, err
	}
	if order == nil {
		return 0, baseErrs.New(payErrs.PayOrderNotExists, in.OrderID)
	}
	if order.Status != payModel.PayOrderStatusPaid {
		return 0, baseErrs.New(payErrs.PayOrderStatusDenied, order.Status)
	}

	refunded, err := s.orders.SumRefunded(ctx, order.ID)
	if err != nil {
		return 0, err
	}
	if refunded+in.RefundPrice > order.Price {
		return 0, baseErrs.New(payErrs.PayRefundPriceInvalid)
	}

	_, ch, err := s.channels.Resolve(ctx, order.ChannelCode, false)
	if err != nil {
		return 0, err
	}

	refund := &payModel.PayRefundModel{
		MerchantRefundNo: strconv.FormatInt(int64(support.NewID()), 10),
		OrderID:          order.ID,
		MerchantOrderNo:  order.MerchantOrderNo,
		ChannelID:        order.ChannelID,
		ChannelCode:      order.ChannelCode,
		RefundPrice:      in.RefundPrice,
		Reason:           in.Reason,
		Status:           payModel.PayRefundStatusPending,
	}
	if err := s.refunds.Insert(ctx, refund); err != nil {
		return 0, err
	}

	result, err := ch.Refund(ctx, payChannel.RefundRequest{
		MerchantRefundNo: refund.MerchantRefundNo,
		MerchantOrderNo:  order.MerchantOrderNo,
		RefundPrice:      in.RefundPrice,
		TotalPrice:       order.Price,
		Reason:           in.Reason,
	})
	if err != nil {
		// 渠道调用失败：退款单标记失败，允许重新发起
		if err := s.refunds.UpdateFields(ctx, refund.ID, map[string]any{
			"status":    int8(payModel.PayRefundStatusFailed),
			"error_msg": truncate(err.Error(), 500),
		}, operatorName(ctx)); err != nil {
			return 0, err
		}

		return 0, baseErrs.New(payErrs.PayRefundFailed, err.Error())
	}

	if err := s.settleRefund(ctx, refund, result, order); err != nil {
		return 0, err
	}

	return refund.ID, nil
}

// Page 分页查询退款单。
// 入参 q 为查询条件；返回列表与总数。
func (s *PayRefundService) Page(ctx context.Context, q PayRefundListQuery) ([]payModel.PayRefundModel, int64, error) {
	return s.refunds.Paginate(ctx, payRepository.PayRefundFilter{
		MerchantRefundNo: q.MerchantRefundNo,
		MerchantOrderNo:  q.MerchantOrderNo,
		Status:           q.Status,
	}, q.PageNo, q.PageSize)
}

// Detail 查询退款单详情。
// 入参 id 为退款单 ID；返回退款单模型，不存在返回业务错误码。
func (s *PayRefundService) Detail(ctx context.Context, id types.ID) (*payModel.PayRefundModel, error) {
	m, err := s.refunds.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if m == nil {
		return nil, baseErrs.New(payErrs.PayRefundNotExists, id)
	}

	return m, nil
}

// SyncRefunding 同步「退款中」的退款单（定时任务调用）：
// 微信查单收敛终态；支付宝退款同步返回终态、不会有退款中状态，查询不支持时跳过。
// 入参 ctx 为请求上下文；返回本次收敛数量与写库错误。
func (s *PayRefundService) SyncRefunding(ctx context.Context) (int, error) {
	// 只处理创建超过 1 分钟的单，避免与刚提交的退款请求撞车
	before := types.DateTime(time.Now().Add(-1 * time.Minute))
	refunding, err := s.refunds.ListRefunding(ctx, before, 200)
	if err != nil {
		return 0, err
	}

	settled := 0
	for _, refund := range refunding {
		_, ch, err := s.channels.Resolve(ctx, refund.ChannelCode, false)
		if err != nil {
			continue
		}

		result, err := ch.QueryRefund(ctx, refund.MerchantRefundNo, refund.MerchantOrderNo)
		if err != nil {
			// 渠道不支持退款查单（如支付宝）：跳过，等待人工确认
			continue
		}

		order, err := s.orders.FindByID(ctx, refund.OrderID)
		if err != nil || order == nil {
			continue
		}

		if err := s.settleRefund(ctx, &refund, result, order); err != nil {
			return settled, err
		}
		settled++
	}

	return settled, nil
}

// settleRefund 按渠道退款结果收敛退款单状态并回写订单退款进度。
// 退款成功时更新 pay_order.refund_status（部分 / 全额）并派发退款成功事件。
// 入参 refund 为退款单、result 为渠道结果、order 为支付订单；返回写库错误。
func (s *PayRefundService) settleRefund(ctx context.Context, refund *payModel.PayRefundModel, result payChannel.RefundResult, order *payModel.PayOrderModel) error {
	updater := operatorName(ctx)
	if updater == "" {
		updater = "system"
	}

	fields := map[string]any{
		"channel_refund_no": result.ChannelRefundNo,
		"channel_extra":     result.Raw,
	}

	switch result.State {
	case payChannel.RefundStateSuccess:
		fields["status"] = int8(payModel.PayRefundStatusSuccess)
		fields["success_time"] = time.Now()
		if err := s.refunds.UpdateFields(ctx, refund.ID, fields, updater); err != nil {
			return err
		}

		// 重算订单整体退款进度
		refunded, err := s.orders.SumRefunded(ctx, order.ID)
		if err != nil {
			return err
		}
		refundStatus := payModel.PayOrderRefundStatusPartial
		if refunded >= order.Price {
			refundStatus = payModel.PayOrderRefundStatusAll
		}
		if err := s.orders.UpdateFields(ctx, order.ID, map[string]any{
			"refund_status": int8(refundStatus),
		}, updater); err != nil {
			return err
		}

		payEvents.DispatchRefundSucceeded(&payEvents.RefundSucceededPayload{
			BizType:     order.BizType,
			BizOrderNo:  order.BizOrderNo,
			OrderNo:     order.MerchantOrderNo,
			RefundNo:    refund.MerchantRefundNo,
			ChannelCode: refund.ChannelCode,
			RefundPrice: refund.RefundPrice,
			Reason:      refund.Reason,
			SucceededAt: time.Now().Format(time.RFC3339),
		})

	case payChannel.RefundStateFailed:
		fields["status"] = int8(payModel.PayRefundStatusFailed)
		if err := s.refunds.UpdateFields(ctx, refund.ID, fields, updater); err != nil {
			return err
		}

	default:
		// 退款受理中：流转 0 → 10，后续由 SyncRefunding 收敛
		if _, err := s.refunds.UpdateStatusCAS(ctx, refund.ID, []int8{
			int8(payModel.PayRefundStatusPending),
		}, payModel.PayRefundStatusRefunding, updater); err != nil {
			return err
		}
	}

	return nil
}

// truncate 截断字符串到指定 rune 长度，用于落库前限制错误信息长度。
// 入参 s 为原字符串、max 为最大长度；返回截断后的字符串。
func truncate(s string, max int) string {
	runes := []rune(s)
	if len(runes) <= max {
		return s
	}

	return string(runes[:max])
}
