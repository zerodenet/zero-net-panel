# Zero Network Panel (ZNP)

Zero Network Panel 旨在以 xboard 的功能体系为基线，提供面向节点运营、用户订阅、套餐计费等全栈后端能力。本项目采用 Go 语言与 [go-zero](https://go-zero.dev/) 微服务框架构建，默认以 RESTful API 的方式对外暴露接口，并结合 GORM、可插拔缓存服务以及自动化 CI/CD，支撑后续协议层和运营扩展。

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

## 开发工具
- Go 1.21+
- go-zero 1.5+
- GORM 1.25+
- 可选：Redis、Docker、golangci-lint

## CI/CD
项目提供 GitHub Actions 工作流，包含 `gofmt`、`go vet`、`go test`、`golangci-lint` 四项检查，并预留 Goreleaser/Docker 等自动发布流程的扩展点。

## 许可证
本项目基于 [MIT License](LICENSE) 开源。
