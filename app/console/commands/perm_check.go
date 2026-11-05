// Package commands 提供运维 / 开发辅助命令。
package commands

import (
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/console/command"

	"eqray-go-starter/app/facades"
)

// permPattern 匹配路由文件里 perm("模块:资源:动作") 的声明。
var permPattern = regexp.MustCompile(`perm\("([^"]+)"\)`)

// PermCheck 校验「代码里声明的权限点」与「system_menu 表里的权限点」是否一致。
// 用途：新增接口忘了加菜单权限点时能提前发现（否则线上直接 403）。
type PermCheck struct{}

// Signature 返回命令名。
func (r *PermCheck) Signature() string {
	return "perm:check"
}

// Description 返回命令说明。
func (r *PermCheck) Description() string {
	return "校验代码中的权限点与 system_menu 是否一致（缺菜单 / 多余权限点）"
}

// Extend 声明命令分类。
func (r *PermCheck) Extend() command.Extend {
	return command.Extend{Category: "app"}
}

// Handle 扫描 app/modules 下的 route.go，与库里的权限点做双向比对并打印结果。
// 入参 ctx 为控制台上下文；返回错误（仅扫描或查询失败时）。
func (r *PermCheck) Handle(ctx console.Context) error {
	if !facades.Schema().HasTable("system_menu") {
		ctx.Warning("system_menu 表不存在，请先执行 go run . artisan migrate")

		return nil
	}

	codeCodes, err := collectRoutePermissions("app/modules")
	if err != nil {
		return err
	}

	rows := make([]struct {
		Permission string `db:"permission"`
	}, 0)
	if err := facades.DB().Table("system_menu").
		Select("permission").
		Where("permission", "<>", "").
		WhereNull("deleted_at").
		Get(&rows); err != nil {
		return err
	}

	menuCodes := make(map[string]struct{}, len(rows))
	for _, row := range rows {
		menuCodes[strings.TrimSpace(row.Permission)] = struct{}{}
	}

	missing := make([]string, 0)
	for code := range codeCodes {
		if _, ok := menuCodes[code]; !ok {
			missing = append(missing, code)
		}
	}

	orphan := make([]string, 0)
	for code := range menuCodes {
		if _, ok := codeCodes[code]; !ok {
			orphan = append(orphan, code)
		}
	}

	sort.Strings(missing)
	sort.Strings(orphan)

	ctx.Info("代码权限点：" + itoa(len(codeCodes)) + " 个，菜单权限点：" + itoa(len(menuCodes)) + " 个")

	if len(missing) == 0 {
		ctx.Info("✓ 代码中用到的权限点都已在 system_menu 注册")
	} else {
		ctx.Warning("✗ 以下权限点代码在用、菜单未注册（线上会 403）：")
		for _, code := range missing {
			ctx.Warning("  - " + code)
		}
	}

	if len(orphan) == 0 {
		ctx.Info("✓ 没有多余的菜单权限点")
	} else {
		ctx.Comment("以下权限点在菜单里存在、代码中未使用（可能是历史遗留）：")
		for _, code := range orphan {
			ctx.Comment("  - " + code)
		}
	}

	return nil
}

// collectRoutePermissions 递归扫描目录下所有 Go 文件，收集 perm("...") 里的权限点。
// 入参 root 为起始目录；返回权限点集合与扫描错误。
func collectRoutePermissions(root string) (map[string]struct{}, error) {
	codes := make(map[string]struct{})

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() || !strings.HasSuffix(path, ".go") {
			return nil
		}

		data, readErr := os.ReadFile(path)
		if readErr != nil {
			return readErr
		}

		for _, match := range permPattern.FindAllStringSubmatch(string(data), -1) {
			if len(match) == 2 {
				codes[strings.TrimSpace(match[1])] = struct{}{}
			}
		}

		return nil
	})

	return codes, err
}

// itoa 把整数转成字符串，避免为了一个格式化再引入 strconv 别名。
// 入参 n 为待转换整数；返回十进制字符串。
func itoa(n int) string {
	if n == 0 {
		return "0"
	}

	digits := make([]byte, 0, 4)
	for n > 0 {
		digits = append([]byte{byte('0' + n%10)}, digits...)
		n /= 10
	}

	return string(digits)
}
