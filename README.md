# Zero Network Panel (ZNP)

Zero Network Panel 旨在以 xboard 的功能体系为基线，提供面向节点运营、用户订阅、套餐计费等全栈后端能力。本项目采用 Go 语言与 [go-zero](https://go-zero.dev/) 微服务框架构建，默认以 RESTful API 的方式对外暴露接口，并结合 GORM、可插拔缓存服务以及自动化 CI/CD，支撑后续协议层和运营扩展。

## 核心模块
- **节点发现 (kernel discovery)**：内置 HTTP 与 gRPC Provider 注册表，可在后台一键触发节点配置同步，确保协议资源与内核保持一致。
- **订阅模板管理**：提供模板 CRUD、版本发布、历史追溯及默认模板切换，变量描述采用 GitHub 风格的分页与字段规范。
- **用户订阅能力**：支持订阅列表查询、模板预览与定制选择，同时输出渲染后的内容、ETag 及内容类型信息，方便前端或客户端下载。
- **仓储抽象层**：节点、模板、订阅均由内存仓储模拟真实数据库行为，后续可平滑替换为 GORM 与实际数据源实现。

## 可用 API 示例
- `GET /api/v1/ping`：健康检查。
- `GET /api/v1/admin/nodes`：按分页/过滤获取节点列表。
- `POST /api/v1/admin/nodes/{id}/kernels/sync`：触发节点与内核的即时同步。
- `GET /api/v1/admin/subscription-templates`：查看模板列表及变量定义。
- `POST /api/v1/admin/subscription-templates/{id}/publish`：发布模板并记录版本历史。
- `GET /api/v1/user/subscriptions`：查询当前用户订阅列表。
- `GET /api/v1/user/subscriptions/{id}/preview`：渲染订阅内容并返回预览。

## 项目结构
```
.
├── api/                  # go-zero API 定义
├── cmd/api/              # HTTP 服务入口
├── etc/                  # 服务配置示例
├── internal/             # 业务代码（config/handler/logic/svc/types）
├── pkg/                  # 公共库（cache/database/auth 等）
├── docs/                 # 文档占位
├── migrations/           # 数据库迁移脚本占位
└── .github/workflows/    # CI 配置
```

## 快速启动
1. 准备配置文件（默认 `etc/znp-api.yaml`）。
2. 启动服务：
   ```bash
   go run ./cmd/api -f etc/znp-api.yaml
   ```
3. 访问健康检查：`GET http://localhost:8888/api/v1/ping`

### 默认账户
- 管理员：`admin@example.com` / `P@ssw0rd!`
- 高级会员：`user@example.com` / `P@ssw0rd!`

登录成功后可取得访问令牌（Bearer Token），用于访问 `/api/v1/admin` 与 `/api/v1/user` 下的受保护接口。

## 开发工具
- Go 1.21+
- go-zero 1.5+
- GORM 1.25+
- 可选：Redis、Docker、golangci-lint

## CI/CD
项目提供 GitHub Actions 工作流，包含 `gofmt`、`go vet`、`go test`、`golangci-lint` 四项检查，并预留 Goreleaser/Docker 等自动发布流程的扩展点。

## 许可证
本项目基于 [MIT License](LICENSE) 开源。
