package config

import (
	"eqray-go-starter/app/facades"
)

func init() {
	config := facades.Config()
	config.Add("sms", map[string]any{
		// 登录场景使用的短信模板代码，模板需在后台「会员管理 → 短信模板」配置并绑定渠道
		"login_template_code": config.GetString("SMS_LOGIN_TEMPLATE_CODE", "member_sms_login"),
	})
}
