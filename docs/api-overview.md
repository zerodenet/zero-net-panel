# API 模块与业务逻辑说明

本文档汇总 Zero Net Panel 已实现的 REST API 模块，并对关键业务逻辑进行说明，方便前后端协作与三方集成。

## 管理端模块

| 模块 | 路径 | 说明 |
| ---- | ---- | ---- |
| 仪表盘 | `/api/v1/{admin}/dashboard` | 展示模块导航、权限控制 |
| 节点管理 | `/api/v1/{admin}/nodes` | 节点查询、协议内核同步 |
| 订阅模板 | `/api/v1/{admin}/subscription-templates` | 模板 CRUD、发布、历史追溯 |
| 套餐管理 | `/api/v1/{admin}/plans` | 套餐列表、创建、更新，字段涵盖价格、时长、流量限制等 |
| 公告中心 | `/api/v1/{admin}/announcements` | 公告列表、创建、发布，支持置顶与可见时间窗 |
| 安全配置 | `/api/v1/{admin}/security-settings` | 读取与更新第三方签名/加密开关、凭据与时间窗口 |
| 订单管理 | `/api/v1/{admin}/orders` | 检索、查看订单，支持多支付方式、外部流水追踪、手动标记支付/取消与余额退款 |

> `{admin}` 为可配置的后台前缀，默认为 `admin`，可通过 `Admin.RoutePrefix` 自定义。

## 用户端模块

- `/api/v1/user/subscriptions`：用户订阅列表、预览、模板切换。
- `/api/v1/user/plans`：面向终端的套餐列表，返回价格、特性、流量限制等字段。
- `/api/v1/user/announcements`：按受众过滤当前有效公告，支持置顶排序与限量返回。
- `/api/v1/user/account/balance`：返回当前余额、币种以及流水历史。
- `/api/v1/user/orders`：创建、查询订单并支持取消待支付或零元订单，返回计划快照、条目与余额快照。

### 订单操作补充说明

- 用户端 `POST /api/v1/user/orders` 新增 `payment_method`、`payment_channel`、`payment_return_url` 字段：
  - 默认 `payment_method = balance`，系统直接扣减余额、记录 `balance_transactions`，订单状态立即变为 `paid`、`payment_status = succeeded`。
  - 当 `payment_method = external` 且金额大于零时，会生成 `pending_payment` 订单，创建 `order_payments` 预订单记录，并返回 `payment_intent_id`、`payments` 列表供前端跳转支付；余额不会变动。
- 用户端 `POST /api/v1/user/orders/{id}/cancel` 仅允许取消待支付或零金额订单，不触发余额回滚。
- 管理端提供 `POST /api/v1/{admin}/orders/{id}/pay`、`/cancel` 与 `/refund`，需管理员角色；退款仅适用于余额支付订单，成功后会写入退款流水并回滚余额。
- 外部支付回调可按以下流程接入：
  1. 网关回调携带支付状态后，通过内部逻辑 `PaymentCallbackLogic`（或后续开放的专用接口）调用 `UpdatePaymentState`、`UpdatePaymentRecord`，将订单状态从 `pending_payment` 更新为 `paid`/`payment_failed`，并填充 `payment_reference`、`payment_failure_*` 字段。
  2. 回调完成后，`GET /api/v1/user/orders/:id` 与 `/admin/orders/:id` 均会返回最新的 `payment_status`、`payments` 明细，方便前端落地扫码/轮询场景。

所有用户端接口默认需要 JWT 鉴权，同时可选启用第三方加密认证中间件，对请求进行签名验证与 AES-GCM 解密。

## 第三方认证与加密

- `security_settings` 表提供全局开关，包含 `ThirdPartyAPIEnabled`、`APIKey`、`APISecret`、`NonceTTLSeconds`，可通过管理端 `GET/PATCH /security-settings` 接口调整。
- 中间件通过 `X-ZNP-API-Key`、`X-ZNP-Timestamp`、`X-ZNP-Nonce`、`X-ZNP-Signature` 校验请求。
- 当 `X-ZNP-Encrypted: true` 时，请求体需要使用 `api_secret` 派生的 AES-256-GCM 加密，IV 通过 `X-ZNP-IV` 传递。

## 业务扩展方向

1. **套餐售卖流程**：已实现余额与外部支付并行的下单流程（含流水记录、回调处理），后续可扩展续费、套餐升级与更多支付渠道。
2. **公告推送渠道**：结合 Webhook、邮件通知，将公告同步到外部 IM 渠道。
3. **余额充值**：配合支付网关实现充值、退款、自动开票功能。
4. **审计日志**：记录套餐、公告、节点变更的操作明细，满足审计与回溯需求。
