package repository

import (
	"context"

	"github.com/goravel/framework/contracts/database/db"

	"eqray-go-starter/app/facades"
	baseTypes "eqray-go-starter/app/modules/base/types"
	payModel "eqray-go-starter/app/modules/pay/model"
)

const payRefundTable = payModel.PayRefundTable

// PayRefundFilter 是退款单列表的筛选条件。
type PayRefundFilter struct {
	// MerchantRefundNo 退款单号精确匹配
	MerchantRefundNo string
	// MerchantOrderNo 支付单号精确匹配
	MerchantOrderNo string
	// OrderID 支付订单过滤
	OrderID baseTypes.ID
	// Status 退款状态过滤，nil 表示全部
	Status *payModel.PayRefundStatus
}

// UpdateFieldsTx 在事务内按 ID 更新退款单字段，供余额退款与渠道退款共用。
// 入参 tx 为事务、id 为退款单 ID、data 为列名 → 值、updater 为操作人；返回写库错误。
func (r *PayRefundRepository) UpdateFieldsTx(ctx context.Context, tx db.Tx, id baseTypes.ID, data map[string]any, updater string) error {
	if data == nil {
		data = make(map[string]any)
	}
	data["updater"] = updater
	data["updated_at"] = baseTypes.Now()

	_, err := tx.Table(payRefundTable).Where("id", id).Update(data)

	return err
}

// PayRefundRepository 是 pay_refund 表的仓储。
type PayRefundRepository struct{}

// NewPayRefundRepository 构造退款单仓储。
func NewPayRefundRepository() *PayRefundRepository {
	return &PayRefundRepository{}
}

// Insert 插入退款单；uk_merchant_refund_no 撞键时由调用方按重复处理。
// 入参 m 为待插入退款单；返回写库错误。
func (r *PayRefundRepository) Insert(ctx context.Context, m *payModel.PayRefundModel) error {
	m.FillCreate(ctx)

	_, err := facades.DB().WithContext(ctx).Table(payRefundTable).Insert(map[string]any{
		"id":                 m.ID,
		"merchant_refund_no": m.MerchantRefundNo,
		"order_id":           m.OrderID,
		"merchant_order_no":  m.MerchantOrderNo,
		"channel_id":         m.ChannelID,
		"channel_code":       m.ChannelCode,
		"refund_price":       m.RefundPrice,
		"reason":             m.Reason,
		"status":             m.Status,
		"channel_refund_no":  m.ChannelRefundNo,
		"channel_extra":      m.ChannelExtra,
		"error_msg":          m.ErrorMsg,
		"success_time":       m.SuccessTime,
		"creator":            m.Creator,
		"created_at":         m.CreatedAt,
		"updater":            m.Updater,
		"updated_at":         m.UpdatedAt,
	})

	return err
}

// UpdateFields 按 ID 更新指定字段，供退款状态回写使用。
// 入参 id 为退款单 ID、data 为列名 → 值、updater 为操作人；返回写库错误。
func (r *PayRefundRepository) UpdateFields(ctx context.Context, id baseTypes.ID, data map[string]any, updater string) error {
	if data == nil {
		data = make(map[string]any)
	}
	data["updater"] = updater
	data["updated_at"] = baseTypes.Now()

	_, err := facades.DB().WithContext(ctx).Table(payRefundTable).
		Where("id", id).
		Update(data)

	return err
}

// UpdateStatusCAS 把退款单从 from 状态集合流转到 to 状态，返回实际影响行数。
// 入参 id 为退款单 ID、from 为允许的当前状态集合、to 为目标状态、updater 为操作人；返回影响行数与写库错误。
func (r *PayRefundRepository) UpdateStatusCAS(ctx context.Context, id baseTypes.ID, from []int8, to payModel.PayRefundStatus, updater string) (int64, error) {
	fromAny := make([]any, 0, len(from))
	for _, s := range from {
		fromAny = append(fromAny, s)
	}

	res, err := facades.DB().WithContext(ctx).Table(payRefundTable).
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

// FindByID 按主键查询退款单；未命中返回 (nil, nil)。
// 入参 id 为退款单 ID；返回模型或查询错误。
func (r *PayRefundRepository) FindByID(ctx context.Context, id baseTypes.ID) (*payModel.PayRefundModel, error) {
	var m payModel.PayRefundModel

	if err := facades.DB().WithContext(ctx).Table(payRefundTable).
		Where("id", id).
		First(&m); err != nil {
		return nil, err
	}
	if m.ID == 0 {
		return nil, nil
	}

	return &m, nil
}

// ExistsByMerchantRefundNo 判断退款单号是否已存在（重试沿用同一单号）。
// 入参 merchantRefundNo 为退款单号；返回是否已存在与查询错误。
func (r *PayRefundRepository) ExistsByMerchantRefundNo(ctx context.Context, merchantRefundNo string) (bool, error) {
	total, err := facades.DB().WithContext(ctx).Table(payRefundTable).
		Where("merchant_refund_no", merchantRefundNo).
		Count()

	return total > 0, err
}

// ListRefunding 扫描「退款中」的退款单，供定时查单任务收敛状态。
// 入参 before 为创建时间上限（只处理创建超过该时间的单，避免与同步请求撞车）、limit 为单批上限；返回退款单列表。
func (r *PayRefundRepository) ListRefunding(ctx context.Context, before baseTypes.DateTime, limit int) ([]payModel.PayRefundModel, error) {
	list := make([]payModel.PayRefundModel, 0)

	if err := facades.DB().WithContext(ctx).Table(payRefundTable).
		Where("status", int8(payModel.PayRefundStatusRefunding)).
		Where("created_at < ?", before).
		Limit(uint64(limit)).
		OrderBy("id").
		Get(&list); err != nil {
		return nil, err
	}

	return list, nil
}

// Paginate 分页查询退款单。
// 入参 filter 为筛选条件、pageNo/pageSize 为分页参数；返回列表与总数。
func (r *PayRefundRepository) Paginate(ctx context.Context, filter PayRefundFilter, pageNo, pageSize int) ([]payModel.PayRefundModel, int64, error) {
	list := make([]payModel.PayRefundModel, 0)
	var total int64

	query := facades.DB().WithContext(ctx).Table(payRefundTable)
	if filter.MerchantRefundNo != "" {
		query = query.Where("merchant_refund_no", filter.MerchantRefundNo)
	}
	if filter.MerchantOrderNo != "" {
		query = query.Where("merchant_order_no", filter.MerchantOrderNo)
	}
	if filter.OrderID > 0 {
		query = query.Where("order_id", filter.OrderID)
	}
	if filter.Status != nil {
		query = query.Where("status", *filter.Status)
	}

	if err := query.OrderByDesc("id").Paginate(pageNo, pageSize, &list, &total); err != nil {
		return nil, 0, err
	}

	return list, total, nil
}
