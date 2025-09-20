# API 模块与业务逻辑说明

本文档汇总 Zero Net Panel 已实现的 REST API 模块，并补充关键业务的端到端流程、错误码与排障建议，方便前后端协作与第三方集成。

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
- 所有用户端接口默认需要 JWT 鉴权，同时可选启用第三方加密认证中间件，对请求进行签名验证与 AES-GCM 解密。
- 外部支付回调可按以下流程接入：
  1. 网关回调携带支付状态后，通过内部逻辑 `PaymentCallbackLogic`（或后续开放的专用接口）调用 `UpdatePaymentState`、`UpdatePaymentRecord`，将订单状态从 `pending_payment` 更新为 `paid`/`payment_failed`，并填充 `payment_reference`、`payment_failure_*` 字段。
  2. 回调完成后，`GET /api/v1/user/orders/:id` 与 `/admin/orders/:id` 均会返回最新的 `payment_status`、`payments` 明细，方便前端落地扫码/轮询场景。

## 端到端流程

### 第三方签名校验流程

1. 管理员通过 `GET /api/v1/{admin}/security-settings` 查询当前开关与密钥。
2. 根据需要调用 `PATCH /api/v1/{admin}/security-settings` 设置 `thirdPartyAPIEnabled`、`apiKey`、`apiSecret` 与 `nonceTTLSeconds`。
3. 第三方客户端在调用任何受保护接口时，按照 `timestamp + "\n" + nonce + "\n" + body` 规则生成 HMAC-SHA256 签名，并随请求携带以下头：
   - `X-ZNP-API-Key`
   - `X-ZNP-Timestamp`（Unix 秒）
   - `X-ZNP-Nonce`（建议 16 字节随机值）
   - `X-ZNP-Signature`
   - `X-ZNP-Encrypted: true` 与 `X-ZNP-IV`（可选，当启用 AES-256-GCM 加密时必填）
4. 服务端校验签名、时间窗口与随机数重复使用情况，必要时进行解密后再继续路由。

| 接口 | 错误码 | 说明 | 排障建议 |
| ---- | ------ | ---- | -------- |
| `PATCH /api/v1/{admin}/security-settings` | `400100` | 参数缺失或 TTL 小于 60 秒 | 校验请求体字段是否齐全，确认 `nonceTTLSeconds >= 60`。 |
| 同上 | `409002` | 存在并发更新冲突 | 使用最新版 `updatedAt` 再次提交，或开启重试机制。 |
| 受保护接口（任意） | `401001` | 签名不一致 | 确保使用 `apiSecret` 计算 HMAC，检查换行与大小写是否匹配。 |
| 受保护接口（任意） | `403001` | 时间戳超出窗口 | 对齐客户端时间，必要时缩短网络传输延迟或增大 `nonceTTLSeconds`。 |
| 受保护接口（任意） | `403002` | Nonce 重复使用 | 确认客户端在重试时生成全新随机数。 |

### 节点同步流程

1. 管理端列表接口 `GET /api/v1/{admin}/nodes` 返回节点详情与协议能力。
2. 运维人员选择目标节点，调用 `POST /api/v1/{admin}/nodes/{id}/kernels/sync` 触发与内核的即时同步。
3. 服务端异步排队同步任务，返回排队结果，并通过日志或指标追踪执行情况。
4. 若开启 Prometheus，观察 `znp_node_sync_operations_total` 与 `znp_node_sync_duration_seconds` 判断成功率与耗时。

| 接口 | 错误码 | 说明 | 排障建议 |
| ---- | ------ | ---- | -------- |
| `GET /api/v1/{admin}/nodes` | `400300` | 过滤条件非法 | 确认查询参数（如 `protocol`、`status`）是否在允许范围内。 |
| `POST /api/v1/{admin}/nodes/{id}/kernels/sync` | `404004` | 节点不存在 | 检查节点是否被删除，确认 `Admin.RoutePrefix` 与 URL 中的 `{id}` 是否正确。 |
| 同上 | `409101` | 同步任务正在进行 | 等待上一次任务完成或在 30 秒后重试。 |
| 同上 | `500101` | 内核握手失败 | 检查内核服务地址、令牌是否正确，必要时查看 `Kernel` 配置或抓取 gRPC/HTTP 日志。 |

### 套餐发布流程

1. 管理端 `GET /api/v1/{admin}/subscription-templates` 列表展示模板与版本号。
2. 调用 `POST /api/v1/{admin}/subscription-templates/{id}/publish` 生成新的渲染版本。
3. 使用 `POST /api/v1/{admin}/plans` 创建套餐或 `PATCH` 更新既有套餐，关联最新模板版本、价格与流量限制。
4. 前端或第三方调用 `GET /api/v1/user/plans` 验证套餐是否对终端可见。
5. 订单创建时，`POST /api/v1/user/orders` 会读取套餐快照、扣减余额并返回结果。

| 接口 | 错误码 | 说明 | 排障建议 |
| ---- | ------ | ---- | -------- |
| `POST /api/v1/{admin}/subscription-templates/{id}/publish` | `404010` | 模板不存在或无权限 | 校验模板 ID 与管理员角色；检查是否已归档。 |
| 同上 | `409001` | 模板存在未发布草稿 | 先保存最新草稿，再重新发起发布或删除旧草稿。 |
| `POST /api/v1/{admin}/plans` | `400201` | 套餐字段缺失或价格非法 | 核对必填字段（`name`、`price`、`durationDays`、`templateId`），确保价格 > 0。 |
| 同上 | `409201` | 套餐名称已存在 | 更换名称或在更新接口中使用已有套餐 ID。 |
| `GET /api/v1/user/plans` | `503001` | 套餐缓存构建失败 | 查看缓存服务状态，必要时执行 `znp cache purge`（后续计划）或重启服务。 |
| `POST /api/v1/user/orders` | `402001` | 余额不足 | 提示用户充值或调整套餐价格。 |
| 同上 | `409301` | 套餐不可用 | 确认套餐状态为 `published` 且未过期，或检查权限配置。 |

## 第三方认证与加密

- `security_settings` 表提供全局开关，包含 `ThirdPartyAPIEnabled`、`APIKey`、`APISecret`、`NonceTTLSeconds`，可通过管理端 `GET/PATCH /security-settings` 接口调整。
- 中间件通过 `X-ZNP-API-Key`、`X-ZNP-Timestamp`、`X-ZNP-Nonce`、`X-ZNP-Signature` 校验请求。
- 当 `X-ZNP-Encrypted: true` 时，请求体需要使用 `api_secret` 派生的 AES-256-GCM 加密，IV 通过 `X-ZNP-IV` 传递。

## 业务扩展方向

1. **套餐售卖流程**：已实现余额与外部支付并行的下单流程（含流水记录、回调处理），后续可扩展续费、套餐升级与更多支付渠道。
2. **公告推送渠道**：结合 Webhook、邮件通知，将公告同步到外部 IM 渠道。
3. **余额充值**：配合支付网关实现充值、退款、自动开票功能。
4. **审计日志**：记录套餐、公告、节点变更的操作明细，满足审计与回溯需求。
