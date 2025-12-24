# 运维操作手册

本手册面向日常值班与集成测试场景，提供 Zero Network Panel 的常见 CLI 流程、HTTP 接口调用步骤以及样例请求/响应，帮助快速定位故障并完成节点、套餐等核心业务操作。

## CLI 工作流

| 场景 | 命令 | 说明 |
| ---- | ---- | ---- |
| 查看迁移状态 | `go run ./cmd/znp migrate --config <file> --status` | 输出 `executed`/`pending` 版本列表，方便核对环境是否落后。 |
| 执行数据库迁移 | `go run ./cmd/znp migrate --config <file> --apply --to <version>` | 在运维窗口中逐步升级至指定版本，命令完成后会打印 `before/after/target`。 |
| 回滚最近一次迁移 | `go run ./cmd/znp migrate --config <file> --apply --rollback --to <prev>` | 回退前需手动确认备份可用，执行后请检查 `schema_migrations`。 |
| 检查配置摘要 | `go run ./cmd/znp tools check-config --config <file>` | 校验数据库、缓存、内核配置是否可用。 |
| 启动带观察窗口的服务 | `go run ./cmd/znp serve --config <file> --migrate-to latest --graceful-timeout 30s` | 常用于灰度发布或临时演练，确保迁移与服务启动一体化执行。 |

## HTTP 操作流程

### 1. 节点同步巡检

1. 以管理员身份获取 Bearer Token。
2. 触发节点同步：
   ```bash
   curl -X POST http://127.0.0.1:8888/api/v1/admin/nodes/42/kernels/sync \
     -H "Authorization: Bearer <ADMIN_TOKEN>"
   ```
3. 期待响应：
   ```json
   {
     "code": 0,
     "message": "sync scheduled",
     "data": {
       "nodeId": 42,
       "protocol": "grpc",
       "triggeredAt": "2024-06-30T16:55:00Z"
     }
   }
   ```
4. 若返回 `code=404004` 并提示 `node not found`，请检查节点是否被删除或路由前缀是否正确。
5. 若长时间无同步结果，可通过 `goctl`/日志确认内核连接是否异常。

### 2. 套餐发布流程

1. 查询现有套餐模板：
   ```bash
   curl http://127.0.0.1:8888/api/v1/admin/subscription-templates \
     -H "Authorization: Bearer <ADMIN_TOKEN>"
   ```
2. 发布指定模板：
   ```bash
   curl -X POST http://127.0.0.1:8888/api/v1/admin/subscription-templates/7/publish \
     -H "Authorization: Bearer <ADMIN_TOKEN>"
   ```
3. 创建或更新套餐：
   ```bash
   curl -X POST http://127.0.0.1:8888/api/v1/admin/plans \
     -H "Authorization: Bearer <ADMIN_TOKEN>" \
     -H "Content-Type: application/json" \
     -d '{
       "name": "Pro-30",
       "price": 29.9,
       "currency": "CNY",
       "durationDays": 30,
       "traffic": "300 GB",
       "templateId": 7
     }'
   ```
4. 终端用户校验套餐列表：
   ```bash
   curl http://127.0.0.1:8888/api/v1/user/plans \
     -H "Authorization: Bearer <ACCESS_TOKEN>"
   ```
5. 典型响应：
   ```json
   {
     "code": 0,
     "message": "OK",
     "data": [
       {
         "id": 10001,
         "name": "Pro-30",
         "price": 29.9,
         "currency": "CNY",
         "durationDays": 30,
         "traffic": "300 GB",
         "features": ["优先节点", "高速带宽"],
         "status": "published"
       }
     ]
   }
   ```
6. 如遇 `code=409001`（版本冲突），说明模板存在未发布的草稿，请先重新发布模板再创建套餐。

### 3. 开启第三方签名校验

1. 管理员查询现有配置：
   ```bash
   curl http://127.0.0.1:8888/api/v1/admin/security-settings \
     -H "Authorization: Bearer <ADMIN_TOKEN>"
   ```
2. 启用签名与加密：
   ```bash
   curl -X PATCH http://127.0.0.1:8888/api/v1/admin/security-settings \
     -H "Authorization: Bearer <ADMIN_TOKEN>" \
     -H "Content-Type: application/json" \
     -d '{
       "thirdPartyAPIEnabled": true,
       "apiKey": "demo-key",
       "apiSecret": "demo-secret",
       "nonceTTLSeconds": 300
     }'
   ```
3. 期待响应：
   ```json
   {
     "code": 0,
     "message": "updated",
     "data": {
       "thirdPartyAPIEnabled": true,
       "nonceTTLSeconds": 300,
       "updatedAt": "2024-06-30T17:05:00Z"
     }
   }
   ```
4. 客户端后续请求需携带 `X-ZNP-API-Key`、`X-ZNP-Timestamp`、`X-ZNP-Nonce` 与 `X-ZNP-Signature`，并在开启加密时附加 `X-ZNP-IV` 与 `X-ZNP-Encrypted: true`。
5. 支付回调建议使用 Webhook 配置：Stripe 使用 `Stripe-Signature`（在 `Webhook.Stripe.SigningSecret` 配置），或通过 `Webhook.SharedToken` 携带 `X-ZNP-Webhook-Token`。
6. 若收到 `code=401001`（signature mismatch），请检查第三方签名顺序是否为 `timestamp + "\n" + nonce + "\n" + body`，并确保时间戳处于允许窗口内。

更多巡检、升级与排障方案请继续阅读 [docs/service-upgrade.md](service-upgrade.md)。

## 运维工具与脚本
- 配置校验：`go run ./cmd/znp tools check-config --config <file>`，输出 HTTP/GRPC/DB/缓存/Webhook/管理入口摘要。
- 探活与错误扫描：`scripts/healthcheck.sh`，可覆盖 `ZNP_HEALTH_URL`、`ZNP_LOG_FILE`、`ZNP_ERROR_PATTERNS`，用于 cron 或探针。
- 数据库备份：`scripts/backup-db.sh <output.sql>`，通过 `ZNP_DB_DRIVER=mysql|postgres` 等 env 选择驱动/凭据。
- 进程托管：`deploy/systemd/znp.service`、`deploy/docker/Dockerfile*` 提供最小示例；可结合 `/api/v1/ping` 和 `/metrics` 做健康/指标采集。
