package service

import (
	"context"

	"github.com/goravel/framework/contracts/database/db"

	"eqray-go-starter/app/facades"
	baseErrs "eqray-go-starter/app/modules/base/errs"
	"eqray-go-starter/app/modules/base/support"
	"eqray-go-starter/app/modules/base/types"
	payErrs "eqray-go-starter/app/modules/pay/errs"
	payModel "eqray-go-starter/app/modules/pay/model"
	payRepository "eqray-go-starter/app/modules/pay/repository"
)

// WalletFreezeInput 是后台冻结 / 解冻钱包资金的入参。
type WalletFreezeInput struct {
	// UserID 会员 ID
	UserID types.ID
	// Amount 金额（分），必须大于 0
	Amount int64
	// Remark 备注，会写入流水
	Remark string
}

// WalletAdjustInput 是后台调整钱包余额的入参。
type WalletAdjustInput struct {
	// UserID 会员 ID
	UserID types.ID
	// Amount 调整金额（分）：正数充值、负数扣减，不能为 0
	Amount int64
	// Remark 备注，会写入流水
	Remark string
}

// PayWalletListQuery 是后台钱包列表的查询条件。
type PayWalletListQuery struct {
	// UserID 会员 ID 精确匹配，0 表示全部
	UserID types.ID
	// PageNo / PageSize 分页参数
	PageNo   int
	PageSize int
}

// PayWalletService 负责钱包账户、流水与后台调账。
// 余额支付 / 退款回钱包的事务编排在 PayOrderService / PayRefundService 中，
// 这里只提供账户读写与调账能力。
type PayWalletService struct {
	wallets *payRepository.PayWalletRepository
}

// NewPayWalletService 构造钱包服务。
func NewPayWalletService() *PayWalletService {
	return &PayWalletService{wallets: payRepository.NewPayWalletRepository()}
}

// Ensure 取钱包账户，不存在则创建一个空钱包（并发创建撞唯一键时回查）。
// 入参 userID 为会员 ID；返回钱包模型或业务错误码。
func (s *PayWalletService) Ensure(ctx context.Context, userID types.ID) (*payModel.PayWalletModel, error) {
	if userID == 0 {
		return nil, baseErrs.WithMsg(baseErrs.BadRequest, "缺少会员信息")
	}

	wallet, err := s.wallets.FindByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if wallet != nil {
		return wallet, nil
	}

	now := types.Now()
	operator := operatorName(ctx)
	wallet = &payModel.PayWalletModel{
		ID:        support.NewID(),
		UserID:    userID,
		Balance:   0,
		Status:    payModel.PayWalletStatusEnabled,
		Creator:   operator,
		CreatedAt: now,
		Updater:   operator,
		UpdatedAt: now,
	}
	if err := s.wallets.Create(ctx, wallet); err != nil {
		// 并发首次进入：唯一键冲突后回查已创建的账户
		if support.IsDuplicate(err) {
			return s.wallets.FindByUserID(ctx, userID)
		}

		return nil, err
	}

	return wallet, nil
}

// Detail 查询会员钱包（不存在时返回余额为 0 的空账户，不落库）。
// 入参 userID 为会员 ID；返回钱包模型。
func (s *PayWalletService) Detail(ctx context.Context, userID types.ID) (*payModel.PayWalletModel, error) {
	wallet, err := s.Ensure(ctx, userID)
	if err != nil {
		return nil, err
	}
	if wallet == nil {
		return &payModel.PayWalletModel{UserID: userID, Status: payModel.PayWalletStatusEnabled}, nil
	}

	return wallet, nil
}

// Page 分页查询钱包账户（后台）。
// 入参 q 为查询条件；返回列表与总数。
func (s *PayWalletService) Page(ctx context.Context, q PayWalletListQuery) ([]payModel.PayWalletModel, int64, error) {
	return s.wallets.Paginate(ctx, payRepository.PayWalletFilter{UserID: q.UserID}, q.PageNo, q.PageSize)
}

// Transactions 分页查询会员钱包流水（后台与 C 端共用）。
// 入参 userID 为会员 ID、pageNo/pageSize 为分页参数；返回列表与总数。
func (s *PayWalletService) Transactions(ctx context.Context, userID types.ID, pageNo, pageSize int) ([]payModel.PayWalletTransactionModel, int64, error) {
	return s.wallets.Transactions(ctx, userID, pageNo, pageSize)
}

// Adjust 后台调整余额：正数充值、负数扣减，事务内更新余额并写流水。
// 入参 in 为调账入参；返回写库错误或业务错误码。
func (s *PayWalletService) Adjust(ctx context.Context, in WalletAdjustInput) error {
	if in.Amount == 0 {
		return baseErrs.New(payErrs.PayWalletAmountInvalid)
	}
	if in.UserID == 0 {
		return baseErrs.WithMsg(baseErrs.BadRequest, "请先选择会员")
	}

	// 先确保账户存在（并发安全由唯一键兜底），再在事务里加锁改余额
	if _, err := s.Ensure(ctx, in.UserID); err != nil {
		return err
	}

	operator := operatorName(ctx)
	if operator == "" {
		operator = "system"
	}

	return facades.DB().WithContext(ctx).Transaction(func(tx db.Tx) error {
		wallet, err := s.wallets.FindByUserIDForUpdate(ctx, tx, in.UserID)
		if err != nil {
			return err
		}
		if wallet == nil {
			return baseErrs.New(payErrs.PayWalletNotExists, in.UserID)
		}
		if wallet.Status != payModel.PayWalletStatusEnabled {
			return baseErrs.New(payErrs.PayWalletDisabled)
		}
		if wallet.Balance+in.Amount < 0 {
			return baseErrs.New(payErrs.PayWalletInsufficient, wallet.Balance)
		}

		wallet.Balance += in.Amount
		if in.Amount > 0 {
			wallet.TotalRecharge += in.Amount
		}
		wallet.Updater = operator
		wallet.UpdatedAt = types.Now()
		if err := s.wallets.Update(ctx, tx, wallet); err != nil {
			return err
		}

		txnType := payModel.PayWalletTypeAdjust
		if in.Amount > 0 {
			txnType = payModel.PayWalletTypeRecharge
		}

		return s.wallets.InsertTransaction(ctx, tx, &payModel.PayWalletTransactionModel{
			ID:           support.NewID(),
			WalletID:     wallet.ID,
			UserID:       in.UserID,
			Type:         txnType,
			Amount:       in.Amount,
			BalanceAfter: wallet.Balance,
			Remark:       in.Remark,
			Creator:      operator,
			CreatedAt:    types.Now(),
		})
	})
}

// Freeze 冻结会员钱包资金：可用余额转入冻结金额，仍属于会员资产但不可消费。
// 入参 in 为冻结入参；返回写库错误或业务错误码。
func (s *PayWalletService) Freeze(ctx context.Context, in WalletFreezeInput) error {
	return s.changeFrozen(ctx, in, true)
}

// Unfreeze 解冻会员钱包资金：冻结金额转回可用余额。
// 入参 in 为解冻入参；返回写库错误或业务错误码。
func (s *PayWalletService) Unfreeze(ctx context.Context, in WalletFreezeInput) error {
	return s.changeFrozen(ctx, in, false)
}

// changeFrozen 冻结 / 解冻的公共实现：事务内加锁改余额并写流水。
// 入参 in 为资金操作入参、freeze 为 true 表示冻结；返回写库错误或业务错误码。
func (s *PayWalletService) changeFrozen(ctx context.Context, in WalletFreezeInput, freeze bool) error {
	if in.Amount <= 0 {
		return baseErrs.WithMsg(baseErrs.BadRequest, "金额必须大于 0")
	}
	if in.UserID == 0 {
		return baseErrs.WithMsg(baseErrs.BadRequest, "请先选择会员")
	}

	if _, err := s.Ensure(ctx, in.UserID); err != nil {
		return err
	}

	operator := operatorName(ctx)
	if operator == "" {
		operator = "system"
	}

	return facades.DB().WithContext(ctx).Transaction(func(tx db.Tx) error {
		wallet, err := s.wallets.FindByUserIDForUpdate(ctx, tx, in.UserID)
		if err != nil {
			return err
		}
		if wallet == nil {
			return baseErrs.New(payErrs.PayWalletNotExists, in.UserID)
		}
		if wallet.Status != payModel.PayWalletStatusEnabled {
			return baseErrs.New(payErrs.PayWalletDisabled)
		}

		txnType := payModel.PayWalletTypeFreeze
		amount := -in.Amount
		remark := "冻结：" + in.Remark

		if freeze {
			// 只能占用可用余额，不能把余额冻成负数
			if wallet.Balance < in.Amount {
				return baseErrs.New(payErrs.PayWalletInsufficient, wallet.Balance)
			}
			wallet.Balance -= in.Amount
			wallet.FrozenBalance += in.Amount
		} else {
			if wallet.FrozenBalance < in.Amount {
				return baseErrs.New(payErrs.PayWalletFrozenExceed, wallet.FrozenBalance)
			}
			wallet.Balance += in.Amount
			wallet.FrozenBalance -= in.Amount
			txnType = payModel.PayWalletTypeUnfreeze
			amount = in.Amount
			remark = "解冻：" + in.Remark
		}

		wallet.Updater = operator
		wallet.UpdatedAt = types.Now()
		if err := s.wallets.Update(ctx, tx, wallet); err != nil {
			return err
		}

		return s.wallets.InsertTransaction(ctx, tx, &payModel.PayWalletTransactionModel{
			ID:           support.NewID(),
			WalletID:     wallet.ID,
			UserID:       in.UserID,
			Type:         txnType,
			Amount:       amount,
			BalanceAfter: wallet.Balance,
			Remark:       remark,
			Creator:      operator,
			CreatedAt:    types.Now(),
		})
	})
}
