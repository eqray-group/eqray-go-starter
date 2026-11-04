// Package errs 定义 pay 模块的业务错误码，编码规则见 docs 8.3。
package errs

import baseErrs "eqray-go-starter/app/modules/base/errs"

// pay 模块业务错误码：1-004-{子模块}-{错误码}，
// channel 占用 001，order 占用 002，refund 占用 003，notify 占用 004。
var (
	// ---------- 渠道配置 ----------
	PayChannelNotExists    = baseErrs.ErrorCode{Code: 1_004_001_001, Msg: "支付渠道【%v】不存在"}
	PayChannelDisabled     = baseErrs.ErrorCode{Code: 1_004_001_002, Msg: "支付渠道【%v】已停用"}
	PayChannelConfigBad    = baseErrs.ErrorCode{Code: 1_004_001_003, Msg: "渠道配置不合法：%v"}
	PayAppUrlNotConfigured = baseErrs.ErrorCode{Code: 1_004_001_004, Msg: "未配置 APP_URL，无法生成异步通知地址"}
	PayChannelCodeExists   = baseErrs.ErrorCode{Code: 1_004_001_005, Msg: "渠道编码【%v】已存在"}
	PayChannelUnsupported  = baseErrs.ErrorCode{Code: 1_004_001_006, Msg: "不支持的支付渠道【%v】"}

	// ---------- 支付订单 ----------
	PayOrderNotExists    = baseErrs.ErrorCode{Code: 1_004_002_001, Msg: "支付订单【%v】不存在"}
	PayOrderStatusDenied = baseErrs.ErrorCode{Code: 1_004_002_002, Msg: "订单当前状态【%v】不允许该操作"}
	PayOrderDuplicated   = baseErrs.ErrorCode{Code: 1_004_002_003, Msg: "重复下单，请稍后重试"}
	PayOrderCreateFailed = baseErrs.ErrorCode{Code: 1_004_002_004, Msg: "渠道下单失败：%v"}
	PayNotifyVerifyFail  = baseErrs.ErrorCode{Code: 1_004_002_005, Msg: "回调验签失败"}
	PayNotifyPriceDiff   = baseErrs.ErrorCode{Code: 1_004_002_006, Msg: "回调金额与订单金额不一致"}

	// ---------- 退款 ----------
	PayRefundNotExists    = baseErrs.ErrorCode{Code: 1_004_003_001, Msg: "退款单【%v】不存在"}
	PayRefundPriceInvalid = baseErrs.ErrorCode{Code: 1_004_003_002, Msg: "退款金额超过可退金额"}
	PayRefundFailed       = baseErrs.ErrorCode{Code: 1_004_003_003, Msg: "渠道退款失败：%v"}

	// ---------- C 端支付 ----------
	PayMemberOpenidMissing = baseErrs.ErrorCode{Code: 1_004_004_001, Msg: "请先授权微信登录后再使用微信支付"}
	PayReturnUrlInvalid    = baseErrs.ErrorCode{Code: 1_004_004_002, Msg: "支付完成跳转地址必须与本站同域"}

	// ---------- 钱包 ----------
	PayWalletNotExists     = baseErrs.ErrorCode{Code: 1_004_005_001, Msg: "钱包账户【%v】不存在"}
	PayWalletInsufficient  = baseErrs.ErrorCode{Code: 1_004_005_002, Msg: "钱包余额不足，当前可用余额 %v 分"}
	PayWalletAmountInvalid = baseErrs.ErrorCode{Code: 1_004_005_003, Msg: "调整金额不能为 0"}
	PayWalletDisabled      = baseErrs.ErrorCode{Code: 1_004_005_004, Msg: "钱包已被冻结，暂不可用"}
	PayWalletFrozenExceed  = baseErrs.ErrorCode{Code: 1_004_005_005, Msg: "解冻金额不能超过冻结金额（当前冻结 %v 分）"}

	// ---------- 模拟支付 ----------
	PayMockDisabled      = baseErrs.ErrorCode{Code: 1_004_006_001, Msg: "模拟支付未开启，请先在参数配置把 pay.mock.enabled 置为 1"}
	PayMockChannelDenied = baseErrs.ErrorCode{Code: 1_004_006_002, Msg: "该订单不是模拟支付渠道，不能模拟支付"}
)
