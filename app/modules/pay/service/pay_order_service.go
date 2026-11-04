package service

import (
	"context"
	"encoding/json"
	"errors"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/goravel/framework/contracts/database/db"

	"eqray-go-starter/app/facades"
	baseErrs "eqray-go-starter/app/modules/base/errs"
	"eqray-go-starter/app/modules/base/support"
	"eqray-go-starter/app/modules/base/types"
	infraService "eqray-go-starter/app/modules/infra/service"
	payChannel "eqray-go-starter/app/modules/pay/channel"
	payErrs "eqray-go-starter/app/modules/pay/errs"
	payEvents "eqray-go-starter/app/modules/pay/events"
	payModel "eqray-go-starter/app/modules/pay/model"
	payRepository "eqray-go-starter/app/modules/pay/repository"
)

// OrderExpireMinutes 是支付单默认有效期（分钟），超时由定时任务关闭。
const OrderExpireMinutes = 30

// 支付联调测试单参数：金额固定 1 分，业务类型独立，便于和真实业务单区分。
const (
	// TestBizType 联调测试单的业务类型
	TestBizType = "pay_test"
	// TestPrice 联调测试单金额（分）
	TestPrice int64 = 1
)

// CreateTestOrderInput 是后台支付联调下单的入参。
type CreateTestOrderInput struct {
	// UserID 付款会员 ID
	UserID types.ID
	// ChannelCode 渠道编码
	ChannelCode string
}

// TestOrderOutcome 是批量联调中单个渠道的结果。
// 某个渠道失败（如钱包余额不足）不影响其他渠道，失败原因放在 Error 里。
type TestOrderOutcome struct {
	// ChannelCode 渠道编码
	ChannelCode string `json:"channelCode"`
	// OrderID 支付单 ID，失败时为 0
	OrderID types.ID `json:"id"`
	// MerchantOrderNo 支付单号，失败时为空
	MerchantOrderNo string `json:"merchantOrderNo"`
	// Price 金额（分）
	Price int64 `json:"price"`
	// Status 订单状态
	Status payModel.PayOrderStatus `json:"status"`
	// PayInfo 前端拉起参数
	PayInfo payChannel.CreateOrderResult `json:"payInfo"`
	// Error 失败原因，成功时为空
	Error string `json:"error"`
}

// CreateOrderInput 是创建支付订单的入参。
// 金额 / 标题由业务侧提供（服务间调用）或由反查函数得出（C 端调用），pay 模块不信任前端金额。
type CreateOrderInput struct {
	// BizType 业务类型，如 demo_order
	BizType string
	// BizOrderNo 业务方订单号
	BizOrderNo string
	// ChannelCode 渠道编码
	ChannelCode string
	// Subject 商品标题
	Subject string
	// Description 商品描述
	Description string
	// Price 金额（分）
	Price int64
	// UserID 下单会员 ID
	UserID types.ID
	// ClientIP 下单 IP
	ClientIP string
	// OpenID 渠道侧付款人标识（微信 JSAPI / 小程序必传）
	OpenID string
	// ReturnURL 同步跳转地址（支付宝网页类支付用），可空
	ReturnURL string
}

// CreateOrderResult 是下单结果：支付单 + 前端拉起参数。
type CreateOrderResult struct {
	// Order 支付单模型
	Order *payModel.PayOrderModel
	// PayInfo 前端拉起支付所需的统一结果（kind / url / params）
	PayInfo payChannel.CreateOrderResult
}

// PayOrderListQuery 是支付订单列表的查询条件。
type PayOrderListQuery struct {
	// MerchantOrderNo 支付单号
	MerchantOrderNo string
	// BizType 业务类型
	BizType string
	// BizOrderNo 业务单号
	BizOrderNo string
	// ChannelCode 渠道编码
	ChannelCode string
	// Status 订单状态过滤
	Status *payModel.PayOrderStatus
	// BeginTime / EndTime 创建时间区间（yyyy-MM-dd HH:mm:ss）
	BeginTime string
	EndTime   string
	// PageNo / PageSize 分页参数
	PageNo   int
	PageSize int
}

// PayOrderService 负责支付订单的下单、查单、关单与状态机流转。
type PayOrderService struct {
	orders   *payRepository.PayOrderRepository
	wallets  *payRepository.PayWalletRepository
	channels *PayChannelService
	configs  *infraService.ConfigService
}

// NewPayOrderService 构造支付订单服务。
func NewPayOrderService() *PayOrderService {
	return &PayOrderService{
		orders:   payRepository.NewPayOrderRepository(),
		wallets:  payRepository.NewPayWalletRepository(),
		channels: NewPayChannelService(),
		configs:  infraService.NewConfigService(),
	}
}

// CreateOrder 创建支付订单并向渠道下单：
//  1. 同一「业务单 + 渠道」已有未支付订单则复用（重试沿用同一商户单号）；
//  2. 已支付订单直接拒绝，防止重复收款；
//  3. 新订单先落库（status=0）再调渠道，渠道失败订单保留，允许重试。
//
// 入参 in 为下单入参；返回支付单与前端拉起参数。
func (s *PayOrderService) CreateOrder(ctx context.Context, in CreateOrderInput) (*CreateOrderResult, error) {
	if in.Price <= 0 {
		return nil, baseErrs.WithMsg(baseErrs.BadRequest, "支付金额必须大于 0")
	}

	channelModel, ch, err := s.channels.Resolve(ctx, in.ChannelCode, true)
	if err != nil {
		return nil, err
	}

	// 钱包 / 模拟支付是站内渠道，不产生异步通知地址，也不要求配置 APP_URL
	notifyURL := ""
	if !payChannel.IsInternal(in.ChannelCode) {
		notifyURL, err = s.channels.NotifyURL(in.ChannelCode)
		if err != nil {
			return nil, err
		}
	}

	// uk_biz_order 幂等：同业务单同渠道复用未支付订单，重试不换单号
	existing, err := s.orders.FindByBiz(ctx, in.BizType, in.BizOrderNo, in.ChannelCode)
	if err != nil {
		return nil, err
	}

	order := existing
	if order == nil {
		order = &payModel.PayOrderModel{
			MerchantOrderNo: strconv.FormatInt(int64(support.NewID()), 10),
			BizType:         in.BizType,
			BizOrderNo:      in.BizOrderNo,
			ChannelID:       channelModel.ID,
			ChannelCode:     channelModel.Code,
			Subject:         in.Subject,
			Description:     in.Description,
			Price:           in.Price,
			Currency:        "CNY",
			Status:          payModel.PayOrderStatusPending,
			UserID:          in.UserID,
			ClientIP:        in.ClientIP,
			NotifyURL:       notifyURL,
			ExpireTime:      types.DateTime(time.Now().Add(OrderExpireMinutes * time.Minute)),
		}
		if err := s.orders.Insert(ctx, order); err != nil {
			// 并发重复下单撞唯一键：回查复用
			var be *baseErrs.BusinessError
			if errors.As(err, &be) && be.Code() == payErrs.PayOrderDuplicated.Code {
				order, err = s.orders.FindByBiz(ctx, in.BizType, in.BizOrderNo, in.ChannelCode)
				if err != nil || order == nil {
					return nil, baseErrs.New(payErrs.PayOrderDuplicated)
				}
			} else {
				return nil, err
			}
		}
	}

	if order.Status == payModel.PayOrderStatusPaid {
		return nil, baseErrs.New(payErrs.PayOrderStatusDenied, order.Status)
	}
	if order.Status == payModel.PayOrderStatusClosed {
		return nil, baseErrs.New(payErrs.PayOrderStatusDenied, order.Status)
	}

	// 已过期的支付中订单先关闭，避免向渠道重复下单
	if order.Status == payModel.PayOrderStatusPaying && time.Time(order.ExpireTime).Before(time.Now()) {
		return nil, baseErrs.New(payErrs.PayOrderStatusDenied, order.Status)
	}

	// 余额支付：站内扣款并直接完成，不走渠道下单
	if payChannel.IsWallet(order.ChannelCode) {
		return s.payByWallet(ctx, order, in.UserID)
	}

	result, err := ch.CreateOrder(ctx, payChannel.CreateOrderRequest{
		MerchantOrderNo: order.MerchantOrderNo,
		Subject:         order.Subject,
		Description:     order.Description,
		Price:           order.Price,
		NotifyURL:       notifyURL,
		ReturnURL:       in.ReturnURL,
		OpenID:          in.OpenID,
		ClientIP:        in.ClientIP,
	})
	if err != nil {
		return nil, baseErrs.New(payErrs.PayOrderCreateFailed, err.Error())
	}

	// 下单成功进入「支付中」，回写渠道返回的拉起参数留档
	payInfoJSON, _ := json.Marshal(result)
	if err := s.orders.UpdateFields(ctx, order.ID, map[string]any{
		"status":        int8(payModel.PayOrderStatusPaying),
		"channel_extra": string(payInfoJSON),
		"notify_url":    notifyURL,
	}, operatorName(ctx)); err != nil {
		return nil, err
	}
	order.Status = payModel.PayOrderStatusPaying
	order.NotifyURL = notifyURL

	return &CreateOrderResult{Order: order, PayInfo: result}, nil
}

// CreateOrderForClient 是 C 端下单入口：金额与标题由业务模块反查得出，请求体不接受金额。
// 入参 bizType/bizOrderNo 定位业务单、channelCode 为渠道、userID 为当前会员、clientIP/returnURL 透传渠道；
// 返回支付单与前端拉起参数。
func (s *PayOrderService) CreateOrderForClient(ctx context.Context, bizType, bizOrderNo, channelCode string, userID types.ID, clientIP, returnURL string) (*CreateOrderResult, error) {
	info, err := resolveBizOrder(ctx, bizType, bizOrderNo, userID)
	if err != nil {
		return nil, err
	}
	if info.UserID > 0 && info.UserID != userID {
		return nil, baseErrs.WithMsg(baseErrs.BadRequest, "无权支付该订单")
	}

	// 同步跳转地址只允许回本站，避免支付完成后被导流到外站
	if err := validateReturnURL(returnURL); err != nil {
		return nil, err
	}

	// 小程序与公众号的 openid 不通用，按渠道选择对应社交类型反查
	openID := ""
	if socialType := payChannel.SocialTypeForCode(channelCode); socialType != "" {
		openID, err = s.resolveOpenID(ctx, userID, socialType)
		if err != nil {
			return nil, err
		}
	}

	return s.CreateOrder(ctx, CreateOrderInput{
		BizType:     bizType,
		BizOrderNo:  bizOrderNo,
		ChannelCode: channelCode,
		Subject:     info.Subject,
		Description: info.Description,
		Price:       info.Price,
		UserID:      userID,
		ClientIP:    clientIP,
		OpenID:      openID,
		ReturnURL:   returnURL,
	})
}

// CreateTestOrder 在后台创建一笔 1 分钱的联调测试单并发起支付。
// 业务类型固定 pay_test，不关联任何业务表；钱包渠道会真实扣会员余额。
// 入参 in 为会员与渠道；返回支付单与前端拉起参数。
func (s *PayOrderService) CreateTestOrder(ctx context.Context, in CreateTestOrderInput) (*CreateOrderResult, error) {
	if in.UserID == 0 {
		return nil, baseErrs.WithMsg(baseErrs.BadRequest, "请输入付款会员 ID")
	}
	if strings.TrimSpace(in.ChannelCode) == "" {
		return nil, baseErrs.WithMsg(baseErrs.BadRequest, "请选择支付渠道")
	}

	return s.CreateOrder(ctx, CreateOrderInput{
		BizType:     TestBizType,
		BizOrderNo:  "T" + strconv.FormatInt(int64(support.NewID()), 10),
		ChannelCode: strings.TrimSpace(in.ChannelCode),
		Subject:     "支付联调测试（1 分）",
		Price:       TestPrice,
		UserID:      in.UserID,
		ClientIP:    "127.0.0.1",
	})
}

// CreateTestOrders 批量创建联调测试单：channelCodes 为空时取全部启用渠道。
// 单个渠道失败不影响其他渠道，失败原因写在对应结果的 Error 字段。
// 入参 userID 为付款会员、channelCodes 为渠道编码列表；返回每个渠道的结果。
func (s *PayOrderService) CreateTestOrders(ctx context.Context, userID types.ID, channelCodes []string) ([]TestOrderOutcome, error) {
	if userID == 0 {
		return nil, baseErrs.WithMsg(baseErrs.BadRequest, "请输入付款会员 ID")
	}

	codes := make([]string, 0, len(channelCodes))
	for _, code := range channelCodes {
		if code = strings.TrimSpace(code); code != "" {
			codes = append(codes, code)
		}
	}

	// 不指定渠道时覆盖全部启用渠道，便于一键回归
	if len(codes) == 0 {
		list, err := s.channels.EnabledChannels(ctx)
		if err != nil {
			return nil, err
		}
		for _, m := range list {
			codes = append(codes, m.Code)
		}
	}
	if len(codes) == 0 {
		return nil, baseErrs.WithMsg(baseErrs.BadRequest, "没有启用中的支付渠道，请先到渠道页启用")
	}

	outcomes := make([]TestOrderOutcome, 0, len(codes))
	for _, code := range codes {
		created, err := s.CreateTestOrder(ctx, CreateTestOrderInput{UserID: userID, ChannelCode: code})
		if err != nil {
			outcomes = append(outcomes, TestOrderOutcome{ChannelCode: code, Error: err.Error()})
			continue
		}

		outcomes = append(outcomes, TestOrderOutcome{
			ChannelCode:     code,
			OrderID:         created.Order.ID,
			MerchantOrderNo: created.Order.MerchantOrderNo,
			Price:           created.Order.Price,
			Status:          created.Order.Status,
			PayInfo:         created.PayInfo,
		})
	}

	return outcomes, nil
}

// AdminOrder 按支付单号查询订单（后台联调，不校验会员归属）。
// 入参 merchantOrderNo 为支付单号；返回订单模型，不存在返回业务错误码。
func (s *PayOrderService) AdminOrder(ctx context.Context, merchantOrderNo string) (*payModel.PayOrderModel, error) {
	m, err := s.orders.FindByMerchantOrderNo(ctx, merchantOrderNo)
	if err != nil {
		return nil, err
	}
	if m == nil {
		return nil, baseErrs.New(payErrs.PayOrderNotExists, merchantOrderNo)
	}

	return m, nil
}

// SyncByMerchantOrderNo 按单号主动向渠道查单并同步（后台联调用，不校验会员归属）。
// 入参 merchantOrderNo 为支付单号；返回同步后的订单。
func (s *PayOrderService) SyncByMerchantOrderNo(ctx context.Context, merchantOrderNo string) (*payModel.PayOrderModel, error) {
	order, err := s.AdminOrder(ctx, merchantOrderNo)
	if err != nil {
		return nil, err
	}

	return s.syncOrder(ctx, order)
}

// MockPayByMerchantOrderNo 后台对模拟渠道订单执行「模拟支付成功」（联调用，不校验会员归属）。
// 入参 merchantOrderNo 为支付单号；返回最新订单。
func (s *PayOrderService) MockPayByMerchantOrderNo(ctx context.Context, merchantOrderNo string) (*payModel.PayOrderModel, error) {
	if s.configs.GetString(ctx, "pay.mock.enabled", "0") != "1" {
		return nil, baseErrs.New(payErrs.PayMockDisabled)
	}

	order, err := s.AdminOrder(ctx, merchantOrderNo)
	if err != nil {
		return nil, err
	}
	if !payChannel.IsMock(order.ChannelCode) {
		return nil, baseErrs.New(payErrs.PayMockChannelDenied)
	}
	if order.Status == payModel.PayOrderStatusPaid {
		return order, nil
	}

	if err := s.markPaid(ctx, order, paidInfo{
		ChannelOrderNo: "MOCK" + order.MerchantOrderNo,
		ChannelUser:    "mock",
		Raw:            `{"mock":true}`,
		SuccessTime:    time.Now().Format(time.RFC3339),
	}); err != nil {
		return nil, err
	}

	return s.orders.FindByID(ctx, order.ID)
}

// resolveOpenID 取当前会员在指定社交渠道的 openid（小程序 / 公众号）。
// pay 模块不 import member 模型层，通过启动时注册的反查函数取（见 RegisterOpenIDResolver）。
// 入参 userID 为会员 ID、socialType 为社交渠道类型；返回 openid 或业务错误码。
func (s *PayOrderService) resolveOpenID(ctx context.Context, userID types.ID, socialType string) (string, error) {
	value, ok := openIDResolvers.Load(socialType)
	if !ok {
		return "", baseErrs.New(payErrs.PayMemberOpenidMissing)
	}

	openID, err := value.(OpenIDResolver)(ctx, userID)
	if err != nil {
		return "", err
	}
	if strings.TrimSpace(openID) == "" {
		return "", baseErrs.New(payErrs.PayMemberOpenidMissing)
	}

	return openID, nil
}

// validateReturnURL 校验同步跳转地址：必须是 http/https 且与 APP_URL 同域。
// 入参 raw 为空表示不需要跳转（扫码 / 小程序 / APP 场景）；返回业务错误码。
func validateReturnURL(raw string) error {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}

	target, err := url.Parse(raw)
	if err != nil || target.Host == "" || (target.Scheme != "http" && target.Scheme != "https") {
		return baseErrs.New(payErrs.PayReturnUrlInvalid)
	}

	appURL := strings.TrimRight(strings.TrimSpace(facades.Config().GetString("http.url")), "/")
	if appURL == "" {
		// 未配置 APP_URL 时无法判断同域，交给渠道侧配置兜底
		return nil
	}

	site, err := url.Parse(appURL)
	if err != nil || site.Host == "" {
		return nil
	}
	if !strings.EqualFold(site.Host, target.Host) {
		return baseErrs.New(payErrs.PayReturnUrlInvalid)
	}

	return nil
}

// OpenIDResolver 由会员模块注册：按会员 ID 取某社交渠道的 openid。
// 入参 userID 为会员 ID；返回 openid（未绑定返回空串与业务错误）。
type OpenIDResolver func(ctx context.Context, userID types.ID) (string, error)

// openIDResolvers 是社交类型 → openid 反查函数的注册表。
var openIDResolvers sync.Map

// RegisterOpenIDResolver 注册 openid 反查函数，由会员模块在启动时调用。
// 入参 socialType 为社交渠道类型（wechat_mini）、resolver 为反查函数。
func RegisterOpenIDResolver(socialType string, resolver OpenIDResolver) {
	openIDResolvers.Store(socialType, resolver)
}

// Page 分页查询支付订单。
// 入参 q 为查询条件；返回列表与总数。
func (s *PayOrderService) Page(ctx context.Context, q PayOrderListQuery) ([]payModel.PayOrderModel, int64, error) {
	return s.orders.Paginate(ctx, payRepository.PayOrderFilter{
		MerchantOrderNo: q.MerchantOrderNo,
		BizType:         q.BizType,
		BizOrderNo:      q.BizOrderNo,
		ChannelCode:     q.ChannelCode,
		Status:          q.Status,
		BeginTime:       q.BeginTime,
		EndTime:         q.EndTime,
	}, q.PageNo, q.PageSize)
}

// Detail 查询支付订单详情。
// 入参 id 为订单 ID；返回订单模型，不存在返回业务错误码。
func (s *PayOrderService) Detail(ctx context.Context, id types.ID) (*payModel.PayOrderModel, error) {
	m, err := s.orders.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if m == nil {
		return nil, baseErrs.New(payErrs.PayOrderNotExists, id)
	}

	return m, nil
}

// UserOrder 按单号查询本人支付订单，C 端轮询支付结果用。
// 入参 merchantOrderNo 为支付单号、userID 为会员 ID；返回订单模型，不存在返回业务错误码。
func (s *PayOrderService) UserOrder(ctx context.Context, merchantOrderNo string, userID types.ID) (*payModel.PayOrderModel, error) {
	m, err := s.orders.FindUserOrder(ctx, merchantOrderNo, userID)
	if err != nil {
		return nil, err
	}
	if m == nil {
		return nil, baseErrs.New(payErrs.PayOrderNotExists, merchantOrderNo)
	}

	return m, nil
}

// Sync 主动向渠道查单并同步本地状态（管理端「同步」按钮与 C 端轮询共用）。
// 入参 id 为订单 ID；返回同步后的订单模型。
func (s *PayOrderService) Sync(ctx context.Context, id types.ID) (*payModel.PayOrderModel, error) {
	order, err := s.Detail(ctx, id)
	if err != nil {
		return nil, err
	}

	return s.syncOrder(ctx, order)
}

// SyncUserOrder 主动查单并校验归属，C 端轮询支付结果用。
// 入参 merchantOrderNo 为支付单号、userID 为会员 ID；返回同步后的订单模型。
func (s *PayOrderService) SyncUserOrder(ctx context.Context, merchantOrderNo string, userID types.ID) (*payModel.PayOrderModel, error) {
	order, err := s.UserOrder(ctx, merchantOrderNo, userID)
	if err != nil {
		return nil, err
	}

	return s.syncOrder(ctx, order)
}

// syncOrder 按渠道查单结果推进本地状态机。
// 入参 order 为本地订单；返回同步后的订单模型。
func (s *PayOrderService) syncOrder(ctx context.Context, order *payModel.PayOrderModel) (*payModel.PayOrderModel, error) {
	if order.Status == payModel.PayOrderStatusPaid || order.Status == payModel.PayOrderStatusClosed {
		return order, nil
	}
	// 待支付说明还没向渠道下单，没有可查的单号，直接返回本地状态
	if order.Status == payModel.PayOrderStatusPending {
		return order, nil
	}

	_, ch, err := s.channels.Resolve(ctx, order.ChannelCode, false)
	if err != nil {
		return nil, err
	}

	result, err := ch.QueryOrder(ctx, order.MerchantOrderNo)
	if err != nil {
		return nil, err
	}

	switch result.TradeState {
	case payChannel.TradeStatePaid:
		if err := s.markPaid(ctx, order, paidInfo{
			ChannelOrderNo: result.ChannelOrderNo,
			ChannelUser:    result.ChannelUser,
			Raw:            result.Raw,
			SuccessTime:    result.SuccessTime,
		}); err != nil {
			return nil, err
		}
	case payChannel.TradeStateClosed:
		if _, err := s.orders.UpdateStatusCAS(ctx, order.ID, []int8{
			int8(payModel.PayOrderStatusPending), int8(payModel.PayOrderStatusPaying),
		}, payModel.PayOrderStatusClosed, operatorName(ctx)); err != nil {
			return nil, err
		}
	}

	return s.orders.FindByID(ctx, order.ID)
}

// Close 手动关闭订单：仅待支付 / 支付中可关，先关渠道侧再关本地。
// 入参 id 为订单 ID；返回写库错误或业务错误码。
func (s *PayOrderService) Close(ctx context.Context, id types.ID) error {
	order, err := s.Detail(ctx, id)
	if err != nil {
		return err
	}
	if order.Status != payModel.PayOrderStatusPending && order.Status != payModel.PayOrderStatusPaying {
		return baseErrs.New(payErrs.PayOrderStatusDenied, order.Status)
	}

	_, ch, err := s.channels.Resolve(ctx, order.ChannelCode, false)
	if err != nil {
		return err
	}
	if err := ch.CloseOrder(ctx, order.MerchantOrderNo); err != nil {
		return err
	}

	affected, err := s.orders.UpdateStatusCAS(ctx, order.ID, []int8{
		int8(payModel.PayOrderStatusPending), int8(payModel.PayOrderStatusPaying),
	}, payModel.PayOrderStatusClosed, operatorName(ctx))
	if err != nil {
		return err
	}
	if affected == 0 {
		return baseErrs.New(payErrs.PayOrderStatusDenied, order.Status)
	}

	return nil
}

// CloseExpired 关闭全部超时未完成订单（渠道关单失败也继续关本地，由渠道侧超时兜底）。
// 入参 ctx 为请求上下文；返回关闭数量与写库错误。
func (s *PayOrderService) CloseExpired(ctx context.Context) (int, error) {
	expired, err := s.orders.ListExpired(ctx, types.Now(), 200)
	if err != nil {
		return 0, err
	}

	closed := 0
	for _, order := range expired {
		if _, ch, err := s.channels.Resolve(ctx, order.ChannelCode, false); err == nil {
			_ = ch.CloseOrder(ctx, order.MerchantOrderNo)
		}

		affected, err := s.orders.UpdateStatusCAS(ctx, order.ID, []int8{
			int8(payModel.PayOrderStatusPending), int8(payModel.PayOrderStatusPaying),
		}, payModel.PayOrderStatusClosed, "system")
		if err != nil {
			return closed, err
		}
		if affected > 0 {
			closed++
		}
	}

	return closed, nil
}

// payByWallet 余额支付：在事务内扣减余额、写流水并把订单置为已支付。
// 扣款与订单状态必须同事务，任一步失败整体回滚，避免「扣了钱订单没变」。
// 入参 order 为本地订单、userID 为付款会员；返回支付单与前端结果（kind=wallet）。
func (s *PayOrderService) payByWallet(ctx context.Context, order *payModel.PayOrderModel, userID types.ID) (*CreateOrderResult, error) {
	if userID == 0 {
		userID = order.UserID
	}
	if userID == 0 {
		return nil, baseErrs.WithMsg(baseErrs.BadRequest, "缺少付款会员信息")
	}

	paidAt := time.Now()
	channelOrderNo := "WALLET" + order.MerchantOrderNo

	err := facades.DB().WithContext(ctx).Transaction(func(tx db.Tx) error {
		wallet, err := s.wallets.FindByUserIDForUpdate(ctx, tx, userID)
		if err != nil {
			return err
		}
		if wallet == nil {
			return baseErrs.New(payErrs.PayWalletNotExists, userID)
		}
		if wallet.Status != payModel.PayWalletStatusEnabled {
			return baseErrs.New(payErrs.PayWalletDisabled)
		}
		if wallet.Balance < order.Price {
			return baseErrs.New(payErrs.PayWalletInsufficient, wallet.Balance)
		}

		wallet.Balance -= order.Price
		wallet.TotalExpense += order.Price
		wallet.Updater = "system"
		wallet.UpdatedAt = types.DateTime(paidAt)
		if err := s.wallets.Update(ctx, tx, wallet); err != nil {
			return err
		}

		if err := s.wallets.InsertTransaction(ctx, tx, &payModel.PayWalletTransactionModel{
			ID:           support.NewID(),
			WalletID:     wallet.ID,
			UserID:       userID,
			Type:         payModel.PayWalletTypeConsume,
			Amount:       -order.Price,
			BalanceAfter: wallet.Balance,
			BizType:      order.BizType,
			BizOrderNo:   order.BizOrderNo,
			Remark:       "钱包余额支付",
			Creator:      "system",
			CreatedAt:    types.DateTime(paidAt),
		}); err != nil {
			return err
		}

		affected, err := s.orders.UpdatePaidTx(ctx, tx, order.ID, channelOrderNo, "", paidAt)
		if err != nil {
			return err
		}
		if affected == 0 {
			return baseErrs.New(payErrs.PayOrderStatusDenied, order.Status)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	order.Status = payModel.PayOrderStatusPaid
	order.ChannelOrderNo = channelOrderNo
	successTime := types.DateTime(paidAt)
	order.SuccessTime = &successTime

	s.dispatchPaid(order, paidAt.Format(time.RFC3339))

	return &CreateOrderResult{
		Order:   order,
		PayInfo: payChannel.CreateOrderResult{Kind: payChannel.KindWallet},
	}, nil
}

// MockPay 对模拟支付渠道的订单执行「模拟支付成功」，仅联调使用。
// 需要参数配置 pay.mock.enabled=1，且订单必须是当前会员的 mock 渠道订单。
// 入参 merchantOrderNo 为支付单号、userID 为当前会员；返回最新订单。
func (s *PayOrderService) MockPay(ctx context.Context, merchantOrderNo string, userID types.ID) (*payModel.PayOrderModel, error) {
	if s.configs.GetString(ctx, "pay.mock.enabled", "0") != "1" {
		return nil, baseErrs.New(payErrs.PayMockDisabled)
	}

	order, err := s.UserOrder(ctx, merchantOrderNo, userID)
	if err != nil {
		return nil, err
	}
	if !payChannel.IsMock(order.ChannelCode) {
		return nil, baseErrs.New(payErrs.PayMockChannelDenied)
	}
	if order.Status == payModel.PayOrderStatusPaid {
		return order, nil
	}

	if err := s.markPaid(ctx, order, paidInfo{
		ChannelOrderNo: "MOCK" + order.MerchantOrderNo,
		ChannelUser:    "mock",
		Raw:            `{"mock":true}`,
		SuccessTime:    time.Now().Format(time.RFC3339),
	}); err != nil {
		return nil, err
	}

	return s.orders.FindByID(ctx, order.ID)
}

// dispatchPaid 记录日志并派发支付成功事件（必须在事务提交后调用）。
// 入参 order 为已支付订单、paidAt 为渠道支付时间（RFC3339）。
func (s *PayOrderService) dispatchPaid(order *payModel.PayOrderModel, paidAt string) {
	if paidAt == "" {
		paidAt = time.Now().Format(time.RFC3339)
	}

	facades.Log().Infof("支付成功：%s 业务单 %s 金额 %d 分", order.MerchantOrderNo, order.BizOrderNo, order.Price)
	payEvents.DispatchOrderPaid(&payEvents.OrderPaidPayload{
		BizType:     order.BizType,
		BizOrderNo:  order.BizOrderNo,
		OrderNo:     order.MerchantOrderNo,
		ChannelCode: order.ChannelCode,
		Price:       order.Price,
		PaidAt:      paidAt,
	})
}

// markPaid 把订单流转为已支付并派发支付成功事件。
// 只有 CAS 影响行数为 1（首次流转）才派发事件，重复通知 / 重复查单天然幂等。
// 入参 order 为本地订单、info 为渠道侧支付信息；返回写库错误。
func (s *PayOrderService) markPaid(ctx context.Context, order *payModel.PayOrderModel, info paidInfo) error {
	affected, err := s.orders.UpdateStatusCAS(ctx, order.ID, []int8{
		int8(payModel.PayOrderStatusPending), int8(payModel.PayOrderStatusPaying),
	}, payModel.PayOrderStatusPaid, operatorName(ctx))
	if err != nil {
		return err
	}

	fields := map[string]any{
		"channel_order_no": info.ChannelOrderNo,
		"channel_user":     info.ChannelUser,
	}
	if info.Raw != "" {
		fields["channel_extra"] = info.Raw
	}
	if paidAt, ok := parseChannelTime(info.SuccessTime); ok {
		fields["success_time"] = paidAt
	}
	if err := s.orders.UpdateFields(ctx, order.ID, fields, operatorName(ctx)); err != nil {
		return err
	}

	if affected == 1 {
		s.dispatchPaid(order, info.SuccessTime)
	}

	return nil
}

// HandleChannelPaid 按渠道回调 / 查单结果推进订单为已支付，供回调服务复用。
// 只有 CAS 首次流转成功才派发支付成功事件，重复通知天然幂等。
// 入参 order 为本地订单、result 为渠道回调解析结果；返回写库错误。
func (s *PayOrderService) HandleChannelPaid(ctx context.Context, order *payModel.PayOrderModel, result payChannel.NotifyResult) error {
	return s.markPaid(ctx, order, paidInfo{
		ChannelOrderNo: result.ChannelOrderNo,
		ChannelUser:    result.ChannelUser,
		Raw:            result.Decrypted,
		SuccessTime:    result.SuccessTime,
	})
}

// paidInfo 是推进订单为已支付所需的渠道侧信息。
type paidInfo struct {
	// ChannelOrderNo 渠道侧单号
	ChannelOrderNo string
	// ChannelUser 渠道侧付款人标识
	ChannelUser string
	// Raw 渠道原始报文 / 响应 JSON
	Raw string
	// SuccessTime 支付完成时间（渠道格式）
	SuccessTime string
}

// parseChannelTime 解析渠道返回的时间（微信 RFC3339 / 支付宝 yyyy-MM-dd HH:mm:ss）。
// 入参 raw 为渠道时间字符串；返回时间与是否解析成功。
func parseChannelTime(raw string) (types.DateTime, bool) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return types.DateTime{}, false
	}

	for _, layout := range []string{time.RFC3339, "2006-01-02 15:04:05"} {
		if t, err := time.Parse(layout, raw); err == nil {
			return types.DateTime(t), true
		}
	}

	return types.DateTime{}, false
}
