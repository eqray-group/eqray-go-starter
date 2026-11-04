package channel

// 渠道厂商。
const (
	VendorWechat = "wx"
	VendorAlipay = "alipay"
	// VendorWallet 站内钱包（余额支付）
	VendorWallet = "wallet"
	// VendorMock 联调用模拟支付
	VendorMock = "mock"
)

// 支持的渠道编码，一行 pay_channel 对应一种支付方式。
const (
	CodeWxNative = "wx_native"
	CodeWxH5     = "wx_h5"
	CodeWxJsapi  = "wx_jsapi"
	CodeWxMini   = "wx_mini"
	CodeWxApp    = "wx_app"

	CodeAliPc  = "ali_pc"
	CodeAliWap = "ali_wap"
	CodeAliApp = "ali_app"
	CodeAliQr  = "ali_qr"

	// CodeWallet 余额支付：站内扣款，无第三方请求
	CodeWallet = "wallet"
	// CodeMock 模拟支付：仅联调使用，需在参数配置开启 pay.mock.enabled
	CodeMock = "mock"
)

// 前端拉起方式（CreateOrderResult.Kind）。
const (
	KindRedirect = "redirect"
	KindQR       = "qr"
	KindJSAPI    = "jsapi"
	KindApp      = "app"
	// KindWallet 余额支付成功，前端直接提示支付完成
	KindWallet = "wallet"
	// KindMock 模拟支付，前端展示「模拟支付」按钮
	KindMock = "mock"
)

// IsWallet 判断是否为余额支付渠道。
// 入参 code 为渠道编码；返回是否为钱包渠道。
func IsWallet(code string) bool {
	return code == CodeWallet
}

// IsMock 判断是否为模拟支付渠道。
// 入参 code 为渠道编码；返回是否为模拟渠道。
func IsMock(code string) bool {
	return code == CodeMock
}

// IsInternal 判断是否为站内渠道（钱包 / 模拟支付）。
// 站内渠道不产生异步通知地址，下单时不需要 APP_URL。
// 入参 code 为渠道编码；返回是否为站内渠道。
func IsInternal(code string) bool {
	return IsWallet(code) || IsMock(code)
}

// 支付下单取会员 openid 时使用的社交渠道类型：
// wx_mini 用小程序 openid，wx_jsapi 用公众号 openid，两者不通用。
const (
	SocialTypeWechatMini = "wechat_mini"
	SocialTypeWechatMp   = "wechat_mp"
)

// Field 是渠道 config JSON 的一个字段定义，由 GET /channels/schema 下发，
// 前端据此动态渲染表单，不硬编码 9 套映射。
type Field struct {
	// Key config JSON 中的键名
	Key string `json:"key"`
	// Label 展示名
	Label string `json:"label"`
	// Type 渲染类型：input / password / textarea / switch
	Type string `json:"type"`
	// Required 是否必填（公钥 / 证书模式二选一的字段由后端校验，这里只标建议值）
	Required bool `json:"required"`
	// Secret 敏感字段：列表 / 详情只回脱敏值，编辑留空表示不修改
	Secret bool `json:"secret"`
	// Rows textarea 行数
	Rows int `json:"rows,omitempty"`
	// Desc 字段说明，展示在表单下方
	Desc string `json:"desc,omitempty"`
}

// Meta 是一个渠道编码的静态元数据。
type Meta struct {
	// Vendor 支付厂商：wx / alipay
	Vendor string
	// PayMethod 支付方式：mini/jsapi/app/native/h5/page/wap/qr
	PayMethod string
	// Fields config 动态表单字段定义
	Fields []Field
	// Build 由渠道编码与 config JSON 明文构造渠道客户端
	Build func(code, configJSON string) (Channel, error)
}

// wxFields 是微信 V3 渠道的 config 字段定义。
// 证书模式与「微信支付公钥」模式二选一：填了 publicKey 走公钥模式，否则平台证书自动下载。
func wxFields() []Field {
	return []Field{
		{Key: "appId", Label: "AppID", Type: "input", Required: true, Desc: "商户对应的公众号 / 小程序 / 应用 AppID"},
		{Key: "mchId", Label: "商户号", Type: "input", Required: true},
		{Key: "apiV3Key", Label: "APIv3 密钥", Type: "password", Required: true, Secret: true, Desc: "32 位 APIv3 密钥，用于回调报文解密"},
		{Key: "serialNo", Label: "商户证书序列号", Type: "input", Required: true},
		{Key: "privateKey", Label: "商户私钥", Type: "textarea", Required: true, Secret: true, Rows: 6, Desc: "apiclient_key.pem 的完整内容"},
		{Key: "merchantCertContent", Label: "商户 API 证书", Type: "textarea", Secret: true, Rows: 6, Desc: "apiclient_cert.pem 的完整内容，选填；填写后保存时会校验证书序列号是否与上方一致"},
		{Key: "publicKeyId", Label: "微信支付公钥 ID", Type: "input", Desc: "公钥模式必填，如 PUB_KEY_ID_xxx"},
		{Key: "publicKey", Label: "微信支付公钥", Type: "textarea", Secret: true, Rows: 6, Desc: "公钥模式必填；留空则走平台证书自动下载"},
	}
}

// aliFields 是支付宝 V3 渠道的 config 字段定义。
// 公钥模式只填 alipayPublicKey；证书模式填三个证书内容。
func aliFields() []Field {
	return []Field{
		{Key: "appId", Label: "应用 AppID", Type: "input", Required: true},
		{Key: "privateKey", Label: "应用私钥", Type: "textarea", Required: true, Secret: true, Rows: 6, Desc: "PKCS8 格式应用私钥"},
		{Key: "alipayPublicKey", Label: "支付宝公钥", Type: "textarea", Secret: true, Rows: 4, Desc: "公钥模式必填"},
		{Key: "appCertContent", Label: "应用公钥证书", Type: "textarea", Secret: true, Rows: 4, Desc: "证书模式必填"},
		{Key: "alipayRootCertContent", Label: "支付宝根证书", Type: "textarea", Secret: true, Rows: 4, Desc: "证书模式必填"},
		{Key: "alipayPublicCertContent", Label: "支付宝公钥证书", Type: "textarea", Secret: true, Rows: 4, Desc: "证书模式必填"},
		{Key: "isProd", Label: "生产环境", Type: "switch", Required: true, Desc: "关闭时走支付宝沙箱网关"},
	}
}

// Metas 是全部受支持渠道编码的元数据表；新增渠道只需在此追加条目。
var Metas = map[string]Meta{
	CodeWxNative: {Vendor: VendorWechat, PayMethod: "native", Fields: wxFields(), Build: newWechatChannel},
	CodeWxH5:     {Vendor: VendorWechat, PayMethod: "h5", Fields: wxFields(), Build: newWechatChannel},
	CodeWxJsapi:  {Vendor: VendorWechat, PayMethod: "jsapi", Fields: wxFields(), Build: newWechatChannel},
	CodeWxMini:   {Vendor: VendorWechat, PayMethod: "mini", Fields: wxFields(), Build: newWechatChannel},
	CodeWxApp:    {Vendor: VendorWechat, PayMethod: "app", Fields: wxFields(), Build: newWechatChannel},

	CodeAliPc:  {Vendor: VendorAlipay, PayMethod: "page", Fields: aliFields(), Build: newAlipayChannel},
	CodeAliWap: {Vendor: VendorAlipay, PayMethod: "wap", Fields: aliFields(), Build: newAlipayChannel},
	CodeAliApp: {Vendor: VendorAlipay, PayMethod: "app", Fields: aliFields(), Build: newAlipayChannel},
	CodeAliQr:  {Vendor: VendorAlipay, PayMethod: "qr", Fields: aliFields(), Build: newAlipayChannel},

	// 站内渠道：没有凭据字段，表单为空；钱包余额不足时下单会被拒绝
	CodeWallet: {Vendor: VendorWallet, PayMethod: "wallet", Fields: nil, Build: newWalletChannel},
	CodeMock:   {Vendor: VendorMock, PayMethod: "mock", Fields: nil, Build: newMockChannel},
}

// SocialTypeForCode 返回某渠道下单所需 openid 对应的社交渠道类型。
// 入参 code 为渠道编码；无需 openid 的渠道返回空串。
func SocialTypeForCode(code string) string {
	switch code {
	case CodeWxMini:
		return SocialTypeWechatMini
	case CodeWxJsapi:
		return SocialTypeWechatMp
	default:
		return ""
	}
}

// SupportedCodes 返回全部受支持的渠道编码，按编码排序，供后台下拉与校验。
func SupportedCodes() []string {
	codes := make([]string, 0, len(Metas))
	for code := range Metas {
		codes = append(codes, code)
	}
	for i := 0; i < len(codes); i++ {
		for j := i + 1; j < len(codes); j++ {
			if codes[j] < codes[i] {
				codes[i], codes[j] = codes[j], codes[i]
			}
		}
	}
	return codes
}
