# 安装向导使用指南

Zero Network Panel 提供了交互式安装向导，帮助您快速完成首次部署配置。

## 功能概述

安装向导 (`znp install`) 将引导您完成以下配置步骤：

1. **数据库配置** - 选择数据库类型（SQLite/MySQL/PostgreSQL）并配置连接参数
2. **服务配置** - 设置 HTTP 服务监听地址和端口
3. **JWT 认证** - 自动生成安全的 JWT 密钥
4. **管理员账户** - 创建初始管理员账户
5. **可选功能** - 配置 Prometheus 监控和 gRPC 服务
6. **自动初始化** - 自动执行数据库迁移并创建管理员用户

## 快速开始

### 运行安装向导

```bash
go run ./cmd/znp install
```

或指定输出配置文件路径：

```bash
go run ./cmd/znp install --output /etc/znp/production.yaml
```

### 交互式安装流程

#### 步骤 1: 数据库配置

向导会提示您选择数据库类型：

```
═══ Step 1: Database Configuration ═══

Select database driver:
  1) SQLite (Development/Testing)
  2) MySQL (Recommended for Production)
  3) PostgreSQL
Enter choice [1-3]: 
```

**选择建议：**
- **SQLite (选项 1)**: 适合开发环境和小规模部署，无需额外安装数据库服务
- **MySQL (选项 2)**: 推荐用于生产环境，性能稳定，支持高并发
- **PostgreSQL (选项 3)**: 适合需要高级功能的场景

根据选择的数据库类型，向导会提供相应的 DSN 示例并要求您输入连接字符串。

**SQLite 示例：**
```
SQLite database file path [file:znp.db?cache=shared&mode=rwc]: /var/lib/znp/data.db
```

**MySQL 示例：**
```
Example DSN for mysql:
  root:password@tcp(127.0.0.1:3306)/znp?parseTime=true&loc=UTC

Enter database DSN: znp_user:secure_password@tcp(db.example.com:3306)/znp?parseTime=true&loc=UTC
```

**PostgreSQL 示例：**
```
Example DSN for postgres:
  host=localhost port=5432 user=znp password=password dbname=znp sslmode=disable

Enter database DSN: host=db.example.com port=5432 user=znp_user password=secure_password dbname=znp_prod sslmode=require
```

#### 步骤 2: 服务配置

配置 HTTP 服务的监听地址和端口：

```
═══ Step 2: Service Configuration ═══

Service host [0.0.0.0]: 0.0.0.0
Service port [8888]: 8888
Kernel HTTP base URL (optional) [http://kernel.local]: http://kernel.internal.local
```

- **Service host**: 通常使用 `0.0.0.0` 监听所有网络接口，生产环境可考虑使用 `127.0.0.1` 配合反向代理
- **Service port**: 默认 8888，确保端口未被占用
- **Kernel HTTP base URL**: 可选，配置内核服务的 HTTP 端点

#### 步骤 3: JWT 认证配置

系统会自动生成安全的 JWT 密钥：

```
═══ Step 3: JWT Authentication Configuration ═══

Generating secure JWT secrets...
✓ JWT secrets generated successfully
```

无需手动输入，向导会生成加密安全的随机密钥。

#### 步骤 4: 管理员账户配置

创建初始管理员账户：

```
═══ Step 4: Admin Account Configuration ═══

Admin email [admin@example.com]: admin@mycompany.com
Admin password (min 8 characters): ********
Confirm admin password: ********

✓ Admin account configuration completed
```

**注意事项：**
- 邮箱必须符合标准格式
- 密码至少 8 个字符
- 两次密码输入必须一致

#### 步骤 5: 可选功能配置

配置 Prometheus 监控和 gRPC 服务：

```
═══ Step 5: Optional Features ═══

Enable Prometheus metrics [Y/n]: y
Use separate port for metrics [Y/n]: y
Metrics port [0.0.0.0:9100]: 9100
Admin route prefix [admin]: admin
Enable gRPC server [Y/n]: y
gRPC listen address [0.0.0.0:8890]: 0.0.0.0:8890

✓ Optional features configured
```

**配置建议：**
- **Prometheus metrics**: 建议启用，便于监控服务运行状态
- **Separate port for metrics**: 生产环境建议使用独立端口
- **gRPC server**: 如需使用 gRPC 协议与内核通信，请启用

#### 步骤 6-8: 自动初始化

向导会自动完成以下操作：

```
═══ Step 6: Saving Configuration ═══

✓ Configuration saved to: etc/znp-installed.yaml

═══ Step 7: Initializing Database ═══

Running database migrations...
✓ Migrations applied: 5 version(s)
  Current schema version: 2024063005

═══ Step 8: Creating Admin User ═══

✓ Admin user created successfully
```

### 安装完成

安装成功后，会显示下一步操作指引：

```
╔════════════════════════════════════════════════════════════════╗
║              Installation Completed Successfully!             ║
╚════════════════════════════════════════════════════════════════╝

✓ Configuration file created
✓ Database initialized
✓ Admin user created

Next steps:
  1. Review the configuration file: etc/znp-installed.yaml
  2. Start the service: go run ./cmd/znp serve --config etc/znp-installed.yaml
  3. Access the API at: http://0.0.0.0:8888/api/v1/ping
  4. Login with: admin@mycompany.com

Thank you for using Zero Network Panel!
```

## 命令行选项

```bash
Usage:
  znp install [flags]

Flags:
  -h, --help              help for install
      --non-interactive   Run in non-interactive mode with defaults (未来支持)
  -o, --output string     Output configuration file path (default "etc/znp-installed.yaml")

Global Flags:
  -f, --config string   Path to configuration file (default "etc/znp-api.yaml")
```

### 指定输出路径

使用 `-o` 或 `--output` 参数指定配置文件保存位置：

```bash
# 保存到自定义路径
go run ./cmd/znp install --output /etc/znp/production.yaml

# 保存到当前目录
go run ./cmd/znp install --output ./my-config.yaml
```

## 启动服务

安装完成后，使用生成的配置文件启动服务：

```bash
# 使用默认配置文件启动
go run ./cmd/znp serve --config etc/znp-installed.yaml

# 启动时自动迁移到最新版本
go run ./cmd/znp serve --config etc/znp-installed.yaml --migrate-to latest

# 仅启动 HTTP 服务（禁用 gRPC）
go run ./cmd/znp serve --config etc/znp-installed.yaml --disable-grpc
```

## 验证安装

### 检查服务健康状态

```bash
curl http://localhost:8888/api/v1/ping
```

预期响应：
```json
{"code":0,"message":"pong","data":null}
```

### 验证配置文件

```bash
go run ./cmd/znp tools check-config --config etc/znp-installed.yaml
```

### 获取管理员令牌

使用创建的管理员账户登录：

```bash
curl -X POST http://localhost:8888/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "admin@mycompany.com",
    "password": "your_password"
  }'
```

## 常见问题

### 1. 数据库连接失败

**问题**: 安装时提示数据库连接失败

**解决方案**:
- 确认数据库服务已启动
- 检查 DSN 连接字符串是否正确
- 对于 MySQL/PostgreSQL，确保数据库已创建且用户有足够权限
- 检查防火墙和网络连接

### 2. 端口已被占用

**问题**: 提示端口 8888 或其他端口已被占用

**解决方案**:
```bash
# 检查端口占用
lsof -i :8888

# 使用其他端口
# 在步骤 2 中输入不同的端口号
```

### 3. 重新运行安装向导

**问题**: 需要重新配置

**解决方案**:
- 删除或备份现有配置文件
- 如果数据库已初始化，考虑清空数据库或使用新数据库
- 重新运行 `znp install` 命令

### 4. 修改已生成的配置

**问题**: 想要修改某些配置项

**解决方案**:
- 直接编辑生成的 YAML 配置文件
- 参考 `etc/znp-api.yaml` 和 `etc/znp-prod.example.yaml` 中的配置示例
- 使用 `znp tools check-config` 验证修改后的配置

## 生产环境建议

### 1. 使用环境变量

对于敏感信息（如密码、密钥），建议使用环境变量：

```bash
export ZNP_DB_PASSWORD="secure_password"
export ZNP_JWT_SECRET="generated_secret"
```

### 2. 配置文件权限

确保配置文件权限合理：

```bash
chmod 600 /etc/znp/production.yaml
chown znp:znp /etc/znp/production.yaml
```

### 3. 使用反向代理

建议在前端使用 Nginx 或 Traefik 等反向代理：

```nginx
server {
    listen 80;
    server_name panel.example.com;
    
    location / {
        proxy_pass http://127.0.0.1:8888;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
}
```

### 4. 启用 HTTPS

生产环境务必启用 HTTPS，可使用 Let's Encrypt 免费证书。

### 5. 定期备份

设置数据库定期备份：

```bash
# 使用项目提供的备份脚本
./scripts/backup-db.sh /backup/znp-$(date +%Y%m%d).sql
```

## 相关文档

- [快速入门指南](getting-started.md)
- [API 文档](api-overview.md)
- [运维手册](operations.md)
- [配置参考](../etc/znp-api.yaml)

## 技术支持

如遇到问题，请查看：
- [项目 README](../README.md)
- [GitHub Issues](https://github.com/Khamxmmermax/zero-net-panel/issues)
- [贡献指南](CONTRIBUTING.md)
