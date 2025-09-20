# 快速入门指南

本指南帮助开发者与运维人员在本地或集成环境中快速启动 Zero Network Panel（ZNP），覆盖依赖准备、数据库初始化、常见 CLI 操作以及 HTTP 接口调试示例。

## 环境依赖

| 组件 | 版本建议 | 说明 |
| ---- | -------- | ---- |
| 操作系统 | Linux/macOS/WSL2 | 推荐使用带有 systemd/launchd 的发行版，以便托管 ZNP 服务进程。 |
| Go | 1.22+ | 构建与运行 ZNP 的必备环境，升级后请执行 `go env -w GONOSUMDB=off` 以避免拉取失败。 |
| 数据库 | SQLite 3.39+ / MySQL 8.0+ / PostgreSQL 14+ | SQLite 适合本地开发；测试/生产推荐使用 MySQL 或 PostgreSQL，需启用时区与外键支持。 |
| Redis（可选） | 6.2+ | 如需替换默认内存缓存，可在配置文件中启用 Redis Provider。 |
| goctl（可选） | 1.5+ | 用于在修改 `.api` 文件后重新生成 handler/logic/types。 |

### 依赖安装示例

```bash
# 安装 Go（以 macOS Homebrew 为例）
brew install go

# 安装 goctl 及 lint 工具（可选）
go install github.com/zeromicro/go-zero/tools/goctl@latest
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
```

如果使用 MySQL，请确保创建 `znp` 数据库并授予账户 DDL、DML 权限：

```sql
CREATE DATABASE znp DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci;
CREATE USER 'znp'@'%' IDENTIFIED BY 'StrongPassword!';
GRANT ALL PRIVILEGES ON znp.* TO 'znp'@'%';
FLUSH PRIVILEGES;
```

## 配置文件准备

项目提供两份示例配置，可作为不同部署阶段的模板：

| 场景 | 推荐模板 | 数据库 | 额外建议 |
| ---- | -------- | ------ | -------- |
| 开发环境 | `etc/znp-sqlite.yaml` | SQLite | 开启 `--seed-demo` 以注入演示数据；`Metrics.ListenOn` 可保持空值复用主服务端口。 |
| 测试/集成环境 | `etc/znp-api.yaml` | MySQL | 自定义 `Database.DSN` 指向独立测试库，建议为 CI/CD 准备专用凭据。 |
| 生产环境 | 基于 `etc/znp-api.yaml` 衍生 | MySQL 或 PostgreSQL | 调整 `Auth` 密钥、缓存 Provider 及 `Kernel` 地址，启用 HTTPS/TLS 与只读副本。 |

将配置副本放置在安全位置（如 `/etc/znp/production.yaml`），并通过 `--config`/`-f` 参数显式指定。

## 初始化数据库

1. 确认配置中的数据库可以连通，并已创建目标库。
2. 执行迁移并可选注入演示数据：
   ```bash
   go run ./cmd/znp migrate --config etc/znp-sqlite.yaml --apply --seed-demo
   ```
   常见参数：
   - `--to <version>`：仅执行至指定迁移版本号，例如 `--to 2024063001`。
   - `--rollback`：回滚到较早版本，通常用于演练或问题排查。
3. 迁移完成后，可在日志中看到如下输出，表示状态同步成功：
   ```text
   applying migrations: before=2024063001 target=2024063003 versions=[2024063002 2024063003]
   seed demo data completed
   ```

## 启动服务

启动一体化服务（HTTP + gRPC）：

```bash
go run ./cmd/znp serve --config etc/znp-sqlite.yaml --migrate-to latest
```

- 仅需 HTTP 时，可追加 `--disable-grpc`。
- 如需提前验证配置，可运行 `go run ./cmd/znp tools check-config --config etc/znp-sqlite.yaml`，输出将列出数据库驱动、缓存 Provider 等摘要。

## 常见 CLI 操作示例

| 命令 | 用途 | 示例输出 |
| ---- | ---- | -------- |
| `go run ./cmd/znp migrate --config <file> --status` | 查看已执行/待执行迁移版本 | `pending versions: [] executed: [2024063001 2024063002]` |
| `go run ./cmd/znp tools check-config --config <file>` | 校验配置并打印关键信息 | `database=sqlite cache=memory grpc=disabled (flag) metrics=http://0.0.0.0:8888/metrics` |
| `go run ./cmd/znp serve --config <file> --graceful-timeout 10s` | 启动服务并自定义优雅退出时间 | `http listen on 0.0.0.0:8888 grpc disabled (flag)` |

## HTTP 调试示例

### 健康检查

```bash
curl http://127.0.0.1:8888/api/v1/ping
```

示例响应：

```json
{"code":0,"message":"pong","data":null}
```

### 创建订单

1. 通过登录接口获取 Bearer Token（示例中的 `<ACCESS_TOKEN>` 需替换为实际值）。
2. 调用用户订单创建接口：

```bash
curl -X POST http://127.0.0.1:8888/api/v1/user/orders \
  -H "Authorization: Bearer <ACCESS_TOKEN>" \
  -H "Content-Type: application/json" \
  -d '{
    "planId": 10001,
    "couponCode": "",
    "autoRenew": true
  }'
```

示例响应：

```json
{
  "code": 0,
  "message": "OK",
  "data": {
    "orderId": "ORD-20240630-001",
    "status": "PAID",
    "amount": 29.9,
    "currency": "CNY",
    "items": [
      {
        "name": "Pro-30",
        "quantity": 1,
        "unitPrice": 29.9,
        "traffic": "300 GB"
      }
    ]
  }
}
```

### 节点同步

```bash
curl -X POST http://127.0.0.1:8888/api/v1/admin/nodes/42/kernels/sync \
  -H "Authorization: Bearer <ADMIN_TOKEN>"
```

示例响应：

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

更多端到端流程、错误码与排障建议可参考 [docs/api-overview.md](api-overview.md) 与 [docs/operations.md](operations.md)。
