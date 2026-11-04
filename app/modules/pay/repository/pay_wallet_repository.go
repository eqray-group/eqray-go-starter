package repository

import (
	"context"

	"github.com/goravel/framework/contracts/database/db"

	"eqray-go-starter/app/facades"
	baseTypes "eqray-go-starter/app/modules/base/types"
	payModel "eqray-go-starter/app/modules/pay/model"
)

const (
	walletTable            = payModel.PayWalletTable
	walletTransactionTable = payModel.PayWalletTransactionTable
)

// PayWalletFilter 是后台钱包列表的筛选条件。
type PayWalletFilter struct {
	// UserID 会员 ID 精确匹配，0 表示全部
	UserID baseTypes.ID
}

// PayWalletRepository 是 pay_wallet / pay_wallet_transaction 的仓储。
type PayWalletRepository struct{}

// NewPayWalletRepository 构造钱包仓储。
func NewPayWalletRepository() *PayWalletRepository {
	return &PayWalletRepository{}
}

// FindByUserID 按会员查询钱包；未命中返回 (nil, nil)。
// 入参 userID 为会员 ID；返回钱包模型或查询错误。
func (r *PayWalletRepository) FindByUserID(ctx context.Context, userID baseTypes.ID) (*payModel.PayWalletModel, error) {
	var m payModel.PayWalletModel

	if err := facades.DB().WithContext(ctx).Table(walletTable).
		Where("user_id", userID).
		First(&m); err != nil {
		return nil, err
	}
	if m.ID == 0 {
		return nil, nil
	}

	return &m, nil
}

// FindByUserIDForUpdate 在事务内按会员查询钱包并加行锁，用于余额变动。
// 入参 tx 为事务、userID 为会员 ID；返回钱包模型（可能为 nil）。
func (r *PayWalletRepository) FindByUserIDForUpdate(ctx context.Context, tx db.Tx, userID baseTypes.ID) (*payModel.PayWalletModel, error) {
	var m payModel.PayWalletModel

	if err := tx.Table(walletTable).
		Where("user_id", userID).
		LockForUpdate().
		First(&m); err != nil {
		return nil, err
	}
	if m.ID == 0 {
		return nil, nil
	}

	return &m, nil
}

// Create 新建钱包账户（余额 0）；唯一键冲突由调用方兜底重查。
// 入参 m 为钱包模型；返回写库错误。
func (r *PayWalletRepository) Create(ctx context.Context, m *payModel.PayWalletModel) error {
	_, err := facades.DB().WithContext(ctx).Table(walletTable).Insert(map[string]any{
		"id":             m.ID,
		"user_id":        m.UserID,
		"balance":        m.Balance,
		"frozen_balance": m.FrozenBalance,
		"total_recharge": m.TotalRecharge,
		"total_expense":  m.TotalExpense,
		"status":         m.Status,
		"creator":        m.Creator,
		"created_at":     m.CreatedAt,
		"updater":        m.Updater,
		"updated_at":     m.UpdatedAt,
	})

	return err
}

// Update 在事务内按主键整行更新余额与累计值。
// 入参 tx 为事务、m 为已改好的钱包模型；返回写库错误。
func (r *PayWalletRepository) Update(ctx context.Context, tx db.Tx, m *payModel.PayWalletModel) error {
	_, err := tx.Table(walletTable).
		Where("id", m.ID).
		Update(map[string]any{
			"balance":        m.Balance,
			"frozen_balance": m.FrozenBalance,
			"total_recharge": m.TotalRecharge,
			"total_expense":  m.TotalExpense,
			"status":         m.Status,
			"updater":        m.Updater,
			"updated_at":     m.UpdatedAt,
		})

	return err
}

// Paginate 分页查询钱包账户。
// 入参 filter 为筛选条件、pageNo/pageSize 为分页参数；返回列表与总数。
func (r *PayWalletRepository) Paginate(ctx context.Context, filter PayWalletFilter, pageNo, pageSize int) ([]payModel.PayWalletModel, int64, error) {
	list := make([]payModel.PayWalletModel, 0)
	var total int64

	query := facades.DB().WithContext(ctx).Table(walletTable)
	if filter.UserID > 0 {
		query = query.Where("user_id", filter.UserID)
	}

	if err := query.OrderByDesc("updated_at").Paginate(pageNo, pageSize, &list, &total); err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

// InsertTransaction 在事务内追加一条流水。
// 入参 tx 为事务、m 为流水模型；返回写库错误。
func (r *PayWalletRepository) InsertTransaction(ctx context.Context, tx db.Tx, m *payModel.PayWalletTransactionModel) error {
	_, err := tx.Table(walletTransactionTable).Insert(map[string]any{
		"id":            m.ID,
		"wallet_id":     m.WalletID,
		"user_id":       m.UserID,
		"type":          m.Type,
		"amount":        m.Amount,
		"balance_after": m.BalanceAfter,
		"biz_type":      m.BizType,
		"biz_order_no":  m.BizOrderNo,
		"remark":        m.Remark,
		"creator":       m.Creator,
		"created_at":    m.CreatedAt,
	})

	return err
}

// Transactions 分页查询某会员的钱包流水，按时间倒序。
// 入参 userID 为会员 ID、pageNo/pageSize 为分页参数；返回列表与总数。
func (r *PayWalletRepository) Transactions(ctx context.Context, userID baseTypes.ID, pageNo, pageSize int) ([]payModel.PayWalletTransactionModel, int64, error) {
	list := make([]payModel.PayWalletTransactionModel, 0)
	var total int64

	if err := facades.DB().WithContext(ctx).Table(walletTransactionTable).
		Where("user_id", userID).
		OrderByDesc("id").
		Paginate(pageNo, pageSize, &list, &total); err != nil {
		return nil, 0, err
	}

	return list, total, nil
}
