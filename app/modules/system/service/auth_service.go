// Package service 存放 system 模块的业务逻辑，事务只在这一层开启。
package service

import (
	"context"
	"strings"

	"golang.org/x/crypto/bcrypt"

	baseAuth "eqray-go-starter/app/modules/base/auth"
	baseEnums "eqray-go-starter/app/modules/base/enums"
	baseErrs "eqray-go-starter/app/modules/base/errs"
	"eqray-go-starter/app/modules/base/types"
	systemErrs "eqray-go-starter/app/modules/system/errs"
	"eqray-go-starter/app/modules/system/repository"
)

// LoginResult 是登录 / 续期成功后返回给前端的一对令牌。
type LoginResult = baseAuth.TokenPair

// LoginInput 是登录入参，用结构体承接避免 Login 的参数列表过长。
type LoginInput struct {
	IP          string
	Username    string
	Password    string
	CaptchaID   string
	CaptchaCode string
}

// AuthService 负责后台登录认证（docs 9.1）。
type AuthService struct {
	users      *repository.UserRepository
	permission *PermissionService
}

// NewAuthService 构造 Service。
func NewAuthService() *AuthService {
	return &AuthService{
		users:      repository.NewUserRepository(),
		permission: NewPermissionService(),
	}
}

// Login 依次校验验证码、账号密码与账号状态，通过后签发令牌并刷新登录信息。
// 账号不存在与密码错误返回同一个错误码，避免被用来枚举账号。
func (s *AuthService) Login(ctx context.Context, in LoginInput) (*LoginResult, error) {
	if !baseAuth.VerifyCaptcha(in.CaptchaID, in.CaptchaCode) {
		return nil, baseErrs.New(systemErrs.SystemCaptchaWrong)
	}

	user, err := s.users.FindByUsername(ctx, strings.ToLower(strings.TrimSpace(in.Username)))
	if err != nil {
		return nil, err
	}
	if user == nil || bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(in.Password)) != nil {
		return nil, baseErrs.New(systemErrs.SystemLoginFailed)
	}
	if user.Status != baseEnums.CommonStatusEnabled {
		return nil, baseErrs.New(systemErrs.SystemAccountDisabled)
	}

	tokens, err := baseAuth.Issue(user.ID)
	if err != nil {
		return nil, err
	}

	// 登录信息只用于展示，回写失败不阻断登录。
	_ = s.users.UpdateLoginInfo(ctx, user.ID, in.IP, types.Now())

	return tokens, nil
}

// Refresh 用刷新令牌换一对新令牌，旧刷新令牌随之作废（docs 9.2）。
func (s *AuthService) Refresh(refreshToken string) (*LoginResult, error) {
	return baseAuth.Refresh(refreshToken)
}

// Logout 删除服务端保存的访问令牌与刷新令牌，使两者立即失效。
func (s *AuthService) Logout(accessToken, refreshToken string) {
	baseAuth.Revoke(accessToken, refreshToken)
}

// Profile 返回当前登录者的基础信息、角色标识与权限标识（docs 9.3）。
func (s *AuthService) Profile(ctx context.Context, id types.ID) (map[string]any, error) {
	user, err := s.users.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, baseErrs.New(systemErrs.SystemUserNotExists, id)
	}

	roles, err := s.permission.RoleCodes(ctx, user.ID)
	if err != nil {
		return nil, err
	}
	buttons, err := s.permission.Permissions(ctx, user.ID)
	if err != nil {
		return nil, err
	}

	return map[string]any{
		"userId":   user.ID,
		"userName": user.Username,
		"nickname": user.Nickname,
		"avatar":   user.Avatar,
		"email":    user.Email,
		"roles":    roles,
		"buttons":  buttons,
	}, nil
}
