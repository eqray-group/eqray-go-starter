// Package repository 是演示业务模块的数据访问层。
package repository

import (
	"context"

	"eqray-go-starter/app/facades"
	baseTypes "eqray-go-starter/app/modules/base/types"
	demoModel "eqray-go-starter/app/modules/demo/model"
)

const demoOrderTable = demoModel.DemoOrderTable

// DemoOrderRepository 是 demo_order 表的仓储。
type DemoOrderRepository struct{}

// NewDemoOrderRepository 构造演示订单仓储。
func NewDemoOrderRepository() *DemoOrderRepository {
	return &DemoOrderRepository{}
}

// Insert 插入演示订单；uk_order_no 保证单号唯一。
// 入参 m 为待插入订单；返回写库错误。
func (r *DemoOrderRepository) Insert(ctx context.Context, m *demoModel.DemoOrderModel) error {
	m.FillCreate(ctx)

	_, err := facades.DB().WithContext(ctx).Table(demoOrderTable).Insert(map[string]any{
		"id":           m.ID,
		"order_no":     m.OrderNo,
		"user_id":      m.UserID,
		"subject":      m.Subject,
		"price":        m.Price,
		"status":       m.Status,
		"pay_order_no": m.PayOrderNo,
		"paid_at":      m.PaidAt,
		"creator":      m.Creator,
		"created_at":   m.CreatedAt,
		"updater":      m.Updater,
		"updated_at":   m.UpdatedAt,
	})

	return err
}

// FindByOrderNo 按业务单号查询演示订单；未命中返回 (nil, nil)。
// 入参 orderNo 为业务单号；返回订单模型或查询错误。
func (r *DemoOrderRepository) FindByOrderNo(ctx context.Context, orderNo string) (*demoModel.DemoOrderModel, error) {
	var m demoModel.DemoOrderModel

	if err := facades.DB().WithContext(ctx).Table(demoOrderTable).
		Where("order_no", orderNo).
		First(&m); err != nil {
		return nil, err
	}
	if m.ID == 0 {
		return nil, nil
	}

	return &m, nil
}

// PageMine 分页查询某会员的演示订单，按创建时间倒序。
// 入参 userID 为会员 ID、pageNo/pageSize 为分页参数；返回列表与总数。
func (r *DemoOrderRepository) PageMine(ctx context.Context, userID baseTypes.ID, pageNo, pageSize int) ([]demoModel.DemoOrderModel, int64, error) {
	list := make([]demoModel.DemoOrderModel, 0)
	var total int64

	if err := facades.DB().WithContext(ctx).Table(demoOrderTable).
		Where("user_id", userID).
		OrderByDesc("id").
		Paginate(pageNo, pageSize, &list, &total); err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

// MarkPayOrderNo 回写业务单对应的支付单号，便于联调时对账。
// 入参 orderNo 为业务单号、payOrderNo 为支付单号；返回写库错误。
func (r *DemoOrderRepository) MarkPayOrderNo(ctx context.Context, orderNo, payOrderNo string) error {
	_, err := facades.DB().WithContext(ctx).Table(demoOrderTable).
		Where("order_no", orderNo).
		Update(map[string]any{
			"pay_order_no": payOrderNo,
			"updater":      "system",
			"updated_at":   baseTypes.Now(),
		})

	return err
}

// MarkPaid 把演示订单标记为已支付，天然幂等：仅当当前状态为待支付时生效。
// 入参 orderNo 为业务单号、payOrderNo 为支付单号、paidAt 为支付完成时间；返回写库错误。
func (r *DemoOrderRepository) MarkPaid(ctx context.Context, orderNo, payOrderNo string, paidAt baseTypes.DateTime) error {
	_, err := facades.DB().WithContext(ctx).Table(demoOrderTable).
		Where("order_no", orderNo).
		Where("status", demoModel.DemoOrderStatusPending).
		Update(map[string]any{
			"status":       demoModel.DemoOrderStatusPaid,
			"pay_order_no": payOrderNo,
			"paid_at":      paidAt,
			"updater":      "system",
			"updated_at":   baseTypes.Now(),
		})

	return err
}
