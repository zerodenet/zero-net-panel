# Zero Network Panel (ZNP)

Zero Network Panel 旨在以 xboard 的功能体系为基线，提供面向节点运营、用户订阅、套餐计费等全栈后端能力。本项目采用 Go 语言与 [go-zero](https://go-zero.dev/) 微服务框架构建，默认以 RESTful API 的方式对外暴露接口，并结合 GORM、可插拔缓存服务以及自动化 CI/CD，支撑后续协议层和运营扩展。

## 核心模块
- **节点发现 (kernel discovery)**：内置 HTTP 与 gRPC Provider 注册表，可在后台一键触发节点配置同步，确保协议资源与内核保持一致。
- **订阅模板管理**：提供模板 CRUD、版本发布、历史追溯及默认模板切换，变量描述采用 GitHub 风格的分页与字段规范。
- **用户订阅能力**：支持订阅列表查询、模板预览与定制选择，同时输出渲染后的内容、ETag 及内容类型信息，方便前端或客户端下载。
- **套餐/公告/余额**：实现 `plans`、`announcements`、`user_balances` 等核心表，对齐 xboard 套餐管理、公告通知与钱包查询能力，并支持第三方加密校验开关。
- **计费订单**：新增 `orders`/`order_items` 模型，支持用户下单、余额扣费与取消，管理端可检索订单并执行手动支付、取消与余额退款，支撑支付与开票扩展。
- **第三方安全配置**：提供 `security_settings` 仓储与管理端接口，可动态开启/关闭签名与加密、维护 API Key/Secret 及时间窗口。
- **仓储抽象层**：全部领域模型已迁移至 GORM，兼容 MySQL/PostgreSQL/SQLite，配合版本化迁移 (`schema_migrations`) 与演示数据脚本快速初始化环境。

## 可用 API 示例
- `GET /api/v1/ping`：健康检查。
- `GET /api/v1/{AdminPrefix}/dashboard`：获取管理后台模块概览（默认 `AdminPrefix=admin`）。
- `GET /api/v1/{AdminPrefix}/nodes`：按分页/过滤获取节点列表。
- `POST /api/v1/{AdminPrefix}/nodes/{id}/kernels/sync`：触发节点与内核的即时同步。
- `GET /api/v1/{AdminPrefix}/subscription-templates`：查看模板列表及变量定义。
- `POST /api/v1/{AdminPrefix}/subscription-templates/{id}/publish`：发布模板并记录版本历史。
- `GET /api/v1/user/subscriptions`：查询当前用户订阅列表。
- `GET /api/v1/user/subscriptions/{id}/preview`：渲染订阅内容并返回预览。
- `GET /api/v1/{AdminPrefix}/plans`：管理套餐列表，支持分页检索与多条件过滤。
- `POST /api/v1/{AdminPrefix}/announcements`：创建并发布面向用户的公告，支持置顶和可见时间窗。
- `GET /api/v1/user/plans`：终端可用套餐列表，返回价格、流量与特性描述。
- `GET /api/v1/user/announcements`：按受众过滤当前有效公告。
- `GET /api/v1/user/account/balance`：查询用户余额与最近流水，默认受第三方安全中间件保护。
- `POST /api/v1/user/orders`：创建套餐订单并自动从余额扣费。
- `GET /api/v1/user/orders`：分页查看订单记录及订单条目。
- `GET /api/v1/user/orders/{id}`：获取单个订单详情及余额快照。
- `POST /api/v1/user/orders/{id}/cancel`：取消待支付或零元订单。
- `GET /api/v1/{AdminPrefix}/security-settings`：查看第三方 API 安全开关与凭据配置。
- `PATCH /api/v1/{AdminPrefix}/security-settings`：更新开关、密钥与时间窗口信息。
- `GET /api/v1/{AdminPrefix}/orders`：按状态、用户、支付方式筛选订单。
- `GET /api/v1/{AdminPrefix}/orders/{id}`：查看订单详情、关联用户及条目。
- `POST /api/v1/{AdminPrefix}/orders/{id}/pay`：手动标记订单已支付，可覆盖支付方式与时间。
- `POST /api/v1/{AdminPrefix}/orders/{id}/cancel`：管理员取消订单，支持记录原因。
- `POST /api/v1/{AdminPrefix}/orders/{id}/refund`：对余额支付订单执行退款并写入流水。

## 项目结构
```
.
├── api/                  # go-zero API 定义（按 shared/auth/admin/user 模块拆分）
├── cmd/api/              # HTTP 服务入口
├── etc/                  # 服务配置示例
├── internal/             # 业务代码（config/handler/logic/svc/types）
├── pkg/                  # 公共库（cache/database/auth 等）
├── docs/                 # 文档占位
├── migrations/           # 数据库迁移脚本占位
└── .github/workflows/    # CI 配置
```

### API 定义与代码生成

API 入口文件位于 `api/znp.api`，该文件通过 `import` 聚合 `api/shared/*.api`、`api/auth/*.api`、`api/admin/*.api`、`api/user/*.api` 等领域定义，便于按模块维护路由、请求/响应结构与复用类型。

使用 [goctl](https://go-zero.dev/docs/goctl/)（1.5+）即可一次性解析上述多文件结构，并输出 `internal/handler`、`internal/logic`、`internal/types` 等目录中的模板代码。项目内提供脚本帮助开发者统一执行：

```bash
./scripts/gen-api.sh            # 默认读取 api/znp.api 并输出到 internal/
./scripts/gen-api.sh api/znp.api build/internal  # 自定义输出目录
```

脚本会先运行 `goctl api format -dir api` 对全部 `.api` 文件进行格式化，然后使用 `goctl api go` 生成最新的 handler/logic/types 代码。生成后请根据实际业务手动调整逻辑层实现，并执行 `go fmt` 与测试校验。

## 快速启动
1. 选择配置文件（示例提供 `etc/znp-sqlite.yaml` 便于本地使用 SQLite）。
2. 初始化数据库（包含 schema 迁移以及可选的演示数据）：
   ```bash
   go run ./cmd/znp migrate --config etc/znp-sqlite.yaml --apply --seed-demo
   ```
   支持通过 `--to <version>` 仅执行至指定迁移版本，默认为最新。
3. 启动一体化服务：
   ```bash
   go run ./cmd/znp serve --config etc/znp-sqlite.yaml
   ```
   若仅需 HTTP，可追加 `--disable-grpc`；亦可继续使用兼容入口：`go run ./cmd/api -f etc/znp-sqlite.yaml`
4. 访问健康检查：`GET http://localhost:8888/api/v1/ping`

## 监控与指标

`Metrics` 配置块控制 Prometheus 指标的导出方式：

```yaml
Metrics:
  Enable: true            # 是否开启指标采集
  Path: /metrics          # 暴露指标的 HTTP 路径
  ListenOn: 0.0.0.0:9100  # 可选：独立监听地址，留空则复用主 HTTP 服务
```

- 当 `ListenOn` 留空时，指标会随主服务一起暴露，例如 `curl http://127.0.0.1:8888/metrics`。
- 指定 `ListenOn` 后，CLI 会额外启动独立的 Prometheus HTTP Server，并在终止或收到 `SIGTERM` 时优雅关闭，可通过 `curl http://127.0.0.1:9100/metrics` 校验。

核心链路已接入以下指标：

- **节点同步**：`znp_node_sync_operations_total`、`znp_node_sync_duration_seconds`，按协议与结果标签区分成功/失败。
- **订单创建**：`znp_order_create_requests_total`、`znp_order_create_duration_seconds`，按支付方式与结果标签统计。

可将对应地址加入 Prometheus `scrape_config` 采集，也可以通过 Grafana 等工具构建可视化看板。

### 默认账户
- 管理员：`admin@example.com` / `P@ssw0rd!`
- 高级会员：`user@example.com` / `P@ssw0rd!`

登录成功后可取得访问令牌（Bearer Token），用于访问 `/api/v1/{AdminPrefix}` 与 `/api/v1/user` 下的受保护接口。

## CLI 工具集
项目内置 `znp` 命令行用于统一管理服务：

- `go run ./cmd/znp serve`：同时启动 HTTP 与 gRPC 服务，可配合 `--disable-grpc` 仅运行 HTTP。
- `go run ./cmd/znp migrate`：执行数据库迁移与种子数据注入，支持 `--to` 指定目标版本。
- `go run ./cmd/znp tools check-config`：校验配置文件并输出摘要。

> 配置提示：`Admin.RoutePrefix` 支持自定义管理端路由前缀；`GRPCServer` 配置块用于控制监听地址、开关及 reflection。

所有子命令均可通过 `--config`/`-f` 指定配置文件，`serve` 亦支持 `--migrate-to` 控制启动前的迁移目标版本。

## 开发工具
- Go 1.22+
- go-zero 1.5+
- GORM 1.25+
- 可选：Redis、Docker、golangci-lint

## 开发规范
项目采用 `main`/`develop` 双分支模型：

- `main`：发布稳定版本，仅接收经过验收的变更。
- `develop`：日常开发主干，功能分支需先合并至此分支再合入 `main`。

更多贡献指引、代码风格及分支策略请参考 [docs/CONTRIBUTING.md](docs/CONTRIBUTING.md)。

## Roadmap
与 xboard 能力对齐的阶段性目标及进度请参阅 [docs/ROADMAP.md](docs/ROADMAP.md)。

## CI/CD
项目提供 GitHub Actions 工作流：常规 CI (`ci.yml`) 执行 `gofmt`、`go vet`、`go test`、`golangci-lint`；发布流水线 (`release.yml`) 会在 `main` 分支与版本标签上构建多平台二进制并上传制品，便于版本管理与分发。

## 许可证
本项目基于 [MIT License](LICENSE) 开源。
