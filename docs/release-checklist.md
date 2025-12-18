# 发布前检查清单（v0.1.0-beta）

适用于首个可对接前端的演示/验证版本，不包含内核对接。

## 版本与范围
- 目标版本：`v0.1.0-beta`。
- 范围：API（登录/刷新、套餐、公告、订阅、订单/退款、余额查询、节点同步、模板发布），CLI（serve/migrate/tools），轻量运维脚本（探活、备份）。

## 预发布验证
- 配置：基于 `etc/znp-prod.example.yaml` 生成最终配置，替换 Auth/DB/缓存/Webhook 秘钥，确认 Admin.RoutePrefix、Webhook 校验策略。
- 迁移：`go run ./cmd/znp migrate --config <prod.yaml> --apply --to latest`，并在预生产演练 `--rollback --to <prev>`。
- 测试：`CGO_ENABLED=1 go test ./...`（需要 sqlite3 驱动）；可选 `golangci-lint run` 如本地有安装。
- 自检：`go run ./cmd/znp tools check-config --config <prod.yaml>`，确认 HTTP/GRPC/DB/缓存/Webhook/管理入口摘要。
- 健康检查：本地/预发跑 `go run ./cmd/znp serve --config <prod.yaml> --migrate-to latest`，验证 `/api/v1/ping` 与关键 API（套餐列表、下单、退款、公告）。

## 构建与制品
- 二进制：`make build`（输出 `bin/znp`，包含版本/commit/buildDate）。
- Docker（无 SQLite）：`docker build -t znp:latest -f deploy/docker/Dockerfile .`
- Docker（含 SQLite/CGO）：`docker build -t znp:cgo -f deploy/docker/Dockerfile.cgo .`
- 附带文件：`etc/znp-prod.example.yaml`、`deploy/systemd/znp.service`、`deploy/docker/README.md`、`scripts/healthcheck.sh`、`scripts/backup-db.sh`。

## 打 tag 与发布
- 生成 changelog（可在 PR/发布说明中概述）并创建 tag：`git tag v0.1.0-beta && git push origin v0.1.0-beta`。
- GitHub Actions 自动发布：推送 tag 后 `.github/workflows/release.yml` 会构建多平台二进制，生成自动 Release Notes，并上传 dist/** 到对应 Release。
- 在发布页补充配置示例、已知限制（未接入真实支付网关、未实现注册/找回/管理员 CRUD、无 CORS/限流配置项）。

## 发布后验证
- 生产环境迁移并重启后，验证 `/api/v1/ping`、套餐/订阅/下单/退款/余额核心链路；确认 `/metrics` 可访问。
- 备份与监控：配置定时执行 `scripts/backup-db.sh`，部署探活脚本或探针；观察日志/指标，确保无频繁错误。
