# 服务升级与迁移指南

本指南描述 Zero Net Panel 在迭代过程中的升级策略、注意事项与数据库迁移流程，帮助运营人员安全上线新版本。

## 发行说明

### Go 1.22 升级

- **破坏性变更**：构建与运行时环境需升级至 Go 1.22 或更高版本，Go 1.21 将无法通过新的 CI/Release 工作流。升级后请本地执行 `go mod tidy`、`go fmt`, `go vet`, `go test ./...` 及 `golangci-lint` 以确保兼容。
- **依赖验证**：现有依赖（`github.com/zeromicro/go-zero v1.5.3`、`google.golang.org/grpc v1.55.0` 等）已在 Go 1.22 下通过编译与测试，无需额外调整。如需自定义升级，可参考官方发行说明确认兼容性。

## 版本策略

- **分支规范**：遵循 `develop` 作为日常开发分支，所有功能分支先合并至 `develop`，经验证后再进入 `main`。
- **版本号**：采用 `MAJOR.MINOR.PATCH` 语义化版本，当出现破坏性变更或数据库结构调整时提升 MAJOR 版本。
- **预发布**：在 `develop` 上每日构建 Preview 版本，通过 CLI `znp migrate` 与 `znp serve api --config etc/znp-sqlite.yaml` 进行联调。

## 数据库迁移

1. **注册迁移**：所有结构定义位于 `internal/repository`，每次 Schema 变更需在 `internal/bootstrap/migrations/registry.go` 新增一个 `Migration` 结构体，并使用递增的时间戳（如 `2024063001`）作为版本号。
2. **执行迁移**：CLI `znp migrate --apply` 与 `znp serve --auto-migrate` 会调用 `internal/bootstrap.ApplyMigrations`。可通过 `--to <version>`（或 `serve` 子命令的 `--migrate-to`）限制执行到指定版本；迁移元数据会记录在 `schema_migrations` 表中，重复运行将自动跳过已执行版本。
3. **演示数据**：`internal/bootstrap/seed` 提供基础演示数据，可根据 `ZNP_SEED` 环境变量或 CLI `--seed-demo` 参数控制是否执行。
4. **数据备份**：生产环境升级前务必对核心表（`users`、`subscriptions`、`plans`、`orders`、`announcements` 等）进行备份，推荐使用数据库快照或备份脚本。

## 回滚策略

- **配置回滚**：保留上一版本的配置文件与 `security_settings`、`plans` 等核心表的快照。
- **代码回滚**：若新版本出现问题，可快速切换至上一个发布标签并执行 `znp migrate --rollback`（规划中）。
- **流量切换**：建议通过 Nginx/Ingress 逐步放量，实时监控 gRPC 与 REST 健康检查。

## 常见问题

- **迁移失败**：确认数据库用户具备 DDL 权限，查看日志了解具体 SQL 错误。
- **第三方加密开关**：`security_settings` 默认关闭，如开启需同步客户端密钥，并验证签名是否正确。
- **缓存一致性**：变更节点、套餐等高频数据后建议清理 Redis/本地缓存，保证新配置即时生效。
