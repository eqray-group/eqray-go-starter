package service

import (
	"go/parser"
	"go/token"
	"strings"
	"testing"
)

// TestBuildFilesCompiles 校验模板拼出的 Go 文件语法合法，且没有残留占位符。
// 生成器最容易出的问题是模板拼错导致产出无法编译，这里用标准库 parser 兜住。
func TestBuildFilesCompiles(t *testing.T) {
	in := GeneratorInput{Table: "demo_order", Module: "demo", Pascal: "DemoOrder"}
	columns := []ColumnInfo{
		{Name: "id", Type: "bigint", PrimaryKey: true, GoType: "types.ID", FieldName: "ID", JsonName: "id", FormType: "number"},
		{Name: "order_no", Type: "varchar", GoType: "string", FieldName: "OrderNo", JsonName: "orderNo", FormType: "input", Comment: "订单号"},
		{Name: "price", Type: "bigint", GoType: "int64", FieldName: "Price", JsonName: "price", FormType: "number", Comment: "金额（分）"},
		{Name: "status", Type: "tinyint", GoType: "int8", FieldName: "Status", JsonName: "status", FormType: "switch", Comment: "状态"},
		{Name: "remark", Type: "text", GoType: "string", FieldName: "Remark", JsonName: "remark", FormType: "textarea"},
		{Name: "created_at", Type: "datetime", GoType: "types.DateTime", FieldName: "CreatedAt", JsonName: "createdAt", FormType: "datetime"},
	}

	files := (&GeneratorService{}).buildFiles(in, columns)
	if len(files) == 0 {
		t.Fatal("没有生成任何文件")
	}

	fileSet := token.NewFileSet()
	for _, file := range files {
		if strings.Contains(file.Content, "@@") {
			t.Fatalf("%s 残留模板占位符：\n%s", file.Path, file.Content)
		}
		if !file.Writable || !strings.HasSuffix(file.Path, ".go") {
			continue
		}
		if _, err := parser.ParseFile(fileSet, file.Path, file.Content, parser.AllErrors); err != nil {
			t.Fatalf("%s 语法不合法：%v\n%s", file.Path, err, file.Content)
		}
	}

	// 前端必须拆成「列表页 + 表单弹窗」，列表页里不再内嵌 ElDialog
	byPath := make(map[string]string, len(files))
	for _, file := range files {
		byPath[file.Path] = file.Content
	}

	indexPath := "ui/src/views/demo/demo_order/index.vue"
	dialogPath := "ui/src/views/demo/demo_order/modules/demo_order-form-dialog.vue"
	index, ok := byPath[indexPath]
	if !ok {
		t.Fatalf("缺少列表页：%s", indexPath)
	}
	dialog, ok := byPath[dialogPath]
	if !ok {
		t.Fatalf("缺少表单弹窗：%s", dialogPath)
	}
	if strings.Contains(index, "<ElDialog") {
		t.Fatal("列表页不应内嵌表单弹窗")
	}
	if !strings.Contains(index, "DemoOrderFormDialog") {
		t.Fatal("列表页未引用表单弹窗组件")
	}
	if !strings.Contains(dialog, "<ElDialog") || !strings.Contains(dialog, "fetchUpdateDemoOrder") {
		t.Fatal("表单弹窗内容不完整")
	}
	if _, ok := byPath["ui/src/types/api/demo/demo_order.d.ts"]; !ok {
		t.Fatal("缺少前端类型声明文件")
	}
}

// TestBuildFilesWithoutAuditColumns 校验没有审计字段的表不会生成引用 Creator 等不存在的字段。
func TestBuildFilesWithoutAuditColumns(t *testing.T) {
	in := GeneratorInput{Table: "demo_tag", Module: "demo", Pascal: "DemoTag"}
	columns := []ColumnInfo{
		{Name: "id", Type: "bigint", PrimaryKey: true, GoType: "types.ID", FieldName: "ID", JsonName: "id", FormType: "number"},
		{Name: "name", Type: "varchar", GoType: "string", FieldName: "Name", JsonName: "name", FormType: "input"},
	}

	files := (&GeneratorService{}).buildFiles(in, columns)
	for _, file := range files {
		if strings.Contains(file.Content, "@@") {
			t.Fatalf("%s 残留模板占位符", file.Path)
		}
		if strings.Contains(file.Path, "/service/") {
			if strings.Contains(file.Content, "Creator:") || strings.Contains(file.Content, "UpdatedAt:") {
				t.Fatalf("无审计字段的表不应生成审计字段赋值：\n%s", file.Content)
			}
			if strings.Contains(file.Content, "now := types.Now()") {
				t.Fatalf("无时间字段的表不应声明 now：\n%s", file.Content)
			}
		}
	}
}

// TestNamingHelpers 校验命名转换。
func TestNamingHelpers(t *testing.T) {
	if got := pascalCase("demo_order"); got != "DemoOrder" {
		t.Fatalf("pascalCase 结果不符：%s", got)
	}
	if got := camelCase("order_no"); got != "orderNo" {
		t.Fatalf("camelCase 结果不符：%s", got)
	}
	if got := pascalToSnake("DemoOrder"); got != "demo_order" {
		t.Fatalf("pascalToSnake 结果不符：%s", got)
	}
	if got := goTypeOf("datetime", false); got != "types.DateTime" {
		t.Fatalf("goTypeOf 结果不符：%s", got)
	}
}
