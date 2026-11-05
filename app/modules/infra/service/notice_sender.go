package service

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"mime"
	"net"
	"net/http"
	"net/smtp"
	"net/url"
	"strconv"
	"strings"
	"time"

	"eqray-go-starter/app/modules/infra/model"
)

// NoticeFieldSchema 描述渠道配置里的一个字段，前端据此动态渲染表单。
type NoticeFieldSchema struct {
	// Key 字段名，落库时作为 JSON 的 key
	Key string `json:"key"`
	// Label 显示名
	Label string `json:"label"`
	// Type 输入类型：text / password / number / select
	Type string `json:"type"`
	// Required 是否必填
	Required bool `json:"required"`
	// Placeholder 占位提示
	Placeholder string `json:"placeholder"`
	// Desc 字段说明
	Desc string `json:"desc"`
	// Options select 类型的选项
	Options []NoticeFieldOption `json:"options,omitempty"`
}

// NoticeFieldOption 是 select 字段的选项。
type NoticeFieldOption struct {
	// Label 显示名
	Label string `json:"label"`
	// Value 值
	Value string `json:"value"`
}

// NoticeTypeSchema 描述一种渠道类型及其配置字段。
type NoticeTypeSchema struct {
	// Type 类型标识：email / wecom / dingtalk
	Type string `json:"type"`
	// Name 类型名称
	Name string `json:"name"`
	// Desc 类型说明
	Desc string `json:"desc"`
	// TargetLabel 接收目标在页面上的叫法（邮箱地址 / 无需填写）
	TargetLabel string `json:"targetLabel"`
	// Fields 配置字段
	Fields []NoticeFieldSchema `json:"fields"`
}

// NoticeTypes 返回支持的渠道类型与配置字段定义，供后台表单渲染。
// 入参无；返回类型列表。
func NoticeTypes() []NoticeTypeSchema {
	return []NoticeTypeSchema{
		{
			Type:        model.NoticeTypeEmail,
			Name:        "邮件（SMTP）",
			Desc:        "直接连接 SMTP 服务器投递，支持 SSL / STARTTLS",
			TargetLabel: "收件人邮箱",
			Fields: []NoticeFieldSchema{
				{Key: "host", Label: "SMTP 服务器", Type: "text", Required: true, Placeholder: "smtp.qq.com"},
				{Key: "port", Label: "端口", Type: "number", Required: true, Placeholder: "465", Desc: "SSL 常用 465，STARTTLS 常用 587"},
				{Key: "username", Label: "账号", Type: "text", Required: true, Placeholder: "noreply@example.com"},
				{Key: "password", Label: "密码 / 授权码", Type: "password", Required: true},
				{Key: "from", Label: "发件地址", Type: "text", Required: true, Placeholder: "noreply@example.com"},
				{Key: "fromName", Label: "发件人名称", Type: "text", Required: false, Placeholder: "系统通知"},
				{
					Key: "encryption", Label: "加密方式", Type: "select", Required: true,
					Desc: "465 端口选 SSL；587 端口选 STARTTLS",
					Options: []NoticeFieldOption{
						{Label: "SSL（隐式 TLS）", Value: "ssl"},
						{Label: "STARTTLS", Value: "tls"},
						{Label: "不加密", Value: "none"},
					},
				},
			},
		},
		{
			Type:        model.NoticeTypeWeCom,
			Name:        "企业微信群机器人",
			Desc:        "群设置里添加机器人后拿到的 Webhook 地址",
			TargetLabel: "无需填写（发到机器人所在群）",
			Fields: []NoticeFieldSchema{
				{Key: "webhook", Label: "Webhook 地址", Type: "password", Required: true, Placeholder: "https://qyapi.weixin.qq.com/cgi-bin/webhook/send?key=..."},
			},
		},
		{
			Type:        model.NoticeTypeDingTalk,
			Name:        "钉钉群机器人",
			Desc:        "支持加签校验；填了密钥会按时间戳签名",
			TargetLabel: "无需填写（发到机器人所在群）",
			Fields: []NoticeFieldSchema{
				{Key: "webhook", Label: "Webhook 地址", Type: "password", Required: true, Placeholder: "https://oapi.dingtalk.com/robot/send?access_token=..."},
				{Key: "secret", Label: "加签密钥", Type: "password", Required: false, Placeholder: "SEC...", Desc: "机器人安全设置选「加签」时填写"},
			},
		},
	}
}

// noticeTypeByName 按类型标识取定义。
// 入参 noticeType 为类型标识；返回定义与是否存在。
func noticeTypeByName(noticeType string) (NoticeTypeSchema, bool) {
	for _, item := range NoticeTypes() {
		if item.Type == noticeType {
			return item, true
		}
	}

	return NoticeTypeSchema{}, false
}

// validateNoticeConfig 校验渠道配置的必填项是否齐全。
// 入参 noticeType 为类型、cfg 为配置；返回业务错误。
func validateNoticeConfig(noticeType string, cfg map[string]string) error {
	schema, ok := noticeTypeByName(noticeType)
	if !ok {
		return errors.New("不支持的渠道类型：" + noticeType)
	}

	missing := make([]string, 0)
	for _, field := range schema.Fields {
		if !field.Required {
			continue
		}
		if strings.TrimSpace(cfg[field.Key]) == "" {
			missing = append(missing, field.Label)
		}
	}
	if len(missing) > 0 {
		return errors.New("缺少必填配置：" + strings.Join(missing, "、"))
	}

	return nil
}

// sendNotice 按渠道类型分发到对应发送实现。
// 入参 noticeType 为类型、cfg 为配置、target 为接收目标、title/content 为内容；返回发送错误。
func sendNotice(ctx context.Context, noticeType string, cfg map[string]string, target, title, content string) error {
	switch noticeType {
	case model.NoticeTypeEmail:
		return sendByEmail(ctx, cfg, target, title, content)
	case model.NoticeTypeWeCom:
		return sendByWeCom(ctx, cfg, title, content)
	case model.NoticeTypeDingTalk:
		return sendByDingTalk(ctx, cfg, title, content)
	}

	return errors.New("不支持的渠道类型：" + noticeType)
}

// sendByEmail 通过 SMTP 投递邮件，支持 SSL / STARTTLS / 明文三种连接方式。
// 入参 cfg 为渠道配置、target 为收件人（逗号分隔可多个）、title/content 为内容；返回发送错误。
func sendByEmail(ctx context.Context, cfg map[string]string, target, title, content string) error {
	host := strings.TrimSpace(cfg["host"])
	port := strings.TrimSpace(cfg["port"])
	username := strings.TrimSpace(cfg["username"])
	password := cfg["password"]
	from := strings.TrimSpace(cfg["from"])
	fromName := strings.TrimSpace(cfg["fromName"])
	encryption := strings.ToLower(strings.TrimSpace(cfg["encryption"]))
	if encryption == "" {
		encryption = "ssl"
	}

	if host == "" || port == "" || from == "" {
		return errors.New("SMTP 配置不完整")
	}

	recipients := splitEmails(target)
	if len(recipients) == 0 {
		return errors.New("收件人不能为空")
	}

	if _, err := strconv.Atoi(port); err != nil {
		return errors.New("SMTP 端口不合法")
	}

	addr := net.JoinHostPort(host, port)
	dialer := &net.Dialer{Timeout: 10 * time.Second}

	var (
		conn net.Conn
		err  error
	)
	if encryption == "ssl" {
		conn, err = tls.DialWithDialer(dialer, "tcp", addr, &tls.Config{ServerName: host})
	} else {
		conn, err = dialer.DialContext(ctx, "tcp", addr)
	}
	if err != nil {
		return fmt.Errorf("连接 SMTP 失败：%w", err)
	}
	defer conn.Close()

	client, err := smtp.NewClient(conn, host)
	if err != nil {
		return fmt.Errorf("初始化 SMTP 失败：%w", err)
	}
	defer client.Close()

	if encryption == "tls" {
		if err := client.StartTLS(&tls.Config{ServerName: host}); err != nil {
			return fmt.Errorf("STARTTLS 失败：%w", err)
		}
	}

	if username != "" {
		if err := client.Auth(smtp.PlainAuth("", username, password, host)); err != nil {
			return fmt.Errorf("SMTP 认证失败：%w", err)
		}
	}

	if err := client.Mail(from); err != nil {
		return fmt.Errorf("设置发件人失败：%w", err)
	}
	for _, recipient := range recipients {
		if err := client.Rcpt(recipient); err != nil {
			return fmt.Errorf("设置收件人 %s 失败：%w", recipient, err)
		}
	}

	writer, err := client.Data()
	if err != nil {
		return fmt.Errorf("开始写入邮件失败：%w", err)
	}
	if _, err := writer.Write(buildMailMessage(from, fromName, recipients, title, content)); err != nil {
		return fmt.Errorf("写入邮件内容失败：%w", err)
	}
	if err := writer.Close(); err != nil {
		return fmt.Errorf("提交邮件失败：%w", err)
	}

	return client.Quit()
}

// buildMailMessage 组装 MIME 邮件原文（UTF-8 纯文本）。
// 入参 from/fromName/recipients/title/content；返回字节内容。
func buildMailMessage(from, fromName string, recipients []string, title, content string) []byte {
	var buf bytes.Buffer

	fromHeader := from
	if fromName != "" {
		fromHeader = fmt.Sprintf("%s <%s>", mime.QEncoding.Encode("UTF-8", fromName), from)
	}

	buf.WriteString("From: " + fromHeader + "\r\n")
	buf.WriteString("To: " + strings.Join(recipients, ", ") + "\r\n")
	buf.WriteString("Subject: " + mime.QEncoding.Encode("UTF-8", title) + "\r\n")
	buf.WriteString("MIME-Version: 1.0\r\n")
	buf.WriteString("Content-Type: text/plain; charset=UTF-8\r\n")
	buf.WriteString("Content-Transfer-Encoding: 8bit\r\n")
	buf.WriteString("Date: " + time.Now().Format(time.RFC1123Z) + "\r\n")
	buf.WriteString("\r\n")
	buf.WriteString(content)

	return buf.Bytes()
}

// splitEmails 拆分逗号 / 分号 / 空白分隔的收件人并去掉空项。
// 入参 raw 为原始字符串；返回收件人列表。
func splitEmails(raw string) []string {
	fields := strings.FieldsFunc(raw, func(r rune) bool {
		return r == ',' || r == ';' || r == '\n' || r == ' ' || r == '\t'
	})

	recipients := make([]string, 0, len(fields))
	for _, field := range fields {
		if field = strings.TrimSpace(field); field != "" {
			recipients = append(recipients, field)
		}
	}

	return recipients
}

// sendByWeCom 发送企业微信群机器人 markdown 消息。
// 入参 cfg 为配置、title/content 为内容；返回发送错误。
func sendByWeCom(ctx context.Context, cfg map[string]string, title, content string) error {
	webhook := strings.TrimSpace(cfg["webhook"])
	if webhook == "" {
		return errors.New("企业微信 Webhook 不能为空")
	}

	payload := map[string]any{
		"msgtype": "markdown",
		"markdown": map[string]string{
			"content": joinTitleContent(title, content),
		},
	}

	return postNoticeJSON(ctx, webhook, payload, "企业微信")
}

// sendByDingTalk 发送钉钉群机器人 markdown 消息，配置了密钥时按加签规则补参数。
// 入参 cfg 为配置、title/content 为内容；返回发送错误。
func sendByDingTalk(ctx context.Context, cfg map[string]string, title, content string) error {
	webhook := strings.TrimSpace(cfg["webhook"])
	if webhook == "" {
		return errors.New("钉钉 Webhook 不能为空")
	}

	if secret := strings.TrimSpace(cfg["secret"]); secret != "" {
		timestamp := strconv.FormatInt(time.Now().UnixMilli(), 10)
		sign, err := dingTalkSign(timestamp, secret)
		if err != nil {
			return err
		}
		separator := "&"
		if !strings.Contains(webhook, "?") {
			separator = "?"
		}
		webhook = webhook + separator + "timestamp=" + timestamp + "&sign=" + url.QueryEscape(sign)
	}

	payload := map[string]any{
		"msgtype": "markdown",
		"markdown": map[string]string{
			"title": title,
			"text":  joinTitleContent(title, content),
		},
	}

	return postNoticeJSON(ctx, webhook, payload, "钉钉")
}

// dingTalkSign 按钉钉加签规则计算签名：HmacSHA256(timestamp+"\n"+secret, secret) 再 Base64。
// 入参 timestamp 为毫秒时间戳、secret 为密钥；返回签名与错误。
func dingTalkSign(timestamp, secret string) (string, error) {
	mac := hmac.New(sha256.New, []byte(secret))
	if _, err := mac.Write([]byte(timestamp + "\n" + secret)); err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(mac.Sum(nil)), nil
}

// postNoticeJSON 向机器人 Webhook POST JSON 并校验返回码。
// 入参 url 为地址、payload 为请求体、scene 为场景名（用于错误提示）；返回发送错误。
func postNoticeJSON(ctx context.Context, url string, payload map[string]any, scene string) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("组装%s消息失败：%w", scene, err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("创建%s请求失败：%w", scene, err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("请求%s失败：%w", scene, err)
	}
	defer resp.Body.Close()

	var result struct {
		ErrCode int    `json:"errcode"`
		ErrMsg  string `json:"errmsg"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("解析%s响应失败（HTTP %d）：%w", scene, resp.StatusCode, err)
	}
	if result.ErrCode != 0 {
		return fmt.Errorf("%s返回错误：%d %s", scene, result.ErrCode, result.ErrMsg)
	}

	return nil
}

// joinTitleContent 把标题与正文拼成 markdown 文本；正文已含标题时不重复。
// 入参 title/content；返回拼接后的文本。
func joinTitleContent(title, content string) string {
	title = strings.TrimSpace(title)
	content = strings.TrimSpace(content)

	if title == "" {
		return content
	}
	if content == "" {
		return "### " + title
	}
	if strings.HasPrefix(content, title) {
		return content
	}

	return "### " + title + "\n" + content
}
