# Docker 最小示例

## 构建
```bash
docker build -t znp:latest -f deploy/docker/Dockerfile .
```

## 运行（无 SQLite）
- 准备生产配置 `/path/to/znp.yaml`（可从镜像内的 `/etc/znp/znp.yaml.sample` 复制并修改）。
- 启动容器（默认禁用 SQLite，依赖外部数据库）：
```bash
docker run --rm \
  -v /path/to/znp.yaml:/etc/znp/znp.yaml:ro \
  -p 8888:8888 \
  --name znp \
  znp:latest \
  serve --config /etc/znp/znp.yaml --migrate-to latest
```

## 运行（需要 SQLite）
- 使用 CGO 版镜像（体积更大但支持 SQLite）：
```bash
docker build -t znp:cgo -f deploy/docker/Dockerfile.cgo .
docker run --rm \
  -v /path/to/znp.yaml:/etc/znp/znp.yaml:ro \
  -p 8888:8888 \
  --name znp \
  znp:cgo \
  serve --config /etc/znp/znp.yaml --migrate-to latest
```

## 健康检查
- HTTP 探针：`curl -f http://127.0.0.1:8888/api/v1/ping`（容器外部端口视映射而定）。
- 日志：容器日志即可；如需文件日志，请挂载宿主机目录并在配置中指定路径。

> 说明：Dockerfile 使用 `CGO_ENABLED=0`，因此不包含 SQLite 驱动；请使用 MySQL/PostgreSQL。
