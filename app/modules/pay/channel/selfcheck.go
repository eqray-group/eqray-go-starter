package channel

import (
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"strings"

	"eqray-go-starter/app/modules/base/errs"
	payErrs "eqray-go-starter/app/modules/pay/errs"
)

// SelfCheck 对渠道凭据做一次离线自检：不发起网络请求，只检查格式与字段间的配套关系。
// 目的是把「密钥长度不对 / 私钥格式不对 / 序列号与证书不配套」这类问题在保存时就暴露，
// 而不是等到用户付款时才发现。
// 入参 code 为渠道编码、configJSON 为 config 明文；返回第一个问题的业务错误，全部通过返回 nil。
func SelfCheck(code, configJSON string) error {
	meta, ok := Metas[code]
	if !ok {
		return errs.New(payErrs.PayChannelUnsupported, code)
	}

	switch meta.Vendor {
	case VendorWechat:
		var cfg WechatConfig
		if err := json.Unmarshal([]byte(configJSON), &cfg); err != nil {
			return errs.New(payErrs.PayChannelConfigBad, "微信渠道配置不是合法 JSON")
		}
		return selfCheckWechat(cfg)
	case VendorAlipay:
		var cfg AlipayConfig
		if err := json.Unmarshal([]byte(configJSON), &cfg); err != nil {
			return errs.New(payErrs.PayChannelConfigBad, "支付宝渠道配置不是合法 JSON")
		}
		return selfCheckAlipay(cfg)
	default:
		return errs.New(payErrs.PayChannelUnsupported, code)
	}
}

// selfCheckWechat 校验微信 V3 凭据：APIv3 密钥长度、私钥格式、证书序列号是否与证书配套。
// 入参 cfg 为微信渠道配置；返回第一个问题的业务错误。
func selfCheckWechat(cfg WechatConfig) error {
	if length := len(strings.TrimSpace(cfg.APIv3Key)); length != 32 {
		return errs.New(payErrs.PayChannelConfigBad, "APIv3 密钥必须为 32 位字符")
	}

	block, _ := pem.Decode([]byte(cfg.PrivateKey))
	if block == nil {
		return errs.New(payErrs.PayChannelConfigBad, "商户私钥不是合法的 PEM 内容，请粘贴 apiclient_key.pem 全文")
	}
	if _, err := x509.ParsePKCS8PrivateKey(block.Bytes); err != nil {
		return errs.New(payErrs.PayChannelConfigBad, "商户私钥解析失败，仅支持 PKCS8 格式（apiclient_key.pem）")
	}

	// 序列号推导必须依赖证书：私钥里没有该信息，所以只有填了证书才做配套校验
	if strings.TrimSpace(cfg.MerchantCertContent) == "" {
		return nil
	}

	certBlock, _ := pem.Decode([]byte(cfg.MerchantCertContent))
	if certBlock == nil {
		return errs.New(payErrs.PayChannelConfigBad, "商户 API 证书不是合法的 PEM 内容，请粘贴 apiclient_cert.pem 全文")
	}
	cert, err := x509.ParseCertificate(certBlock.Bytes)
	if err != nil {
		return errs.New(payErrs.PayChannelConfigBad, "商户 API 证书解析失败")
	}

	serial := strings.ToUpper(hex.EncodeToString(cert.SerialNumber.Bytes()))
	if !strings.EqualFold(serial, strings.TrimSpace(cfg.SerialNo)) {
		return errs.New(payErrs.PayChannelConfigBad, "商户证书序列号与证书内容不一致，请核对 apiclient_cert.pem 对应的序列号")
	}

	return nil
}

// selfCheckAlipay 校验支付宝凭据：应用私钥可解析、验签材料（公钥或三份证书）齐备且可解析。
// 入参 cfg 为支付宝渠道配置；返回第一个问题的业务错误。
func selfCheckAlipay(cfg AlipayConfig) error {
	if err := parseRSAKey(cfg.PrivateKey); err != nil {
		return errs.New(payErrs.PayChannelConfigBad, "应用私钥解析失败，请粘贴 PKCS8 格式私钥")
	}

	if strings.TrimSpace(cfg.AlipayPublicKey) != "" {
		return nil
	}

	// 证书模式：三份证书必须齐备，且都是合法的 X509 证书
	certs := map[string]string{
		"应用公钥证书":  cfg.AppCertContent,
		"支付宝根证书":  cfg.AlipayRootCertContent,
		"支付宝公钥证书": cfg.AlipayPublicCertContent,
	}
	for label, content := range certs {
		if strings.TrimSpace(content) == "" {
			return errs.New(payErrs.PayChannelConfigBad, "证书模式缺少"+label)
		}
		if err := parseCertificate(content); err != nil {
			return errs.New(payErrs.PayChannelConfigBad, label+"解析失败")
		}
	}

	return nil
}

// parseRSAKey 解析 RSA 私钥，兼容 PEM 包裹与纯 base64 两种常见粘贴方式。
// 入参 raw 为私钥内容；返回解析错误。
func parseRSAKey(raw string) error {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return errs.New(payErrs.PayChannelConfigBad, "应用私钥为空")
	}

	var der []byte
	if block, _ := pem.Decode([]byte(raw)); block != nil {
		der = block.Bytes
	} else {
		decoded, err := base64.StdEncoding.DecodeString(strings.ReplaceAll(raw, "\n", ""))
		if err != nil {
			return err
		}
		der = decoded
	}

	if _, err := x509.ParsePKCS8PrivateKey(der); err != nil {
		// 兼容 PKCS1（部分工具导出的老格式）
		if _, pkcs1Err := x509.ParsePKCS1PrivateKey(der); pkcs1Err != nil {
			return err
		}
	}

	return nil
}

// parseCertificate 解析 X509 证书，兼容 PEM 与纯 base64。
// 入参 raw 为证书内容；返回解析错误。
func parseCertificate(raw string) error {
	raw = strings.TrimSpace(raw)

	var der []byte
	if block, _ := pem.Decode([]byte(raw)); block != nil {
		der = block.Bytes
	} else {
		decoded, err := base64.StdEncoding.DecodeString(strings.ReplaceAll(raw, "\n", ""))
		if err != nil {
			return err
		}
		der = decoded
	}

	_, err := x509.ParseCertificate(der)

	return err
}
