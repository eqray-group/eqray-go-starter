package request

import (
	"github.com/goravel/framework/contracts/http"
)

// GeneratorRequest 是代码生成的入参。
type GeneratorRequest struct {
	// Table 表名
	Table string `form:"table" json:"table"`
	// Module 目标模块
	Module string `form:"module" json:"module"`
	// Pascal 结构体名前缀，留空由表名推导
	Pascal string `form:"pascal" json:"pascal"`
	// Overwrite 是否覆盖已存在文件
	Overwrite bool `form:"overwrite" json:"overwrite"`
}

// Rules 返回校验规则。
func (r *GeneratorRequest) Rules(ctx http.Context) map[string]any {
	return map[string]any{
		"table":  "required|max:64",
		"module": "required|max:32",
		"pascal": "max:64",
	}
}

// Messages 覆盖默认提示。
func (r *GeneratorRequest) Messages(ctx http.Context) map[string]string {
	return map[string]string{
		"table.required":  "请选择表",
		"module.required": "请填写目标模块，如 demo",
	}
}
