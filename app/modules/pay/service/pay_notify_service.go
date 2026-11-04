package service

import (
	"context"
	"encoding/json"
	nethttp "net/http"
	"time"

	"eqray-go-starter/app/facades"
	payChannel "eqray-go-starter/app/modules/pay/channel"
	payModel "eqray-go-starter/app/modules/pay/model"
	payRepository "eqray-go-starter/app/modules/pay/repository"
)

// NotifyLogRetentionDays 是回调日志保留天数，由定时任务按此清理。
const NotifyLogRetentionDays = 30

// PayNotifyService 负责渠道异步通知的验签、幂等落单与日志留痕。
type PayNotifyService struct {
	orders   *payRepository.PayOrderRepository
	logs     *payRepository.PayNotifyLogRepository
	channels *PayChannelService
	paySvc   *PayOrderService
}

// NewPayNotifyService 构造回调处理服务。
func NewPayNotifyService() *PayNotifyService {
	return &PayNotifyService{
		orders:   payRepository.NewPayOrderRepository(),
		logs:     payRepository.NewPayNotifyLogRepository(),
		channels: NewPayChannelService(),
		paySvc:   NewPayOrderService(),
	}
}

// HandlePayNotify 处理渠道支付异步通知，六步闭环：
//  1. 解析渠道定位客户端（渠道停用也处理，避免漏单）；
//  2. 验签 + 解密，失败落日志并返回失败应答；
//  3. 金额与订单比对，不一致落异常日志；
//  4. CAS 把订单流转为已支付，影响行数 = 1 判定首次；
//  5. 首次成功后派发 pay.order.paid 事件（幂等由监听器兜底）；
//  6. 按渠道格式返回成功 / 失败应答。
//
// 入参 channelCode 为路由中的渠道编码、r 为原始请求；返回应答报文。
func (s *PayNotifyService) HandlePayNotify(ctx context.Context, channelCode string, r *nethttp.Request) payChannel.NotifyAck {
	log := &payModel.PayNotifyLogModel{
		ChannelCode:    channelCode,
		Type:           payModel.PayNotifyTypePay,
		RequestHeaders: pickNotifyHeaders(r),
	}

	_, ch, err := s.channels.Resolve(ctx, channelCode, false)
	if err != nil {
		return s.fail(ctx, log, ch, err.Error())
	}

	result, err := ch.ParseNotify(ctx, r)
	if err != nil {
		return s.fail(ctx, log, ch, err.Error())
	}

	// 留档原始报文（微信为密文报文、支付宝为表单），便于排查
	log.MerchantOrderNo = result.MerchantOrderNo
	log.RequestBody = truncate(result.RawRequest, 65535)
	log.DecryptedBody = truncate(result.Decrypted, 65535)
	log.VerifyStatus = payModel.PayNotifyVerifyPassed

	order, err := s.orders.FindByMerchantOrderNo(ctx, result.MerchantOrderNo)
	if err != nil {
		return s.fail(ctx, log, ch, "查询订单失败："+err.Error())
	}
	if order == nil {
		return s.fail(ctx, log, ch, "订单不存在："+result.MerchantOrderNo)
	}

	// 金额一致性：回调金额与订单金额不符直接拒绝，标记人工介入
	if result.TradeState == payChannel.TradeStatePaid && result.Amount != 0 && result.Amount != order.Price {
		return s.fail(ctx, log, ch, "回调金额与订单金额不一致")
	}

	// 非支付成功通知确认收到即可，不改进订单状态
	if result.TradeState != payChannel.TradeStatePaid {
		log.HandleStatus = payModel.PayNotifyHandleOK
		s.saveLog(ctx, log)
		return ch.NotifyAck()
	}

	// 重复通知 / 回调与查单并发：CAS 影响行数判定，只有首次流转派发事件
	if err := s.paySvc.HandleChannelPaid(ctx, order, result); err != nil {
		return s.fail(ctx, log, ch, err.Error())
	}

	log.HandleStatus = payModel.PayNotifyHandleOK
	s.saveLog(ctx, log)

	return ch.NotifyAck()
}

// PurgeNotifyLogs 物理清理过期回调日志（定时任务调用）。
// 入参 ctx 为请求上下文；返回删除行数与写库错误。
func (s *PayNotifyService) PurgeNotifyLogs(ctx context.Context) (int64, error) {
	return s.logs.DeleteBefore(ctx, time.Now().AddDate(0, 0, -NotifyLogRetentionDays))
}

// fail 落一条失败日志并返回渠道要求的失败应答（触发渠道重试）。
// 入参 ctx 为上下文、log 为日志模型、ch 为渠道客户端（可能为 nil）、reason 为失败原因；返回失败应答。
func (s *PayNotifyService) fail(ctx context.Context, log *payModel.PayNotifyLogModel, ch payChannel.Channel, reason string) payChannel.NotifyAck {
	facades.Log().Warningf("支付回调处理失败（%s）：%s", log.ChannelCode, reason)

	log.ErrorMsg = truncate(reason, 500)
	s.saveLog(ctx, log)

	if ch != nil {
		return ch.NotifyFailAck(reason)
	}

	return payChannel.NotifyAck{Body: "fail", ContentType: "text/plain"}
}

// pickNotifyHeaders 提取回调的关键请求头留档，避免把整组头（cookie / token）写进日志。
// 入参 r 为原始请求；返回 JSON 字符串，序列化失败返回空串。
func pickNotifyHeaders(r *nethttp.Request) string {
	keys := []string{
		"Wechatpay-Serial", "Wechatpay-Timestamp", "Wechatpay-Nonce",
		"Wechatpay-Signature", "Content-Type", "User-Agent",
	}

	headers := make(map[string]string, len(keys))
	for _, key := range keys {
		if value := r.Header.Get(key); value != "" {
			headers[key] = value
		}
	}

	data, err := json.Marshal(headers)
	if err != nil {
		return ""
	}

	return string(data)
}

// saveLog 追加回调日志；日志写失败不影响主流程，只记应用日志。
// 入参 ctx 为上下文、log 为日志模型。
func (s *PayNotifyService) saveLog(ctx context.Context, log *payModel.PayNotifyLogModel) {
	if err := s.logs.Insert(ctx, log); err != nil {
		facades.Log().Errorf("支付回调日志写入失败：%v", err)
	}
}
