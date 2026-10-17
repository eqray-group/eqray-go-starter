// Package result 提供 Controller 统一的响应输出助手。
package result

import (
	"errors"

	"github.com/goravel/framework/contracts/http"

	"eqray-go-starter/app/facades"
	"eqray-go-starter/app/modules/base/errs"
	"eqray-go-starter/app/modules/base/response"
)

// Success 输出成功响应，HTTP 状态码固定 200，data 无内容时传 nil（docs 8.1）。
func Success(ctx http.Context, data any) http.Response {
	return ctx.Response().Success().Json(response.Success(data))
}

// Page 输出分页成功响应，只含 list 与 total（docs 8.2）。
func Page(ctx http.Context, list any, total int64) http.Response {
	return ctx.Response().Success().Json(response.Page(list, total))
}

// Fail 统一输出失败响应：业务异常按自身错误码返回，非预期错误先记日志再按 500 返回。
// HTTP 状态码始终为 200，业务结果只由响应体的 code 表达（docs 8.1）。
func Fail(ctx http.Context, err error) http.Response {
	var be *errs.BusinessError
	if !errors.As(err, &be) {
		facades.Log().WithContext(ctx.Context()).Error(err.Error())
	}
	return ctx.Response().Success().Json(response.Fail(err))
}
