package channel

import (
	"context"
	"encoding/json"
	"fmt"
	nethttp "net/http"
	"net/url"
	"strings"

	gopay "github.com/go-pay/gopay"
	wechatv3 "github.com/go-pay/gopay/wechat/v3"

	"eqray-go-starter/app/modules/base/errs"
	payErrs "eqray-go-starter/app/modules/pay/errs"
)

// WechatConfig 是微信 V3 渠道的 config JSON 结构。
// 证书模式与「微信支付公钥」模式二选一：填了 PublicKey 走公钥模式，否则平台证书自动下载。
type WechatConfig struct {
	// AppID 商户对应的公众号 / 小程序 / 应用 AppID
	AppID string `json:"appId"`
	// MchID 直连商户号
	MchID string `json:"mchId"`
	// APIv3Key 32 位 APIv3 密钥
	APIv3Key string `json:"apiV3Key"`
	// SerialNo 商户证书序列号
	SerialNo string `json:"serialNo"`
	// PrivateKey 商户私钥（apiclient_key.pem 内容）
	PrivateKey string `json:"privateKey"`
	// MerchantCertContent 商户 API 证书内容（apiclient_cert.pem，选填）
	// 仅用于自检「serialNo 与私钥是否配套」：私钥本身推导不出序列号，需要证书参与比对
	MerchantCertContent string `json:"merchantCertContent"`
	// PublicKeyID 微信支付公钥 ID（公钥模式）
	PublicKeyID string `json:"publicKeyId"`
	// PublicKey 微信支付公钥内容（公钥模式）
	PublicKey string `json:"publicKey"`
}

// WechatChannel 是微信 V3 渠道适配器，覆盖 native / h5 / jsapi / mini / app 五种方式。
type WechatChannel struct {
	// cfg 渠道配置快照
	cfg WechatConfig
	// client gopay 微信 V3 客户端
	client *wechatv3.ClientV3
	// payMethod 支付方式（wx_native 等 code 的后半段）
	payMethod string
	// code 渠道编码
	code string
}

// newWechatChannel 由 config JSON 明文构造微信 V3 渠道客户端。
// 入参 code 为渠道编码（决定下单方式）、configJSON 为 pay_channel.config 明文；
// 配置缺项或 SDK 初始化失败返回业务错误码。
func newWechatChannel(code, configJSON string) (Channel, error) {
	var cfg WechatConfig
	if err := json.Unmarshal([]byte(configJSON), &cfg); err != nil {
		return nil, errs.New(payErrs.PayChannelConfigBad, "微信渠道配置不是合法 JSON")
	}
	if strings.TrimSpace(cfg.AppID) == "" || strings.TrimSpace(cfg.MchID) == "" ||
		strings.TrimSpace(cfg.APIv3Key) == "" || strings.TrimSpace(cfg.SerialNo) == "" ||
		strings.TrimSpace(cfg.PrivateKey) == "" {
		return nil, errs.New(payErrs.PayChannelConfigBad, "微信渠道缺少 appId / mchId / apiV3Key / serialNo / privateKey")
	}

	client, err := wechatv3.NewClientV3(cfg.MchID, cfg.SerialNo, cfg.APIv3Key, cfg.PrivateKey)
	if err != nil {
		return nil, errs.New(payErrs.PayChannelConfigBad, "微信商户私钥解析失败")
	}

	if strings.TrimSpace(cfg.PublicKey) != "" {
		// 公钥模式：直接用微信支付公钥验签，无需平台证书
		if err := client.AutoVerifySignByPublicKey([]byte(cfg.PublicKey), cfg.PublicKeyID); err != nil {
			return nil, errs.New(payErrs.PayChannelConfigBad, "微信支付公钥加载失败")
		}
	} else {
		// 平台证书模式：首次调用自动下载并周期刷新
		if err := client.AutoVerifySign(true); err != nil {
			return nil, errs.New(payErrs.PayChannelConfigBad, "微信平台证书初始化失败")
		}
	}

	return &WechatChannel{cfg: cfg, client: client, payMethod: code[strings.IndexByte(code, '_')+1:], code: code}, nil
}

// Code 返回渠道编码。
func (c *WechatChannel) Code() string {
	return c.code
}

// CreateOrder 按渠道编码的支付方式向微信下单，返回前端拉起支付所需参数。
// 入参 req 为统一下单请求；金额单位分，内部不换算。
func (c *WechatChannel) CreateOrder(ctx context.Context, req CreateOrderRequest) (CreateOrderResult, error) {
	bm := gopay.BodyMap{}
	bm.Set("appid", c.cfg.AppID).
		Set("mchid", c.cfg.MchID).
		Set("description", req.Subject).
		Set("out_trade_no", req.MerchantOrderNo).
		Set("notify_url", req.NotifyURL).
		SetBodyMap("amount", func(b gopay.BodyMap) {
			b.Set("total", req.Price).Set("currency", "CNY")
		})

	switch c.payMethod {
	case "native":
		wxRsp, err := c.client.V3TransactionNative(ctx, bm)
		if err = c.checkRsp(err, wxRsp.Code, wxRsp.Error); err != nil {
			return CreateOrderResult{}, err
		}
		return CreateOrderResult{Kind: KindQR, URL: wxRsp.Response.CodeUrl}, nil

	case "h5":
		bm.SetBodyMap("scene_info", func(b gopay.BodyMap) {
			b.Set("payer_client_ip", req.ClientIP).
				SetBodyMap("h5_info", func(h gopay.BodyMap) { h.Set("type", "Wap") })
		})
		wxRsp, err := c.client.V3TransactionH5(ctx, bm)
		if err = c.checkRsp(err, wxRsp.Code, wxRsp.Error); err != nil {
			return CreateOrderResult{}, err
		}
		// 微信 H5 的回跳地址通过 h5_url 的 redirect_url 参数传递，且域名需在商户平台配置
		h5URL := wxRsp.Response.H5Url
		if req.ReturnURL != "" {
			h5URL += "&redirect_url=" + url.QueryEscape(req.ReturnURL)
		}
		return CreateOrderResult{Kind: KindRedirect, URL: h5URL}, nil

	case "jsapi", "mini":
		if strings.TrimSpace(req.OpenID) == "" {
			return CreateOrderResult{}, errs.New(payErrs.PayMemberOpenidMissing)
		}
		bm.SetBodyMap("payer", func(b gopay.BodyMap) { b.Set("openid", req.OpenID) })
		wxRsp, err := c.client.V3TransactionJsapi(ctx, bm)
		if err = c.checkRsp(err, wxRsp.Code, wxRsp.Error); err != nil {
			return CreateOrderResult{}, err
		}
		// 小程序与公众号的调起签名函数不同，不能混用
		if c.payMethod == "mini" {
			return c.buildAppletParams(wxRsp.Response.PrepayId)
		}
		return c.buildJSAPIParams(wxRsp.Response.PrepayId)

	case "app":
		wxRsp, err := c.client.V3TransactionApp(ctx, bm)
		if err = c.checkRsp(err, wxRsp.Code, wxRsp.Error); err != nil {
			return CreateOrderResult{}, err
		}
		return c.buildAppParams(wxRsp.Response.PrepayId)

	default:
		return CreateOrderResult{}, errs.New(payErrs.PayChannelUnsupported, c.code)
	}
}

// buildJSAPIParams 生成公众号（WeixinJSBridge）拉起支付所需签名参数。
// 入参 prepayId 为微信预支付会话标识；返回 kind=jsapi 的统一结果。
func (c *WechatChannel) buildJSAPIParams(prepayId string) (CreateOrderResult, error) {
	jsapi, err := c.client.PaySignOfJSAPI(c.cfg.AppID, prepayId)
	if err != nil {
		return CreateOrderResult{}, errs.New(payErrs.PayOrderCreateFailed, "生成公众号调起参数失败")
	}

	return CreateOrderResult{Kind: KindJSAPI, Params: map[string]any{
		"appId":     jsapi.AppId,
		"timeStamp": jsapi.TimeStamp,
		"nonceStr":  jsapi.NonceStr,
		"package":   jsapi.Package,
		"signType":  jsapi.SignType,
		"paySign":   jsapi.PaySign,
	}}, nil
}

// buildAppletParams 生成小程序 wx.requestPayment 所需签名参数。
// 入参 prepayId 为微信预支付会话标识；返回 kind=jsapi 的统一结果。
func (c *WechatChannel) buildAppletParams(prepayId string) (CreateOrderResult, error) {
	applet, err := c.client.PaySignOfApplet(c.cfg.AppID, prepayId)
	if err != nil {
		return CreateOrderResult{}, errs.New(payErrs.PayOrderCreateFailed, "生成小程序调起参数失败")
	}

	return CreateOrderResult{Kind: KindJSAPI, Params: map[string]any{
		"timeStamp": applet.TimeStamp,
		"nonceStr":  applet.NonceStr,
		"package":   applet.Package,
		"signType":  applet.SignType,
		"paySign":   applet.PaySign,
	}}, nil
}

// buildAppParams 生成原生 App 拉起微信支付所需签名参数。
// 入参 prepayId 为微信预支付会话标识；返回 kind=app 的统一结果。
func (c *WechatChannel) buildAppParams(prepayId string) (CreateOrderResult, error) {
	app, err := c.client.PaySignOfApp(c.cfg.AppID, prepayId)
	if err != nil {
		return CreateOrderResult{}, errs.New(payErrs.PayOrderCreateFailed, "生成 App 调起参数失败")
	}

	return CreateOrderResult{Kind: KindApp, Params: map[string]any{
		"appid":     app.Appid,
		"partnerid": app.Partnerid,
		"prepayid":  app.Prepayid,
		"package":   app.Package,
		"noncestr":  app.Noncestr,
		"timestamp": app.Timestamp,
		"sign":      app.Sign,
	}}, nil
}

// QueryOrder 按商户单号查询微信侧订单，并统一映射交易状态。
func (c *WechatChannel) QueryOrder(ctx context.Context, merchantOrderNo string) (QueryOrderResult, error) {
	wxRsp, err := c.client.V3TransactionQueryOrder(ctx, wechatv3.OutTradeNo, merchantOrderNo)
	if err = c.checkRsp(err, wxRsp.Code, wxRsp.Error); err != nil {
		return QueryOrderResult{}, err
	}

	resp := wxRsp.Response
	result := QueryOrderResult{
		TradeState:     mapWechatTradeState(resp.TradeState),
		ChannelOrderNo: resp.TransactionId,
		SuccessTime:    resp.SuccessTime,
		Raw:            marshalRaw(resp),
	}
	if resp.Payer != nil {
		result.ChannelUser = resp.Payer.Openid
	}
	if resp.Amount != nil {
		result.Amount = int64(resp.Amount.Total)
	}

	return result, nil
}

// CloseOrder 关闭微信侧订单；订单不存在或已关闭视为成功。
func (c *WechatChannel) CloseOrder(ctx context.Context, merchantOrderNo string) error {
	wxRsp, err := c.client.V3TransactionCloseOrder(ctx, merchantOrderNo)
	if err != nil {
		return err
	}
	if wxRsp.Code == nethttp.StatusOK {
		return nil
	}
	// ORDER_NOT_EXIST / ORDER_CLOSED 视为已关闭
	if strings.Contains(wxRsp.Error, "ORDER_NOT_EXIST") || strings.Contains(wxRsp.Error, "ORDER_CLOSED") {
		return nil
	}

	return errs.New(payErrs.PayOrderCreateFailed, "微信关单失败："+wxRsp.Error)
}

// Refund 向微信发起退款；微信退款是异步的，返回「受理中」状态，由定时任务查单收敛。
func (c *WechatChannel) Refund(ctx context.Context, req RefundRequest) (RefundResult, error) {
	bm := gopay.BodyMap{}
	bm.Set("out_trade_no", req.MerchantOrderNo).
		Set("out_refund_no", req.MerchantRefundNo).
		Set("reason", req.Reason).
		SetBodyMap("amount", func(b gopay.BodyMap) {
			b.Set("refund", req.RefundPrice).
				Set("total", req.TotalPrice).
				Set("currency", "CNY")
		})

	wxRsp, err := c.client.V3Refund(ctx, bm)
	if err = c.checkRsp(err, wxRsp.Code, wxRsp.Error); err != nil {
		return RefundResult{}, err
	}

	resp := wxRsp.Response
	state := RefundStatePending
	switch resp.Status {
	case "SUCCESS":
		state = RefundStateSuccess
	case "ABNORMAL", "CLOSED":
		state = RefundStateFailed
	}

	return RefundResult{State: state, ChannelRefundNo: resp.RefundId, Raw: marshalRaw(resp)}, nil
}

// QueryRefund 按商户退款单号查询微信侧退款状态。
// 微信退款查询只认 out_refund_no（路径参数），额外拼 out_trade_no 会被判为非法参数，故不传 bm。
// 入参 merchantRefundNo 为商户退款单号；merchantOrderNo 仅为接口统一保留，微信侧不使用。
func (c *WechatChannel) QueryRefund(ctx context.Context, merchantRefundNo, merchantOrderNo string) (RefundResult, error) {
	_ = merchantOrderNo

	wxRsp, err := c.client.V3RefundQuery(ctx, merchantRefundNo, nil)
	if err = c.checkRsp(err, wxRsp.Code, wxRsp.Error); err != nil {
		return RefundResult{}, err
	}

	resp := wxRsp.Response
	state := RefundStatePending
	switch resp.Status {
	case "SUCCESS":
		state = RefundStateSuccess
	case "ABNORMAL", "CLOSED":
		state = RefundStateFailed
	}

	return RefundResult{State: state, ChannelRefundNo: resp.RefundId, Raw: marshalRaw(resp)}, nil
}

// ParseNotify 解析微信支付异步通知：验签 → 解密，二者任一失败即返回错误。
// 入参 r 为原始请求；返回统一通知结果（含原始密文与解密报文，供日志留档）。
func (c *WechatChannel) ParseNotify(ctx context.Context, r *nethttp.Request) (NotifyResult, error) {
	notifyReq, err := wechatv3.V3ParseNotify(r)
	if err != nil {
		return NotifyResult{}, errs.New(payErrs.PayNotifyVerifyFail)
	}

	// 优先按序列号查公钥（平台证书模式），退化为单一公钥（公钥模式）
	if pkMap := c.client.WxPublicKeyMap(); len(pkMap) > 0 {
		if err := notifyReq.VerifySignByPKMap(pkMap); err != nil {
			return NotifyResult{}, errs.New(payErrs.PayNotifyVerifyFail)
		}
	} else if err := notifyReq.VerifySignByPK(c.client.WxPublicKey()); err != nil {
		return NotifyResult{}, errs.New(payErrs.PayNotifyVerifyFail)
	}

	result, err := notifyReq.DecryptPayCipherText(c.cfg.APIv3Key)
	if err != nil {
		return NotifyResult{}, errs.New(payErrs.PayNotifyVerifyFail)
	}

	rawReq, _ := json.Marshal(notifyReq)
	return NotifyResult{
		MerchantOrderNo: result.OutTradeNo,
		ChannelOrderNo:  result.TransactionId,
		TradeState:      mapWechatTradeState(result.TradeState),
		Amount:          int64(result.Amount.Total),
		ChannelUser:     payerOpenid(result.Payer),
		SuccessTime:     result.SuccessTime,
		RawRequest:      string(rawReq),
		Decrypted:       marshalRaw(result),
	}, nil
}

// NotifyAck 返回微信要求的成功应答报文。
func (c *WechatChannel) NotifyAck() NotifyAck {
	return NotifyAck{Body: `{"code":"SUCCESS","message":"成功"}`, ContentType: "application/json", Success: true}
}

// NotifyFailAck 返回微信要求的失败应答报文，触发渠道按策略重试。
func (c *WechatChannel) NotifyFailAck(reason string) NotifyAck {
	body, _ := json.Marshal(map[string]string{"code": "FAIL", "message": reason})
	return NotifyAck{Body: string(body), ContentType: "application/json"}
}

// checkRsp 把 gopay 的响应码 / 错误信息归一为业务错误。
// 入参 err 为 SDK 错误、code 为 HTTP 状态、errBody 为微信错误体。
func (c *WechatChannel) checkRsp(err error, code int, errBody string) error {
	if err != nil {
		return errs.New(payErrs.PayOrderCreateFailed, "微信接口调用失败："+err.Error())
	}
	if code == nethttp.StatusOK {
		return nil
	}

	msg := strings.TrimSpace(errBody)
	if msg == "" {
		msg = fmt.Sprintf("HTTP %d", code)
	}

	return errs.New(payErrs.PayOrderCreateFailed, "微信返回错误："+msg)
}

// mapWechatTradeState 把微信交易状态映射为统一状态。
func mapWechatTradeState(state string) string {
	switch state {
	case "SUCCESS", "REFUND":
		return TradeStatePaid
	case "NOTPAY", "USERPAYING", "PAYERROR":
		return TradeStateUnpaid
	case "CLOSED", "REVOKED":
		return TradeStateClosed
	default:
		return TradeStateUnknown
	}
}

// payerOpenid 从付款人结构取 openid，空结构返回空串。
func payerOpenid(payer *wechatv3.Payer) string {
	if payer == nil {
		return ""
	}
	return payer.Openid
}

// marshalRaw 把渠道响应序列化为 JSON 字符串，失败返回空串。
func marshalRaw(v any) string {
	data, err := json.Marshal(v)
	if err != nil {
		return ""
	}
	return string(data)
}
