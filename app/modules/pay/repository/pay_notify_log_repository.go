package repository

import (
	"context"
	"time"

	"eqray-go-starter/app/facades"
	"eqray-go-starter/app/modules/base/support"
	baseTypes "eqray-go-starter/app/modules/base/types"
	payModel "eqray-go-starter/app/modules/pay/model"
)

const payNotifyLogTable = payModel.PayNotifyLogTable

// PayNotifyLogFilter 是回调日志列表的筛选条件。
type PayNotifyLogFilter struct {
	// ChannelCode 渠道编码精确匹配
	ChannelCode string
	// Type 类型过滤：pay / refund
	Type string
	// MerchantOrderNo 支付单号精确匹配
	MerchantOrderNo string
	// VerifyStatus 验签结果过滤，nil 表示全部
	VerifyStatus *int8
	// HandleStatus 处理结果过滤，nil 表示全部
	HandleStatus *int8
}

// PayNotifyLogRepository 是 pay_notify_log 表的仓储。
// 日志只追加不更新，过期数据由定时任务物理清理。
type PayNotifyLogRepository struct{}

// NewPayNotifyLogRepository 构造回调日志仓储。
func NewPayNotifyLogRepository() *PayNotifyLogRepository {
	return &PayNotifyLogRepository{}
}

// Insert 追加一条回调日志。
// 入参 m 为日志模型（只填业务列，created_at 由本方法填充）；返回写库错误。
func (r *PayNotifyLogRepository) Insert(ctx context.Context, m *payModel.PayNotifyLogModel) error {
	m.ID = support.NewID()
	m.CreatedAt = baseTypes.Now()

	_, err := facades.DB().WithContext(ctx).Table(payNotifyLogTable).Insert(map[string]any{
		"id":                m.ID,
		"channel_code":      m.ChannelCode,
		"type":              m.Type,
		"merchant_order_no": m.MerchantOrderNo,
		"request_headers":   m.RequestHeaders,
		"request_body":      m.RequestBody,
		"decrypted_body":    m.DecryptedBody,
		"verify_status":     m.VerifyStatus,
		"handle_status":     m.HandleStatus,
		"error_msg":         m.ErrorMsg,
		"created_at":        m.CreatedAt,
	})

	return err
}

// Paginate 分页查询回调日志。
// 入参 filter 为筛选条件、pageNo/pageSize 为分页参数；返回列表与总数。
func (r *PayNotifyLogRepository) Paginate(ctx context.Context, filter PayNotifyLogFilter, pageNo, pageSize int) ([]payModel.PayNotifyLogModel, int64, error) {
	list := make([]payModel.PayNotifyLogModel, 0)
	var total int64

	query := facades.DB().WithContext(ctx).Table(payNotifyLogTable)
	if filter.ChannelCode != "" {
		query = query.Where("channel_code", filter.ChannelCode)
	}
	if filter.Type != "" {
		query = query.Where("type", filter.Type)
	}
	if filter.MerchantOrderNo != "" {
		query = query.Where("merchant_order_no", filter.MerchantOrderNo)
	}
	if filter.VerifyStatus != nil {
		query = query.Where("verify_status", *filter.VerifyStatus)
	}
	if filter.HandleStatus != nil {
		query = query.Where("handle_status", *filter.HandleStatus)
	}

	if err := query.OrderByDesc("id").Paginate(pageNo, pageSize, &list, &total); err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

// DeleteBefore 物理删除指定时间点之前的回调日志，返回删除行数。
// 入参 before 为时间点；返回删除行数与写库错误。
func (r *PayNotifyLogRepository) DeleteBefore(ctx context.Context, before time.Time) (int64, error) {
	res, err := facades.DB().WithContext(ctx).Table(payNotifyLogTable).
		Where("created_at < ?", before).
		Delete()
	if err != nil {
		return 0, err
	}

	return res.RowsAffected, nil
}
