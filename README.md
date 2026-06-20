# OpsWeaver

OpsWeaver 是面向运维、研发和测试团队的工程智能平台。本阶段提供三个 Go 服务、双逻辑 PostgreSQL 数据库、Redis/Asynq 队列、安全配置、健康检查和 Prometheus 指标基础。

## 前置条件

- Go 1.25 或更高版本
- Docker 与 Docker Compose
- `curl`

## 本地配置

Compose 使用仅限本地开发的默认服务令牌和主密钥。可在启动前通过环境变量覆盖：

```bash
export INTERNAL_SERVICE_TOKEN='replace-with-a-random-token'
export MASTER_KEY_BASE64='replace-with-a-base64-encoded-32-byte-key'
```

其他可用环境变量见 `.env.example`。不要在生产环境使用仓库中的开发默认值。

## 构建和启动

```bash
docker compose -f deploy/docker-compose.yml config
docker compose -f deploy/docker-compose.yml build
docker compose -f deploy/docker-compose.yml up -d --wait
docker compose -f deploy/docker-compose.yml ps
```

## 数据库迁移

```bash
export OPSWEAVER_SERVER_DATABASE_DSN='postgres://opsweaver:opsweaver@localhost:5432/opsweaver_server_db?sslmode=disable'
export OPSWEAVER_GATEWAY_DATABASE_DSN='postgres://opsweaver:opsweaver@localhost:5432/opsweaver_gateway_db?sslmode=disable'

go run ./cmd/migrate -db server -command up
go run ./cmd/migrate -db gateway -command up
```

查询迁移版本：

```bash
go run ./cmd/migrate -db server -command version
go run ./cmd/migrate -db gateway -command version
```

## 服务端点

| 服务 | 地址 |
|---|---|
| opsweaver-server | `http://localhost:8080` |
| opsweaver-worker | `http://localhost:8081` |
| opsweaver-gateway | `http://localhost:8082` |

每个服务都提供 `/healthz`、`/readyz` 和 `/metrics`：

```bash
for port in 8080 8081 8082; do
  curl --fail "http://localhost:${port}/healthz"
  curl --fail "http://localhost:${port}/readyz"
  curl --fail "http://localhost:${port}/metrics"
done
```

## 日志与停止

查看日志：

```bash
docker compose -f deploy/docker-compose.yml logs -f
```

停止服务并保留数据：

```bash
docker compose -f deploy/docker-compose.yml down
```

停止服务并删除本地数据卷：

```bash
docker compose -f deploy/docker-compose.yml down -v
```

## 本地 Go 验证

```bash
go test ./...
go vet ./...
```
