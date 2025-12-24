# 上线准备与运行指南

## 轻量化原则
- 单进程/单二进制优先：`go build -o bin/znp ./cmd/znp` 可直接运行；仅在需要时使用 Docker/systemd。
- 依赖最小化：默认仅需 Go + 数据库（生产建议 MySQL/PostgreSQL），缓存可先用内存实现；监控与日志保持内置/轻量脚本即可。
- 配置显式化：一份 `<env>.yaml` 即可启动，不强依赖外部配置中心。

## 上线前必做
- 环境：Go 1.25；生产数据库建议 MySQL/PostgreSQL；测试与 CI 需 CGO 开启以跑 SQLite 相关测试。
- 配置：复制 `etc/znp-prod.example.yaml` 为环境专用文件，替换 `Auth` 密钥、数据库/缓存凭据、`Kernel` 端点；禁用/删除演示账户与 `--seed-demo` 数据。可用 `envsubst < etc/znp-prod.example.yaml > /etc/znp/znp.yaml` 由环境变量生成最终配置。
- 目录约定：业务代码在 `internal/`（handler/logic/svc/repository），API 定义在 `api/`，CLI 入口 `cmd/znp/`，示例配置 `etc/`，文档 `docs/`。

## 数据迁移与回滚
- 生成/调整迁移：在 `internal/bootstrap/migrations` 添加迁移，保证可逆。
- 预演：`go run ./cmd/znp migrate --config etc/znp-sqlite.yaml --apply --seed-demo` 仅限本地验证。
- 生产：`go run ./cmd/znp migrate --config <prod.yaml> --apply --to <version>`；发布前执行一次 `--rollback` 演练，确认可逆。
- 常见故障排查：确认数据库连通/权限；如报“beyond latest”检查目标版本；失败后先 `--rollback --to <prev>` 再重试；保留失败日志用于比对 schema 状态。

## 备份与导出
- 使用 `scripts/backup-db.sh <output.sql>`，通过环境变量选择驱动与凭据（`ZNP_DB_DRIVER=mysql|postgres` 等）；建议发布前和每日定时执行。

## 启动与运维
- 启动：`go run ./cmd/znp serve --config <prod.yaml> --migrate-to latest [--disable-grpc]`；可直接前台运行或用 systemd/Docker 托管，保持探针简单。
- 健康检查：`GET /api/v1/ping`；配置自检：`go run ./cmd/znp tools check-config --config <prod.yaml>`.
- 运行时关注日志与退出码；必要时启用 `Metrics.ListenOn` 分离指标端口。
- 版本输出：`znp version` 打印版本、提交、构建时间（Makefile 已注入 ldflags）。

## 安全加固
- 秘钥轮换：定期更新 JWT 秘钥；将配置放入环境变量/密钥管控服务，不写入仓库。
- 网络：为 HTTP/GRPC 增加 TLS/反向代理，限定管理端 `Admin.RoutePrefix` 与防火墙策略。
- 数据：最小权限的 DB/缓存账户；禁用默认管理员、关闭演示数据；对外接口启用限流/重试策略。
- 入口防护：使用 `Admin.Access` 配置 IP 白名单与每 IP 限速，防止管理端暴露。
- 回调校验：支付回调可配置 `Webhook.AllowCIDRs`、`Webhook.SharedToken` 或 `Webhook.Stripe.SigningSecret`（`Stripe-Signature`）进行校验。

## 测试与验收
- 基线：`gofmt -w .`, `go vet ./...`, `golangci-lint run`, `go test ./...`（CGO 开启）；涉及数据库的改动新增/更新迁移测试。
- 手工验收：节点同步、模板发布、用户订阅预览、下单/取消/退款、余额查询、公告发布；跨 DB（MySQL/PostgreSQL）抽样验证核心链路。

## 监控与告警
- 轻量监控：保持 `/metrics` 可用（便于后续扩展）；最小化方案用健康检查 `GET /api/v1/ping` + 进程存活探针 + 日志采集。
- 告警建议：基于日志关键词/计数或简单探活脚本，对错误率、延迟、迁移失败、余额异常设置通知；如需指标再按需接入 Prometheus，不强依赖 Grafana。
- 示例：`scripts/healthcheck.sh` 可用于 cron/探针，支持覆盖 `ZNP_HEALTH_URL`、`ZNP_LOG_FILE`、`ZNP_ERROR_PATTERNS`。

## 上线验收清单
- 配置与健康：`znp tools check-config --config <prod.yaml>` 通过；健康检查 `GET /api/v1/ping` 通过。
- 节点与模板：节点同步触发成功；模板发布成功并可查询历史。
- 订阅链路：用户订阅列表与预览返回 200；ETag/Content-Type 正确。
- 订单闭环：下单成功并扣费，取消/退款接口返回预期结果，余额同步更新；创建时可传 `idempotency_key`，重复请求返回同一订单防止重复扣费。
- 公告与可见性：创建/过滤公告成功，受众过滤生效。
- 迁移可逆：`migrate --apply` 与 `--rollback --to <prev>` 在非生产环境验证通过。

## 发布与回滚
- 分支：`develop` 合并到 `main` 发布；提交沿用 `[feat]/[fix]/[docs]/[ci]` 前缀。
- 构建：`go build -o bin/znp ./cmd/znp`（或镜像构建）；发布前生成 tag 并记录版本与迁移版本号。
- 回滚：按版本恢复二进制/镜像，同时执行 `migrate --rollback --to <prev>`；发布说明需标记不兼容变更。

## 运行手册（速查）
- 启动卡住：先跑 `tools check-config`；确认 DB 连通与凭据。
- 迁移异常：检查目标版本与锁，重试 `--rollback` 后再 `--apply`；保留失败日志。
- 请求异常：查看网关/代理日志与 `Admin.RoutePrefix` 配置，使用健康检查验证存活。
