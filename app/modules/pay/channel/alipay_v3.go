package channel

import (
	"context"
	"encoding/json"
	"fmt"
	nethttp "net/http"
	"strconv"
	"strings"

	gopay "github.com/go-pay/gopay"
	alipay "github.com/go-pay/gopay/alipay"
	alipayv3 "github.com/go-pay/gopay/alipay/v3"

	"eqray-go-starter/app/modules/base/errs"
	payErrs "eqray-go-starter/app/modules/pay/errs"
)

// AlipayConfig 是支付宝 V3 渠道的 config JSON 结构。
// 公钥模式只填 alipayPublicKey；证书模式填三个证书内容，走 SetCert。
type AlipayConfig struct {
	// AppID 开放平台应用 AppID
	AppID string `json:"appId"`
	// PrivateKey 应用私钥（PKCS8）
	PrivateKey string `json:"privateKey"`
	// AlipayPublicKey 支付宝公钥（公钥模式）
	AlipayPublicKey string `json:"alipayPublicKey"`
	// AppCertContent 应用公钥证书内容（证书模式）
	AppCertContent string `json:"appCertContent"`
	// AlipayRootCertContent 支付宝根证书内容（证书模式）
	AlipayRootCertContent string `json:"alipayRootCertContent"`
	// AlipayPublicCertContent 支付宝公钥证书内容（证书模式）
	AlipayPublicCertContent string `json:"alipayPublicCertContent"`
	// IsProd 是否生产环境；false 走沙箱网关
	IsProd bool `json:"isProd"`
}

// AlipayChannel 是支付宝 V3 渠道适配器，覆盖 pc / wap / app / qr 四种方式。
type AlipayChannel struct {
	// cfg 渠道配置快照
	cfg AlipayConfig
	// client gopay 支付宝 V3 客户端
	client *alipayv3.ClientV3
	// payMethod 支付方式（ali_pc 等 code 的后半段）
	payMethod string
	// code 渠道编码
	code string
}

// newAlipayChannel 由 config JSON 明文构造支付宝 V3 渠道客户端。
// 入参 code 为渠道编码（决定下单方式）、configJSON 为 pay_channel.config 明文；
// 配置缺项或 SDK 初始化失败返回业务错误码。
func newAlipayChannel(code, configJSON string) (Channel, error) {
	var cfg AlipayConfig
	if err := json.Unmarshal([]byte(configJSON), &cfg); err != nil {
		return nil, errs.New(payErrs.PayChannelConfigBad, "支付宝渠道配置不是合法 JSON")
	}
	if strings.TrimSpace(cfg.AppID) == "" || strings.TrimSpace(cfg.PrivateKey) == "" {
		return nil, errs.New(payErrs.PayChannelConfigBad, "支付宝渠道缺少 appId / privateKey")
	}
	if cfg.AlipayPublicKey == "" && (cfg.AppCertContent == "" || cfg.AlipayRootCertContent == "" || cfg.AlipayPublicCertContent == "") {
		return nil, errs.New(payErrs.PayChannelConfigBad, "支付宝渠道需填写 alipayPublicKey（公钥模式）或三份证书内容（证书模式）")
	}

	client, err := alipayv3.NewClientV3(cfg.AppID, cfg.PrivateKey, cfg.IsProd)
	if err != nil {
		return nil, errs.New(payErrs.PayChannelConfigBad, "支付宝私钥解析失败")
	}

	// 证书模式需要显式装载三份证书；公钥模式 SDK 直接用 alipayPublicKey 验签
	if cfg.AppCertContent != "" {
		if err := client.SetCert([]byte(cfg.AppCertContent), []byte(cfg.AlipayRootCertContent), []byte(cfg.AlipayPublicCertContent)); err != nil {
			return nil, errs.New(payErrs.PayChannelConfigBad, "支付宝证书装载失败")
		}
	}

	return &AlipayChannel{cfg: cfg, client: client, payMethod: code[strings.IndexByte(code, '_')+1:], code: code}, nil
}

// Code 返回渠道编码。
func (c *AlipayChannel) Code() string {
	return c.code
}

// CreateOrder 按渠道编码的支付方式向支付宝下单，返回前端拉起支付所需参数。
// 入参 req 为统一下单请求；金额从「分」换算为支付宝要求的「元」字符串。
func (c *AlipayChannel) CreateOrder(ctx context.Context, req CreateOrderRequest) (CreateOrderResult, error) {
	bm := gopay.BodyMap{}
	bm.Set("out_trade_no", req.MerchantOrderNo).
		Set("subject", req.Subject).
		Set("total_amount", centsToYuan(req.Price)).
		Set("notify_url", req.NotifyURL)
	if req.ReturnURL != "" {
		bm.Set("return_url", req.ReturnURL)
	}

	switch c.payMethod {
	case "qr":
		aliRsp, err := c.client.TradePrecreate(ctx, bm)
		if err = c.checkRsp(err, aliRsp.StatusCode, aliRsp.ErrResponse); err != nil {
			return CreateOrderResult{}, err
		}
		return CreateOrderResult{Kind: KindQR, URL: aliRsp.QrCode}, nil

	case "page":
		payURL, err := c.client.TradePagePay(ctx, bm)
		if err != nil {
			return CreateOrderResult{}, errs.New(payErrs.PayOrderCreateFailed, "支付宝下单失败："+err.Error())
		}
		return CreateOrderResult{Kind: KindRedirect, URL: payURL}, nil

	case "wap":
		payURL, err := c.client.TradeWapPay(ctx, bm)
		if err != nil {
			return CreateOrderResult{}, errs.New(payErrs.PayOrderCreateFailed, "支付宝下单失败："+err.Error())
		}
		return CreateOrderResult{Kind: KindRedirect, URL: payURL}, nil

	case "app":
		orderStr, err := c.client.TradeAppPay(ctx, bm)
		if err != nil {
			return CreateOrderResult{}, errs.New(payErrs.PayOrderCreateFailed, "支付宝下单失败："+err.Error())
		}
		return CreateOrderResult{Kind: KindApp, Params: map[string]any{"orderStr": orderStr}}, nil

	default:
		return CreateOrderResult{}, errs.New(payErrs.PayChannelUnsupported, c.code)
	}
}

// QueryOrder 按商户单号查询支付宝侧订单，并统一映射交易状态。
func (c *AlipayChannel) QueryOrder(ctx context.Context, merchantOrderNo string) (QueryOrderResult, error) {
	bm := gopay.BodyMap{}
	bm.Set("out_trade_no", merchantOrderNo)

	aliRsp, err := c.client.TradeQuery(ctx, bm)
	if err = c.checkRsp(err, aliRsp.StatusCode, aliRsp.ErrResponse); err != nil {
		return QueryOrderResult{}, err
	}

	result := QueryOrderResult{
		TradeState:     mapAlipayTradeStatus(aliRsp.TradeStatus),
		ChannelOrderNo: aliRsp.TradeNo,
		ChannelUser:    aliRsp.BuyerUserId,
		SuccessTime:    aliRsp.SendPayDate,
		Raw:            marshalRaw(aliRsp),
	}
	if amount, err := yuanToCents(aliRsp.TotalAmount); err == nil {
		result.Amount = amount
	}

	return result, nil
}

// CloseOrder 关闭支付宝侧订单；订单不存在视为已关闭。
func (c *AlipayChannel) CloseOrder(ctx context.Context, merchantOrderNo string) error {
	bm := gopay.BodyMap{}
	bm.Set("out_trade_no", merchantOrderNo)

	aliRsp, err := c.client.TradeClose(ctx, bm)
	if err = c.checkRsp(err, aliRsp.StatusCode, aliRsp.ErrResponse); err != nil {
		// 订单不存在时渠道返回 ACQ.TRADE_NOT_EXIST，视为已关闭
		if strings.Contains(err.Error(), "TRADE_NOT_EXIST") {
			return nil
		}
		return err
	}

	return nil
}

// Refund 向支付宝发起退款；支付宝退款同步返回结果（fund_change=Y 表示退款成功）。
func (c *AlipayChannel) Refund(ctx context.Context, req RefundRequest) (RefundResult, error) {
	bm := gopay.BodyMap{}
	bm.Set("out_trade_no", req.MerchantOrderNo).
		Set("out_request_no", req.MerchantRefundNo).
		Set("refund_amount", centsToYuan(req.RefundPrice)).
		Set("refund_reason", req.Reason)

	aliRsp, err := c.client.TradeRefund(ctx, bm)
	if err = c.checkRsp(err, aliRsp.StatusCode, aliRsp.ErrResponse); err != nil {
		return RefundResult{}, err
	}

	state := RefundStateFailed
	if aliRsp.FundChange == "Y" {
		state = RefundStateSuccess
	}

	return RefundResult{State: state, ChannelRefundNo: aliRsp.TradeNo, Raw: marshalRaw(aliRsp)}, nil
}

// QueryRefund 支付宝退款同步返回终态，不提供查单能力；调用方应跳过退款中的支付宝单。
func (c *AlipayChannel) QueryRefund(ctx context.Context, merchantRefundNo, merchantOrderNo string) (RefundResult, error) {
	return RefundResult{}, errs.New(payErrs.PayChannelUnsupported, c.code+" 退款查单")
}

// ParseNotify 解析并验签支付宝异步通知（表单报文 + RSA2 验签）。
// 入参 r 为原始请求；返回统一通知结果（原始与解密内容相同，均留档）。
func (c *AlipayChannel) ParseNotify(ctx context.Context, r *nethttp.Request) (NotifyResult, error) {
	bm, err := alipay.ParseNotifyToBodyMap(r)
	if err != nil {
		return NotifyResult{}, errs.New(payErrs.PayNotifyVerifyFail)
	}

	// 公钥模式用支付宝公钥字符串验签；证书模式改用「支付宝公钥证书」验签
	var ok bool
	if strings.TrimSpace(c.cfg.AlipayPublicKey) != "" {
		ok, err = alipay.VerifySign(c.cfg.AlipayPublicKey, bm)
	} else {
		ok, err = alipay.VerifySignWithCert(c.cfg.AlipayPublicCertContent, bm)
	}
	if err != nil || !ok {
		return NotifyResult{}, errs.New(payErrs.PayNotifyVerifyFail)
	}

	body, _ := json.Marshal(bm)
	return NotifyResult{
		MerchantOrderNo: bm.GetString("out_trade_no"),
		ChannelOrderNo:  bm.GetString("trade_no"),
		TradeState:      mapAlipayTradeStatus(bm.GetString("trade_status")),
		Amount:          centsFromYuanOrZero(bm.GetString("total_amount")),
		ChannelUser:     firstNonEmpty(bm.GetString("buyer_open_id"), bm.GetString("buyer_id")),
		SuccessTime:     bm.GetString("gmt_payment"),
		RawRequest:      string(body),
		Decrypted:       string(body),
	}, nil
}

// NotifyAck 返回支付宝要求的成功应答报文（纯文本 success）。
func (c *AlipayChannel) NotifyAck() NotifyAck {
	return NotifyAck{Body: "success", ContentType: "text/plain", Success: true}
}

// NotifyFailAck 返回支付宝要求的失败应答报文，触发渠道按策略重试。
func (c *AlipayChannel) NotifyFailAck(reason string) NotifyAck {
	return NotifyAck{Body: "fail", ContentType: "text/plain"}
}

// checkRsp 把 gopay 支付宝 V3 的响应码 / 错误结构归一为业务错误。
// 入参 err 为 SDK 错误、statusCode 为 HTTP 状态、errResp 为支付宝错误结构。
func (c *AlipayChannel) checkRsp(err error, statusCode int, errResp alipayv3.ErrResponse) error {
	if err != nil {
		return errs.New(payErrs.PayOrderCreateFailed, "支付宝接口调用失败："+err.Error())
	}
	if statusCode < 200 || statusCode >= 300 {
		return errs.New(payErrs.PayOrderCreateFailed,
			fmt.Sprintf("支付宝返回错误：%s %s", errResp.Code, errResp.Message))
	}

	return nil
}

// mapAlipayTradeStatus 把支付宝交易状态映射为统一状态。
func mapAlipayTradeStatus(status string) string {
	switch status {
	case "TRADE_SUCCESS", "TRADE_FINISHED":
		return TradeStatePaid
	case "WAIT_BUYER_PAY":
		return TradeStateUnpaid
	case "TRADE_CLOSED":
		return TradeStateClosed
	case "TRADE_NOT_EXIST":
		return TradeStateUnknown
	default:
		return TradeStateUnknown
	}
}

// centsToYuan 把「分」转换为支付宝要求的「元」字符串，如 1 → "0.01"。
func centsToYuan(cents int64) string {
	return fmt.Sprintf("%d.%02d", cents/100, cents%100)
}

// yuanToCents 把支付宝「元」字符串转换为「分」；非法输入返回错误。
func yuanToCents(yuan string) (int64, error) {
	yuan = strings.TrimSpace(yuan)
	if yuan == "" {
		return 0, fmt.Errorf("empty amount")
	}

	f, err := strconv.ParseFloat(yuan, 64)
	if err != nil {
		return 0, err
	}

	return int64(f*100 + 0.5), nil
}

// centsFromYuanOrZero 把支付宝「元」字符串转换为「分」，非法输入返回 0（通知金额校验会拦下不一致）。
func centsFromYuanOrZero(yuan string) int64 {
	cents, err := yuanToCents(yuan)
	if err != nil {
		return 0
	}
	return cents
}

// firstNonEmpty 返回第一个非空字符串，全部为空时返回空串。
func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}
