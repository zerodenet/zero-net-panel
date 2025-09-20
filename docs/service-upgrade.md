# 服务升级与迁移指南

本指南描述 Zero Net Panel 在迭代过程中的升级策略、注意事项与数据库迁移流程，并补充升级前后必须完成的端到端校验场景，帮助运营人员安全上线新版本。

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
2. **执行迁移/回滚**：CLI `znp migrate --apply` 与 `znp serve --auto-migrate` 会调用 `internal/bootstrap.ApplyMigrations`。可通过 `--to <version>`（或 `serve` 子命令的 `--migrate-to`）限制执行到指定版本；若需回退到旧版本，必须显式执行 `znp migrate --apply --rollback --to <version>`。命令行会在完成后输出 `before/after/target` 以及 `versions=[...]` 列表，便于核对实际执行的升级或回滚序列。迁移元数据记录在 `schema_migrations` 表中，重复运行会自动跳过已执行版本。
3. **演示数据**：`internal/bootstrap/seed` 提供基础演示数据，可根据 `ZNP_SEED` 环境变量或 CLI `--seed-demo` 参数控制是否执行。
4. **数据备份**：生产环境升级前务必对核心表（`users`、`subscriptions`、`plans`、`orders`、`announcements` 等）进行备份，推荐使用数据库快照或备份脚本。

## 升级前后验证清单

### 签名校验回归

1. 在预发布或灰度环境中使用管理员账户调用 `GET /api/v1/{admin}/security-settings`，记录当前配置。
2. 若版本中包含安全配置变更，执行 `PATCH /api/v1/{admin}/security-settings` 更新密钥，并确认响应中的 `updatedAt` 为最新时间。
3. 使用第三方客户端按 `timestamp + "\n" + nonce + "\n" + body` 规则生成签名，调用一个受保护接口（例如 `GET /api/v1/user/account/balance`），校验返回结果。
4. 针对失败场景，确认系统返回的错误码符合预期并按下表排障。

| 接口 | 错误码 | 说明 | 排障建议 |
| ---- | ------ | ---- | -------- |
| `PATCH /api/v1/{admin}/security-settings` | `400100` | 参数缺失或 TTL 小于 60 秒 | 校验部署脚本是否正确下发 `nonceTTLSeconds`、`apiKey` 与 `apiSecret`。 |
| `GET /api/v1/{admin}/security-settings` | `500120` | 配置读取失败 | 确认数据库连接可用，必要时查看 `security_settings` 表结构。 |
| 受保护接口（任意） | `401001` | 签名不一致 | 检查客户端 HMAC 密钥、换行符与请求体是否被额外转义。 |
| 受保护接口（任意） | `403001` | 时间戳超出窗口 | 同步客户端 NTP，或在测试阶段暂时增大 `nonceTTLSeconds`。 |
| 受保护接口（任意） | `403002` | Nonce 重复使用 | 确认重试策略会刷新随机数，避免使用缓存的请求副本。 |

### 节点同步回归

1. 升级前记录关键节点列表与上次同步时间（`GET /api/v1/{admin}/nodes`）。
2. 升级后随机抽查 2~3 个节点，调用 `POST /api/v1/{admin}/nodes/{id}/kernels/sync` 并观察响应。
3. 检查服务日志或 Prometheus 指标 `znp_node_sync_operations_total`，确保无大规模失败。
4. 若节点同步失败，参考下表排障：

| 接口 | 错误码 | 说明 | 排障建议 |
| ---- | ------ | ---- | -------- |
| `POST /api/v1/{admin}/nodes/{id}/kernels/sync` | `404004` | 节点不存在 | 环境差异导致 ID 变化，重新确认节点清单或重新导入数据。 |
| 同上 | `409101` | 同步任务正在进行 | 等待当前任务完成，必要时调整同步间隔或并发度。 |
| 同上 | `500101` | 内核握手失败 | 验证 `Kernel` 配置、内核 token 与网络连通性；可使用 `curl`/`grpcurl` 手动探测。 |
| `GET /api/v1/{admin}/nodes` | `400300` | 查询条件非法 | 升级脚本可能注入了旧版参数，移除无效过滤项后重试。 |

### 套餐发布回归

1. 升级前在 staging 环境导出关键套餐与模板 ID。
2. 升级后调用 `POST /api/v1/{admin}/subscription-templates/{id}/publish` 发布新模板。
3. 通过 `POST /api/v1/{admin}/plans` 或 `PATCH /api/v1/{admin}/plans/{id}` 创建/更新套餐。
4. 使用测试账号访问 `GET /api/v1/user/plans`，确认新套餐可见，并执行一次 `POST /api/v1/user/orders` 验证扣费链路。

| 接口 | 错误码 | 说明 | 排障建议 |
| ---- | ------ | ---- | -------- |
| `POST /api/v1/{admin}/subscription-templates/{id}/publish` | `409001` | 模板存在未发布草稿 | 先保存草稿或清理旧版本，再次发布。 |
| `POST /api/v1/{admin}/plans` | `400201` | 套餐字段缺失或价格非法 | 检查部署脚本中的 JSON 字段，确保价格、时长、模板 ID 等必填项已设置。 |
| `PATCH /api/v1/{admin}/plans/{id}` | `409201` | 乐观锁冲突 | 前端或脚本使用了过期的版本号，重新获取详情后重试。 |
| `GET /api/v1/user/plans` | `503001` | 套餐缓存构建失败 | 查看缓存服务状态，必要时重启服务或手动清理缓存。 |
| `POST /api/v1/user/orders` | `402001` | 余额不足 | 使用测试账户充值，或暂时设置套餐为零元套餐验证流程。 |

## 回滚策略

- **配置回滚**：保留上一版本的配置文件与 `security_settings`、`plans` 等核心表的快照。
- **代码回滚**：若新版本出现问题，可快速切换至上一个发布标签，并执行 `znp migrate --apply --rollback --to <旧版本号>` 将数据库恢复到目标结构。操作前务必确认 `schema_migrations` 中的版本号，并核对 CLI 输出的 `before/after/target` 及 `versions=[...]` 信息。
- **执行前备份**：在回滚或执行破坏性迁移之前，必须完成全量备份（如快照、`mysqldump`/`sqlite3 .backup`）。只有在确认备份可用后，才应继续执行 CLI 回滚命令，避免在恢复过程中发生不可逆的数据丢失。
- **备份优先**：在执行回滚或任何破坏性迁移前，务必对数据库进行物理备份或至少导出关键表（`users`、`subscriptions`、`orders` 等）。建议在备份成功后，通过只读副本或 staging 环境演练一次回滚流程，确保 `--rollback` 参数和版本号正确无误。
- **风险提示**：回滚会删除新建表或字段，可能导致最新数据不可恢复。确认业务已停机或无写入流量，避免在回滚过程中产生新的业务数据。若目标版本早于第一条迁移，可将 `--to` 设置为更小的时间戳（例如 `--to 2024060100`）以清空所有结构，但更应提前做好全量备份。
- **流量切换**：建议通过 Nginx/Ingress 逐步放量，实时监控 gRPC 与 REST 健康检查。

## 常见问题

- **迁移失败**：确认数据库用户具备 DDL 权限，查看日志了解具体 SQL 错误。
- **第三方加密开关**：`security_settings` 默认关闭，开启需同步客户端密钥，并验证签名是否正确。若仍失败，请对照上文错误码表排查。
- **缓存一致性**：变更节点、套餐等高频数据后建议清理 Redis/本地缓存，保证新配置即时生效。
- **指标异常**：Prometheus 抓取不到指标时，确认 `Metrics.ListenOn` 与防火墙规则是否正确，并检查 `znp_node_sync_operations_total`、`znp_order_create_requests_total` 等关键指标是否持续增长。
