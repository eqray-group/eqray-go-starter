// Package errs 定义 member 模块的业务错误码，编码规则见 docs 8.3(2)。
package errs

import baseErrs "eqray-go-starter/app/modules/base/errs"

// member 模块业务错误码：1-003-{子模块}-{错误码}，
// user 占用 001，level 占用 002，tag 占用 003，group 占用 004，social 占用 005，auth 占用 006。
var (
	MemberNotExists    = baseErrs.ErrorCode{Code: 1_003_001_001, Msg: "会员【%v】不存在"}
	MemberMobileExists = baseErrs.ErrorCode{Code: 1_003_001_002, Msg: "手机号【%v】已被其他会员使用"}
	// MemberLevelUnavailable 用于会员档案引用了不存在或已停用的等级
	MemberLevelUnavailable = baseErrs.ErrorCode{Code: 1_003_001_003, Msg: "会员等级【%v】不存在或已停用"}
	MemberPointNotEnough   = baseErrs.ErrorCode{Code: 1_003_001_004, Msg: "积分不足，当前余额【%v】"}

	MemberLevelNotExists = baseErrs.ErrorCode{Code: 1_003_002_001, Msg: "会员等级【%v】不存在"}
	MemberLevelExists    = baseErrs.ErrorCode{Code: 1_003_002_002, Msg: "等级序号【%v】已被占用"}
	MemberLevelInUse     = baseErrs.ErrorCode{Code: 1_003_002_003, Msg: "等级【%v】仍有会员在使用，不可删除"}

	MemberTagNotExists   = baseErrs.ErrorCode{Code: 1_003_003_001, Msg: "会员标签【%v】不存在"}
	MemberGroupNotExists = baseErrs.ErrorCode{Code: 1_003_004_001, Msg: "会员分组【%v】不存在"}

	// social 占用 005 子模块：第三方登录渠道、授权码与绑定关系
	MemberSocialClientUnavailable = baseErrs.ErrorCode{Code: 1_003_005_001, Msg: "第三方登录渠道【%v】未配置或已停用"}
	MemberSocialCodeInvalid       = baseErrs.ErrorCode{Code: 1_003_005_002, Msg: "第三方授权码无效或已过期"}
	MemberSocialAccountNotBound   = baseErrs.ErrorCode{Code: 1_003_005_003, Msg: "该第三方账号尚未绑定会员"}
	MemberSocialAlreadyBound      = baseErrs.ErrorCode{Code: 1_003_005_004, Msg: "该第三方账号已绑定其他会员"}
	MemberSocialBindingNotExists  = baseErrs.ErrorCode{Code: 1_003_005_005, Msg: "绑定关系【%v】不存在"}
	MemberSocialLastBinding       = baseErrs.ErrorCode{Code: 1_003_005_006, Msg: "至少保留一个第三方绑定，不能全部解绑"}
	MemberSocialRequestFailed     = baseErrs.ErrorCode{Code: 1_003_005_007, Msg: "第三方接口调用失败：%v"}
	MemberDisabled                = baseErrs.ErrorCode{Code: 1_003_005_008, Msg: "会员账号已被禁用"}
	MemberSocialTypeDuplicated    = baseErrs.ErrorCode{Code: 1_003_005_009, Msg: "该会员已绑定【%v】，请先解绑"}
	MemberSocialClientNotExists   = baseErrs.ErrorCode{Code: 1_003_005_010, Msg: "第三方登录配置【%v】不存在"}
	MemberSocialClientTypeExists  = baseErrs.ErrorCode{Code: 1_003_005_011, Msg: "第三方登录渠道【%v】已配置"}

	// auth 占用 006 子模块：手机号短信验证码登录 / 注册
	MemberMobileInvalid         = baseErrs.ErrorCode{Code: 1_003_006_001, Msg: "手机号格式不正确"}
	MemberSmsCodeTooFrequent    = baseErrs.ErrorCode{Code: 1_003_006_002, Msg: "验证码发送过于频繁，请稍后再试"}
	MemberSmsCodeDailyLimit     = baseErrs.ErrorCode{Code: 1_003_006_003, Msg: "该手机号今日验证码发送次数已达上限"}
	MemberSmsSendFailed         = baseErrs.ErrorCode{Code: 1_003_006_004, Msg: "短信验证码发送失败"}
	MemberSmsCodeWrong          = baseErrs.ErrorCode{Code: 1_003_006_005, Msg: "短信验证码不正确或已过期"}
	MemberSmsChannelNotExists   = baseErrs.ErrorCode{Code: 1_003_006_006, Msg: "短信渠道【%v】不存在或已停用"}
	MemberSmsTemplateNotExists  = baseErrs.ErrorCode{Code: 1_003_006_007, Msg: "短信模板【%v】不存在或已停用"}
	MemberSmsChannelCodeExists  = baseErrs.ErrorCode{Code: 1_003_006_008, Msg: "短信渠道代码【%v】已存在"}
	MemberSmsTemplateCodeExists = baseErrs.ErrorCode{Code: 1_003_006_009, Msg: "短信模板代码【%v】已存在"}
	MemberSmsChannelInUse       = baseErrs.ErrorCode{Code: 1_003_006_010, Msg: "渠道【%v】下仍有模板，不可删除"}
)
