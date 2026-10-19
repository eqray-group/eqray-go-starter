# AGENTS.md — 开发指南

本文件面向 AI 助手与开发者，汇总本项目的约定、模板与踩坑记录。完整设计说明见
[`docs/framework-design.md`](docs/framework-design.md)，两者冲突时以设计文档为准，并请同步修正本文件。

---

## 1. 项目速览

| 项       | 说明                                                                                            |
| -------- | ----------------------------------------------------------------------------------------------- |
| 后端     | Go 1.25、Goravel 1.18（底层 Gin）、MySQL 8、Redis                                               |
| 前端     | Vue 3.5、Vite 7、TypeScript、Element Plus 2.11、Tailwind 4、Pinia（`ui/`，Art Design Pro 风格） |
| 端口     | 后端 `3000`，前端 `3006`（开发环境代理 `/api` 到 3000）                                         |
| 认证     | 账号密码 + 图形验证码，访问令牌 / 刷新令牌双令牌                                                |
| 权限     | RBAC，权限点形如 `system:user:create`；超管角色 `SUPER_ADMIN` 直接放行                          |
| 菜单     | `VITE_ACCESS_MODE=backend`，前端路由由后端 `system_menu` 下发                                   |
| 默认账号 | `admin` / `admin123`（仅本地初始化用）                                                          |

已上线的业务模块：`system`（认证、用户、角色、菜单、字典、登录日志、操作日志）、
`infra`（参数配置、文件管理、文件配置、API 访问日志、API 错误日志）。

> 定时任务功能已整体移除（代码、迁移、菜单、数据表），不要再引入，除非明确要求。

---

## 2. 常用命令

```bash
# 后端
go run .                              # 启动，http://127.0.0.1:3000
go run . artisan migrate              # 执行迁移
go run . artisan migrate:status       # 查看迁移状态
go run . artisan key:generate         # 生成 APP_KEY
go build ./...                        # 编译校验
go vet ./...                          # 静态检查

# 前端（ui/）
pnpm install
pnpm dev                              # http://127.0.0.1:3006
pnpm build                            # vue-tsc --noEmit && vite build
npx vue-tsc --noEmit                  # 仅类型检查
npx eslint "src/views/xxx" --fix      # 格式化 + 规范修复
```

**提交前至少跑**：`go build ./...`、`go vet ./...`、`npx vue-tsc --noEmit`、`eslint`。

---

## 3. 目录结构

```
app/
  facades/                 # 框架门面（DB / Schema / Cache / Storage / Config ...）
  http/                    # 中间件、异常处理等装配
  modules/
    base/                  # 认证、响应、错误码、参数、枚举、类型、权限中间件
    system/                # 登录认证、用户、角色、菜单、字典、日志
    infra/                 # 参数配置、文件、文件配置、API 日志
      route.go             # 模块路由入口（Register）
      controller/          # 参数绑定 + 调 service + 统一响应
      service/             # 业务规则、事务编排
      repository/          # 数据访问
      model/               # 表模型
      request/             # 入参结构 + 校验规则 + 中文提示
      errs/                # 模块业务错误码
      middleware/          # 请求日志等
bootstrap/
  app.go                   # 应用装配（中间件、Provider、路由）
  migrations.go            # 迁移注册表（★ 必须手动注册）
  providers.go
config/                    # 配置文件
database/migrations/       # 迁移（建表 + 种子数据）
routes/                    # web / grpc / modules 入口
ui/src/
  api/{system,infra}/      # 接口封装，按模块分文件
  views/{system,infra}/    # 页面，`页面名/index.vue` + `modules/*.vue`
  hooks/core/              # useTable、useAuth、useTheme 等
  components/core/         # ArtTable、ArtTableHeader、ArtSearchBar、ArtButtonTable 等
  types/api/api.d.ts       # 全局接口类型（namespace Api）
  router/                  # 静态路由 + 后端动态路由装配
  store/modules/           # Pinia
  directives/core/         # v-auth、v-roles 等指令
```

---

## 4. 后端开发

### 4.1 分层职责

| 层           | 职责                                     | 禁止                 |
| ------------ | ---------------------------------------- | -------------------- |
| `controller` | 取参、绑定校验、调 service、返回统一响应 | 写业务规则、直接查库 |
| `service`    | 业务校验、组合仓储、事务、缓存           | 依赖 `http.Context`  |
| `repository` | 单表/多表查询，返回 model                | 写业务分支           |
| `request`    | 入参结构、`Rules`、`Messages`            | 访问数据库           |
| `errs`       | 模块业务错误码                           | 拼接文案             |

### 4.2 新增一个接口的步骤

1. `request/xxx_request.go`：定义入参 + `Rules()` + `Messages()`（中文提示必须补）。
2. `service/xxx_service.go`：写业务方法，错误用 `baseErrs.New(errCode, args...)`。
3. `repository/xxx_repository.go`：数据访问（如需新查询）。
4. `controller/xxx_controller.go`：`param.Bind` / `param.PathID`，调 service，`result.Success/Fail/Page`。
5. `route.go`：在对应 `perm("模块:资源:动作")` 下注册路由。
6. `errs/infra_error.go`（或 system 的 errs）：新增错误码。
7. 迁移：新表/新字段/种子数据（见 4.6）。
8. 菜单与权限点：如新增页面，必须补 seed 迁移（见 4.7）。
9. 校验：`go build ./... && go vet ./...`。

### 4.3 统一响应与错误码

```go
result.Success(ctx, data)          // { code: 0, message: "success", data }
result.Fail(ctx, err)              // 业务错误，HTTP 200，code != 0
result.Page(ctx, list, total)      // { list, total }
```

- 错误码规则：`1_{模块}_{子模块}_{序号}`，例如 `1_001_002_001`。
- 定义：`baseErrs.ErrorCode{Code: ..., Msg: "文件【%v】不存在"}`，用 `baseErrs.New(code, id)` 填充。
- 需要覆盖文案时用 `baseErrs.WithMsg(baseErrs.BadRequest, "……")`。
- 鉴权中间件这类不经过 `result.Fail` 的中断，**必须调用 `result.MarkFail(ctx, code)`**，否则 API 错误日志的 `message` 为空。

### 4.4 参数校验

```go
// request
func (r *XxxRequest) Rules(ctx http.Context) map[string]any {
    return map[string]any{
        "name":   "required|max:64",
        "status": "integer|in:0,1",
    }
}

func (r *XxxRequest) Messages(ctx http.Context) map[string]string {
    return map[string]string{"name.required": "名称不能为空"}
}

// controller
var req request.XxxRequest
if err := param.Bind(ctx, &req); err != nil {
    return result.Fail(ctx, err)
}
```

分页统一用 `param.Page{PageNo, PageSize}` + `Normalize()`，默认页大小见 `param.DefaultPageSize`。

### 4.5 权限与路由

```go
facades.Route().Prefix("/api/v1/infra").Group(func(r route.Router) {
    auth := r.Middleware(&baseMiddleware.TokenAuth{})
    perm := func(code string) route.Router {
        return auth.Middleware(&baseMiddleware.Permission{Code: code})
    }

    perm("infra:file-config:list").Get("/file-configs", fileConfigController.Index)
})
```

- 权限点命名：`模块:资源:动作`（`list/create/update/delete`，特殊动作另起，如 `infra:api-log:update`）。
- 路径参数用 `{id}`，`param.PathID(ctx)` 解析。
- 更具体的静态路径（如 `/jobs/handlers`）要注册在 `/jobs/{id}` 之前，避免被路径参数吞掉。
- 上传这类自助接口只挂 `TokenAuth`，不挂具体权限点。

### 4.6 迁移（最容易踩坑）

**文件名**：`YYYYMMDDHHMMSS_snake_case.go`，例如 `20261020000013_seed_infra_file_config.go`。

**必须在 `bootstrap/migrations.go` 注册**，否则 `artisan migrate` 提示 `Nothing to migrate`，数据和表都不会创建（本项目曾因此 `system_menu` 缺数据）。

```go
type M20261020000013SeedInfraFileConfig struct{}

func (r *M20261020000013SeedInfraFileConfig) Signature() string {
    return "20261020000013_seed_infra_file_config"
}

func (r *M20261020000013SeedInfraFileConfig) Up() error {
    if !facades.Schema().HasTable("infra_file_config") {
        return nil
    }
    // ...幂等写入
    return nil
}

func (r *M20261020000013SeedInfraFileConfig) Down() error {
    return nil
}
```

约定：

- `Up` 必须幂等（`HasTable` / 先查再插）。
- 建表用 `facades.DB().Statement("CREATE TABLE ...")`，风格参考 `20261020000008_create_infra_api_access_log_table.go`。
- **已执行的迁移不要修改**，需求变更时新增迁移。
- 迁移文件删除后，要同步删除 `migrations` 表中对应的执行记录，避免"有记录无文件"。
- 删除功能时，若文件已执行过，删除迁移文件不会撤销已做的 DB 变更，需另写清理迁移或手工处理。
- 权限/菜单变更后前移缓存版本，避免管理员等待 30 分钟：

```go
_ = facades.Cache().Put("rbac:version", strconv.FormatInt(time.Now().UnixNano(), 10), 30*time.Minute)
```

### 4.7 菜单与权限点种子

菜单表 `system_menu`：`type=1` 目录、`type=2` 菜单、`type=3` 按钮。
新增页面时，向 `system_menu` 插入页面菜单 + 按钮权限点，并授权给内置角色（`system_role_menu`）。

参考 `20261020000012_seed_infra_ops_menus.go` 的 `ensureMenu` / `grantBuiltinRoles` 写法：
先按 `path`/`permission` 查是否存在（注意 `deleted_at`），不存在再插入，最后给所有内置角色补齐授权。

**组件路径规则**：`component` 填 `/infra/file-config`，前端 `ComponentLoader` 会解析到
`ui/src/views/infra/file-config/index.vue`。路径填错 → 菜单点击白屏。

### 4.8 文件存储

- 配置表 `infra_file_config`，`storage` 支持 `local` / `s3`；主配置唯一（业务层保证）。
- `config` 为 JSON：
  - local：`{"dir":"uploads"}`
  - s3：`{"bucket","endpoint","region","accessKey","secretKey","url","usePathStyle","dir"}`
- 上传路径：`{dir}/{YYYY}/{MM}/{DD}/{雪花ID}.{ext}`，原文件名保留在 `infra_file.name`。
- local 地址：`/storage/{path}`；s3 地址：`driver.Url(path)`（依赖 `url` 前缀与 Bucket 公开读）。
- 删除物理文件时按 `infra_file.config_id` 找到原配置选择磁盘，配置缺失才回落 local。

---

## 5. 后端模板

### 5.1 Controller

```go
// Index GET /api/v1/infra/xxx
func (c *XxxController) Index(ctx http.Context) http.Response {
    page := param.Page{
        PageNo:   ctx.Request().QueryInt("pageNo", 1),
        PageSize: ctx.Request().QueryInt("pageSize", param.DefaultPageSize),
    }
    if err := page.Normalize(); err != nil {
        return result.Fail(ctx, err)
    }

    list, total, err := c.service.Page(ctx.Context(), service.XxxQuery{
        Keyword:  ctx.Request().Query("keyword"),
        PageNo:   page.PageNo,
        PageSize: page.PageSize,
    })
    if err != nil {
        return result.Fail(ctx, err)
    }

    return result.Page(ctx, list, total)
}

// Store POST /api/v1/infra/xxx
func (c *XxxController) Store(ctx http.Context) http.Response {
    var req request.XxxRequest
    if err := param.Bind(ctx, &req); err != nil {
        return result.Fail(ctx, err)
    }
    if err := c.service.Create(ctx.Context(), req.ToInput()); err != nil {
        return result.Fail(ctx, err)
    }
    return result.Success(ctx, nil)
}
```

### 5.2 Service

```go
type XxxService struct {
    items *repository.XxxRepository
}

func NewXxxService() *XxxService {
    return &XxxService{items: repository.NewXxxRepository()}
}

func (s *XxxService) Create(ctx context.Context, in XxxInput) error {
    if err := validateXxx(in); err != nil {
        return err
    }

    return s.items.Create(ctx, &model.InfraXxxModel{
        Name: strings.TrimSpace(in.Name),
    })
}
```

### 5.3 Repository

```go
func (r *XxxRepository) Paginate(ctx context.Context, filter XxxFilter, pageNo, pageSize int) ([]model.InfraXxxModel, int64, error) {
    var list []model.InfraXxxModel
    var total int64

    query := facades.DB().WithContext(ctx).Table(xxxTable).WhereNull("deleted_at")
    if filter.Keyword != "" {
        query = query.WhereLike("name", "%"+filter.Keyword+"%")
    }
    if filter.Status != nil {
        query = query.Where("status", *filter.Status)
    }

    if err := query.OrderByDesc("id").Paginate(pageNo, pageSize, &list, &total); err != nil {
        return nil, 0, err
    }

    return list, total, nil
}
```

### 5.4 Request

```go
type XxxRequest struct {
    Name   string `form:"name" json:"name"`
    Status int8   `form:"status" json:"status"`
}

func (r *XxxRequest) Rules(ctx http.Context) map[string]any {
    return map[string]any{
        "name":   "required|max:64",
        "status": "integer|in:0,1",
    }
}

func (r *XxxRequest) Messages(ctx http.Context) map[string]string {
    return map[string]string{
        "name.required": "名称不能为空",
        "status.in":     "状态取值不合法",
    }
}
```

### 5.5 错误码

```go
var (
    InfraXxxNotExists = baseErrs.ErrorCode{Code: 1_001_005_001, Msg: "记录【%v】不存在"}
)
```

---

## 6. 前端开发

### 6.1 页面目录与路由

```
views/infra/file-config/
  index.vue                  # 列表页
  modules/file-config-dialog.vue
```

- 路由由后端菜单驱动：`component` 为 `/infra/file-config` → `views/infra/file-config/index.vue`。
- 不需要在前端路由文件里手写业务页面。
- 页面根元素建议 `class="xxx-page art-full-height"`。

### 6.2 列表页统一范式

必须使用项目封装的组件，不要直接写 `ElTable`（样式会与全站不一致）：

```
ArtSearchBar
ElCard.art-table-card
  ArtTableHeader（v-model:columns、加载态、刷新、左侧新增按钮）
  ArtTable（data、columns、pagination、分页事件）
```

`useTable` 返回值：

| 字段                                                | 用途                           |
| --------------------------------------------------- | ------------------------------ |
| `columns` / `columnChecks`                          | 列配置 / 列显隐双向绑定        |
| `data` / `loading` / `pagination`                   | 表格数据、加载态、分页         |
| `getData()`                                         | 按当前条件重新查询             |
| `searchParams`                                      | 实际请求参数                   |
| `resetSearchParams()`                               | 重置搜索并查询                 |
| `handleSizeChange` / `handleCurrentChange`          | 分页事件                       |
| `refreshData`                                       | 保持条件刷新（工具条刷新按钮） |
| `refreshCreate` / `refreshUpdate` / `refreshRemove` | 新增/编辑/删除后的刷新         |

**项目约定**：

- 表格不显示「序号」列（已统一移除，不要加回来）。
- 操作按钮用 `ArtButtonTable`（`type: edit/delete/action/view`），不要用文字链接。
- 操作列权限用 `hasAuth('模块:资源:动作')`；页面按钮可用 `v-auth`（已兼容后端 `buttons`）。
- 金额/状态/时间用 `formatter` 渲染 `ElTag` 等，避免模板里堆 `v-if`。

### 6.3 弹窗

- 用 `v-model:visible` 双向绑定，`type: 'add' | 'edit'`，`xxx-data` 传当前行。
- 提交成功 `emit('success')`，父页面按类型调用 `refreshCreate` / `refreshUpdate`。
- 打开时在 `watch(visible)` 中回填并 `clearValidate()`。
- 关闭确认框：`try { await ElMessageBox.confirm(...) } catch { return }`。
- 请求错误提示由 http 层统一弹出，`catch` 块留空并注释即可，不要重复 `ElMessage.error`。

### 6.4 API 封装

```ts
// ui/src/api/infra/xxx.ts
import request from "@/utils/http";

export function fetchGetXxxList(params: Api.Infra.XxxSearchParams) {
  return request.get<Api.Infra.XxxList>({ url: "/api/v1/infra/xxxs", params });
}

export function fetchCreateXxx(data: Api.Infra.XxxParams) {
  return request.post<null>({ url: "/api/v1/infra/xxxs", params: data });
}

export function fetchUpdateXxx(id: string, data: Api.Infra.XxxParams) {
  return request.put<null>({ url: `/api/v1/infra/xxxs/${id}`, params: data });
}

export function fetchDeleteXxx(id: string) {
  return request.del<null>({ url: `/api/v1/infra/xxxs/${id}` });
}
```

- `params` 在 POST/PUT 时会被 http 层转成 JSON body，与后端 `param.Bind` 对应。
- `request.get<T>` 返回的是后端 `data` 字段（已解包），不是完整响应体。
- ID 为雪花 ID，统一用 `string`。

### 6.5 类型定义

全部集中在 `ui/src/types/api/api.d.ts` 的 `namespace Api` 下，按模块分：

```ts
namespace Api {
  namespace Infra {
    interface XxxItem {
      id: string;
      name: string;
      createdAt: string;
    }

    interface XxxList {
      list: XxxItem[];
      total: number;
    }

    interface XxxParams {
      name: string;
    }

    interface XxxSearchParams {
      keyword?: string;
      pageNo?: number;
      pageSize?: number;
    }
  }
}
```

### 6.6 样式

- Tailwind 4 原子类为主，间距用 `mt-1` / `gap-2` / `flex items-center`。
- 次要文字用 `text-xs text-g-600`（项目自定义色阶）。
- 卡片用 `ElCard.art-table-card`，页面根容器加 `art-full-height`。
- 需要 CSS 变量时用 Element 变量：`text-[var(--el-color-primary)]`、`bg-[var(--el-fill-color-light)]`。

---

## 7. 前端模板

### 7.1 列表页

```vue
<template>
  <div class="xxx-page art-full-height">
    <ArtSearchBar
      v-model="searchForm"
      :items="searchItems"
      :rules="{}"
      @reset="resetSearchParams"
      @search="handleSearch"
    />

    <ElCard class="art-table-card">
      <ArtTableHeader
        v-model:columns="columnChecks"
        :loading="loading"
        @refresh="refreshData"
      >
        <template #left>
          <ElSpace wrap>
            <ElButton
              v-auth="'infra:xxx:create'"
              @click="showDialog('add')"
              v-ripple
            >
              新增
            </ElButton>
          </ElSpace>
        </template>
      </ArtTableHeader>

      <ArtTable
        :loading="loading"
        :data="data"
        :columns="columns"
        :pagination="pagination"
        @pagination:size-change="handleSizeChange"
        @pagination:current-change="handleCurrentChange"
      />
    </ElCard>

    <XxxDialog
      v-model:visible="dialogVisible"
      :type="dialogType"
      :xxx-data="currentRow"
      @success="handleDialogSuccess"
    />
  </div>
</template>

<script setup lang="ts">
import { useTable } from "@/hooks/core/useTable";
import { useAuth } from "@/hooks/core/useAuth";
import { fetchGetXxxList } from "@/api/infra/xxx";
import { ElButton, ElCard, ElSpace, ElTag } from "element-plus";
import ArtButtonTable from "@/components/core/forms/art-button-table/index.vue";
import XxxDialog from "./modules/xxx-dialog.vue";

defineOptions({ name: "InfraXxx" });

type XxxItem = Api.Infra.XxxItem;
const { hasAuth } = useAuth();

const searchForm = ref<Api.Infra.XxxSearchParams>({ keyword: undefined });
const searchItems = [
  {
    label: "关键字",
    key: "keyword",
    type: "input",
    placeholder: "请输入",
    clearable: true,
  },
];

const dialogVisible = ref(false);
const dialogType = ref<"add" | "edit">("add");
const currentRow = ref<XxxItem | null>(null);

const {
  columns,
  columnChecks,
  data,
  loading,
  pagination,
  getData,
  searchParams,
  resetSearchParams,
  handleSizeChange,
  handleCurrentChange,
  refreshData,
  refreshCreate,
  refreshUpdate,
  refreshRemove,
} = useTable({
  core: {
    apiFn: fetchGetXxxList,
    apiParams: { pageNo: 1, pageSize: 20, ...searchForm.value },
    columnsFactory: () => [
      { prop: "name", label: "名称", minWidth: 160, showOverflowTooltip: true },
      {
        prop: "status",
        label: "状态",
        width: 100,
        formatter: (row: XxxItem) =>
          h(ElTag, { type: row.status === 1 ? "success" : "info" }, () =>
            row.status === 1 ? "启用" : "禁用",
          ),
      },
      {
        prop: "operation",
        label: "操作",
        width: 180,
        fixed: "right",
        formatter: (row: XxxItem) =>
          h("div", { class: "flex items-center" }, [
            ...(hasAuth("infra:xxx:update")
              ? [
                  h(ArtButtonTable, {
                    type: "edit",
                    text: "编辑",
                    onClick: () => showDialog("edit", row),
                  }),
                ]
              : []),
            ...(hasAuth("infra:xxx:delete")
              ? [
                  h(ArtButtonTable, {
                    type: "delete",
                    text: "删除",
                    onClick: () => handleDelete(row),
                  }),
                ]
              : []),
          ]),
      },
    ],
  },
});

const handleSearch = (params: Api.Infra.XxxSearchParams) => {
  Object.assign(searchParams, params);
  getData();
};

const showDialog = (type: "add" | "edit", row?: XxxItem) => {
  dialogType.value = type;
  currentRow.value = type === "edit" ? (row ?? null) : null;
  dialogVisible.value = true;
};

const handleDialogSuccess = () => {
  dialogType.value === "edit" ? refreshUpdate() : refreshCreate();
};
</script>
```

### 7.2 弹窗

```vue
<template>
  <ElDialog
    v-model="dialogVisible"
    :title="isEdit ? '编辑' : '新增'"
    width="560px"
    align-center
    :close-on-click-modal="false"
  >
    <ElForm ref="formRef" :model="form" :rules="rules" label-width="110px">
      <ElFormItem label="名称" prop="name">
        <ElInput v-model="form.name" placeholder="请输入名称" maxlength="64" />
      </ElFormItem>
    </ElForm>

    <template #footer>
      <ElButton @click="dialogVisible = false">取消</ElButton>
      <ElButton type="primary" :loading="submitting" @click="handleSubmit"
        >提交</ElButton
      >
    </template>
  </ElDialog>
</template>

<script setup lang="ts">
import type { FormInstance, FormRules } from "element-plus";
import { ElDialog, ElMessage } from "element-plus";
import { fetchCreateXxx, fetchUpdateXxx } from "@/api/infra/xxx";

defineOptions({ name: "XxxDialog" });

const props = withDefaults(
  defineProps<{
    visible: boolean;
    type: "add" | "edit";
    xxxData?: Api.Infra.XxxItem | null;
  }>(),
  { xxxData: null },
);
const emit = defineEmits<{
  (e: "update:visible", value: boolean): void;
  (e: "success"): void;
}>();

const dialogVisible = computed({
  get: () => props.visible,
  set: (value) => emit("update:visible", value),
});
const isEdit = computed(() => props.type === "edit");

const formRef = ref<FormInstance>();
const submitting = ref(false);
const form = reactive({ name: "" });
const rules: FormRules = {
  name: [{ required: true, message: "请输入名称", trigger: "blur" }],
};

watch(
  () => props.visible,
  (visible) => {
    if (!visible) return;
    Object.assign(form, { name: props.xxxData?.name ?? "" });
    nextTick(() => formRef.value?.clearValidate());
  },
  { immediate: true },
);

const handleSubmit = async () => {
  const valid = await formRef.value?.validate().catch(() => false);
  if (!valid) return;

  submitting.value = true;
  try {
    const payload = { name: form.name.trim() };
    if (isEdit.value && props.xxxData) {
      await fetchUpdateXxx(props.xxxData.id, payload);
    } else {
      await fetchCreateXxx(payload);
    }
    ElMessage.success("保存成功");
    dialogVisible.value = false;
    emit("success");
  } catch {
    // 错误提示由请求层统一处理
  } finally {
    submitting.value = false;
  }
};
</script>
```

---

## 8. 已知坑（务必先看）

### 8.1 Goravel 的 `Where` 三参数写法会失效

```go
// ❌ 会被编译成 id IN ('!=', 123)，条件恒不成立
query.Where("id", "!=", id)

// ✅ 用占位符
query.Where("id <> ?", id)

// ✅ 模糊查询用 WhereLike
query.WhereLike("name", "%"+keyword+"%")

// ✅ WhereIn 需要 []any
args := make([]any, 0, len(ids))
for _, id := range ids { args = append(args, id) }
query.WhereIn("id", args)
```

本项目曾因此出现「同时存在多个主文件配置」「编辑参数误报键已存在」。

### 8.2 迁移不注册 = 不执行

新增迁移文件后必须加进 `bootstrap/migrations.go`，否则 `migrate` 直接提示 `Nothing to migrate`。

### 8.3 `v-auth` 与后端动态菜单

`v-auth` 指令同时读取 `user.info.buttons`（后端下发）与路由 `meta.authList`。
后端动态菜单不生成 `meta.authList`，如果只读它，所有按钮都会被移除（曾导致「新增」按钮消失）。

### 8.4 权限/菜单缓存

权限结果缓存 30 分钟（`rbac:version`）。修改菜单或权限点的迁移必须前移该版本号；
前端本地还持久化了菜单与路由，改完需要重新登录或清缓存才能看到变化。

### 8.5 表格样式与序号列

- 列表页统一 `ArtTable`/`ArtTableHeader`/`ArtSearchBar`，不要用原生 `ElTable`。
- 不要加「序号」列（项目统一约定）。

### 8.6 错误日志 `message` 为空

`RequestLog` 中间件从上下文读取 `result.Fail` 写入的文案。直接 `Abort(401/403)` 的中间件
必须调用 `result.MarkFail(ctx, code)`；框架直接返回的 404/405 会走兜底文案。

### 8.7 上传与存储

- 上传路径 `uploads/YYYY/MM/DD/{雪花ID}.{ext}`，原文件名存 `infra_file.name`。
- S3/MinIO 需要 `usePathStyle=true`；`url` 前缀决定文件访问地址。
- Bucket 未开放公开读时图片会 403，需配置 Bucket 策略或填 CDN 域名。
- 主配置唯一由业务层保证（`master` 无数据库唯一约束）。

### 8.8 更新 RowsAffected 为 0

MySQL 在"新旧值相同"时 `RowsAffected` 为 0，不要直接判定记录不存在，
需要再查一次存在性（参考 `FileConfigRepository.SetMaster`）。

### 8.9 迁移文件删除

删除已执行的迁移文件不会撤销数据库变更，需另写清理迁移或手工处理；
同时要删除 `migrations` 表中的执行记录，保持状态一致。

---

## 9. 常见问题排查

| 现象                      | 排查顺序                                                                                        |
| ------------------------- | ----------------------------------------------------------------------------------------------- |
| 菜单点击白屏              | `system_menu.component` 是否为 `/模块/页面`；对应的 `views/模块/页面/index.vue` 是否存在        |
| 菜单不显示                | 迁移是否注册并执行；`system_role_menu` 是否授权；`rbac:version` 是否前移；是否重新登录          |
| 按钮不显示                | `v-auth` 的权限点是否在 `system_menu.permission`；当前用户角色是否授权；`info.buttons` 是否刷新 |
| 接口 403                  | 路由上的 `perm(...)` 与前端权限点是否一致；超管外角色是否授权                                   |
| 接口 500 且错误日志无信息 | 是否走了 `Abort` 而未 `MarkFail`；查看 `storage/logs`                                           |
| 文件配置多主配置          | 检查 `ClearMaster` 是否用了 `Where("id <> ?", id)`                                              |
| 上传 404 / 图片 403       | local 检查 `/storage` 静态路由；s3 检查 `url` 前缀与 Bucket 公开读                              |

---

## 10. 提交前检查清单

**后端**

- [ ] 新增文件已加入对应分层目录，命名与既有文件一致
- [ ] 请求参数有 `Rules` 与中文 `Messages`
- [ ] 错误使用业务错误码，未硬编码文案
- [ ] 路由挂了正确的 `perm(...)`，权限点已写入菜单种子
- [ ] 新迁移已注册到 `bootstrap/migrations.go`，且幂等
- [ ] 菜单/权限变更已前移 `rbac:version`
- [ ] `go build ./...`、`go vet ./...` 通过

**前端**

- [ ] 页面使用 `ArtSearchBar + ArtTableHeader + ArtTable + useTable`
- [ ] 没有「序号」列
- [ ] 按钮权限：页面按钮 `v-auth`，行内操作 `hasAuth`
- [ ] 新增/编辑/删除后调用 `refreshCreate / refreshUpdate / refreshRemove`
- [ ] 接口封装在 `api/{模块}`，类型写在 `Api.{模块}` 命名空间
- [ ] 弹窗用 `v-model:visible` + `emit('success')`
- [ ] `npx vue-tsc --noEmit`、`eslint` 通过
