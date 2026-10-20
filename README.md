# eqray-go-starter

基于 [Goravel](https://www.goravel.dev) 的后台管理脚手架：Go 后端提供 JWT 双令牌认证与 RBAC 权限，Vue 3 前端提供用户、角色、菜单管理，开箱即用。

## 技术栈

- 后端：Go 1.25、Goravel 1.18、Gin、MySQL 8、Redis
- 前端：Vue 3、Vite、TypeScript、Element Plus、Pinia（`ui/`，基于 Art Design Pro）

## 功能

- 账号密码登录 + 图形验证码，访问令牌 / 刷新令牌双令牌
- 用户、角色、菜单的增删改查
- RBAC：角色授权菜单、用户分配角色、接口级权限校验
- 内置超管角色 `SUPER_ADMIN`，拥有全部菜单与接口权限
- 后端动态路由（`VITE_ACCESS_MODE=backend`）：按当前用户权限返回菜单树
- 统一响应结构、业务错误码、参数校验、分页

## 目录结构

```
app/modules/
  base/     # 公共能力：认证、响应、错误码、参数、枚举、类型
  system/   # 系统模块：登录认证、用户、角色、菜单
bootstrap/  # 应用启动、迁移与 Provider 注册
config/     # 配置
database/   # 数据库迁移（建表 + 初始化数据）
routes/     # 路由入口
docs/       # 设计文档
ui/         # 前端工程
```

## 快速开始

### 环境要求

Go 1.25+、MySQL 8、Redis 6+、Node 20.19+、pnpm 8+

### 1. 配置

```bash
cp .env.example .env          # Windows PowerShell: Copy-Item .env.example .env
# 修改 DB_* 与 REDIS_*，然后生成 APP_KEY
go run . artisan key:generate
```

### 2. 初始化数据库

```bash
go run . artisan migrate
# 或使用仓库自带脚本：./artisan migrate（Windows: artisan.bat migrate）
```

### 3. 启动后端

```bash
go run .
# http://127.0.0.1:11111
```

### 4. 启动前端

```bash
cd ui
pnpm install
pnpm dev
# http://127.0.0.1:22222，开发环境代理到 11111
```

## 默认账号

| 账号 | 密码 | 说明 |
|---|---|---|
| `admin` | `admin123` | 初始管理员，仅用于本地初始化，上线前必须修改 |

内置超管角色标识为 `SUPER_ADMIN`，权限校验命中后直接放行。

## 主要接口

统一前缀 `/api/v1/system`，除登录、验证码、刷新、登出外均需携带访问令牌。

| 方法 | 路径 | 说明 |
|---|---|---|
| GET | `/auth/captcha` | 图形验证码 |
| POST | `/auth/login` | 登录 |
| POST | `/auth/refresh` | 刷新令牌 |
| POST | `/auth/logout` | 登出 |
| GET | `/profile` | 当前用户信息、角色与权限 |
| GET | `/menus/routes` | 当前用户的动态路由菜单 |
| GET/POST | `/users` | 用户列表 / 新增 |
| GET/PUT/DELETE | `/users/{id}` | 用户详情 / 编辑 / 删除 |
| GET/PUT | `/users/{id}/roles` | 查询 / 分配用户角色 |
| GET/POST | `/roles` | 角色列表 / 新增 |
| GET/PUT/DELETE | `/roles/{id}` | 角色详情 / 编辑 / 删除 |
| GET/PUT | `/roles/{id}/menus` | 查询 / 分配角色菜单 |
| GET/POST | `/menus` | 菜单树 / 新增 |
| GET/PUT/DELETE | `/menus/{id}` | 菜单详情 / 编辑 / 删除 |

接口权限标识形如 `system:user:create`，与 `system_menu.permission` 一一对应。

## 文档

详细设计（分层约定、权限模型、编码规范、开发模板）见 [docs/framework-design.md](docs/framework-design.md)。

## License

[MIT](LICENSE)
