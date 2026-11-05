package auth

import (
	"context"
	"fmt"
	"strings"
	"time"

	"eqray-go-starter/app/facades"
)

// loginFailPrefix 是登录失败计数的缓存前缀；scene 区分后台 / C 端，避免互相影响。
const loginFailPrefix = "security:login:fail:"

// loginFailKey 生成失败计数键，账号统一小写去空格，避免大小写绕过锁定。
// 入参 scene 为登录场景、account 为账号；返回缓存键。
func loginFailKey(scene, account string) string {
	return loginFailPrefix + scene + ":" + strings.ToLower(strings.TrimSpace(account))
}

// LoginFailCount 返回当前累计失败次数（窗口内）。
// 入参 scene 为登录场景、account 为账号；返回失败次数。
func LoginFailCount(ctx context.Context, scene, account string) int64 {
	value := facades.Cache().Get(loginFailKey(scene, account))
	switch count := value.(type) {
	case int64:
		return count
	case int:
		return int64(count)
	}

	return 0
}

// LoginLocked 判断账号是否已被锁定，并返回剩余锁定时间。
// 入参 scene 为登录场景、account 为账号；返回是否锁定与剩余时长。
func LoginLocked(ctx context.Context, scene, account string) (bool, time.Duration) {
	maxFail := facades.Config().GetInt("security.login.max_fail", 5)
	lockMinutes := facades.Config().GetInt("security.login.lock_minutes", 15)
	if maxFail <= 0 || LoginFailCount(ctx, scene, account) < int64(maxFail) {
		return false, 0
	}

	return true, time.Duration(lockMinutes) * time.Minute
}

// RecordLoginFailure 累加失败次数，首次失败时设置过期窗口。
// 入参 scene 为登录场景、account 为账号；返回累计失败次数。
func RecordLoginFailure(ctx context.Context, scene, account string) int64 {
	lockMinutes := facades.Config().GetInt("security.login.lock_minutes", 15)
	key := loginFailKey(scene, account)

	count, err := facades.Cache().Increment(key, 1)
	if err != nil {
		return 0
	}
	if count == 1 {
		_ = facades.Cache().Put(key, count, time.Duration(lockMinutes)*time.Minute)
	}

	return count
}

// ClearLoginFailure 登录成功后清空失败计数。
// 入参 scene 为登录场景、account 为账号；无返回值。
func ClearLoginFailure(ctx context.Context, scene, account string) {
	facades.Cache().Forget(loginFailKey(scene, account))
}

// LoginLockedMessage 生成带剩余分钟的锁定提示，便于前端直接展示。
// 入参 d 为剩余锁定时间；返回提示文案。
func LoginLockedMessage(d time.Duration) string {
	minutes := int(d.Minutes())
	if minutes <= 0 {
		minutes = 1
	}

	return fmt.Sprintf("登录失败次数过多，请 %d 分钟后再试", minutes)
}
