/**
 * API 接口类型总入口（仅索引用，类型声明都在同目录的模块文件里）
 *
 * 类型按业务模块拆分到独立文件，全部声明在全局命名空间 Api 下，
 * TypeScript 会把同名 namespace 自动合并，业务代码无需 import 即可直接使用：
 *
 * const params: Api.Auth.LoginParams = { userName: 'admin', password: '123456' }
 * const channels: Api.Pay.ChannelItem[] = await fetchGetPayChannelList()
 *
 * 约定：
 * - 一个后端模块对应一个文件，新增模块时新建 模块名.d.ts 并在此登记；
 * - 文件内统一写成 declare namespace Api 包裹模块子命名空间；
 * - 类型文件里不要写 import / export，否则文件会变成模块，全局命名空间合并不再生效；
 * - 在 .vue 文件里使用需在 eslint.config.mjs 配置 globals: { Api: 'readonly' }。
 *
 * 文件清单（tsconfig 会加载 src 下全部 d.ts，此处仅作索引）：
 * - common.d.ts：分页参数、统一响应结构等通用类型
 * - auth.d.ts：登录、令牌、当前用户信息
 * - system.d.ts：系统管理（用户、角色、菜单、字典、日志）
 * - infra.d.ts：基础设施（文件、参数配置、API 日志、短信）
 * - member.d.ts：会员中心（档案、等级、标签、签到、社交）
 * - pay.d.ts：支付模块（渠道、订单、退款、回调日志）
 * - dashboard.d.ts：工作台
 *
 * @module types/api
 */
