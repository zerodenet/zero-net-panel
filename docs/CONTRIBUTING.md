# 贡献指南

为了保持 Zero Network Panel 的演进节奏与质量，请在提交新特性或缺陷修复前阅读以下约定。

## 分支模型

项目采用 Git Flow 的简化形式：

- `main`：稳定发布分支，仅在发布版本或热修复完成验收后更新。
- `develop`：默认开发主干。所有功能分支需从 `develop` 派生，并在代码评审通过后合并回 `develop`。需要发布时再从 `develop` 向 `main` 发起合并请求。
- 功能分支命名建议使用 `feature/<topic>` 或 `fix/<topic>`，并保持提交粒度清晰。

## 提交前检查清单

1. 执行 `gofmt -w .`, `go vet ./...`, `go test ./...`, `golangci-lint run`，确保本地环境与 CI 检查一致。
2. 对涉及数据库的变更，确保编写或更新迁移脚本，并使用 `go run ./cmd/znp migrate --apply` 验证迁移可重复执行。
3. 若引入新的 CLI 子命令或配置项，请同步更新文档及示例配置。
4. 提交信息需包含变更目的、关键实现与可能影响。

## 数据库与迁移

仓储层基于 GORM，所有结构变更都需通过迁移实现。推荐流程：

1. 更新模型结构（位于 `internal/repository`）。
2. 编写或更新迁移与种子逻辑（`internal/bootstrap`）。
3. 在本地运行：
   ```bash
   go run ./cmd/znp migrate --config etc/znp-sqlite.yaml --apply
   go run ./cmd/znp migrate --config etc/znp-sqlite.yaml --seed-demo
   ```
   确认迁移与种子数据可以幂等执行。

## 代码风格

- 遵循 Go 官方格式化与 linter 规范。
- 避免将业务逻辑散落在 handler 层，优先封装至 logic、svc 与 repository。
- 对外暴露的公共方法需编写注释，描述行为与返回值。

## API 定义与 goctl 生成

- `api/znp.api` 作为入口文件，按 `shared/`、`auth/`、`admin/`、`user/` 等子目录拆分接口与复用类型，可根据领域独立维护。
- 运行 `./scripts/gen-api.sh` 即可格式化全部 `.api` 文件并调用 `goctl api go` 重新生成 `internal/handler`、`internal/logic`、`internal/types` 模板代码。脚本允许通过参数覆盖入口文件或输出目录，方便在自定义路径上预览生成结果。
- 生成后的 handler/logic 需要根据实际业务补充实现，提交前请执行 `gofmt -w .` 与 `go test ./...`。

## 问题反馈

在提交 Issue 时，请尽可能提供以下信息：

- 操作系统与 Go 版本；
- 使用的配置文件片段；
- 复现步骤与期望行为；
- 相关日志或错误堆栈。

感谢你的贡献！
