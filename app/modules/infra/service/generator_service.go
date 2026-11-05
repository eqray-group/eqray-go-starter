package service

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"unicode"

	"eqray-go-starter/app/facades"
	baseErrs "eqray-go-starter/app/modules/base/errs"
)

// generatorNamePattern 限制表名 / 模块名，避免路径穿越与非法标识符。
var generatorNamePattern = regexp.MustCompile(`^[a-z][a-z0-9_]*$`)

// generatorRoot 是允许写入的仓库根目录（相对 artisan 启动目录）。
const generatorRoot = "."

// TableSummary 是库中的一张表。
type TableSummary struct {
	// Name 表名
	Name string `json:"name"`
	// Comment 表注释
	Comment string `json:"comment"`
	// Rows 行数（information_schema 估算值）
	Rows int64 `json:"rows"`
}

// ColumnInfo 是表字段的元信息。
type ColumnInfo struct {
	// Name 列名
	Name string `json:"name"`
	// Type 数据库类型（不含长度）
	Type string `json:"type"`
	// ColumnType 完整类型（含长度）
	ColumnType string `json:"columnType"`
	// Nullable 是否可空
	Nullable bool `json:"nullable"`
	// PrimaryKey 是否主键
	PrimaryKey bool `json:"primaryKey"`
	// AutoIncrement 是否自增
	AutoIncrement bool `json:"autoIncrement"`
	// Comment 列注释
	Comment string `json:"comment"`
	// GoType 映射后的 Go 类型
	GoType string `json:"goType"`
	// FieldName 结构体字段名
	FieldName string `json:"fieldName"`
	// JsonName JSON 字段名
	JsonName string `json:"jsonName"`
	// FormType 前端表单控件类型：input / number / textarea / switch / datetime
	FormType string `json:"formType"`
}

// GeneratorInput 是预览 / 生成的入参。
type GeneratorInput struct {
	// Table 表名
	Table string
	// Module 目标模块（app/modules/{module}、ui/src/views/{module}）
	Module string
	// Pascal 结构体名前缀，留空由表名推导
	Pascal string
	// Overwrite 是否覆盖已存在的文件
	Overwrite bool
}

// GeneratedFile 是一个生成结果文件。
type GeneratedFile struct {
	// Path 相对仓库根的路径
	Path string `json:"path"`
	// Language 语言，供前端高亮（go / ts / vue / text）
	Language string `json:"language"`
	// Writable 是否允许落盘（路由片段等只读预览项为 false）
	Writable bool `json:"writable"`
	// Content 文件内容
	Content string `json:"content"`
}

// GenerateResult 是落盘结果。
type GenerateResult struct {
	// Written 已写入的路径
	Written []string `json:"written"`
	// Skipped 因已存在且未开启覆盖而跳过的路径
	Skipped []string `json:"skipped"`
}

// GeneratorService 读取 information_schema，按项目分层生成 model / repository / service /
// controller / request 与前端 api / 页面，模板与现有模块保持同一风格。
type GeneratorService struct{}

// NewGeneratorService 构造代码生成服务。
func NewGeneratorService() *GeneratorService {
	return &GeneratorService{}
}

// ListTables 列出当前数据库的业务表（排除迁移与日志类系统表）。
// 入参 ctx 为请求上下文；返回表摘要列表。
func (s *GeneratorService) ListTables(ctx context.Context) ([]TableSummary, error) {
	rows := make([]struct {
		Name    string `db:"name"`
		Comment string `db:"comment"`
		Rows    int64  `db:"rows_count"`
	}, 0)

	// 直接读 information_schema：跨库名用 DATABASE() 限定，避免生成到其他库的表
	err := facades.DB().WithContext(ctx).Select(&rows, `
SELECT TABLE_NAME AS name, TABLE_COMMENT AS comment, TABLE_ROWS AS rows_count
FROM information_schema.tables
WHERE TABLE_SCHEMA = DATABASE() AND TABLE_TYPE = 'BASE TABLE'
ORDER BY TABLE_NAME`)
	if err != nil {
		return nil, baseErrs.WithMsg(baseErrs.BadRequest, "读取表列表失败："+err.Error())
	}

	list := make([]TableSummary, 0, len(rows))
	for _, row := range rows {
		if strings.HasSuffix(row.Name, "_log") {
			continue
		}
		list = append(list, TableSummary{Name: row.Name, Comment: row.Comment, Rows: row.Rows})
	}

	return list, nil
}

// DescribeTable 读取表字段并映射成 Go / 前端类型。
// 入参 table 为表名；返回字段列表。
func (s *GeneratorService) DescribeTable(ctx context.Context, table string) ([]ColumnInfo, error) {
	if !generatorNamePattern.MatchString(table) {
		return nil, baseErrs.WithMsg(baseErrs.BadRequest, "表名不合法")
	}

	rows := make([]struct {
		Name       string `db:"name"`
		Type       string `db:"data_type"`
		ColumnType string `db:"column_type"`
		Nullable   string `db:"is_nullable"`
		ColumnKey  string `db:"column_key"`
		Extra      string `db:"extra"`
		Comment    string `db:"comment"`
	}, 0)

	err := facades.DB().WithContext(ctx).Select(&rows, `
SELECT COLUMN_NAME AS name, DATA_TYPE AS data_type, COLUMN_TYPE AS column_type,
       IS_NULLABLE AS is_nullable, COLUMN_KEY AS column_key, EXTRA AS extra,
       COLUMN_COMMENT AS comment
FROM information_schema.columns
WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = ?
ORDER BY ORDINAL_POSITION`, table)
	if err != nil {
		return nil, baseErrs.WithMsg(baseErrs.BadRequest, "读取表结构失败："+err.Error())
	}
	if len(rows) == 0 {
		return nil, baseErrs.WithMsg(baseErrs.BadRequest, "表不存在："+table)
	}

	columns := make([]ColumnInfo, 0, len(rows))
	for _, row := range rows {
		primary := row.ColumnKey == "PRI"
		columns = append(columns, ColumnInfo{
			Name:          row.Name,
			Type:          row.Type,
			ColumnType:    row.ColumnType,
			Nullable:      strings.EqualFold(row.Nullable, "YES"),
			PrimaryKey:    primary,
			AutoIncrement: strings.Contains(strings.ToLower(row.Extra), "auto_increment"),
			Comment:       row.Comment,
			GoType:        goTypeOf(row.Type, primary),
			FieldName:     pascalCase(row.Name),
			JsonName:      camelCase(row.Name),
			FormType:      formTypeOf(row.Type, row.Name),
		})
	}

	return columns, nil
}

// Preview 生成文件内容但不落盘。
// 入参 in 为生成参数；返回文件列表。
func (s *GeneratorService) Preview(ctx context.Context, in GeneratorInput) ([]GeneratedFile, error) {
	normalized, columns, err := s.prepare(ctx, in)
	if err != nil {
		return nil, err
	}

	return s.buildFiles(normalized, columns), nil
}

// Generate 生成文件并写入磁盘；已存在的文件在未开启覆盖时跳过。
// 入参 in 为生成参数；返回写入 / 跳过清单。
func (s *GeneratorService) Generate(ctx context.Context, in GeneratorInput) (*GenerateResult, error) {
	normalized, columns, err := s.prepare(ctx, in)
	if err != nil {
		return nil, err
	}

	result := &GenerateResult{Written: make([]string, 0), Skipped: make([]string, 0)}
	for _, file := range s.buildFiles(normalized, columns) {
		if !file.Writable {
			continue
		}

		target := filepath.Clean(filepath.Join(generatorRoot, filepath.FromSlash(file.Path)))
		if !strings.HasPrefix(target, generatorRoot) {
			return nil, baseErrs.WithMsg(baseErrs.BadRequest, "非法输出路径："+file.Path)
		}

		if _, statErr := os.Stat(target); statErr == nil && !normalized.Overwrite {
			result.Skipped = append(result.Skipped, file.Path)
			continue
		}

		if mkErr := os.MkdirAll(filepath.Dir(target), 0o755); mkErr != nil {
			return nil, baseErrs.WithMsg(baseErrs.BadRequest, "创建目录失败："+mkErr.Error())
		}
		if writeErr := os.WriteFile(target, []byte(file.Content), 0o644); writeErr != nil {
			return nil, baseErrs.WithMsg(baseErrs.BadRequest, "写入文件失败："+writeErr.Error())
		}

		result.Written = append(result.Written, file.Path)
	}

	return result, nil
}

// prepare 校验入参并读取字段，同时推导命名。
// 入参 in 为原始入参；返回规范化参数与字段列表。
func (s *GeneratorService) prepare(ctx context.Context, in GeneratorInput) (GeneratorInput, []ColumnInfo, error) {
	table := strings.TrimSpace(in.Table)
	module := strings.TrimSpace(in.Module)

	if !generatorNamePattern.MatchString(table) {
		return in, nil, baseErrs.WithMsg(baseErrs.BadRequest, "表名不合法（只允许小写字母、数字、下划线）")
	}
	if !generatorNamePattern.MatchString(module) {
		return in, nil, baseErrs.WithMsg(baseErrs.BadRequest, "模块名不合法（只允许小写字母、数字、下划线）")
	}

	columns, err := s.DescribeTable(ctx, table)
	if err != nil {
		return in, nil, err
	}

	pascal := strings.TrimSpace(in.Pascal)
	if pascal == "" {
		pascal = pascalCase(strings.TrimSuffix(table, "s"))
	}
	if pascal == "" {
		return in, nil, baseErrs.WithMsg(baseErrs.BadRequest, "结构体名不能为空")
	}

	return GeneratorInput{
		Table:     table,
		Module:    module,
		Pascal:    pascal,
		Overwrite: in.Overwrite,
	}, columns, nil
}

// buildFiles 按项目分层拼装全部生成文件。
// 入参 in 为规范化参数、columns 为字段；返回文件列表。
func (s *GeneratorService) buildFiles(in GeneratorInput, columns []ColumnInfo) []GeneratedFile {
	name := pascalToSnake(in.Pascal)

	return []GeneratedFile{
		{
			Path:     "app/modules/" + in.Module + "/model/" + name + "_model.go",
			Language: "go",
			Writable: true,
			Content:  renderModel(in, columns),
		},
		{
			Path:     "app/modules/" + in.Module + "/repository/" + name + "_repository.go",
			Language: "go",
			Writable: true,
			Content:  renderRepository(in, columns),
		},
		{
			Path:     "app/modules/" + in.Module + "/service/" + name + "_service.go",
			Language: "go",
			Writable: true,
			Content:  renderService(in, columns),
		},
		{
			Path:     "app/modules/" + in.Module + "/admin/request/" + name + "_request.go",
			Language: "go",
			Writable: true,
			Content:  renderRequest(in, columns),
		},
		{
			Path:     "app/modules/" + in.Module + "/admin/controller/" + name + "_controller.go",
			Language: "go",
			Writable: true,
			Content:  renderController(in, columns),
		},
		{
			Path:     "app/modules/" + in.Module + "/admin/route.go",
			Language: "text",
			Writable: false,
			Content:  renderRouteSnippet(in),
		},
		{
			Path:     "ui/src/api/" + in.Module + "/" + name + ".ts",
			Language: "ts",
			Writable: true,
			Content:  renderFrontendAPI(in),
		},
		{
			Path:     "ui/src/views/" + in.Module + "/" + name + "/index.vue",
			Language: "vue",
			Writable: true,
			Content:  renderFrontendPage(in, columns),
		},
	}
}

// goTypeOf 把 MySQL 类型映射成 Go 类型，主键统一用 types.ID。
// 入参 dbType 为字段类型、primary 是否主键；返回 Go 类型。
func goTypeOf(dbType string, primary bool) string {
	if primary {
		return "types.ID"
	}

	switch strings.ToLower(dbType) {
	case "bigint", "int", "smallint", "mediumint":
		return "int64"
	case "tinyint":
		return "int8"
	case "decimal", "numeric", "float", "double":
		return "float64"
	case "datetime", "timestamp", "date", "time":
		return "types.DateTime"
	case "json":
		return "string"
	default:
		return "string"
	}
}

// formTypeOf 推断前端表单控件类型。
// 入参 dbType 为字段类型、name 为列名；返回控件类型。
func formTypeOf(dbType, name string) string {
	switch strings.ToLower(dbType) {
	case "tinyint":
		if strings.Contains(name, "status") || strings.Contains(name, "enabled") {
			return "switch"
		}

		return "number"
	case "bigint", "int", "smallint", "mediumint", "decimal", "float", "double":
		return "number"
	case "text", "mediumtext", "longtext", "json":
		return "textarea"
	case "datetime", "timestamp", "date":
		return "datetime"
	}

	return "input"
}

// pascalCase 把下划线命名转成大驼峰。
// 入参 value 为原始名；返回大驼峰。
func pascalCase(value string) string {
	parts := strings.FieldsFunc(value, func(r rune) bool { return r == '_' || r == '-' })
	var builder strings.Builder
	for _, part := range parts {
		runes := []rune(part)
		if len(runes) == 0 {
			continue
		}
		runes[0] = unicode.ToUpper(runes[0])
		builder.WriteString(string(runes))
	}

	return builder.String()
}

// camelCase 把下划线命名转成小驼峰。
// 入参 value 为原始名；返回小驼峰。
func camelCase(value string) string {
	pascal := pascalCase(value)
	if pascal == "" {
		return ""
	}

	runes := []rune(pascal)
	runes[0] = unicode.ToLower(runes[0])

	return string(runes)
}

// pascalToSnake 把大驼峰转成下划线命名，用于文件名。
// 入参 value 为大驼峰；返回下划线命名。
func pascalToSnake(value string) string {
	var builder strings.Builder
	for index, r := range value {
		if unicode.IsUpper(r) {
			if index > 0 {
				builder.WriteRune('_')
			}
			builder.WriteRune(unicode.ToLower(r))

			continue
		}
		builder.WriteRune(r)
	}

	return builder.String()
}

// goDoc 生成字段注释，优先用列注释。
// 入参 col 为字段；返回注释文本（不含 //）。
func goDoc(col ColumnInfo) string {
	if strings.TrimSpace(col.Comment) != "" {
		return col.Comment
	}

	return col.FieldName
}

// fmtTag 给结构体字段拼上 db 与 json 标签。
// 入参 col 为字段；返回标签字面量（含反引号）。
func fmtTag(col ColumnInfo) string {
	return "`db:\"" + col.Name + "\" json:\"" + col.JsonName + "\"`"
}

// searchColumns 取用于关键字模糊匹配的文本列（最多两个）。
// 入参 columns 为全部字段；返回列名列表。
func searchColumns(columns []ColumnInfo) []string {
	result := make([]string, 0, 2)
	for _, col := range columns {
		if col.PrimaryKey {
			continue
		}
		switch strings.ToLower(col.Type) {
		case "varchar", "char", "text":
			result = append(result, col.Name)
			if len(result) == 2 {
				return result
			}
		}
	}

	return result
}

// editableColumns 取表单可写字段（排除主键与审计字段）。
// 入参 columns 为全部字段；返回可写字段。
func editableColumns(columns []ColumnInfo) []ColumnInfo {
	result := make([]ColumnInfo, 0, len(columns))
	for _, col := range columns {
		if col.PrimaryKey || isAuditColumn(col.Name) {
			continue
		}
		result = append(result, col)
	}

	return result
}

// hasColumn 判断表里是否存在指定列。
// 入参 columns 为全部字段、name 为列名；返回是否存在。
func hasColumn(columns []ColumnInfo, name string) bool {
	for _, col := range columns {
		if col.Name == name {
			return true
		}
	}

	return false
}

// isAuditColumn 判断是否框架统一的审计字段。
// 入参 name 为列名；返回是否审计字段。
func isAuditColumn(name string) bool {
	switch name {
	case "creator", "created_at", "updater", "updated_at", "deleted_at":
		return true
	}

	return false
}

// renderModel 生成 model 文件。
func renderModel(in GeneratorInput, columns []ColumnInfo) string {
	var body strings.Builder
	for _, col := range columns {
		fmt.Fprintf(&body, "\t// %s\n\t%s %s %s\n", goDoc(col), col.FieldName, col.GoType, fmtTag(col))
		if col.Name == "created_at" || col.Name == "updated_at" {
			continue
		}
	}

	tpl := `package model

import (
	"eqray-go-starter/app/modules/base/types"
)

// @@PASCAL@@Table 是表名常量。
const @@PASCAL@@Table = "@@TABLE@@"

// @@PASCAL@@Model 对应 @@TABLE@@ 表。
type @@PASCAL@@Model struct {
@@BODY@@}

// TableName 返回表名。
func (m *@@PASCAL@@Model) TableName() string {
	return @@PASCAL@@Table
}
`
	return replaceTokens(tpl, map[string]string{
		"@@PASCAL@@": in.Pascal,
		"@@TABLE@@":  in.Table,
		"@@BODY@@":   body.String(),
	})
}

// renderRepository 生成 repository 文件。
func renderRepository(in GeneratorInput, columns []ColumnInfo) string {
	search := searchColumns(columns)
	keywordExpr := ""
	if len(search) > 0 {
		keywordExpr = "\t\tquery = query.Where(\"CONCAT(" + strings.Join(search, ", ' ', ") +
			") LIKE ?\", \"%\"+keyword+\"%\")"
	}

	insertLines := make([]string, 0, len(columns))
	updateLines := make([]string, 0, len(columns))
	for _, col := range columns {
		if col.PrimaryKey {
			insertLines = append(insertLines, fmt.Sprintf("\t\t\"%s\": m.ID,", col.Name))
			continue
		}
		insertLines = append(insertLines, fmt.Sprintf("\t\t\"%s\": m.%s,", col.Name, col.FieldName))
		if !isAuditColumn(col.Name) {
			updateLines = append(updateLines, fmt.Sprintf("\t\t\"%s\": in.%s,", col.Name, col.FieldName))
		}
	}
	// 审计字段按表结构生成，避免对没有 creator / updated_at 的表生成无法编译的代码
	if hasColumn(columns, "updater") {
		updateLines = append(updateLines, "\t\t\"updater\":    operator,")
	}
	if hasColumn(columns, "updated_at") {
		updateLines = append(updateLines, "\t\t\"updated_at\": types.Now(),")
	}

	tpl := `package repository

import (
	"context"
	"strings"

	"github.com/goravel/framework/contracts/database/db"

	"eqray-go-starter/app/facades"
	"eqray-go-starter/app/modules/base/types"
	"eqray-go-starter/app/modules/@@MODULE@@/model"
)

// @@PASCAL@@Filter 是列表筛选条件。
type @@PASCAL@@Filter struct {
	// Keyword 关键字模糊匹配
	Keyword string
}

// @@PASCAL@@Input 是可写字段入参（由 service 组装）。
type @@PASCAL@@Input struct {
@@INPUT_FIELDS@@}

// @@PASCAL@@Repository 是 @@TABLE@@ 表的仓储。
type @@PASCAL@@Repository struct{}

// New@@PASCAL@@Repository 构造仓储。
func New@@PASCAL@@Repository() *@@PASCAL@@Repository {
	return &@@PASCAL@@Repository{}
}

// applyFilter 套用筛选条件。
func (r *@@PASCAL@@Repository) applyFilter(query db.Query, filter @@PASCAL@@Filter) db.Query {
	if keyword := strings.TrimSpace(filter.Keyword); keyword != "" {
@@KEYWORD_EXPR@@
	}

	return query
}

// Paginate 分页查询。
func (r *@@PASCAL@@Repository) Paginate(ctx context.Context, filter @@PASCAL@@Filter, pageNo, pageSize int) ([]model.@@PASCAL@@Model, int64, error) {
	total, err := r.applyFilter(facades.DB().WithContext(ctx).Table(model.@@PASCAL@@Table), filter).Count()
	if err != nil {
		return nil, 0, err
	}

	list := make([]model.@@PASCAL@@Model, 0)
	if total == 0 {
		return list, 0, nil
	}

	if err := r.applyFilter(facades.DB().WithContext(ctx).Table(model.@@PASCAL@@Table), filter).
		OrderBy("@@ORDER_BY@@").
		Offset(uint64((pageNo - 1) * pageSize)).
		Limit(uint64(pageSize)).
		Get(&list); err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

// FindByID 按主键查询，未命中返回 (nil, nil)。
func (r *@@PASCAL@@Repository) FindByID(ctx context.Context, id types.ID) (*model.@@PASCAL@@Model, error) {
	m := &model.@@PASCAL@@Model{}
	if err := facades.DB().WithContext(ctx).Table(model.@@PASCAL@@Table).Where("id", id).First(m); err != nil {
		return nil, nil
	}

	return m, nil
}

// Insert 新增一条记录。
func (r *@@PASCAL@@Repository) Insert(ctx context.Context, m *model.@@PASCAL@@Model) error {
	_, err := facades.DB().WithContext(ctx).Table(model.@@PASCAL@@Table).Insert(map[string]any{
@@INSERT_FIELDS@@	})

	return err
}

// Update 按主键更新可写字段。
func (r *@@PASCAL@@Repository) Update(ctx context.Context, id types.ID, in @@PASCAL@@Input, operator string) error {
	_, err := facades.DB().WithContext(ctx).Table(model.@@PASCAL@@Table).Where("id", id).Update(map[string]any{
@@UPDATE_FIELDS@@	})

	return err
}

// Delete 按主键删除。
func (r *@@PASCAL@@Repository) Delete(ctx context.Context, id types.ID) error {
	_, err := facades.DB().WithContext(ctx).Table(model.@@PASCAL@@Table).Where("id", id).Delete()

	return err
}
`

	return replaceTokens(tpl, map[string]string{
		"@@MODULE@@":        in.Module,
		"@@PASCAL@@":        in.Pascal,
		"@@TABLE@@":         in.Table,
		"@@KEYWORD_EXPR@@":  keywordExpr,
		"@@INPUT_FIELDS@@":  renderInputFields(columns),
		"@@INSERT_FIELDS@@": strings.Join(insertLines, "\n") + "\n",
		"@@UPDATE_FIELDS@@": strings.Join(updateLines, "\n") + "\n",
		"@@ORDER_BY@@":      orderByColumn(columns),
	})
}

// renderInputFields 生成 repository 入参结构体字段。
// 入参 columns 为全部字段；返回字段行。
func renderInputFields(columns []ColumnInfo) string {
	var body strings.Builder
	for _, col := range editableColumns(columns) {
		fmt.Fprintf(&body, "\t// %s\n\t%s %s\n", goDoc(col), col.FieldName, col.GoType)
	}

	return body.String()
}

// orderByColumn 选择排序列：优先自增主键，其次 created_at。
// 入参 columns 为全部字段；返回列名。
func orderByColumn(columns []ColumnInfo) string {
	for _, col := range columns {
		if col.PrimaryKey {
			return col.Name + " desc"
		}
	}

	return "id desc"
}

// renderService 生成 service 文件。
func renderService(in GeneratorInput, columns []ColumnInfo) string {
	modelLines := make([]string, 0, len(columns))
	for _, col := range editableColumns(columns) {
		modelLines = append(modelLines, fmt.Sprintf("\t\t%s: in.%s,", col.FieldName, col.FieldName))
	}
	nowDecl := ""
	if hasColumn(columns, "creator") {
		modelLines = append(modelLines, "\t\tCreator:   operator,")
	}
	if hasColumn(columns, "created_at") {
		modelLines = append(modelLines, "\t\tCreatedAt: now,")
		nowDecl = "\tnow := types.Now()\n"
	}
	if hasColumn(columns, "updater") {
		modelLines = append(modelLines, "\t\tUpdater:   operator,")
	}
	if hasColumn(columns, "updated_at") {
		modelLines = append(modelLines, "\t\tUpdatedAt: now,")
		nowDecl = "\tnow := types.Now()\n"
	}

	tpl := `package service

import (
	"context"
	"strings"

	"eqray-go-starter/app/modules/base/errs"
	"eqray-go-starter/app/modules/base/support"
	"eqray-go-starter/app/modules/base/types"
	"eqray-go-starter/app/modules/@@MODULE@@/model"
	"eqray-go-starter/app/modules/@@MODULE@@/repository"
)

// @@PASCAL@@ListQuery 是列表查询条件。
type @@PASCAL@@ListQuery struct {
	// Keyword 关键字
	Keyword string
	// PageNo / PageSize 分页
	PageNo   int
	PageSize int
}

// @@PASCAL@@Service 提供 @@TABLE@@ 的增删改查。
type @@PASCAL@@Service struct {
	repo *repository.@@PASCAL@@Repository
}

// New@@PASCAL@@Service 构造服务。
func New@@PASCAL@@Service() *@@PASCAL@@Service {
	return &@@PASCAL@@Service{repo: repository.New@@PASCAL@@Repository()}
}

// Page 分页查询。
func (s *@@PASCAL@@Service) Page(ctx context.Context, q @@PASCAL@@ListQuery) ([]model.@@PASCAL@@Model, int64, error) {
	pageNo, pageSize := q.PageNo, q.PageSize
	if pageNo <= 0 {
		pageNo = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}

	return s.repo.Paginate(ctx, repository.@@PASCAL@@Filter{Keyword: q.Keyword}, pageNo, pageSize)
}

// Create 新增。
func (s *@@PASCAL@@Service) Create(ctx context.Context, in repository.@@PASCAL@@Input, operator string) (types.ID, error) {
@@NOW_DECL@@	m := &model.@@PASCAL@@Model{
		ID: support.NewID(),
@@MODEL_FIELDS@@	}
	if err := s.repo.Insert(ctx, m); err != nil {
		return 0, err
	}

	return m.ID, nil
}

// Update 编辑。
func (s *@@PASCAL@@Service) Update(ctx context.Context, id types.ID, in repository.@@PASCAL@@Input, operator string) error {
	m, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if m == nil {
		return errs.WithMsg(errs.BadRequest, "记录不存在")
	}

	return s.repo.Update(ctx, id, in, operator)
}

// Delete 删除。
func (s *@@PASCAL@@Service) Delete(ctx context.Context, id types.ID) error {
	m, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if m == nil {
		return errs.WithMsg(errs.BadRequest, "记录不存在")
	}

	return s.repo.Delete(ctx, id)
}

// trimKeyword 去掉首尾空白。
func trimKeyword(value string) string {
	return strings.TrimSpace(value)
}
`

	return replaceTokens(tpl, map[string]string{
		"@@MODULE@@":       in.Module,
		"@@PASCAL@@":       in.Pascal,
		"@@TABLE@@":        in.Table,
		"@@NOW_DECL@@":     nowDecl,
		"@@MODEL_FIELDS@@": strings.Join(modelLines, "\n") + "\n",
	})
}

// renderRequest 生成 request 文件。
func renderRequest(in GeneratorInput, columns []ColumnInfo) string {
	var ruleLines strings.Builder
	for _, col := range editableColumns(columns) {
		rule := "max:255"
		switch strings.ToLower(col.Type) {
		case "text", "mediumtext", "longtext":
			rule = "max:20000"
		case "bigint", "int", "smallint", "mediumint", "tinyint":
			rule = "integer"
		case "decimal", "float", "double":
			rule = "numeric"
		}
		if !col.Nullable {
			rule = "required|" + rule
		}
		fmt.Fprintf(&ruleLines, "\t\t\"%s\": \"%s\",\n", col.JsonName, rule)
	}

	var fieldLines strings.Builder
	for _, col := range editableColumns(columns) {
		fmt.Fprintf(&fieldLines, "\t// %s\n\t%s %s `form:\"%s\" json:\"%s\"`\n",
			goDoc(col), col.FieldName, col.GoType, col.JsonName, col.JsonName)
	}

	tpl := `package request

import (
	"github.com/goravel/framework/contracts/http"
)

// @@PASCAL@@SaveRequest 是新增 / 编辑入参。
type @@PASCAL@@SaveRequest struct {
@@FIELDS@@}

// Rules 返回校验规则。
func (r *@@PASCAL@@SaveRequest) Rules(ctx http.Context) map[string]any {
	return map[string]any{
@@RULES@@	}
}
`

	return replaceTokens(tpl, map[string]string{
		"@@PASCAL@@": in.Pascal,
		"@@FIELDS@@": fieldLines.String(),
		"@@RULES@@":  ruleLines.String(),
	})
}

// renderController 生成 controller 文件。
func renderController(in GeneratorInput, columns []ColumnInfo) string {
	assignLines := make([]string, 0, len(columns))
	for _, col := range editableColumns(columns) {
		assignLines = append(assignLines, fmt.Sprintf("\t\t%s: req.%s,", col.FieldName, col.FieldName))
	}

	tpl := `package controller

import (
	"strconv"

	"github.com/goravel/framework/contracts/http"

	"eqray-go-starter/app/modules/base/param"
	"eqray-go-starter/app/modules/base/result"
	"eqray-go-starter/app/modules/base/support"
	"eqray-go-starter/app/modules/base/types"
	"eqray-go-starter/app/modules/@@MODULE@@/admin/request"
	"eqray-go-starter/app/modules/@@MODULE@@/repository"
	@@SERVICE_IMPORT@@
)

// @@PASCAL@@Controller 提供 @@TABLE@@ 的后台接口。
type @@PASCAL@@Controller struct {
	service *@@SERVICE_PKG@@.@@PASCAL@@Service
}

// New@@PASCAL@@Controller 构造控制器。
func New@@PASCAL@@Controller() *@@PASCAL@@Controller {
	return &@@PASCAL@@Controller{service: @@SERVICE_PKG@@.New@@PASCAL@@Service()}
}

// Index 分页查询。
func (c *@@PASCAL@@Controller) Index(ctx http.Context) http.Response {
	page := param.Page{
		PageNo:   ctx.Request().QueryInt("pageNo", 1),
		PageSize: ctx.Request().QueryInt("pageSize", param.DefaultPageSize),
	}
	if err := page.Normalize(); err != nil {
		return result.Fail(ctx, err)
	}

	list, total, err := c.service.Page(ctx.Context(), @@SERVICE_PKG@@.@@PASCAL@@ListQuery{
		Keyword:  ctx.Request().Query("keyword"),
		PageNo:   page.PageNo,
		PageSize: page.PageSize,
	})
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Page(ctx, list, total)
}

// Store 新增。
func (c *@@PASCAL@@Controller) Store(ctx http.Context) http.Response {
	var req request.@@PASCAL@@SaveRequest
	if err := param.Bind(ctx, &req); err != nil {
		return result.Fail(ctx, err)
	}

	id, err := c.service.Create(ctx.Context(), repository.@@PASCAL@@Input{
@@ASSIGN_FIELDS@@	}, operatorName(ctx))
	if err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, map[string]any{"id": id})
}

// Update 编辑。
func (c *@@PASCAL@@Controller) Update(ctx http.Context) http.Response {
	id, err := param.PathID(ctx)
	if err != nil {
		return result.Fail(ctx, err)
	}

	var req request.@@PASCAL@@SaveRequest
	if err := param.Bind(ctx, &req); err != nil {
		return result.Fail(ctx, err)
	}

	if err := c.service.Update(ctx.Context(), id, repository.@@PASCAL@@Input{
@@ASSIGN_FIELDS@@	}, operatorName(ctx)); err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, nil)
}

// Destroy 删除。
func (c *@@PASCAL@@Controller) Destroy(ctx http.Context) http.Response {
	id, err := param.PathID(ctx)
	if err != nil {
		return result.Fail(ctx, err)
	}

	if err := c.service.Delete(ctx.Context(), id); err != nil {
		return result.Fail(ctx, err)
	}

	return result.Success(ctx, nil)
}

// operatorName 取当前登录账号，未登录记 system。
func operatorName(ctx http.Context) string {
	if uid := support.UserID(ctx.Context()); uid != 0 {
		return types.ID(uid).String()
	}

	return "system"
}

// parseIntPtr 把查询串转成数字指针，空串返回 nil。
func parseIntPtr(value string) *int8 {
	if value == "" {
		return nil
	}

	n, err := strconv.Atoi(value)
	if err != nil {
		return nil
	}

	result := int8(n)

	return &result
}
`

	servicePkg := in.Module + "Service"
	if in.Module == "infra" {
		servicePkg = "infraService"
	}

	return replaceTokens(tpl, map[string]string{
		"@@MODULE@@":         in.Module,
		"@@PASCAL@@":         in.Pascal,
		"@@TABLE@@":          in.Table,
		"@@SERVICE_PKG@@":    servicePkg,
		"@@SERVICE_IMPORT@@": "\t" + servicePkg + " \"eqray-go-starter/app/modules/" + in.Module + "/service\"",
		"@@ASSIGN_FIELDS@@":  strings.Join(assignLines, "\n") + "\n",
	})
}

// renderRouteSnippet 生成路由注册片段（只预览，不落盘）。
func renderRouteSnippet(in GeneratorInput) string {
	name := pascalToSnake(in.Pascal)
	code := fmt.Sprintf(`// 在 app/modules/%s/admin/route.go 的 Register() 里加入：
%sController := controller.New%sController()

perm("%s:%s:list").Get("/%ss", %sController.Index)
perm("%s:%s:create").Middleware(operate("%s", "新增")).Post("/%ss", %sController.Store)
perm("%s:%s:update").Middleware(operate("%s", "编辑")).Put("/%ss/{id}", %sController.Update)
perm("%s:%s:delete").Middleware(operate("%s", "删除")).Delete("/%ss/{id}", %sController.Destroy)
`,
		in.Module,
		name, in.Pascal,
		in.Module, name, name, name,
		in.Module, name, name, name, name,
		in.Module, name, name, name, name,
		in.Module, name, name, name, name,
	)

	return code
}

// renderFrontendAPI 生成前端 api 文件。
// 路径统一用字符串拼接，避免生成代码里出现模板字符串与反引号转义问题。
func renderFrontendAPI(in GeneratorInput) string {
	name := pascalToSnake(in.Pascal)
	pascal := in.Pascal
	namespace := pascalCase(in.Module)

	return fmt.Sprintf(`import request from '@/utils/http'

// 分页查询
export function fetchGet%sList(params: { keyword?: string; pageNo?: number; pageSize?: number }) {
  return request.get<{ list: Api.%s.%sItem[]; total: number }>({
    url: '/api/v1/%s/%ss',
    params
  })
}

// 新增
export function fetchCreate%s(data: Partial<Api.%s.%sItem>) {
  return request.post<{ id: string }>({
    url: '/api/v1/%s/%ss',
    params: data
  })
}

// 编辑
export function fetchUpdate%s(id: string, data: Partial<Api.%s.%sItem>) {
  return request.put<null>({
    url: '/api/v1/%s/%ss/' + id,
    params: data
  })
}

// 删除
export function fetchDelete%s(id: string) {
  return request.del<null>({
    url: '/api/v1/%s/%ss/' + id
  })
}
`, pascal, namespace, pascal, in.Module, name,
		pascal, namespace, pascal, in.Module, name,
		pascal, namespace, pascal, in.Module, name,
		pascal, in.Module, name)
}

// renderFrontendPage 生成列表页（含新增 / 编辑弹窗，字段按表结构渲染）。
func renderFrontendPage(in GeneratorInput, columns []ColumnInfo) string {
	fields := editableColumns(columns)
	formItems := make([]string, 0, len(fields))
	formModel := make([]string, 0, len(fields))
	tableCols := make([]string, 0, len(fields))

	for _, col := range fields {
		switch col.FormType {
		case "number":
			formItems = append(formItems, fmt.Sprintf(`      <ElFormItem label="%s" prop="%s">
        <ElInputNumber v-model="form.%s" :min="0" controls-position="right" class="w-full" />
      </ElFormItem>`, goDoc(col), col.JsonName, col.JsonName))
			formModel = append(formModel, fmt.Sprintf("    %s: 0,", col.JsonName))
		case "textarea":
			formItems = append(formItems, fmt.Sprintf(`      <ElFormItem label="%s" prop="%s">
        <ElInput v-model="form.%s" type="textarea" :rows="3" />
      </ElFormItem>`, goDoc(col), col.JsonName, col.JsonName))
			formModel = append(formModel, fmt.Sprintf("    %s: '',", col.JsonName))
		case "switch":
			formItems = append(formItems, fmt.Sprintf(`      <ElFormItem label="%s" prop="%s">
        <ElSwitch v-model="form.%s" :active-value="1" :inactive-value="0" />
      </ElFormItem>`, goDoc(col), col.JsonName, col.JsonName))
			formModel = append(formModel, fmt.Sprintf("    %s: 1,", col.JsonName))
		default:
			formItems = append(formItems, fmt.Sprintf(`      <ElFormItem label="%s" prop="%s">
        <ElInput v-model="form.%s" maxlength="255" />
      </ElFormItem>`, goDoc(col), col.JsonName, col.JsonName))
			formModel = append(formModel, fmt.Sprintf("    %s: '',", col.JsonName))
		}

		tableCols = append(tableCols, fmt.Sprintf(`        <ElTableColumn prop="%s" label="%s" min-width="140" show-overflow-tooltip />`,
			col.JsonName, goDoc(col)))
	}

	tpl := `<!-- @@TABLE@@ 列表页：由代码生成器生成，字段按表结构渲染 -->
<template>
  <div class="@@NAME@@-page art-full-height">
    <ElCard class="art-table-card">
      <div class="toolbar">
        <ElInput v-model="keyword" placeholder="关键字" clearable class="w-52" @keyup.enter="handleSearch" />
        <ElButton @click="handleSearch">查询</ElButton>
        <ElButton type="primary" @click="showDialog()">新增</ElButton>
      </div>

      <ElTable v-loading="loading" :data="list" border>
@@TABLE_COLS@@
        <ElTableColumn label="操作" width="140" fixed="right">
          <template #default="{ row }">
            <ElButton link type="primary" @click="showDialog(row)">编辑</ElButton>
            <ElButton link type="danger" @click="handleDelete(row)">删除</ElButton>
          </template>
        </ElTableColumn>
        <template #empty>
          <ElEmpty description="暂无数据" :image-size="90" />
        </template>
      </ElTable>

      <div class="pager">
        <ElPagination
          v-model:current-page="pageNo"
          v-model:page-size="pageSize"
          :total="total"
          :page-sizes="[10, 20, 50]"
          layout="total, sizes, prev, pager, next"
          @current-change="loadList"
          @size-change="handleSearch"
        />
      </div>
    </ElCard>

    <ElDialog v-model="dialogVisible" :title="form.id ? '编辑' : '新增'" width="560px" align-center>
      <ElForm ref="formRef" :model="form" label-position="top">
@@FORM_ITEMS@@
      </ElForm>
      <template #footer>
        <ElButton @click="dialogVisible = false">取消</ElButton>
        <ElButton type="primary" :loading="submitting" @click="handleSubmit">提交</ElButton>
      </template>
    </ElDialog>
  </div>
</template>

<script setup lang="ts">
  import {
    ElButton,
    ElCard,
    ElDialog,
    ElEmpty,
    ElForm,
    ElFormItem,
    ElInput,
    ElInputNumber,
    ElMessage,
    ElMessageBox,
    ElPagination,
    ElSwitch,
    ElTable,
    ElTableColumn
  } from 'element-plus'
  import {
    fetchCreate@@PASCAL@@,
    fetchDelete@@PASCAL@@,
    fetchGet@@PASCAL@@List,
    fetchUpdate@@PASCAL@@
  } from '@/api/@@MODULE@@/@@NAME@@'

  defineOptions({ name: '@@PASCAL@@Page' })

  const loading = ref(false)
  const submitting = ref(false)
  const list = ref<any[]>([])
  const total = ref(0)
  const pageNo = ref(1)
  const pageSize = ref(10)
  const keyword = ref('')

  const dialogVisible = ref(false)
  const formRef = ref()
  const form = reactive<any>({
    id: undefined,
@@FORM_MODEL@@  })

  const loadList = async (): Promise<void> => {
    loading.value = true
    try {
      const res = await fetchGet@@PASCAL@@List({
        keyword: keyword.value || undefined,
        pageNo: pageNo.value,
        pageSize: pageSize.value
      })
      list.value = res?.list ?? []
      total.value = res?.total ?? 0
    } catch {
      list.value = []
      total.value = 0
    } finally {
      loading.value = false
    }
  }

  const handleSearch = (): void => {
    pageNo.value = 1
    loadList()
  }

  const showDialog = (row?: any): void => {
    Object.assign(form, row ?? { id: undefined })
    dialogVisible.value = true
  }

  const handleSubmit = async (): Promise<void> => {
    submitting.value = true
    try {
      if (form.id) {
        await fetchUpdate@@PASCAL@@(String(form.id), form)
        ElMessage.success('保存成功')
      } else {
        await fetchCreate@@PASCAL@@(form)
        ElMessage.success('新增成功')
      }
      dialogVisible.value = false
      loadList()
    } catch {
      // 错误提示已由请求层统一弹出
    } finally {
      submitting.value = false
    }
  }

  const handleDelete = async (row: any): Promise<void> => {
    try {
      await ElMessageBox.confirm('确定删除这条记录？', '删除', {
        confirmButtonText: '删除',
        cancelButtonText: '取消',
        type: 'warning'
      })
    } catch {
      return
    }

    try {
      await fetchDelete@@PASCAL@@(String(row.id))
      ElMessage.success('已删除')
      loadList()
    } catch {
      // 错误提示已由请求层统一弹出
    }
  }

  onMounted(loadList)
</script>

<style lang="scss" scoped>
  .toolbar {
    display: flex;
    flex-wrap: wrap;
    gap: 8px;
    margin-bottom: 12px;

    .w-52 {
      width: 208px;
    }
  }

  .pager {
    display: flex;
    justify-content: flex-end;
    margin-top: 16px;
  }
</style>
`

	return replaceTokens(tpl, map[string]string{
		"@@TABLE@@":      in.Table,
		"@@NAME@@":       pascalToSnake(in.Pascal),
		"@@MODULE@@":     in.Module,
		"@@PASCAL@@":     in.Pascal,
		"@@TABLE_COLS@@": strings.Join(tableCols, "\n"),
		"@@FORM_ITEMS@@": strings.Join(formItems, "\n"),
		"@@FORM_MODEL@@": strings.Join(formModel, "\n"),
	})
}

// replaceTokens 批量替换模板占位符。
// 入参 tpl 为模板文本、tokens 为占位符 → 值；返回替换结果。
func replaceTokens(tpl string, tokens map[string]string) string {
	for key, value := range tokens {
		tpl = strings.ReplaceAll(tpl, key, value)
	}

	return tpl
}
