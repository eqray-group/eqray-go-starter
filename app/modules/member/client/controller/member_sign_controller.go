package controller

import (
	"github.com/goravel/framework/contracts/http"

	"eqray-go-starter/app/modules/base/result"
	"eqray-go-starter/app/modules/base/support"
	"eqray-go-starter/app/modules/member/service"
)

// MemberSignController 提供 C 端每日签到与签到状态查询。
type MemberSignController struct {
	service *service.MemberSignService
}

func NewMemberSignController() *MemberSignController {
	return &MemberSignController{service: service.NewMemberSignService()}
}

// Sign POST /api/v1/client/member/sign
// 同一会员同一天只能签到一次，重复签到返回业务错误码。
func (c *MemberSignController) Sign(ctx http.Context) http.Response {
	res, err := c.service.Sign(ctx.Context(), support.MemberID(ctx.Context()))
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, res)
}

// Status GET /api/v1/client/member/sign/status
func (c *MemberSignController) Status(ctx http.Context) http.Response {
	res, err := c.service.Status(ctx.Context(), support.MemberID(ctx.Context()))
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, res)
}
