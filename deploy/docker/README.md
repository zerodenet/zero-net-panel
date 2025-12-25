# Docker 部署指南

Zero Network Panel 提供完整的 Docker 部署方案，支持安装向导和多种数据库配置。

## 快速开始

### 方案 1：使用安装向导（推荐首次部署）

#### 步骤 1：构建镜像

```bash
# 构建支持 SQLite 的完整镜像
docker build -t znp:latest -f deploy/docker/Dockerfile.cgo -t znp:cgo .

# 或者构建轻量级镜像（不支持 SQLite，需要 MySQL/PostgreSQL）
docker build -t znp:latest -f deploy/docker/Dockerfile .
```

#### 步骤 2：运行安装向导

```bash
# 创建配置目录
mkdir -p ./deploy/docker/config ./deploy/docker/data

# 运行交互式安装向导
docker run -it --rm \
  -v $(pwd)/deploy/docker/config:/etc/znp \
  -v $(pwd)/deploy/docker/data:/var/lib/znp \
  znp:cgo \
  install --output /etc/znp/znp.yaml
```

安装向导将引导您完成：
- 数据库配置（SQLite/MySQL/PostgreSQL）
- 服务端口和监听地址
- JWT 密钥自动生成
- 管理员账户创建
- 可选功能配置（Prometheus、gRPC）

#### 步骤 3：启动服务

```bash
# 使用生成的配置启动服务
docker run -d \
  --name znp-server \
  -v $(pwd)/deploy/docker/config:/etc/znp:ro \
  -v $(pwd)/deploy/docker/data:/var/lib/znp \
  -p 8888:8888 \
  -p 8890:8890 \
  -p 9100:9100 \
  znp:cgo \
  serve --config /etc/znp/znp.yaml --migrate-to latest
```

### 方案 2：使用 Docker Compose（推荐生产环境）

#### 使用 SQLite（开发/测试）

```bash
cd deploy/docker

# 首次运行：执行安装向导
docker-compose -f docker-compose.sqlite.yml run --rm znp install --output /etc/znp/znp.yaml

# 启动服务
docker-compose -f docker-compose.sqlite.yml up -d

# 查看日志
docker-compose -f docker-compose.sqlite.yml logs -f

# 停止服务
docker-compose -f docker-compose.sqlite.yml down
```

#### 使用 MySQL（生产环境）

```bash
cd deploy/docker

# 编辑 docker-compose.yml，配置 MySQL 连接信息
# 或者使用环境变量覆盖

# 启动 MySQL 和应用
docker-compose up -d

# 首次运行：在容器中执行安装向导
docker-compose exec znp znp install --output /etc/znp/znp.yaml

# 或者预先准备配置文件，放在 config/ 目录
# 然后重启服务
docker-compose restart znp
```

#### 使用 PostgreSQL

```bash
cd deploy/docker

# 使用 PostgreSQL profile 启动
docker-compose --profile postgres up -d postgres znp

# 配置安装时选择 PostgreSQL
```

## 镜像说明

### Dockerfile（轻量级）
- **用途**：生产环境，不需要 SQLite
- **特点**：
  - CGO 禁用，镜像更小
  - 仅支持 MySQL 和 PostgreSQL
  - 适合云端部署
- **构建**：`docker build -f deploy/docker/Dockerfile -t znp:slim .`

### Dockerfile.cgo（完整功能）
- **用途**：需要 SQLite 或开发环境
- **特点**：
  - CGO 启用，支持 SQLite
  - 镜像体积较大
  - 适合本地开发和小型部署
- **构建**：`docker build -f deploy/docker/Dockerfile.cgo -t znp:cgo .`

## 配置管理

### 环境变量

可以通过环境变量覆盖配置文件中的设置：

```bash
docker run -d \
  -e ZNP_DB_DRIVER=mysql \
  -e ZNP_DB_DSN="user:pass@tcp(mysql:3306)/znp?parseTime=true" \
  -e ZNP_JWT_SECRET="your-secret-key" \
  -p 8888:8888 \
  znp:latest
```

### 卷挂载建议

- `/etc/znp` - 配置文件目录
- `/var/lib/znp` - 数据目录（SQLite 数据库文件）
- `/var/log/znp` - 日志目录（可选）

## 常见场景

### 开发环境（SQLite）

```bash
# 快速启动开发环境
cd deploy/docker
docker-compose -f docker-compose.sqlite.yml run --rm znp install --output /etc/znp/znp.yaml
docker-compose -f docker-compose.sqlite.yml up -d
```

### 生产环境（MySQL）

```bash
# 1. 准备 MySQL 数据库
docker-compose up -d mysql

# 2. 等待 MySQL 就绪
docker-compose exec mysql mysqladmin ping

# 3. 运行安装向导或准备配置文件
docker-compose run --rm znp install --output /etc/znp/znp.yaml

# 4. 编辑配置文件，设置 MySQL 连接
# vim deploy/docker/config/znp.yaml

# 5. 启动应用
docker-compose up -d znp
```

### 集群部署

```bash
# 使用外部数据库和负载均衡
docker run -d \
  --name znp-node1 \
  -e ZNP_DB_DSN="user:pass@tcp(db.example.com:3306)/znp" \
  -p 8888:8888 \
  znp:slim \
  serve --config /etc/znp/znp.yaml
```

## 健康检查

```bash
# 容器内健康检查
docker exec znp-server wget -O- http://localhost:8888/api/v1/ping

# 查看容器健康状态
docker inspect --format='{{.State.Health.Status}}' znp-server

# Docker Compose 健康检查
docker-compose ps
```

## 故障排查

### 查看日志

```bash
# Docker
docker logs znp-server

# Docker Compose
docker-compose logs znp
```

### 进入容器

```bash
# Docker
docker exec -it znp-server /bin/bash

# Docker Compose
docker-compose exec znp /bin/bash
```

### 重新运行安装向导

```bash
# 停止服务
docker-compose down

# 清理旧配置
rm -rf deploy/docker/config/*

# 重新运行向导
docker-compose run --rm znp install --output /etc/znp/znp.yaml

# 启动服务
docker-compose up -d
```

### 数据库连接问题

```bash
# 检查数据库容器状态
docker-compose ps

# 测试数据库连接
docker-compose exec mysql mysql -u znp -pznppassword -e "SELECT 1"

# 查看数据库日志
docker-compose logs mysql
```

## 安全建议

1. **配置文件权限**：确保配置文件权限为 600
   ```bash
   chmod 600 deploy/docker/config/znp.yaml
   ```

2. **密钥管理**：使用 Docker secrets 或环境变量管理敏感信息
   ```bash
   docker secret create znp_jwt_secret jwt_secret.txt
   ```

3. **网络隔离**：使用自定义网络隔离容器
   ```yaml
   networks:
     znp-network:
       driver: bridge
   ```

4. **非 root 用户**：镜像已配置为以非 root 用户运行

5. **只读配置**：生产环境挂载配置为只读
   ```bash
   -v /path/to/config:/etc/znp:ro
   ```

## 升级指南

```bash
# 1. 备份数据和配置
docker-compose exec znp tar czf /backup/znp-backup.tar.gz /etc/znp /var/lib/znp

# 2. 拉取新镜像
docker-compose pull

# 3. 重启服务（自动执行数据库迁移）
docker-compose up -d

# 4. 验证升级
docker-compose logs znp | grep "migration"
```

## 监控集成

### Prometheus

```yaml
scrape_configs:
  - job_name: 'znp'
    static_configs:
      - targets: ['znp-server:9100']
```

### Grafana

导入 ZNP Dashboard 模板（如果提供）

## 参考资源

- [安装向导文档](../../docs/installation-wizard.md)
- [快速入门指南](../../docs/getting-started.md)
- [API 文档](../../docs/api-overview.md)
- [主项目 README](../../README.md)

