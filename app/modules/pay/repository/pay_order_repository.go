package repository

import (
	"context"

	"eqray-go-starter/app/facades"
	baseErrs "eqray-go-starter/app/modules/base/errs"
	"eqray-go-starter/app/modules/base/support"
	baseTypes "eqray-go-starter/app/modules/base/types"
	payErrs "eqray-go-starter/app/modules/pay/errs"
	payModel "eqray-go-starter/app/modules/pay/model"
)

const payOrderTable = payModel.PayOrderTable

// PayOrderFilter 是支付订单列表的筛选条件。
type PayOrderFilter struct {
	// MerchantOrderNo 支付单号精确匹配
	MerchantOrderNo string
	// BizType 业务类型精确匹配
	BizType string
	// BizOrderNo 业务单号精确匹配
	BizOrderNo string
	// ChannelCode 渠道编码精确匹配
	ChannelCode string
	// Status 订单状态过滤，nil 表示全部
	Status *payModel.PayOrderStatus
	// UserID 下单会员过滤，0 表示全部
	UserID baseTypes.ID
	// BeginTime 创建时间下限（含），格式 yyyy-MM-dd HH:mm:ss，空表示不限
	BeginTime string
	// EndTime 创建时间上限（含），格式 yyyy-MM-dd HH:mm:ss，空表示不限
	EndTime string
}

// PayOrderRepository 是 pay_order 表的仓储。
type PayOrderRepository struct{}

// NewPayOrderRepository 构造支付订单仓储。
func NewPayOrderRepository() *PayOrderRepository {
	return &PayOrderRepository{}
}

// Insert 插入支付订单；uk_biz_order / uk_merchant_order_no 撞键时转换业务错误码。
// 入参 m 为待插入订单；返回唯一键冲突或写库错误。
func (r *PayOrderRepository) Insert(ctx context.Context, m *payModel.PayOrderModel) error {
	m.FillCreate(ctx)

	_, err := facades.DB().WithContext(ctx).Table(payOrderTable).Insert(map[string]any{
		"id":                m.ID,
		"merchant_order_no": m.MerchantOrderNo,
		"biz_type":          m.BizType,
		"biz_order_no":      m.BizOrderNo,
		"channel_id":        m.ChannelID,
		"channel_code":      m.ChannelCode,
		"subject":           m.Subject,
		"description":       m.Description,
		"price":             m.Price,
		"currency":          m.Currency,
		"status":            m.Status,
		"refund_status":     m.RefundStatus,
		"user_id":           m.UserID,
		"client_ip":         m.ClientIP,
		"channel_order_no":  m.ChannelOrderNo,
		"channel_user":      m.ChannelUser,
		"channel_extra":     m.ChannelExtra,
		"notify_url":        m.NotifyURL,
		"expire_time":       m.ExpireTime,
		"success_time":      m.SuccessTime,
		"creator":           m.Creator,
		"created_at":        m.CreatedAt,
		"updater":           m.Updater,
		"updated_at":        m.UpdatedAt,
	})
	if err != nil && support.IsDuplicate(err) {
		return baseErrs.New(payErrs.PayOrderDuplicated)
	}

	return err
}

// UpdateFields 按 ID 更新指定字段，供状态流转 / 渠道信息回写使用。
// 入参 id 为订单 ID、data 为列名 → 值、updater 为操作人；返回写库错误。
func (r *PayOrderRepository) UpdateFields(ctx context.Context, id baseTypes.ID, data map[string]any, updater string) error {
	if data == nil {
		data = make(map[string]any)
	}
	data["updater"] = updater
	data["updated_at"] = baseTypes.Now()

	_, err := facades.DB().WithContext(ctx).Table(payOrderTable).
		Where("id", id).
		Update(data)

	return err
}

// UpdateStatusCAS 把订单从 from 状态集合流转到 to 状态，返回实际影响行数。
// 状态机并发安全的核心：只有影响行数 = 1 的一方负责派发后续事件。
// 入参 id 为订单 ID、from 为允许的当前状态集合、to 为目标状态、updater 为操作人；返回影响行数与写库错误。
func (r *PayOrderRepository) UpdateStatusCAS(ctx context.Context, id baseTypes.ID, from []int8, to payModel.PayOrderStatus, updater string) (int64, error) {
	fromAny := make([]any, 0, len(from))
	for _, s := range from {
		fromAny = append(fromAny, s)
	}

	res, err := facades.DB().WithContext(ctx).Table(payOrderTable).
		Where("id", id).
		WhereIn("status", fromAny).
		Update(map[string]any{
			"status":     to,
			"updater":    updater,
			"updated_at": baseTypes.Now(),
		})
	if err != nil {
		return 0, err
	}

	return res.RowsAffected, nil
}

// FindByID 按主键查询支付订单；未命中返回 (nil, nil)。
// 入参 id 为订单 ID；返回模型或查询错误。
func (r *PayOrderRepository) FindByID(ctx context.Context, id baseTypes.ID) (*payModel.PayOrderModel, error) {
	var m payModel.PayOrderModel

	if err := facades.DB().WithContext(ctx).Table(payOrderTable).
		Where("id", id).
		First(&m); err != nil {
		return nil, err
	}
	if m.ID == 0 {
		return nil, nil
	}

	return &m, nil
}

// FindByMerchantOrderNo 按支付单号查询订单（回调处理入口）。
// 入参 merchantOrderNo 为支付单号；返回模型，未命中 (nil, nil)。
func (r *PayOrderRepository) FindByMerchantOrderNo(ctx context.Context, merchantOrderNo string) (*payModel.PayOrderModel, error) {
	var m payModel.PayOrderModel

	if err := facades.DB().WithContext(ctx).Table(payOrderTable).
		Where("merchant_order_no", merchantOrderNo).
		First(&m); err != nil {
		return nil, err
	}
	if m.ID == 0 {
		return nil, nil
	}

	return &m, nil
}

// FindByBiz 按「业务类型 + 业务单号 + 渠道」查询订单，对应 uk_biz_order 唯一键。
// 入参 bizType/bizOrderNo/channelCode 组成唯一约束；返回模型，未命中 (nil, nil)。
func (r *PayOrderRepository) FindByBiz(ctx context.Context, bizType, bizOrderNo, channelCode string) (*payModel.PayOrderModel, error) {
	var m payModel.PayOrderModel

	if err := facades.DB().WithContext(ctx).Table(payOrderTable).
		Where("biz_type", bizType).
		Where("biz_order_no", bizOrderNo).
		Where("channel_code", channelCode).
		First(&m); err != nil {
		return nil, err
	}
	if m.ID == 0 {
		return nil, nil
	}

	return &m, nil
}

// FindUserOrder 按单号 + 会员查询订单，C 端轮询支付结果用（防越权）。
// 入参 merchantOrderNo 为支付单号、userID 为会员 ID；返回模型，未命中 (nil, nil)。
func (r *PayOrderRepository) FindUserOrder(ctx context.Context, merchantOrderNo string, userID baseTypes.ID) (*payModel.PayOrderModel, error) {
	var m payModel.PayOrderModel

	if err := facades.DB().WithContext(ctx).Table(payOrderTable).
		Where("merchant_order_no", merchantOrderNo).
		Where("user_id", userID).
		First(&m); err != nil {
		return nil, err
	}
	if m.ID == 0 {
		return nil, nil
	}

	return &m, nil
}

// ListExpired 扫描已超时的未完成订单，供定时关单任务使用。
// 入参 before 为超时时间点、limit 为单批上限；返回订单列表。
func (r *PayOrderRepository) ListExpired(ctx context.Context, before baseTypes.DateTime, limit int) ([]payModel.PayOrderModel, error) {
	list := make([]payModel.PayOrderModel, 0)

	if err := facades.DB().WithContext(ctx).Table(payOrderTable).
		WhereIn("status", []any{int8(payModel.PayOrderStatusPending), int8(payModel.PayOrderStatusPaying)}).
		Where("expire_time < ?", before).
		Limit(uint64(limit)).
		OrderBy("id").
		Get(&list); err != nil {
		return nil, err
	}

	return list, nil
}

// Paginate 分页查询支付订单。
// 入参 filter 为筛选条件、pageNo/pageSize 为分页参数；返回列表与总数。
func (r *PayOrderRepository) Paginate(ctx context.Context, filter PayOrderFilter, pageNo, pageSize int) ([]payModel.PayOrderModel, int64, error) {
	list := make([]payModel.PayOrderModel, 0)
	var total int64

	query := facades.DB().WithContext(ctx).Table(payOrderTable)
	if filter.MerchantOrderNo != "" {
		query = query.Where("merchant_order_no", filter.MerchantOrderNo)
	}
	if filter.BizType != "" {
		query = query.Where("biz_type", filter.BizType)
	}
	if filter.BizOrderNo != "" {
		query = query.Where("biz_order_no", filter.BizOrderNo)
	}
	if filter.ChannelCode != "" {
		query = query.Where("channel_code", filter.ChannelCode)
	}
	if filter.Status != nil {
		query = query.Where("status", *filter.Status)
	}
	if filter.UserID > 0 {
		query = query.Where("user_id", filter.UserID)
	}
	if filter.BeginTime != "" {
		query = query.Where("created_at >= ?", filter.BeginTime)
	}
	if filter.EndTime != "" {
		query = query.Where("created_at <= ?", filter.EndTime)
	}

	if err := query.OrderByDesc("id").Paginate(pageNo, pageSize, &list, &total); err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

// SumRefunded 统计订单已退款成功的总金额（分），用于退款前校验可退余额。
// 入参 orderID 为支付订单 ID；返回金额与查询错误。
func (r *PayOrderRepository) SumRefunded(ctx context.Context, orderID baseTypes.ID) (int64, error) {
	// 用显式聚合 + COALESCE：空结果返回 0，避免 SUM 的 NULL 扫描问题
	row := struct {
		Total int64 `db:"total"`
	}{}

	if err := facades.DB().WithContext(ctx).Table(payRefundTable).
		Select("COALESCE(SUM(refund_price), 0) AS total").
		Where("order_id", orderID).
		Where("status", int8(payModel.PayRefundStatusSuccess)).
		First(&row); err != nil {
		return 0, err
	}

	return row.Total, nil
}
