<div align="center">
  <img src="./ui/src/assets/images/common/logo.svg" width="120" alt="eqray-go-starter logo" />

  <h1>eqray-go-starter</h1>

  <p>
    <strong>Go + Vue 3 中后台全栈脚手架</strong><br />
    认证、权限、菜单、日志、会员，开箱即用
  </p>

  <p>
    <a href="https://go.dev"><img src="https://img.shields.io/badge/Go-1.25-00ADD8?logo=go&logoColor=white" alt="Go" /></a>
    <a href="https://www.goravel.dev"><img src="https://img.shields.io/badge/Goravel-1.18-FF2D20" alt="Goravel" /></a>
    <a href="https://vuejs.org"><img src="https://img.shields.io/badge/Vue-3.5-4FC08D?logo=vuedotjs&logoColor=white" alt="Vue" /></a>
    <a href="https://vite.dev"><img src="https://img.shields.io/badge/Vite-7.1-646CFF?logo=vite&logoColor=white" alt="Vite" /></a>
    <a href="https://www.typescriptlang.org"><img src="https://img.shields.io/badge/TypeScript-5.6-3178C6?logo=typescript&logoColor=white" alt="TypeScript" /></a>
  </p>

  <p>
    <a href="https://element-plus.org"><img src="https://img.shields.io/badge/Element%20Plus-2.11-409EFF" alt="Element Plus" /></a>
    <a href="https://tailwindcss.com"><img src="https://img.shields.io/badge/Tailwind%20CSS-4.1-06B6D4?logo=tailwindcss&logoColor=white" alt="Tailwind CSS" /></a>
    <a href="https://www.mysql.com"><img src="https://img.shields.io/badge/MySQL-8-4479A1?logo=mysql&logoColor=white" alt="MySQL" /></a>
    <a href="https://redis.io"><img src="https://img.shields.io/badge/Redis-6%2B-DC382D?logo=redis&logoColor=white" alt="Redis" /></a>
    <a href="LICENSE"><img src="https://img.shields.io/badge/License-MIT-blue" alt="License" /></a>
    <a href="../../issues"><img src="https://img.shields.io/badge/PRs-welcome-brightgreen" alt="PRs welcome" /></a>
  </p>

  <p>
    <a href="#快速开始">快速开始</a> ·
    <a href="#功能模块">功能模块</a> ·
    <a href="#接口约定">接口约定</a> ·
    <a href="docs/openapi/member.yaml">接口文档</a> ·
    <a href="docs/framework-design.md">设计文档</a>
  </p>
</div>

后端基于 Goravel 1.18（底层 Gin），内置双令牌认证、RBAC 权限与模块化分层；前端基于 Vue 3 + Element Plus，
提供系统管理、基础设施、会员中心三套完整业务页面。

拉下来跑一遍迁移、改掉默认密码，就能直接在业务模块上写代码 —— 不用再从登录页、权限表和日志表开始搭。

---

## 目录

- [特性](#特性)
- [技术栈](#技术栈)
- [功能模块](#功能模块)
- [架构概览](#架构概览)
- [快速开始](#快速开始)
- [默认账号](#默认账号)
- [环境变量](#环境变量)
- [接口约定](#接口约定)
- [文档](#文档)
- [部署](#部署)
- [开发约定](#开发约定)
- [常见问题](#常见问题)

---

## 特性

- **双令牌认证**：访问令牌 30 分钟、刷新令牌 7 天，签发时只把 SHA-256 摘要写入 Redis，读走 Redis 也拿不到可用令牌，天然支持「删除即失效」。
- **登录防护**：账号密码 + 图形验证码，登录日志落库。
- **RBAC 权限**：权限点形如 `system:user:create`，与 `system_menu.permission` 一一对应；内置超管角色 `SUPER_ADMIN` 直接放行。
- **后端动态菜单**：`VITE_ACCESS_MODE=backend`，前端路由完全由后端 `system_menu` 下发，改菜单不用发版。
- **统一响应与错误码**：`{ code, msg, data }`，HTTP 恒为 200，业务结果只看 `code`；错误码按 `1_{模块}_{子模块}_{序号}` 编码。
- **两端接入隔离**：管理端 `/api/v1/{module}` 走令牌 + 权限点，C 端 `/api/v1/client/{module}` 只认会员令牌，作用域互不通用。
- **模块化分层**：`controller / request / service / repository / model / errs` 职责单一，接入层按 `admin` / `client` 分目录，模块之间依赖单向。
- **多存储文件上传**：`local` / `s3` 双驱动，主配置唯一，删除文件自动回到原存储。
- **会员体系**：等级、标签、分组、积分流水、签到奖励档位、收货地址、微信小程序 / 公众号 / 开放平台登录与手机号验证码登录。
- **可运行的默认数据**：迁移自带管理员账号、角色、完整菜单树、默认等级、默认短信渠道与登录模板。

## 技术栈

| 层 | 技术 |
| --- | --- |
| 后端 | Go 1.25、Goravel 1.18（底层 Gin）、MySQL 8、Redis 6+ |
| 前端 | Vue 3.5、Vite 7、TypeScript、Element Plus 2.11、Tailwind CSS 4、Pinia |
| 认证 | 账号密码 + 图形验证码，访问 / 刷新双令牌 |
| 权限 | RBAC（角色 - 菜单 - 权限点） |

前端基于 [Art Design Pro](https://github.com/Daymychen/art-design-pro) 的设计风格改造，`ui/` 为独立工程。

## 功能模块

### system｜系统管理

| 功能 | 说明 |
| --- | --- |
| 登录认证 | 图形验证码、登录、刷新令牌、登出、当前用户信息与权限集合 |
| 用户管理 | 增删改查、启停、重置密码、分配角色 |
| 角色管理 | 增删改查、授权菜单（树形勾选，自动带上父级目录） |
| 菜单管理 | 目录 / 菜单 / 按钮三级，按钮节点即权限点 |
| 字典管理 | 字典类型与字典数据 |
| 日志 | 登录日志、操作日志（写操作自动留痕） |

### infra｜基础设施

| 功能 | 说明 |
| --- | --- |
| 参数配置 | 运行期可改的键值配置 |
| 文件管理 | 上传记录、下载、删除；`local` / `s3` 双驱动 |
| 文件配置 | 存储驱动与凭据维护，主配置唯一 |
| API 日志 | 访问日志、错误日志（全局中间件采集），错误日志可标记已处理 |
| 短信 | 短信渠道（调试日志 / 阿里云 / 腾讯云）与短信模板，会员登录验证码走这里 |

### member｜会员中心

管理端访问「任意会员」，C 端只访问「会员本人」：

| 端 | 能力 |
| --- | --- |
| 管理端 `/api/v1/member` | 会员档案增删改查、启停、积分调整、等级调整、积分 / 等级 / 签到 / 地址 / 三方绑定只读；等级、标签、分组、签到档位、社交客户端配置 |
| C 端 `/api/v1/client/member` | 手机号验证码登录（自动注册）、微信三方登录与绑定、令牌续期、我的资料、收货地址、每日签到、积分 / 等级 / 签到记录 |

### 定时任务

内置一条调度（`bootstrap/schedule.go`）：每天 `03:10` 清理 7 天前的 API 访问日志，避免日志表无限膨胀。保留天数见 `config/infra.go`，可用 `INFRA_API_LOG_RETENTION_DAYS` 覆盖，`<= 0` 表示永久保留。

```bash
go run . artisan schedule:list    # 查看已注册的定时任务
```

## 架构概览

```
app/
  facades/            # 框架门面（DB / Schema / Cache / Storage / Config / Schedule ...）
  http/               # 中间件与异常处理装配
  modules/
    base/             # 认证、统一响应、错误码、参数校验、枚举、类型、TokenAuth/MemberAuth
    system/           # 系统模块
    infra/            # 基础设施模块
    member/           # 会员模块
      route.go        # 聚合入口：admin.Register + client.Register
      admin/          # 管理端：controller / request
      client/         # C 端：controller / request
      service/        # 业务规则、事务编排（两端共用）
      repository/     # 数据访问
      model/          # 表模型
      errs/           # 模块错误码
bootstrap/            # 应用装配、迁移注册表、定时任务
config/               # 配置
database/migrations/  # 建表 + 种子数据
docs/                 # 设计文档与 OpenAPI
routes/               # 路由入口
ui/                   # 前端工程
```

分层约定：`controller` 只做取参 / 校验 / 调 service / 返回统一响应；`service` 负责业务规则与事务；`repository` 只管数据访问；`request` 只声明入参与中文提示。

## 快速开始

### 环境要求

- Go 1.25+
- MySQL 8.0+
- Redis 6+
- Node.js ≥ 20.19、pnpm ≥ 8.8

### 1. 配置

```bash
cp .env.example .env            # Windows：Copy-Item .env.example .env
```

修改 `.env` 中的 `DB_*` 与 `REDIS_*`，然后生成应用密钥：

```bash
go run . artisan key:generate
```

### 2. 初始化数据库

先建库（默认库名 `go_starter`，可在 `.env` 的 `DB_DATABASE` 调整），再执行迁移：

```bash
go run . artisan migrate
go run . artisan migrate:status   # 确认迁移全部执行
```

迁移会建表并写入初始数据：管理员账号、超管角色、完整菜单树、默认会员等级、默认短信渠道与登录短信模板。

### 3. 启动后端

```bash
go run .
# http://127.0.0.1:11111
```

端口由 `.env` 的 `APP_PORT` 控制。

### 4. 启动前端

```bash
cd ui
pnpm install
pnpm dev
# http://127.0.0.1:22222
```

开发环境由 Vite 把 `/api` 代理到 `VITE_API_PROXY_URL`（默认 `http://127.0.0.1:11111`），无需处理跨域。

## 默认账号

| 账号 | 密码 | 说明 |
| --- | --- | --- |
| `admin` | `admin123` | 初始管理员，**上线前必须修改密码** |

内置超管角色标识为 `SUPER_ADMIN`，命中后跳过权限点校验。

## 环境变量

关键项（完整列表见 `.env.example`）：

| 变量 | 默认值 | 说明 |
| --- | --- | --- |
| `APP_ENV` | `local` | 运行环境，生产建议 `production` |
| `APP_DEBUG` | `true` | 调试开关，生产必须关闭 |
| `APP_KEY` | 空 | 应用密钥，`artisan key:generate` 生成 |
| `APP_HOST` / `APP_PORT` | `127.0.0.1` / `11111` | 后端监听地址与端口 |
| `DB_CONNECTION` / `DB_HOST` / `DB_PORT` / `DB_DATABASE` | `mysql` / `127.0.0.1` / `3306` / `go_starter` | 数据库连接 |
| `REDIS_HOST` / `REDIS_PORT` / `REDIS_DB` | `127.0.0.1` / `6379` / `0` | Redis 连接（令牌、验证码、限流计数） |
| `CACHE_DRIVER` | `redis` | 缓存驱动，令牌依赖 Redis |
| `ACCESS_TOKEN_TTL` | `1800` | 访问令牌有效期（秒） |
| `REFRESH_TOKEN_TTL` | `604800` | 刷新令牌有效期（秒） |
| `TOKEN_PREFIX` | `auth` | 令牌在 Redis 中的键前缀 |
| `INFRA_API_LOG_RETENTION_DAYS` | `7` | API 访问日志保留天数 |

短信渠道与模板、第三方登录凭据都配置在**数据库**（后台页面维护），不走环境变量。

前端配置在 `ui/.env`：`VITE_PORT`（开发端口）、`VITE_ACCESS_MODE`（`backend` 表示菜单由后端下发）、`VITE_API_PROXY_URL`（开发代理目标）。

## 接口约定

统一响应：

```json
{ "code": 0, "msg": "", "data": {} }
```

- HTTP 状态码恒为 200，成功与否只看 `code`；`code = 0` 表示成功。
- 失败时省略 `data`：`{ "code": 1003001001, "msg": "会员【1】不存在" }`。
- 错误码规则 `1_{模块}_{子模块}_{序号}`，如 `1_003_001_002`；通用错误使用 HTTP 语义码（`400` / `401` / `403` / `404` / `500`）。
- 分页统一 `{ "list": [], "total": 0 }`，入参 `pageNo`（默认 1）、`pageSize`（默认 10，范围 1~200）。

鉴权与权限：

| 端 | 前缀 | 鉴权 |
| --- | --- | --- |
| 管理端 | `/api/v1/{module}` | `Authorization: Bearer <accessToken>` + 权限点（如 `system:user:create`） |
| C 端 | `/api/v1/client/{module}` | `Authorization: Bearer <memberAccessToken>`，与后台令牌作用域隔离 |

核心接口（管理端）：

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| GET | `/api/v1/system/auth/captcha` | 图形验证码 |
| POST | `/api/v1/system/auth/login` | 登录 |
| POST | `/api/v1/system/auth/refresh` | 刷新令牌 |
| POST | `/api/v1/system/auth/logout` | 登出 |
| GET | `/api/v1/system/profile` | 当前用户、角色与权限 |
| GET | `/api/v1/system/menus/routes` | 当前用户的动态路由菜单 |
| GET/POST | `/api/v1/system/users` | 用户列表 / 新增 |
| GET/PUT/DELETE | `/api/v1/system/users/{id}` | 用户详情 / 编辑 / 删除 |
| GET/PUT | `/api/v1/system/users/{id}/roles` | 查询 / 分配角色 |
| GET/POST | `/api/v1/system/roles` | 角色列表 / 新增 |
| GET/PUT | `/api/v1/system/roles/{id}/menus` | 查询 / 授权菜单 |
| GET/POST | `/api/v1/system/menus` | 菜单树 / 新增 |

会员 C 端 24 个接口的完整定义见 [`docs/openapi/member.yaml`](docs/openapi/member.yaml)，可直接导入 Swagger UI / Apifox / Postman。

## 文档

| 文档 | 内容 |
| --- | --- |
| [`docs/framework-design.md`](docs/framework-design.md) | 项目设计说明：分层约定、权限模型、错误码、编码规范、开发模板 |
| [`docs/openapi/member.yaml`](docs/openapi/member.yaml) | 会员 C 端接口 OpenAPI 3.0 文档 |
| [`AGENTS.md`](AGENTS.md) | 开发指南与踩坑记录（新增接口、迁移、菜单种子、文件存储等） |
| [`docs/framework/`](docs/framework) | 框架使用参考 |

## 部署

### 二进制部署

```bash
# 后端
go build -o eqray-go-starter .
APP_ENV=production APP_DEBUG=false ./eqray-go-starter artisan migrate

# 前端
cd ui && pnpm install && pnpm build   # 产物在 ui/dist
```

上线检查：

1. 修改 `admin` 默认密码，删除或禁用测试账号。
2. `APP_DEBUG=false`，并确认 `.env` 不会被提交到仓库。
3. Nginx 托管 `ui/dist`，把 `/api` 反向代理到后端端口；`ui/.env.production` 中 `VITE_API_URL` 保持 `/` 即可同域部署。
4. 给 `storage/` 目录写权限（本地存储驱动下上传文件落在 `public/storage`）。

### Docker

仓库自带 `Dockerfile` 与 `docker-compose.yml`，是脚手架默认模板（端口 `3000`、构建时复制 `.env`）。使用前请按实际环境调整：

- 端口改成与 `APP_PORT` 一致；
- `.env` 通过挂载或环境变量注入，不要打进镜像；
- 前端建议单独构建后用 Nginx 托管。

## 开发约定

新增一个接口的标准路径：

1. `request/` 定义入参 + `Rules()` + `Messages()`（中文提示必须补全）。
2. `service/` 写业务逻辑，错误统一 `baseErrs.New(errCode, args...)`。
3. `repository/` 补数据访问。
4. `controller/` 绑定 → 调 service → `result.Success / Fail / Page`。
5. 路由挂到 `admin/route.go`（带权限点）或 `client/route.go`（带 `MemberAuth`）。
6. 新增错误码到模块 `errs/`。
7. 建表 / 种子数据写迁移，并**在 `bootstrap/migrations.go` 注册**。
8. 新增页面要补菜单与权限点种子迁移。

提交前至少跑通：

```bash
go build ./... && go vet ./...
cd ui && npx vue-tsc --noEmit && npx eslint "src/views/xxx" --fix
```

## 常见问题

**迁移执行了但表 / 菜单没变化**

迁移必须在 `bootstrap/migrations.go` 注册，否则 `artisan migrate` 会提示 `Nothing to migrate`。

**菜单点开是白屏**

`system_menu.component` 要写成 `/system/user` 这种相对 `ui/src/views` 的路径，路径错了前端加载不到组件。

**接口报 `Unknown column`**

迁移没跑全。执行 `go run . artisan migrate:status` 确认，缺的补跑；历史环境可执行新迁移对齐表结构。

**短信验证码发不出去**

检查「基础设施 → 短信模板」是否存在 `code = member_sms_login` 的模板且已启用，并绑定到启用的渠道（本地默认 `DebugLog` 渠道会把验证码打到日志）。

**改了权限 / 菜单没立即生效**

权限缓存有版本号，改动后可在后台重新登录，或执行对应迁移里的 `rbac:version` 刷新逻辑；源码中相关缓存键为 `rbac:version`。

## License

[MIT](LICENSE)
