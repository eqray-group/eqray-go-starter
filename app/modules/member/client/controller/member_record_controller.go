package controller

import (
	"github.com/goravel/framework/contracts/http"

	"eqray-go-starter/app/modules/base/param"
	"eqray-go-starter/app/modules/base/result"
	"eqray-go-starter/app/modules/base/support"
	"eqray-go-starter/app/modules/member/service"
)

// MemberRecordController 提供 C 端「我的」只读查询：积分流水、签到记录、等级记录与等级概览。
// 记录查询统一按令牌里的会员 ID 过滤，不接受前端传 userId。
type MemberRecordController struct {
	members *service.MemberService
	levels  *service.MemberLevelService
}

func NewMemberRecordController() *MemberRecordController {
	return &MemberRecordController{
		members: service.NewMemberService(),
		levels:  service.NewMemberLevelService(),
	}
}

// PointRecords GET /api/v1/client/member/points/records
func (c *MemberRecordController) PointRecords(ctx http.Context) http.Response {
	memberID := support.MemberID(ctx.Context())
	if _, err := c.members.EnsureActive(ctx.Context(), memberID); err != nil {
		return result.Fail(ctx, err)
	}

	page, err := param.FromQuery(ctx)
	if err != nil {
		return result.Fail(ctx, err)
	}

	list, total, err := c.members.PointRecords(ctx.Context(), memberID, page.PageNo, page.PageSize)
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Page(ctx, list, total)
}

// SignRecords GET /api/v1/client/member/sign/records
func (c *MemberRecordController) SignRecords(ctx http.Context) http.Response {
	memberID := support.MemberID(ctx.Context())
	if _, err := c.members.EnsureActive(ctx.Context(), memberID); err != nil {
		return result.Fail(ctx, err)
	}

	page, err := param.FromQuery(ctx)
	if err != nil {
		return result.Fail(ctx, err)
	}

	list, total, err := c.members.SignRecords(ctx.Context(), memberID, page.PageNo, page.PageSize)
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Page(ctx, list, total)
}

// LevelRecords GET /api/v1/client/member/level/records
func (c *MemberRecordController) LevelRecords(ctx http.Context) http.Response {
	memberID := support.MemberID(ctx.Context())
	if _, err := c.members.EnsureActive(ctx.Context(), memberID); err != nil {
		return result.Fail(ctx, err)
	}

	page, err := param.FromQuery(ctx)
	if err != nil {
		return result.Fail(ctx, err)
	}

	list, total, err := c.members.LevelRecords(ctx.Context(), memberID, page.PageNo, page.PageSize)
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Page(ctx, list, total)
}

// Levels GET /api/v1/client/member/levels
// 返回启用中的等级列表，并带上当前会员的等级与经验，供前端画升级进度。
func (c *MemberRecordController) Levels(ctx http.Context) http.Response {
	memberID := support.MemberID(ctx.Context())
	m, err := c.members.EnsureActive(ctx.Context(), memberID)
	if err != nil {
		return result.Fail(ctx, err)
	}

	list, err := c.levels.Options(ctx.Context())
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, map[string]any{
		"list":           list,
		"currentLevelId": m.LevelID,
		"experience":     m.Experience,
	})
}
