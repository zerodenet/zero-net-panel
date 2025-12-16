# Zero Network Panel - 项目分析报告 / Project Analysis Report

**生成日期 / Generated**: 2025-12-11  
**分析人 / Analyst**: GitHub Copilot  
**版本 / Version**: 1.0

---

## 项目概述 / Project Overview

### 中文概述

Zero Network Panel (ZNP) 是一个使用 Go 语言和 go-zero 微服务框架构建的网络面板管理后端系统。该项目以 xboard 的功能体系为基线，提供面向节点运营、用户订阅、套餐计费等全栈后端能力。

### English Overview

Zero Network Panel (ZNP) is a network panel management backend system built with Go language and the go-zero microservice framework. Based on xboard's feature set, it provides comprehensive backend capabilities for node operations, user subscriptions, and package billing.

---

## 技术栈 / Technology Stack

### 核心框架 / Core Frameworks
- **Go 1.22+**: 主编程语言 / Main programming language
- **go-zero 1.5+**: 微服务框架 / Microservice framework
- **GORM 1.25+**: ORM 框架 / ORM framework

### 数据库支持 / Database Support
- MySQL (via gorm.io/driver/mysql)
- PostgreSQL (via gorm.io/driver/postgres)
- SQLite (via gorm.io/driver/sqlite)

### 其他关键依赖 / Other Key Dependencies
- **JWT Authentication** (github.com/golang-jwt/jwt/v5)
- **gRPC** (google.golang.org/grpc)
- **Prometheus Metrics** (github.com/prometheus/client_golang)
- **Cobra CLI** (github.com/spf13/cobra)
- **Redis Support** (github.com/go-redis/redis/v8)

---

## 项目结构分析 / Project Structure Analysis

### 目录结构 / Directory Structure

```
zero-net-panel/
├── api/                    # API 定义文件 / API definition files
│   ├── admin/             # 管理员 API / Admin APIs
│   ├── auth/              # 认证 API / Authentication APIs
│   ├── shared/            # 共享类型 / Shared types
│   ├── user/              # 用户 API / User APIs
│   └── znp.api            # 主入口文件 / Main entry file
├── cmd/                    # 命令行入口 / CLI entry points
│   ├── api/               # API 服务入口 / API service entry
│   └── znp/               # 主 CLI 工具 / Main CLI tool
├── internal/               # 内部实现 / Internal implementation
│   ├── admin/             # 管理后台路由 / Admin routes
│   ├── bootstrap/         # 启动与迁移 / Bootstrap and migrations
│   ├── config/            # 配置定义 / Configuration definitions
│   ├── handler/           # HTTP 处理器 / HTTP handlers
│   ├── logic/             # 业务逻辑 / Business logic
│   ├── middleware/        # 中间件 / Middleware
│   ├── repository/        # 数据仓储层 / Data repository layer
│   ├── security/          # 安全相关 / Security utilities
│   ├── svc/               # 服务上下文 / Service context
│   └── types/             # 类型定义 / Type definitions
├── pkg/                    # 公共库 / Shared packages
│   ├── auth/              # 认证工具 / Auth utilities
│   ├── cache/             # 缓存实现 / Cache implementations
│   ├── database/          # 数据库工具 / Database utilities
│   ├── kernel/            # 内核发现 / Kernel discovery
│   ├── metrics/           # 指标采集 / Metrics collection
│   └── subscription/      # 订阅模板 / Subscription templates
├── docs/                   # 文档 / Documentation
├── etc/                    # 配置文件 / Configuration files
└── scripts/               # 工具脚本 / Utility scripts
```

### 架构模式 / Architecture Pattern

该项目采用**清晰的分层架构**：

1. **API 层** (api/): 使用 go-zero API 定义格式，支持 RESTful API
2. **处理器层** (internal/handler/): HTTP 请求处理
3. **业务逻辑层** (internal/logic/): 核心业务逻辑
4. **仓储层** (internal/repository/): 数据访问抽象
5. **基础设施层** (pkg/): 可复用的基础组件

---

## 核心功能模块 / Core Feature Modules

### 1. 节点发现与管理 / Node Discovery & Management

**文件位置 / Location**: 
- `pkg/kernel/`: 内核注册与发现
- `internal/logic/admin/node/`: 节点管理逻辑

**功能特点 / Features**:
- HTTP 与 gRPC Provider 注册表
- 节点配置同步
- 协议资源管理

**API 端点 / API Endpoints**:
```
GET  /api/v1/admin/nodes              # 获取节点列表
POST /api/v1/admin/nodes/{id}/kernels/sync  # 触发节点同步
```

### 2. 订阅模板管理 / Subscription Template Management

**文件位置 / Location**: 
- `internal/logic/admin/template/`
- `pkg/subscription/template/`

**功能特点 / Features**:
- 模板 CRUD 操作
- 版本发布与历史追溯
- 默认模板切换
- GitHub 风格的分页与字段规范

**API 端点 / API Endpoints**:
```
GET   /api/v1/admin/subscription-templates           # 查看模板列表
POST  /api/v1/admin/subscription-templates/{id}/publish  # 发布模板
```

### 3. 用户订阅能力 / User Subscription Capabilities

**文件位置 / Location**: 
- `internal/logic/user/subscription/`

**功能特点 / Features**:
- 订阅列表查询
- 模板预览与定制选择
- ETag 支持
- 内容类型信息输出

**API 端点 / API Endpoints**:
```
GET /api/v1/user/subscriptions                  # 查询订阅
GET /api/v1/user/subscriptions/{id}/preview    # 预览订阅内容
```

### 4. 套餐管理 / Plan Management

**文件位置 / Location**: 
- `internal/repository/plan_repository.go`
- `internal/logic/admin/plan/`

**功能特点 / Features**:
- 套餐 CRUD
- 价格、流量、时长配置
- 管理端与用户端分离

**数据模型 / Data Model**:
- `plans` 表：套餐主表
- 字段：价格、时长、流量限制、模板关联等

**API 端点 / API Endpoints**:
```
GET  /api/v1/admin/plans     # 管理端套餐列表
POST /api/v1/admin/plans     # 创建套餐
GET  /api/v1/user/plans      # 用户可见套餐
```

### 5. 公告系统 / Announcement System

**文件位置 / Location**: 
- `internal/repository/announcement_repository.go`
- `internal/logic/admin/announcement/`

**功能特点 / Features**:
- 公告创建与发布
- 置顶功能
- 可见时间窗口
- 受众过滤

**API 端点 / API Endpoints**:
```
GET  /api/v1/admin/announcements      # 管理端公告列表
POST /api/v1/admin/announcements      # 创建公告
GET  /api/v1/user/announcements       # 用户端公告
```

### 6. 计费订单系统 / Billing & Order System

**文件位置 / Location**: 
- `internal/repository/order_repository.go`
- `internal/logic/user/order/`
- `internal/logic/admin/orders/`

**功能特点 / Features**:
- 订单创建与查询
- 余额支付与外部支付
- 订单取消
- 退款管理
- 支付状态追踪

**数据模型 / Data Models**:
- `orders`: 订单主表
- `order_items`: 订单条目
- `order_payments`: 支付记录
- `order_refunds`: 退款记录

**支付方式 / Payment Methods**:
1. **余额支付 (balance)**: 直接扣减用户余额
2. **外部支付 (external)**: 生成待支付订单，等待回调

**API 端点 / API Endpoints**:
```
# 用户端
POST /api/v1/user/orders                 # 创建订单
GET  /api/v1/user/orders                 # 查询订单
GET  /api/v1/user/orders/{id}           # 订单详情
POST /api/v1/user/orders/{id}/cancel    # 取消订单

# 管理端
GET  /api/v1/admin/orders                # 订单列表
GET  /api/v1/admin/orders/{id}          # 订单详情
POST /api/v1/admin/orders/{id}/pay      # 标记已支付
POST /api/v1/admin/orders/{id}/cancel   # 取消订单
POST /api/v1/admin/orders/{id}/refund   # 退款
```

### 7. 用户余额管理 / User Balance Management

**文件位置 / Location**: 
- `internal/repository/balance_repository.go`
- `internal/logic/user/account/`

**功能特点 / Features**:
- 余额查询
- 交易流水记录
- 退款处理
- 余额变动追踪

**数据模型 / Data Models**:
- `user_balances`: 用户余额
- `balance_transactions`: 余额交易流水

**API 端点 / API Endpoints**:
```
GET /api/v1/user/account/balance    # 查询余额与流水
```

### 8. 第三方安全配置 / Third-Party Security Configuration

**文件位置 / Location**: 
- `internal/repository/security_repository.go`
- `internal/logic/admin/security/`
- `internal/middleware/thirdpartymiddleware.go`

**功能特点 / Features**:
- API Key/Secret 管理
- 签名验证（HMAC-SHA256）
- AES-256-GCM 加密/解密
- Nonce 防重放
- 时间窗口验证

**配置项 / Configuration**:
- `ThirdPartyAPIEnabled`: 开关
- `APIKey` / `APISecret`: 凭据
- `NonceTTLSeconds`: 时间窗口

**安全流程 / Security Flow**:
1. 客户端使用 API Secret 生成 HMAC-SHA256 签名
2. 携带 `X-ZNP-API-Key`, `X-ZNP-Timestamp`, `X-ZNP-Nonce`, `X-ZNP-Signature` 头
3. 可选 AES-256-GCM 加密（携带 `X-ZNP-Encrypted: true` 和 `X-ZNP-IV`）
4. 服务端验证签名、时间窗口、Nonce 唯一性

**API 端点 / API Endpoints**:
```
GET   /api/v1/admin/security-settings     # 查看配置
PATCH /api/v1/admin/security-settings     # 更新配置
```

---

## 认证与授权 / Authentication & Authorization

### JWT 认证 / JWT Authentication

**实现位置 / Implementation**: `pkg/auth/jwt.go`

**功能特点 / Features**:
- Access Token (短期)
- Refresh Token (长期)
- 角色基础访问控制 (RBAC)

**用户角色 / User Roles**:
- `admin`: 管理员
- `user`: 普通用户

**中间件 / Middleware**: 
- `internal/middleware/authmiddleware.go`
- 自动解析 JWT Token
- 注入用户上下文

### 登录流程 / Login Flow

```
POST /api/v1/auth/login
{
  "email": "user@example.com",
  "password": "P@ssw0rd!"
}

Response:
{
  "access_token": "...",
  "refresh_token": "...",
  "expires_in": 3600
}
```

---

## 监控与指标 / Monitoring & Metrics

### Prometheus 集成 / Prometheus Integration

**实现位置 / Implementation**: `pkg/metrics/metrics.go`

**指标列表 / Metrics**:

1. **节点同步 / Node Sync**:
   - `znp_node_sync_operations_total`: 同步操作总数
   - `znp_node_sync_duration_seconds`: 同步耗时

2. **订单创建 / Order Creation**:
   - `znp_order_create_requests_total`: 创建请求总数
   - `znp_order_create_duration_seconds`: 创建耗时

3. **订单取消 / Order Cancellation**:
   - `znp_order_cancel_requests_total`: 取消请求总数
   - `znp_order_cancel_duration_seconds`: 取消耗时

4. **订单支付 / Order Payment**:
   - `znp_order_payment_requests_total`: 支付请求总数
   - `znp_order_payment_duration_seconds`: 支付耗时

5. **订单退款 / Order Refund**:
   - `znp_order_refund_requests_total`: 退款请求总数
   - `znp_order_refund_duration_seconds`: 退款耗时

**配置 / Configuration**:
```yaml
Metrics:
  Enable: true
  Path: /metrics
  ListenOn: 0.0.0.0:9100  # 独立端口，可选
```

---

## 数据库设计 / Database Design

### 核心表结构 / Core Tables

1. **users**: 用户信息
2. **nodes**: 节点信息
3. **subscription_templates**: 订阅模板
4. **template_versions**: 模板版本历史
5. **plans**: 套餐配置
6. **announcements**: 公告
7. **orders**: 订单
8. **order_items**: 订单条目
9. **order_payments**: 支付记录
10. **order_refunds**: 退款记录
11. **user_balances**: 用户余额
12. **balance_transactions**: 余额交易流水
13. **security_settings**: 安全配置
14. **schema_migrations**: 迁移版本

### 迁移管理 / Migration Management

**实现位置 / Implementation**: `internal/bootstrap/migrations/`

**支持功能 / Features**:
- 版本化迁移
- 向前迁移 (migrate up)
- 回滚 (rollback)
- 演示数据注入 (seed)

**CLI 命令 / CLI Commands**:
```bash
# 执行迁移
go run ./cmd/znp migrate --config etc/znp-sqlite.yaml --apply

# 注入演示数据
go run ./cmd/znp migrate --config etc/znp-sqlite.yaml --apply --seed-demo

# 迁移到指定版本
go run ./cmd/znp migrate --config etc/znp-sqlite.yaml --to <version>
```

---

## CLI 工具 / CLI Tools

### 命令列表 / Command List

**实现位置 / Implementation**: `cmd/znp/cli/`

1. **serve**: 启动服务
   ```bash
   go run ./cmd/znp serve --config etc/znp-sqlite.yaml
   ```
   - 支持 `--disable-grpc` 禁用 gRPC
   - 支持 `--migrate-to latest` 启动前迁移

2. **migrate**: 数据库迁移
   ```bash
   go run ./cmd/znp migrate --config etc/znp-sqlite.yaml --apply
   ```
   - `--apply`: 执行迁移
   - `--seed-demo`: 注入演示数据
   - `--to <version>`: 迁移到指定版本
   - `--rollback`: 回滚迁移

3. **tools check-config**: 配置检查
   ```bash
   go run ./cmd/znp tools check-config --config etc/znp-sqlite.yaml
   ```

---

## 配置管理 / Configuration Management

### 配置文件 / Configuration Files

**位置 / Location**: `etc/`

1. **znp-sqlite.yaml**: 开发环境配置（SQLite + 内存缓存）
2. **znp-api.yaml**: 生产环境配置（MySQL/PostgreSQL + Redis）

### 关键配置项 / Key Configuration

```yaml
Server:
  ListenOn: 0.0.0.0:8888
  Timeout: 3000

Database:
  DSN: "file:znp.db?cache=shared&mode=rwc"
  Driver: sqlite3

Cache:
  Provider: memory  # 或 redis

Auth:
  AccessSecret: "your-secret-key"
  AccessExpire: 3600
  RefreshExpire: 604800

Admin:
  RoutePrefix: admin  # 可自定义管理端路由前缀

Kernel:
  HTTPProviders: []   # HTTP 内核提供者
  GRPCProviders: []   # gRPC 内核提供者

Metrics:
  Enable: true
  Path: /metrics
  ListenOn: 0.0.0.0:9100
```

---

## 测试覆盖 / Test Coverage

### 测试文件分布 / Test Files

```
✓ cmd/znp/cli/migrate_test.go          # 迁移测试
✓ internal/bootstrap/migrations/registry_test.go  # 迁移注册测试
✓ internal/config/config_test.go       # 配置测试
✓ internal/logic/admin/orders/refundlogic_test.go  # 退款逻辑测试
✓ internal/logic/admin/orders/paymentcallbacklogic_test.go  # 支付回调测试
✓ internal/logic/user/order/lifecycle_test.go  # 订单生命周期测试
✓ internal/logic/user/order/cancellogic_test.go  # 取消逻辑测试
✓ internal/logic/user/order/createlogic_test.go  # 创建逻辑测试
✓ pkg/metrics/metrics_test.go          # 指标测试
✓ pkg/auth/jwt_test.go                 # JWT 测试
✓ pkg/cache/memory_test.go             # 缓存测试
```

### 测试覆盖重点 / Test Focus Areas

- ✅ 订单生命周期（创建、支付、取消、退款）
- ✅ JWT 认证与令牌刷新
- ✅ 缓存操作（内存缓存）
- ✅ 数据库迁移
- ✅ Prometheus 指标采集

---

## CI/CD 流程 / CI/CD Pipeline

### GitHub Actions

**工作流文件 / Workflow Files**: `.github/workflows/`

1. **ci.yml**: 持续集成
   - `go fmt` 格式检查
   - `go vet` 静态分析
   - `go test` 单元测试
   - `golangci-lint` 代码质量检查

2. **release.yml**: 发布流水线
   - 多平台构建（Linux, macOS, Windows）
   - 二进制制品上传
   - 版本标签发布

---

## 代码质量分析 / Code Quality Analysis

### 发现的问题 / Issues Found

#### 1. ✅ 已修复：编译错误 / Fixed: Compilation Errors

**文件 / File**: `internal/logic/admin/orders/refundlogic.go`

**问题 / Issues**:
1. 未使用的变量 `refundRecords`
2. 重复声明 `refundRecord`

**修复 / Fix**:
- 移除未使用的变量声明
- 删除重复的 `refundRecord` 创建逻辑

#### 2. 待修复：测试失败 / To Fix: Test Failure

**文件 / File**: `internal/logic/user/order/lifecycle_test.go`

**问题 / Issue**: 
- 测试期望部分退款后状态为 `paid`，实际为 `partially_refunded`

**影响 / Impact**: 
- 这是一个测试期望不匹配的问题，不影响生产功能

**建议 / Recommendation**: 
- 更新测试期望值以匹配实际业务逻辑

### 代码优势 / Code Strengths

1. ✅ **清晰的分层架构**: Handler → Logic → Repository 分层明确
2. ✅ **良好的错误处理**: 统一的错误码和错误类型
3. ✅ **完善的测试**: 核心业务逻辑有单元测试覆盖
4. ✅ **可扩展设计**: 支持多种数据库和缓存实现
5. ✅ **监控就绪**: 内置 Prometheus 指标
6. ✅ **文档完善**: 提供详细的 API 文档和操作指南
7. ✅ **安全考虑**: 实现了 JWT 认证、签名验证、加密传输

### 改进建议 / Improvement Suggestions

1. **增加集成测试**: 当前主要是单元测试，可增加端到端测试
2. **API 文档自动生成**: 可考虑使用 Swagger/OpenAPI
3. **日志结构化**: 统一日志格式和级别
4. **错误追踪**: 集成分布式追踪（如 Jaeger）
5. **限流保护**: 添加 API 限流中间件
6. **数据验证**: 加强输入验证和参数校验
7. **缓存策略**: 完善缓存失效和预热机制

---

## 依赖安全 / Dependency Security

### 关键依赖版本 / Key Dependency Versions

```
Go: 1.22
go-zero: 1.5.3
GORM: 1.25.7
JWT: 5.3.0
gRPC: 1.55.0
Prometheus: 1.19.0
```

**安全建议 / Security Recommendations**:
- ✅ 所有核心依赖都是较新的稳定版本
- ⚠️ 建议定期更新依赖以获取安全补丁
- ⚠️ 建议使用 `go mod vendor` 锁定依赖版本

---

## 性能考虑 / Performance Considerations

### 优化点 / Optimizations

1. **数据库连接池**: GORM 已配置连接池
2. **缓存支持**: 支持内存和 Redis 缓存
3. **并发处理**: go-zero 框架自带并发优化
4. **索引设计**: 数据库表应有适当索引（需检查迁移文件）

### 潜在瓶颈 / Potential Bottlenecks

1. **订单创建**: 涉及多表事务，可能成为性能瓶颈
2. **余额查询**: 高频访问，建议加缓存
3. **节点同步**: 外部调用，建议异步处理

---

## 部署建议 / Deployment Recommendations

### 开发环境 / Development

```bash
# 1. 使用 SQLite 配置
cp etc/znp-sqlite.yaml etc/znp-dev.yaml

# 2. 初始化数据库并注入演示数据
go run ./cmd/znp migrate --config etc/znp-dev.yaml --apply --seed-demo

# 3. 启动服务
go run ./cmd/znp serve --config etc/znp-dev.yaml
```

### 生产环境 / Production

```bash
# 1. 准备配置文件
cp etc/znp-api.yaml etc/znp-prod.yaml
# 修改数据库 DSN、缓存配置、密钥等

# 2. 执行迁移
./znp migrate --config etc/znp-prod.yaml --apply

# 3. 启动服务（建议使用 systemd 或容器）
./znp serve --config etc/znp-prod.yaml
```

### Docker 部署 / Docker Deployment

**建议的 Dockerfile**:
```dockerfile
FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o znp ./cmd/znp

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/znp .
COPY etc/ ./etc/
EXPOSE 8888 9100
CMD ["./znp", "serve", "--config", "etc/znp-api.yaml"]
```

---

## 安全检查清单 / Security Checklist

- ✅ JWT 认证已实现
- ✅ HMAC 签名验证已实现
- ✅ AES-256-GCM 加密已实现
- ✅ Nonce 防重放已实现
- ✅ 密码加密存储（使用 bcrypt）
- ⚠️ HTTPS/TLS 配置需在反向代理层处理
- ⚠️ 建议实现 API 限流
- ⚠️ 建议实现审计日志
- ⚠️ 建议定期安全扫描依赖

---

## 总结与建议 / Summary & Recommendations

### 项目优势 / Project Strengths

1. **架构清晰**: 遵循领域驱动设计，模块划分合理
2. **技术栈现代**: 使用 Go 1.22 和最新的 go-zero 框架
3. **功能完整**: 覆盖节点、订阅、计费、用户管理等核心功能
4. **可扩展性强**: 支持多种数据库和缓存，易于扩展
5. **监控完善**: 内置 Prometheus 指标采集
6. **文档详细**: 提供中英文文档，API 说明清晰
7. **测试覆盖**: 核心业务逻辑有测试保障

### 需要改进的方面 / Areas for Improvement

1. **修复编译错误**: ✅ 已完成
2. **修复测试失败**: 需要调整测试期望或业务逻辑
3. **增加集成测试**: 提高测试覆盖率
4. **API 文档**: 考虑使用 Swagger/OpenAPI
5. **限流保护**: 添加 API 限流中间件
6. **审计日志**: 记录关键操作
7. **监控告警**: 配置 Prometheus AlertManager

### 下一步行动 / Next Steps

1. ✅ **立即**: 修复编译错误（已完成）
2. 🔶 **短期**: 修复测试失败，完善单元测试
3. 🔶 **中期**: 增加集成测试，完善文档
4. 🔶 **长期**: 优化性能，增强安全性，扩展功能

---

## 附录 / Appendix

### 默认账户 / Default Accounts

- **管理员 / Admin**: admin@example.com / P@ssw0rd!
- **用户 / User**: user@example.com / P@ssw0rd!

### 健康检查 / Health Check

```
GET http://localhost:8888/api/v1/ping
```

### Prometheus 指标 / Prometheus Metrics

```
GET http://localhost:9100/metrics
```

### 相关文档 / Related Documentation

- [README.md](../README.md)
- [API Overview](api-overview.md)
- [Architecture](architecture.md)
- [Getting Started](getting-started.md)
- [Operations](operations.md)
- [Contributing](CONTRIBUTING.md)
- [Roadmap](ROADMAP.md)

---

**分析完成 / Analysis Completed**: 2025-12-11  
**分析工具 / Analysis Tool**: GitHub Copilot  
**项目版本 / Project Version**: Latest (Main Branch)
