# 支付模块（pay）设计方案

> 状态：**已实施（后端 + 后台页面）**，实施偏差见文末「16. 实施记录」
> 目标：基于 [gopay](https://github.com/go-pay/gopay) 实现聚合支付，支持后台可视化配置渠道，跑通「下单 → 支付 → 异步回调 → 订单状态流转 → 退款」闭环。
> 依赖版本：`github.com/go-pay/gopay v1.5.123`（`go list -m -versions` 当前最高版本）

---

## 1. 结论先行

| 决策点 | 方案 |
| --- | --- |
| 模块归属 | 新建 `app/modules/pay`，表前缀 `pay_`，与 system / infra / member 平级 |
| 路由 | 管理端 `/api/v1/pay`（令牌 + 权限点），C 端 `/api/v1/client/pay`（会员令牌），回调 `/api/v1/pay/notify/{code}`（免鉴权） |
| 渠道配置 | 数据库表 `pay_channel` 一行对应一种支付方式（如 `wx_mini`），后台增删改查 + 启停，**改完即时生效** |
| 回调地址 | 不落库、不由用户填写，统一 `APP_URL` + 固定路由生成，后台页面只读展示 + 复制 |
| 密钥存储 | `config` JSON **明文**入库（不加密），接口层只回脱敏值 |
| 业务通知 | 支付成功后派发事件 `pay.order.paid`，业务模块通过 `WithEvents` 订阅，不反向依赖支付模块 |
| 渠道抽象 | 统一 `Channel` 接口（下单 / 查单 / 关单 / 退款 / 回调解析），微信 V3 与支付宝 V3 各一个实现 |
| 金额单位 | `int64` 分，全链路不出现浮点 |
| 一期范围 | **全渠道**：微信 5 种（`wx_mini`/`wx_jsapi`/`wx_app`/`wx_native`/`wx_h5`）+ 支付宝 4 种（`ali_pc`/`ali_wap`/`ali_app`/`ali_qr`），含退款与超时关单 |
| 业务边界 | pay **只做支付**：收单、退款、通知、事件派发；不 import 任何业务模块，不写业务表 |
| 演示闭环 | 附带一张 `demo_order` 演示表 + 一个监听器，用来验证「下单 → 支付 → 事件回写」整条链路，可直接删 |
| 暂不做 | 钱包余额 / 充值到账（属业务侧）、分账（已确认不需要）、对账单下载、跨境支付、优惠与运费计算 |

---

## 2. 目标与非目标

### 2.1 目标

1. 业务侧只关心「谁、多少钱、买什么」，不需要接触微信 / 支付宝 SDK。
2. 渠道凭据全部在后台维护，新增一个商户号不用发版。
3. 支付结果可靠：验签、金额校验、幂等、重复通知都可重入。
4. 所有支付动作可追溯：订单、退款单、原始通知报文都留档。

### 2.2 非目标（本期不做）

- 商品 / 订单 / 优惠券等交易域模型（本期只提供一张 `demo_order` 演示表，真实业务由对应模块自己维护）。
- 会员余额、钱包流水、充值到账（属业务侧，pay 只负责通知「已支付」）。
- 分账（已确认不需要）、对账单自动下载、发票。
- 优惠、运费、金额拆分等计价逻辑（订单金额由业务侧算好，pay 只按 `price` 收单）。

---

## 3. 模块与目录

```
app/modules/pay/
├─ route.go                      # 聚合：admin.Register + client.Register + callback.Register
├─ admin/
│  ├─ route.go
│  ├─ controller/
│  │  ├─ channel_controller.go   # 渠道配置 CRUD / 启停 / 连通性测试
│  │  ├─ order_controller.go     # 支付订单查询、同步、关闭
│  │  ├─ refund_controller.go    # 退款申请与查询
│  │  └─ notify_log_controller.go
│  └─ request/
├─ client/
│  ├─ route.go
│  ├─ controller/
│  │  ├─ order_controller.go     # 创建支付、查询支付状态、可用渠道
│  │  └─ request/
├─ callback/                     # 免鉴权回调，独立分组
│  ├─ route.go
│  └─ controller/notify_controller.go
├─ channel/                      # 渠道适配层（不依赖 http）
│  ├─ channel.go                 # Channel 接口 + 请求/响应结构
│  ├─ registry.go                # code → 构造函数
│  ├─ wechat_v3.go
│  └─ alipay_v3.go
├─ service/
│  ├─ pay_channel_service.go     # 配置读写、client 缓存失效
│  ├─ pay_order_service.go       # 下单、查单、关单、状态机
│  ├─ pay_refund_service.go
│  └─ pay_notify_service.go      # 回调验签、幂等、落单、派发事件
├─ events/                       # 对外事件（OrderPaid / RefundSucceeded）
├─ repository/
├─ model/
└─ errs/
```

依赖方向：`channel` 只依赖模型与配置，不依赖 `service`；`service` 依赖 `channel`；`controller` 只依赖 `service`；业务模块只订阅 `events`，**不反向依赖支付模块**。

配套的演示业务模块（独立于 pay，可整体删除）：

```
app/modules/demo/
├─ route.go                      # 只挂 client 端
├─ client/controller/            # 创建演示订单、查询状态
├─ service/                      # 下单并调 PayService.CreateOrder
├─ repository/
├─ model/
└─ listener/order_paid.go        # 订阅 pay.order.paid，回写 demo_order
```

---

## 4. 数据模型

### 4.1 `pay_channel` 支付渠道

一行 = 一种支付方式（不是「一个商户」）。同一商户号要同时支持扫码和 H5，就配两行。

```sql
CREATE TABLE `pay_channel` (
  `id`          bigint unsigned NOT NULL AUTO_INCREMENT,
  `code`        varchar(32)  NOT NULL COMMENT '渠道编码，如 wx_mini / wx_native / ali_pc',
  `name`        varchar(64)  NOT NULL COMMENT '展示名，如 微信小程序支付',
  `channel`     varchar(16)  NOT NULL COMMENT '支付厂商：wx / alipay',
  `pay_method`  varchar(16)  NOT NULL COMMENT '支付方式：mini/jsapi/app/native/h5/page/wap/qr',
  `config`      json         NOT NULL COMMENT '渠道参数，明文存储（含商户私钥 / APIv3 密钥）',
  `status`      tinyint      NOT NULL DEFAULT 0 COMMENT '0 停用 / 1 启用',
  `sort`        int          NOT NULL DEFAULT 0,
  `remark`      varchar(255) DEFAULT NULL,
  `creator`     varchar(64)  DEFAULT NULL,
  `created_at`  datetime     DEFAULT NULL,
  `updater`     varchar(64)  DEFAULT NULL,
  `updated_at`  datetime     DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_code` (`code`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='支付渠道配置';
```

> 渠道属于配置数据，删除走**物理删除**（避免 `uk_code` 被软删记录占用），操作写入 `system_operate_log`。

### 4.2 `pay_order` 支付订单

```sql
CREATE TABLE `pay_order` (
  `id`                bigint unsigned NOT NULL AUTO_INCREMENT,
  `merchant_order_no` varchar(64)  NOT NULL COMMENT '支付模块生成的商户单号',
  `biz_type`          varchar(32)  NOT NULL COMMENT '业务类型，如 member_recharge',
  `biz_order_no`      varchar(64)  NOT NULL COMMENT '业务方订单号',
  `channel_id`        bigint unsigned NOT NULL,
  `channel_code`      varchar(32)  NOT NULL,
  `subject`           varchar(128) NOT NULL COMMENT '商品标题',
  `description`       varchar(255) DEFAULT NULL,
  `price`             bigint       NOT NULL COMMENT '金额（分）',
  `currency`          varchar(8)   NOT NULL DEFAULT 'CNY',
  `status`            tinyint      NOT NULL DEFAULT 0 COMMENT '0 待支付 / 10 支付中 / 20 已支付 / 30 已关闭',
  `refund_status`     tinyint      NOT NULL DEFAULT 0 COMMENT '0 未退款 / 10 部分退款 / 20 全额退款',
  `user_id`           bigint unsigned DEFAULT NULL COMMENT '下单会员 ID',
  `client_ip`         varchar(64)  DEFAULT NULL,
  `channel_order_no`  varchar(64)  DEFAULT NULL COMMENT '渠道侧单号（微信 transaction_id / 支付宝 trade_no）',
  `channel_user`      varchar(64)  DEFAULT NULL COMMENT '渠道侧付款人标识（openid / buyer_id）',
  `channel_extra`     json         DEFAULT NULL COMMENT '渠道返回的原始数据',
  `notify_url`        varchar(255) NOT NULL COMMENT '下单时实际传给渠道的通知地址，系统按 APP_URL 生成',
  `expire_time`       datetime     NOT NULL COMMENT '支付超时时间',
  `success_time`      datetime     DEFAULT NULL,
  `creator`           varchar(64)  DEFAULT NULL,
  `created_at`        datetime     DEFAULT NULL,
  `updater`           varchar(64)  DEFAULT NULL,
  `updated_at`        datetime     DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_merchant_order_no` (`merchant_order_no`),
  UNIQUE KEY `uk_biz_order` (`biz_type`, `biz_order_no`, `channel_code`),
  KEY `idx_user` (`user_id`, `created_at`),
  KEY `idx_status_expire` (`status`, `expire_time`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='支付订单';
```

`uk_biz_order` 是防重复下单的关键：同一业务单同一渠道重复请求，直接返回已有订单。

### 4.3 `pay_refund` 退款单

```sql
CREATE TABLE `pay_refund` (
  `id`                 bigint unsigned NOT NULL AUTO_INCREMENT,
  `merchant_refund_no` varchar(64) NOT NULL COMMENT '退款单号',
  `order_id`           bigint unsigned NOT NULL,
  `merchant_order_no`  varchar(64) NOT NULL,
  `channel_id`         bigint unsigned NOT NULL,
  `channel_code`       varchar(32) NOT NULL,
  `refund_price`       bigint      NOT NULL COMMENT '退款金额（分）',
  `reason`             varchar(255) NOT NULL,
  `status`             tinyint     NOT NULL DEFAULT 0 COMMENT '0 待退款 / 10 退款中 / 20 已退款 / 30 退款失败',
  `channel_refund_no`  varchar(64) DEFAULT NULL,
  `channel_extra`      json        DEFAULT NULL,
  `success_time`       datetime    DEFAULT NULL,
  `creator`            varchar(64) DEFAULT NULL,
  `created_at`         datetime    DEFAULT NULL,
  `updater`            varchar(64) DEFAULT NULL,
  `updated_at`         datetime    DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_merchant_refund_no` (`merchant_refund_no`),
  KEY `idx_order` (`order_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='支付退款单';
```

### 4.4 `pay_notify_log` 回调日志

```sql
CREATE TABLE `pay_notify_log` (
  `id`                bigint unsigned NOT NULL AUTO_INCREMENT,
  `channel_code`      varchar(32)  NOT NULL,
  `type`              varchar(16)  NOT NULL COMMENT 'pay / refund',
  `merchant_order_no` varchar(64)  DEFAULT NULL,
  `request_headers`   text         DEFAULT NULL COMMENT '关键头信息，脱敏后存储',
  `request_body`      text         DEFAULT NULL COMMENT '原始报文（微信解密前密文 / 支付宝表单）',
  `decrypted_body`    text         DEFAULT NULL COMMENT '解密后的业务报文',
  `verify_status`     tinyint      NOT NULL DEFAULT 0 COMMENT '签名校验：0 失败 / 1 通过',
  `handle_status`     tinyint      NOT NULL DEFAULT 0 COMMENT '业务处理：0 失败 / 1 成功',
  `error_msg`         varchar(500) DEFAULT NULL,
  `created_at`        datetime     DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_order` (`merchant_order_no`),
  KEY `idx_created` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='支付回调日志';
```

保留 30 天，复用现有 `bootstrap/schedule.go` 增加一条清理任务。

### 4.5 `demo_order` 演示业务订单（可删）

pay 模块不碰业务表，但需要一条真实链路验证「支付成功后业务侧能收到通知」。这张表只服务于联调，上线前可直接删。

```sql
CREATE TABLE `demo_order` (
  `id`           bigint unsigned NOT NULL AUTO_INCREMENT,
  `order_no`     varchar(64)  NOT NULL COMMENT '演示业务单号',
  `user_id`      bigint unsigned NOT NULL COMMENT '下单会员 ID',
  `subject`      varchar(128) NOT NULL COMMENT '商品标题',
  `price`        bigint       NOT NULL COMMENT '金额（分）',
  `status`       tinyint      NOT NULL DEFAULT 0 COMMENT '0 待支付 / 10 已支付 / 20 已取消',
  `pay_order_no` varchar(64)  DEFAULT NULL COMMENT '支付单号',
  `paid_at`      datetime     DEFAULT NULL,
  `creator`      varchar(64)  DEFAULT NULL,
  `created_at`   datetime     DEFAULT NULL,
  `updater`      varchar(64)  DEFAULT NULL,
  `updated_at`   datetime     DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_order_no` (`order_no`),
  KEY `idx_user` (`user_id`, `created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='演示业务订单（仅用于联调，上线可删）';
```

demo 模块只做三件事：

1. C 端下单 `POST /api/v1/client/demo/orders`：金额取服务端写死的测试商品（如 `1` 分），**不信任前端传值**；
2. 调 `PayService.CreateOrder` 拿支付参数返回前端；
3. 监听 `pay.order.paid`，按 `bizOrderNo` 回写 `status=10` + `paid_at`（`UPDATE ... WHERE status=0`，天然幂等）。

这样既验证了完整链路，又不会让 pay 反向依赖业务代码。

---

## 5. 渠道配置（「能配置渠道」怎么落地）

### 5.1 支持的渠道编码（一期全部实现）

9 种渠道一次做完，不做「先上几个」的裁剪。

| code | 厂商 | 支付方式 | 典型场景 |
| --- | --- | --- | --- |
| `wx_native` | 微信 V3 | 扫码 | PC 端展示二维码 |
| `wx_h5` | 微信 V3 | H5 | 站外手机浏览器 |
| `wx_jsapi` | 微信 V3 | JSAPI | 公众号内网页 |
| `wx_mini` | 微信 V3 | 小程序 | 小程序（需要 openid） |
| `wx_app` | 微信 V3 | APP | 原生 App |
| `ali_pc` | 支付宝 V3 | 电脑网站 | PC 跳转收银台 |
| `ali_wap` | 支付宝 V3 | 手机网站 | 手机浏览器跳转 |
| `ali_app` | 支付宝 V3 | APP | 原生 App |
| `ali_qr` | 支付宝 V3 | 当面付扫码 | PC 扫码 |

### 5.2 `config` 字段结构

**微信 V3**

```json
{
  "appId": "wx1234567890",
  "mchId": "1900000001",
  "apiV3Key": "32位APIv3密钥",
  "serialNo": "商户证书序列号",
  "privateKey": "-----BEGIN PRIVATE KEY-----\n...",
  "publicKeyId": "PUB_KEY_ID_xxx",
  "publicKey": "-----BEGIN PUBLIC KEY-----\n...",
  "subAppId": "",
  "subMchId": ""
}
```

- 证书模式与「微信支付公钥」模式二选一：填了 `publicKey` 走公钥模式（`AutoVerifySignByPublicKey`），否则走平台证书自动下载（`AutoVerifySign`）。
- 服务商模式预留 `subAppId` / `subMchId`，一期不实现。

**支付宝 V3**

```json
{
  "appId": "2021000000000000",
  "privateKey": "-----BEGIN PRIVATE KEY-----\n...",
  "alipayPublicKey": "支付宝公钥",
  "appCertContent": "",
  "alipayRootCertContent": "",
  "alipayPublicCertContent": "",
  "isProd": true
}
```

- 公钥模式只填 `alipayPublicKey`；证书模式填三个证书内容，走 `SetCert`。

### 5.2.1 渠道字段清单（动态表单的数据来源）

字段定义由**后端**维护，前端不硬编码 9 套映射。`GET /api/v1/pay/channels/schema?code=wx_mini` 返回：

```json
{
  "code": "wx_mini",
  "fields": [
    { "key": "appId",       "label": "小程序 AppID",   "type": "input",    "required": true,  "secret": false },
    { "key": "mchId",       "label": "商户号",         "type": "input",    "required": true,  "secret": false },
    { "key": "apiV3Key",    "label": "APIv3 密钥",     "type": "password", "required": true,  "secret": true },
    { "key": "serialNo",    "label": "证书序列号",     "type": "input",    "required": true,  "secret": false },
    { "key": "privateKey",  "label": "商户私钥",       "type": "textarea", "required": true,  "secret": true, "rows": 6 },
    { "key": "publicKeyId", "label": "公钥 ID",        "type": "input",    "required": false, "secret": false },
    { "key": "publicKey",   "label": "微信支付公钥",   "type": "textarea", "required": false, "secret": true, "rows": 6 }
  ]
}
```

字段清单（`type` 决定前端渲染成 `ElInput` / `ElInput type=password` / `ElInput type=textarea` / `ElSwitch`）：

| 渠道前缀 | 字段 |
| --- | --- |
| `wx_*` | `appId`（必填）、`mchId`（必填）、`apiV3Key`（必填，密钥）、`serialNo`（必填）、`privateKey`（必填，密钥）、`merchantCertContent`（选填，密钥，用于校验序列号与证书配套）、`publicKeyId` / `publicKey`（公钥模式必填，密钥） |
| `ali_*` | `appId`（必填）、`privateKey`（必填，密钥）、`alipayPublicKey`（公钥模式必填，密钥）、`appCertContent` / `alipayRootCertContent` / `alipayPublicCertContent`（证书模式必填）、`isProd`（必填，开关） |

- `secret: true` 的字段在列表 / 详情只回 `wx12****90`，编辑留空表示不修改；
- 新增 / 编辑时后端按 code 白名单校验必填项，缺项返回 `1_004_001_003`；
- 好处：以后加渠道只改后端 schema 常量，前端零改动。

### 5.3 后台操作流

1. 「支付 → 支付渠道 → 新增」，选择渠道编码 → 调 `GET /channels/schema?code=xxx` 拿字段定义，前端用 `ElForm` 动态渲染（input / password / textarea / switch）。
   抽屉顶部固定展示**系统生成的异步通知地址**（见 5.5），只读 + 一键复制，引导用户去渠道后台核对。
2. 保存时**原样明文入库**，`config` 字段直接存 JSON，不做加密、不做编码转换。
3. 列表与详情接口只返回脱敏值（如 `wx12****90`），编辑时留空表示不修改（与现有 `member_social_client`、`infra_sms_channel` 一致）。这是**响应层脱敏**，库里仍是明文。
4. 启用后写入缓存版本号 `pay:channel:version`，运行中的 client 缓存立即失效。
5. 「测试连通性」按钮：用该配置发起一次查单（单号传一个不存在的值），能区分「凭据错误」与「订单不存在」。

### 5.4 存储方式的取舍（已确认：不加密）

`pay_channel.config` 明文存储，换来的是排查简单：直接查库就能确认配置、迁移环境不用处理密钥、换 `APP_KEY` 不受影响。

需要接受的代价与对应补偿：

| 风险 | 补偿措施（实现时必须做） |
| --- | --- |
| 有 DB 读权限的人可直接拿到商户私钥 / APIv3 密钥 | 业务库账号最小权限，不授予导出 / 备份权限；开发环境用沙箱应用 |
| 备份文件泄露即等于密钥泄露 | 备份文件加密存放，不与数据库同机 |
| 密钥出现在前端 / 代理日志 | 接口响应脱敏；前端不缓存详情表单到 localStorage |
| 密钥被打进日志 | `pay_notify_log`、应用日志、异常堆栈中禁止打印 `config` |
| 误提交到仓库 | `.env` 与 SQL 备份加入 `.gitignore`（已有基础，需补备份目录） |

> 如果后续要改成加密存储，只需在 `PayChannelService` 的读写两处加一层编解码，表结构不用动。

### 5.5 回调地址：系统生成 + 页面展示（已确认）

回调地址**不落库、不由用户填写**，统一由 `APP_URL` + 固定路由拼出来：

| 用途 | 地址 |
| --- | --- |
| 支付结果通知 | `{APP_URL}/api/v1/pay/notify/{channelCode}` |

> 退款没有回调路由，状态由定时任务查单收敛（见 7.3 / 7.4）。

- 例：`APP_URL=https://api.example.com`、渠道 `wx_native` → `https://api.example.com/api/v1/pay/notify/wx_native`。
- 下单时把该地址写进渠道参数（微信 `notify_url`、支付宝 `notify_url`），同时写入 `pay_order.notify_url` 留档，事后能查「当时通知发到哪」。
- `APP_URL` 必须是**公网可访问的 HTTPS 域名**（微信 V3 要求 HTTPS，支付宝不接受 localhost / 内网地址）。
- `APP_URL` 配错 → 收不到异步通知，只能靠「主动查单 + 超时关单任务」兜底，所以后台必须有醒目提示。

**后台页面怎么展示**

渠道新增 / 编辑抽屉顶部固定一条只读信息（引导用户复制去渠道后台核对）：

```
异步通知地址（由系统按 APP_URL 生成，无需填写）
支付：https://api.example.com/api/v1/pay/notify/wx_native   [复制]
```

下方跟一句提示文案：

> 该地址由系统按 `APP_URL` 自动生成，本页面无需填写。
> 若显示的域名不是你的线上域名，请先修改 `.env` 中的 `APP_URL` 并重启服务。
> 微信 / 支付宝的异步通知地址在下单时通过参数传递，一般无需在商户平台单独配置；
> 但请确保该域名已备案、可公网访问，且与商户平台登记的授权域名一致。

**地址来源接口**

`GET /api/v1/pay/channels/notify-urls?code=wx_native` 返回地址与当前 `APP_URL`，
前端直接渲染，避免前后端各拼一遍：

```json
{
  "appUrl": "https://api.example.com",
  "payNotifyUrl": "https://api.example.com/api/v1/pay/notify/wx_native"
}
```

- `APP_URL` 为空时返回错误码 `1_004_001_004`，前端提示「请先在 .env 配置 APP_URL」；
- 前端不缓存该结果，避免用户改了 `.env` 仍看到旧地址；
- 路由注册时**必须把 `/channels/notify-urls` 放在 `/channels/{id}` 之前**，否则会被路径参数吞掉（项目已有同类约定）。

**同步跳转地址（`return_url` / `redirect_url`）另说**

它解决的是「支付完成后回到业务页面」，与异步通知无关，同样不在渠道表里配：

- C 端创建支付时可传 `returnUrl`，服务端校验必须与 `APP_URL` 同域（或命中配置的域名白名单），再塞进渠道参数；
- 不传则由前端在支付结果回调里自行跳转 —— 扫码、小程序、APP 场景本来就不需要。

---

## 6. 渠道抽象层

### 6.1 统一接口

```go
// Channel 统一支付渠道能力，屏蔽微信 / 支付宝差异。
type Channel interface {
	// Code 返回渠道编码，如 wx_native
	Code() string
	// CreateOrder 向渠道下单，返回前端拉起支付所需的参数
	CreateOrder(ctx context.Context, req CreateOrderRequest) (CreateOrderResult, error)
	// QueryOrder 查询渠道侧订单状态
	QueryOrder(ctx context.Context, merchantOrderNo string) (QueryOrderResult, error)
	// CloseOrder 关闭渠道侧订单
	CloseOrder(ctx context.Context, merchantOrderNo string) error
	// Refund 发起退款
	Refund(ctx context.Context, req RefundRequest) (RefundResult, error)
	// ParseNotify 解析并验签异步通知
	ParseNotify(ctx context.Context, r *http.Request) (NotifyResult, error)
	// NotifyAck 返回该渠道要求的成功应答
	NotifyAck() NotifyAck
}
```

```go
// CreateOrderResult 是给前端拉起支付的统一结果
type CreateOrderResult struct {
	Kind string            // qr / redirect / jsapi / app / form
	URL  string            // 二维码内容 / 跳转地址
	Params map[string]any  // JSAPI / APP 拉起参数
}
```

`Kind` 让前端只写一套处理逻辑：扫码渲染二维码、`redirect` 直接跳转、`jsapi` 调 `wx.chooseWXPay`、`app` 交原生壳。

### 6.2 gopay 方法映射（按 v1.5.123 源码）

| 能力 | 微信 V3 | 支付宝 V3 |
| --- | --- | --- |
| 创建 client | `wechat.NewClientV3(mchId, serialNo, apiV3Key, privateKey)` | `alipay.NewClientV3(appId, privateKey, isProd)` |
| 验签前置 | `AutoVerifySign()` / `AutoVerifySignByPublicKey(pub, id)` | `SetCert(app, root, aliPub)` 或公钥模式 |
| 扫码下单 | `V3TransactionNative` → `NativeRsp.CodeUrl` | `TradePrecreate` → `qr_code` |
| H5 / WAP | `V3TransactionH5` → `H5Rsp.H5Url` | `TradeWapPay` → `payUrl` |
| 电脑网站 | — | `TradePagePay` → `payUrl` |
| JSAPI | `V3TransactionJsapi` → `PrepayId` → `PaySignOfJSAPI(appid, prepayId)` | `TradeCreate`（`buyer_id`） |
| 小程序 | `V3TransactionJsapi` → `PaySignOfApplet(appid, prepayId)` | — |
| APP | `V3TransactionApp` → `PaySignOfApp(appid, prepayId)` | `TradeAppPay` → `orderStr` |
| 查单 | `V3TransactionQueryOrder(ctx, OutTradeNo, no)` | `TradeQuery` |
| 关单 | `V3TransactionCloseOrder(ctx, no)` | `TradeClose` |
| 退款 | `V3Refund` | `TradeRefund` |
| 退款查询 | `V3RefundQuery` | `TradeFastPayRefundQuery` |
| 回调解析 | `wechat.V3ParseNotify(req)` | `alipay.ParseNotifyToBodyMap(req)`（取自 `gopay/alipay` 根包） |
| 回调验签 | `notify.VerifySignByPK(client.WxPublicKey())` | `alipay.VerifySign(alipayPublicKey, bm)` |
| 回调解密 | `notify.DecryptPayCipherText(apiV3Key)` | 无（回调为明文表单） |
| 成功应答 | `{"code":"SUCCESS","message":"成功"}` | 纯文本 `success` |

> 支付宝 V3 客户端的回调工具函数在 `gopay/alipay`（v2 包）中，实现时同时引入两个包：v3 client 负责下单 / 退款，v2 仅用于回调解析与验签。若后续版本变化，用 `DoAliPayAPISelfV3` 兜底。

### 6.3 client 缓存

- 进程内 `sync.Map`，key 为 `channelId + configHash`；
- 渠道更新时只递增缓存版本号（`facades.Cache().Put("pay:channel:version", ...)`），下次取用时比对版本重建；
- 避免每次下单都解析私钥（RSA 解析成本不低）。

### 6.4 小程序 openid 从哪来

`wx_mini` 下单必须带 `payer.openid`。本项目已有 `member_social_user`（`type=wechat_mini` + `openid`）与 `member_social_binding`，直接按当前会员查：

```sql
SELECT su.openid FROM member_social_user su
JOIN member_social_binding sb ON sb.social_user_id = su.id
WHERE sb.user_id = ? AND su.type = 'wechat_mini' LIMIT 1
```

查不到则返回业务错误码「请先授权微信登录」，引导前端走 `/client/member/auth/social-bind-login`。

---

## 7. 核心流程

### 7.1 下单

```mermaid
sequenceDiagram
    participant B as 业务模块(如充值)
    participant P as PayService
    participant D as DB
    participant C as 渠道(微信/支付宝)
    participant F as 前端

    B->>P: CreateOrder(bizType, bizOrderNo, price, channelCode, userId)
    P->>D: 按 uk_biz_order 查是否已存在
    alt 已存在且未支付
        P-->>B: 复用原订单
    else 不存在
        P->>D: 插入 pay_order(status=0, expire_time=+30min)
    end
    P->>C: CreateOrder
    C-->>P: prepay_id / code_url / payUrl
    P->>D: status=10，写 channel_extra
    P-->>F: CreateOrderResult(kind, url, params)
```

要点：

- 金额只由**业务模块**传入，C 端接口不接受金额字段；
- `uk_biz_order` 兜底并发重复下单，撞键后回查；
- 下单失败订单保留在 `status=0`，允许重试（重试沿用同一 `merchant_order_no`）。

### 7.2 回调

```mermaid
sequenceDiagram
    participant W as 微信/支付宝
    participant N as NotifyController
    participant S as PayNotifyService
    participant D as DB
    participant E as 事件 / 业务监听器

    W->>N: POST /api/v1/pay/notify/{code}
    N->>D: 写 pay_notify_log(原始报文)
    N->>S: ParseNotify(验签+解密)
    S->>S: 校验 appId/mchId、金额、币种
    S->>D: CAS 更新 pay_order(status 10→20)
    alt 更新成功（首次）
        S->>E: 事务提交后派发 pay.order.paid
    else 已支付
        S->>S: 视为重复通知，直接成功
    end
    N-->>W: 渠道要求的成功应答
```

必须处理的六件事：

1. **验签**失败 → 返回非成功应答并记日志，不改进订单；
2. **金额校验**：回调金额与订单 `price` 不一致 → 标记异常，人工介入；
3. **幂等**：用 `UPDATE ... WHERE status = 10` 的影响行数判断是否首次处理；
4. **顺序无关**：先回调后查单、重复通知、乱序通知都要能收敛到正确状态；
5. **事件监听失败不回滚支付状态**：订单已是「已支付」事实，业务侧失败由补偿任务 / 人工处理，日志留痕；
6. **应答格式**：微信要 JSON、支付宝要纯文本 `success`，返回错会触发渠道重试（重试策略各不相同）。

### 7.3 定时任务：超时关单 + 退款状态同步

两条都放 `bootstrap/schedule.go`，各自 `OnOneServer()`。

**1）超时关单**（每 5 分钟，`pay:close-expired-orders`）

```
扫描 status IN (0,10) AND expire_time < NOW() LIMIT 200
  → 调渠道 CloseOrder（失败也继续）
  → status = 30
```

**2）退款状态同步**（每 5 分钟，`pay:sync-refunding`）

```
扫描 status = 10（退款中）AND created_at < NOW() - 1 分钟 LIMIT 200
  → 调渠道退款查询
  → 成功 → status = 20，并回写 pay_order.refund_status
  → 失败 → status = 30，记 error_msg
```

### 7.4 退款（无审核、不收回调）

- 后台在订单详情点「退款」，填写金额与原因 → 创建 `pay_refund`（`status=0`）→ **不做人工审核，提交后直接调渠道**；
- 调渠道退款接口：
  - 支付宝 `TradeRefund` **同步返回结果** → 直接 `status=20`；
  - 微信 `V3Refund` 只返回「受理中」→ `status=10`；
- `status=10` 的单由 7.3 的「退款状态同步」收敛 —— **不单独开退款回调路由**，少一套验签、解析与应答分支；
- 同步成功后回写 `pay_order.refund_status`（部分 / 全额）；
- 退款金额累加不得超过订单金额（事务内校验）。

### 7.5 支付完成后怎么「回到应用」

| 端 | 拉起支付 | 回到应用的方式 | 可否作为发货依据 |
| --- | --- | --- | --- |
| 小程序 | `wx.requestPayment(params)` | 直接 `success/fail` 回调，无跳转 | 否 |
| 公众号 JSAPI | `WeixinJSBridge.invoke('getBrandWCPayRequest')` | 页面内回调 | 否 |
| H5 / 手机浏览器 | `V3TransactionH5` → `h5_url` | 微信 H5 用 `redirect_url`，支付宝 WAP 用 `return_url` 跳回业务页 | 否 |
| PC 扫码 | `V3TransactionNative` → `code_url` | 前端轮询订单状态后自行跳转 | 否 |
| PC 支付宝 | `TradePagePay` → `payUrl` | 收银台按 `return_url` 同步跳回 | 否 |
| 原生 App | 见下 | URL Scheme / SDK 回调 | 否 |

**微信 APP 支付回跳**

- 服务端 `V3TransactionApp` 拿 `prepay_id` → `PaySignOfApp(appid, prepayId)` 生成调起参数，客户端交 `IWXAPI.sendReq`；
- 回 App 靠**客户端配置**，不在服务端参数里：
  - Android：包名下的 `wxapi/WXPayEntryActivity` 接收回执，用户点「返回商家」时微信按**包名 + 签名**拉起；
  - iOS：`URL Types` 配置 App scheme、`LSApplicationQueriesSchemes` 加 `weixin`，在 `handleOpenURL` 接收；
- 微信 V3 App 下单**没有 `return_url` 参数**。

**支付宝 APP 支付回跳**

- 服务端 `TradeAppPay` 返回 `orderStr`，客户端 `PayTask.payV2(orderStr, true)`；
- iOS 靠 `orderStr` 里的 `fromAppUrlScheme` 让支付宝跳回，Android 靠 SDK 的 `PayResult` 回调。

> 客户端回跳只用于「支付中 → 展示结果」的交互。**订单状态一律以 `notify` 或主动查单为准**，客户端返回可伪造、可丢失、可重复。

### 7.6 支付结果怎么「回到业务」：事件（已确认）

渠道通知到达、订单状态 CAS 成功之后，支付模块派发事件，业务模块自行订阅，**支付模块不 import 任何业务模块**。

**事件定义**（`app/modules/pay/events/order_paid.go`）

```go
// OrderPaid 支付成功的领域事件，由通知回调或主动查单在状态流转成功后派发。
// 业务模块只消费该事件，不与支付模块产生代码依赖。
type OrderPaid struct {
	BizType     string // 业务类型，如 member_recharge
	BizOrderNo  string // 业务单号
	OrderNo     string // 支付单号
	ChannelCode string // 渠道编码
	Price       int64  // 实付金额（分）
	PaidAt      string // 支付完成时间
}

// Signature 返回事件唯一标识，供框架识别监听器映射。
func (o *OrderPaid) Signature() string {
	return "pay.order.paid"
}
```

**派发点**（`pay_notify_service.go`，事务提交之后）

```go
// 只在「首次把订单从支付中改为已支付」时派发，重复通知不会重复触发
if affected == 1 {
	facades.Event().Dispatch(&events.OrderPaid{
		BizType:     order.BizType,
		BizOrderNo:  order.BizOrderNo,
		OrderNo:     order.MerchantOrderNo,
		ChannelCode: order.ChannelCode,
		Price:       order.Price,
		PaidAt:      carbon.Now().ToDateTimeString(),
	})
}
```

**注册**（`bootstrap/app.go` 追加，事件映射放 `bootstrap/events.go`）

```go
WithEvents(Events).

// bootstrap/events.go
func Events() map[event.Event][]event.Listener {
	return map[event.Event][]event.Listener{
		&payEvents.OrderPaid{}: {
			&listeners.RechargePaidListener{},
		},
	}
}
```

约定：

1. **监听器必须幂等**。同一 `bizOrderNo` 重复投递只生效一次（用业务表唯一键或状态机兜底）。
2. **监听器不要抛 panic**。Goravel 的事件是同步调用，异常会影响回调应答，渠道会重试；监听器内部自己捕获并记日志。
3. **重活异步化**。如果业务处理耗时（发短信、写多表），监听器里只落一条待处理记录，由定时任务消费。
4. 事件在**事务提交后**派发，避免业务侧读到未提交的订单状态。

> 若后续业务模块希望完全解耦（连事件包都不想引用），可以再引入「支付成功后写 outbox 表 + 业务侧定时拉取」，本期不做。

---

## 8. 接口清单

### 8.1 管理端 `/api/v1/pay`（令牌 + 权限点）

| 方法 | 路径 | 说明 | 权限点 |
| --- | --- | --- | --- |
| GET | `/channels` | 渠道分页 | `pay:channel:list` |
| POST | `/channels` | 新增渠道 | `pay:channel:create` |
| GET | `/channels/notify-urls` | 按渠道编码返回系统生成的回调地址（**必须注册在 `{id}` 之前**） | `pay:channel:list` |
| GET | `/channels/schema` | 按渠道编码返回动态表单字段定义（同样必须在 `{id}` 之前） | `pay:channel:list` |
| GET | `/channels/{id}` | 渠道详情（密钥脱敏） | `pay:channel:list` |
| PUT | `/channels/{id}` | 编辑渠道（密钥留空不改） | `pay:channel:update` |
| PUT | `/channels/{id}/status` | 启停 | `pay:channel:update` |
| DELETE | `/channels/{id}` | 删除 | `pay:channel:delete` |
| POST | `/channels/{id}/test` | 连通性测试 | `pay:channel:update` |
| GET | `/orders` | 支付订单分页（`merchantOrderNo` / `bizOrderNo` / `channelCode` / `status` / `beginTime` / `endTime`） | `pay:order:list` |
| GET | `/orders/{id}` | 订单详情 | `pay:order:list` |
| POST | `/orders/{id}/sync` | 主动查单同步状态 | `pay:order:update` |
| POST | `/orders/{id}/close` | 关闭订单 | `pay:order:update` |
| GET | `/refunds` | 退款单分页 | `pay:refund:list` |
| POST | `/refunds` | 发起退款 | `pay:refund:create` |
| GET | `/refunds/{id}` | 退款详情 | `pay:refund:list` |
| GET | `/notify-logs` | 回调日志分页 | `pay:notify-log:list` |

### 8.2 C 端 `/api/v1/client/pay`（会员令牌）

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| GET | `/channels` | 当前可用渠道（供前端选支付方式） |
| POST | `/orders` | 创建支付：`{ bizType, bizOrderNo, channelCode, returnUrl? }`，金额由服务端按业务单确定 |
| GET | `/orders/{merchantOrderNo}` | 查询支付结果（前端轮询） |
> 演示接口不在 pay 模块下，由 `app/modules/demo` 提供：`POST /api/v1/client/demo/orders`
> （创建演示订单并发起支付）、`GET /api/v1/client/demo/orders`（我的演示订单）、
> `GET /api/v1/client/demo/orders/{orderNo}`（单条状态），仅联调用，上线可整体删除。

> C 端请求体**不含金额、不含 userId**，一律服务端推导，避免篡改。

### 8.3 回调（免鉴权，独立分组）

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| POST | `/api/v1/pay/notify/{code}` | 支付结果通知 |

- 不挂 `TokenAuth`、不写 `system_operate_log`、不参与 RBAC；只做验签 + 入日志 + 幂等处理。
- **退款不提供回调路由**：支付宝退款同步返回结果，微信退款由 7.3 的定时任务查单收敛。

---

## 9. 权限点与菜单

菜单挂在「基础设施」同级的一级目录「支付管理」下：

```
支付管理（目录）
├─ 支付渠道      pay:channel:list/create/update/delete
├─ 支付订单      pay:order:list/update
├─ 退款订单      pay:refund:list/create
└─ 回调日志      pay:notify-log:list
```

新增 seed 迁移：`system_menu` 插入目录 + 4 个菜单 + 按钮权限点，并授权给内置角色；随后刷新 `rbac:version`。

---

## 10. 错误码

模块号取 `004`（现有编号：infra=`001`、system=`002`、member=`003`）。

| 码 | 含义 |
| --- | --- |
| `1_004_001_001` | 支付渠道不存在 |
| `1_004_001_002` | 支付渠道已停用 |
| `1_004_001_003` | 渠道配置格式非法（JSON 解析失败或缺少必填项） |
| `1_004_001_004` | `APP_URL` 未配置，无法生成异步通知地址 |
| `1_004_001_005` | 渠道编码已存在 |
| `1_004_001_006` | 不支持的支付渠道编码 |
| `1_004_002_001` | 支付订单不存在 |
| `1_004_002_002` | 支付订单状态不允许该操作 |
| `1_004_002_003` | 重复下单（唯一键冲突） |
| `1_004_002_004` | 渠道下单失败 |
| `1_004_002_005` | 回调验签失败 |
| `1_004_002_006` | 回调金额与订单不一致 |
| `1_004_003_001` | 退款单不存在 |
| `1_004_003_002` | 退款金额超过可退金额 |
| `1_004_003_003` | 渠道退款失败 |
| `1_004_004_001` | 微信支付缺少 openid，请先授权微信登录 |
| `1_004_004_002` | 支付完成跳转地址必须与本站同域 |

---

## 11. 安全与可靠性清单

| 项 | 做法 |
| --- | --- |
| 金额防篡改 | C 端不传金额；回调金额与订单比对 |
| 重复支付 | `uk_biz_order` 唯一键 + 状态机 CAS |
| 重复通知 | `UPDATE ... WHERE status=10` 影响行数判定 |
| 密钥保护 | 明文入库（已确认不加密）；接口脱敏、日志禁止打印 `config`、DB 账号最小权限、备份文件加密 |
| 回调伪造 | 微信验签 + 解密，支付宝 RSA2 验签 |
| 回调重放 | 订单已支付直接应答成功；日志留档 |
| 请求超时 | gopay 调用统一 `context.WithTimeout(10s)` |
| 证书过期 | 保存渠道与启动巡检各做一次离线自检（`channel.SelfCheck`）：APIv3 密钥 32 位、私钥可解析、填了 `merchantCertContent` 时校验证书序列号与 `serialNo` 一致；启动巡检失败仅告警不阻断启动 |
| 日志脱敏 | `pay_notify_log.request_headers` 只留 `Wechatpay-Serial` 等必要项 |

---

## 12. 前端页面

| 页面 | 路径 | 内容 |
| --- | --- | --- |
| 支付渠道 | `views/pay/channel` | 列表 + 抽屉表单（按渠道编码动态字段）+ **顶部只读回调地址 + 复制** + 启停 + 测试 |
| 支付订单 | `views/pay/order` | 筛选、详情抽屉（含原始回调）、同步、关闭 |
| 退款订单 | `views/pay/refund` | 列表 + 详情 |
| 回调日志 | `views/pay/notify-log` | 列表 + 报文查看（格式化 JSON） |

C 端由业务页面自行调用创建支付接口，按 `kind` 分支处理：

```ts
// qr：渲染二维码；redirect：window.location；jsapi：wx.chooseWXPay；app：交给原生壳
```

demo 不做后台页面：联调时直接用 C 端「创建演示订单」接口 + 管理端「支付订单」页就能验证整条链路。

---

## 13. 分期与工作量（估算）

| 阶段 | 内容 | 预估 |
| --- | --- | --- |
| P1 基础 | 6 张表迁移（5 张 pay 表 + `demo_order`）+ 渠道配置 CRUD（明文存储 / 响应脱敏 / 缓存失效）+ `Channel` 抽象接口 + 事件机制接入 | 2~3 人日 |
| P1 微信 | `wx_native` / `wx_h5` / `wx_jsapi` / `wx_mini` / `wx_app` 五场景 + 回调验签与解密 | 3~4 人日 |
| P1 支付宝 | `ali_pc` / `ali_wap` / `ali_app` / `ali_qr` 四场景 + 回调验签 | 2~3 人日 |
| P1 闭环 | 退款全流程、超时关单、退款查单任务、回调日志页、管理端 4 个页面、demo 业务闭环 | 3~4 人日 |
| **合计** | | **10~13 人日** |
| 后续 | 对账单下载、多商户 / 服务商模式 | 另行评估 |

不含：商户号申请、证书下载、域名备案，以及每个渠道用真实商户号做 1 分钱联调（通常 1~2 天/渠道）。

---

## 14. 已确认的决策与外部依赖

### 14.1 已拍板（不再讨论）

| 决策 | 结论 |
| --- | --- |
| 业务场景 | 先写 `demo_order` 演示表跑通链路，真实业务后续由业务模块接入 |
| 渠道范围 | 9 种全部一期实现，不裁剪 |
| 分账 | 不需要 |
| 优惠 / 运费 / 计价 | pay 不参与，只按业务侧传入的 `price` 收单 |
| 退款 | 后台直接发起，无人工审核流程；不单独开退款回调路由，靠定时任务查单收敛 |
| 渠道表单 | 保留动态表单，字段定义由后端 `GET /channels/schema` 下发，前端不硬编码 |
| 回调日志 | 保留后台页面与 `GET /notify-logs` 接口，便于排查报文 |
| 模块边界 | pay 只做支付，不处理业务；业务侧通过 `pay.order.paid` 事件接入 |

### 14.2 仍需外部准备（不阻塞编码）

1. 微信支付商户号 + APIv3 密钥 + 商户证书（或微信支付公钥 ID 与公钥）；
2. 支付宝开放平台应用（RSA2 密钥对，或应用公钥证书 / 根证书 / 支付宝公钥证书三件套）；
3. 公网 HTTPS 域名，且 `.env` 的 `APP_URL` 与该域名一致；
4. 联调阶段每个渠道需要真实商户号做 1 分钱验证；小程序支付还需要已认证小程序 appid 与会员 openid。

---

## 15. 附：与现有模块的一致性

| 现有约定 | 本方案对应 |
| --- | --- |
| 统一响应 `{code,msg,data}` | 全部接口沿用 `result.Success/Fail/Page` |
| 错误码 `1_{模块}_{子模块}_{序号}` | 使用 `1_004_xxx` |
| 迁移须在 `bootstrap/migrations.go` 注册 | 5 个建表迁移 + 1 个菜单 seed 迁移 |
| 管理端 `TokenAuth` + `Permission` | 渠道/订单/退款/日志全部挂权限点 |
| C 端 `MemberAuth` + 不传 memberId | 下单接口只从令牌取会员身份 |
| 定时任务集中在 `bootstrap/schedule.go` | 新增关单 + 回调日志清理两条 |
| 事件机制（本项目首次使用） | 新增 `bootstrap/events.go` 与 `WithEvents(Events)`，支付成功后派发 `pay.order.paid` |
| 回调地址 | 复用 `.env` 的 `APP_URL` + 固定路由生成，与现有 `config/filesystems.go` 里 public 磁盘 URL 的拼法一致 |
| 演示业务 | `demo_order` + `app/modules/demo`，只订阅事件、不建后台页面，验证完可整体删除 |
| 敏感字段列表不返回、编辑留空不改 | 渠道配置沿用同一套交互 |

---

## 16. 实施记录

### 16.1 落地目录

```
app/modules/pay/
├─ route.go                     # admin + client + callback 聚合
├─ admin/{route.go,controller,request}
├─ client/{route.go,controller,request}
├─ callback/{route.go,controller}
├─ channel/{channel.go,schema.go,wechat_v3.go,alipay_v3.go,registry.go,helpers.go}
├─ events/{order_paid.go,dispatch.go}
├─ service/{pay_channel_service,pay_order_service,pay_refund_service,pay_notify_service,order_resolver}.go
├─ repository/{pay_channel,pay_order,pay_refund,pay_notify_log}_repository.go
├─ model/  errs/
app/modules/demo/               # 演示业务：只依赖 pay 的 service + events
├─ route.go
├─ client/controller/order_controller.go
├─ service/order_service.go
├─ repository/order_repository.go
├─ listener/order_paid.go
└─ model/order_model.go
ui/src/api/pay/{channel,order,refund,notify-log}.ts
ui/src/views/pay/{channel,order,refund,notify-log}/
database/migrations/2026110300000{1..5}_*.go + 20261103000006_seed_pay_menu.go
```

### 16.2 与设计稿的差异（均已按实际框架能力调整）

| 设计稿 | 实际实现 | 原因 |
| --- | --- | --- |
| 事件结构体实现 `Signature() string` | 事件键与载荷分离：`&OrderPaidEvent{}` 作 map key，`OrderPaidPayload` 序列化后经 `event.Arg` 传递，监听器反序列化 | Goravel 1.18 的 `event.Event` 只要求 `Handle(args []Arg) ([]Arg, error)`，`Signature()` 属于 `Listener`；且事件按实例指针查映射，事件本身必须零值 |
| 渠道 client 缓存 + `pay:channel:version` 失效 | 进程内 `sync.Map`，key = 渠道编码 + config 的 SHA256 | key 含配置哈希，配置一改自然换 key，不需要额外的版本号与跨实例失效逻辑 |
| `Channel` 接口含 5 个方法 | 增加 `QueryRefund` 与 `NotifyFailAck` | 微信退款需要查单收敛、失败应答要按渠道格式构造 |
| 退款状态同步对所有渠道查单 | 仅微信需要；支付宝退款同步返回终态，`QueryRefund` 直接返回「不支持」由调用方跳过 | 避免无意义的接口调用与误判 |
| pay 直接查 `member_social_user` 取 openid | pay 定义 `RegisterOpenIDResolver`，`bootstrap/app.go` 启动时注入 member 的 `FindOpenIDByMember` | 保持 pay 不 import member，与权限检查器 / 操作日志的注入方式一致 |
| C 端下单「金额由服务端按业务单确定」 | pay 定义 `RegisterBizOrder(bizType, resolver)`，业务模块注册反查函数（demo 已注册 `demo_order`） | pay 不认识业务表，用注册表拿到金额，请求体仍不接受金额 |
| 应用日志清理任务只有 API 访问日志 | `bootstrap/schedule.go` 新增 `pay:close-expired-orders`（5 分钟）、`pay:sync-refunding`（5 分钟）、`pay:purge-notify-logs`（每天 03:40） | 支付闭环兜底 |

### 16.3 联调步骤

```bash
go run . artisan migrate                 # 建 6 张表 + 支付菜单
go run .                                 # 后端 11111
cd ui && pnpm dev                        # 前端 22222
```

1. 后台「支付管理 → 支付渠道 → 新增渠道」配置商户凭据（密钥明文入库，接口脱敏）；
2. 抽屉顶部核对系统生成的异步通知地址，确保 `APP_URL` 是公网 HTTPS 域名；
3. C 端登录会员后调用 `POST /api/v1/client/demo/orders {"channelCode":"wx_native"}` 拿支付参数；
4. 支付完成后看「支付订单」是否变为已支付、「回调日志」是否有验签通过的记录、「演示订单」是否被事件回写；
5. 需要退款时在「退款订单 → 发起退款」填支付单号，微信退款会先停在「退款中」，约 5 分钟后由定时任务收敛。
