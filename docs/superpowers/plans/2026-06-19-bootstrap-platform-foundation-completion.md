# Bootstrap Platform Foundation Completion Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 完成 `bootstrap-platform-foundation` 的三服务 Compose 运行、开发文档和最终验证，并提交 PR。

**Architecture:** 使用一个多阶段 Dockerfile，通过 `SERVICE` 构建参数生成三个 Go 服务镜像；Compose 负责 PostgreSQL、Redis 和三个服务的依赖及健康检查。敏感配置只通过环境变量传入，非敏感容器网络配置保存在独立 YAML 中。

**Tech Stack:** Go 1.25、Docker、Docker Compose、PostgreSQL 16、Redis 7、Asynq、Prometheus HTTP endpoints、OpenSpec。

---

### Task 1: 增加三服务 Compose 运行能力

**Files:**
- Create: `Dockerfile`
- Create: `.dockerignore`
- Create: `deploy/config/compose.yaml`
- Modify: `deploy/docker-compose.yml`

- [ ] **Step 1: 运行失败的 Compose 服务断言**

Run:

```bash
services=$(docker compose -f deploy/docker-compose.yml config --services | sort)
test "$services" = $'opsweaver-gateway\nopsweaver-server\nopsweaver-worker\npostgres\nredis'
```

Expected: FAIL，因为当前配置只包含 `postgres` 和 `redis`。

- [ ] **Step 2: 创建多阶段 Dockerfile**

```dockerfile
FROM golang:1.25-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
ARG SERVICE
RUN case "$SERVICE" in \
      opsweaver-server|opsweaver-worker|opsweaver-gateway) ;; \
      *) echo "Invalid SERVICE: $SERVICE" >&2; exit 1 ;; \
    esac && \
    CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/service "./cmd/${SERVICE}"

FROM alpine:3.22
RUN apk add --no-cache ca-certificates && \
    addgroup -S opsweaver && adduser -S -G opsweaver opsweaver
WORKDIR /app
COPY --from=build /out/service /app/service
COPY deploy/config/compose.yaml /app/config.yaml
USER opsweaver
ENTRYPOINT ["/app/service"]
CMD ["-config", "/app/config.yaml"]
```

- [ ] **Step 3: 创建 Docker 构建忽略文件**

```text
.git
.github
.idea
.worktrees
docs
openspec
```

- [ ] **Step 4: 创建容器网络配置**

```yaml
server:
  port: 8080
  database:
    dsn: postgres://opsweaver:opsweaver@postgres:5432/opsweaver_server_db?sslmode=disable
worker:
  health_port: 8081
gateway:
  port: 8082
  database:
    dsn: postgres://opsweaver:opsweaver@postgres:5432/opsweaver_gateway_db?sslmode=disable
asynq_redis:
  addr: redis:6379
  db: 0
cache_redis:
  addr: redis:6379
  db: 1
security:
  internal_service_token: ""
  master_key_base64: ""
```

- [ ] **Step 5: 在 Compose 中增加三个服务**

```yaml
name: opsweaver

x-service-environment: &service-environment
  INTERNAL_SERVICE_TOKEN: ${INTERNAL_SERVICE_TOKEN:-local-development-token}
  MASTER_KEY_BASE64: ${MASTER_KEY_BASE64:-MDEyMzQ1Njc4OTAxMjM0NTY3ODkwMTIzNDU2Nzg5MDE=}

services:
  postgres:
    image: postgres:16-alpine
    environment:
      POSTGRES_USER: opsweaver
      POSTGRES_PASSWORD: opsweaver
      POSTGRES_DB: opsweaver_server_db
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data
      - ./postgres/init:/docker-entrypoint-initdb.d:ro
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U opsweaver -d opsweaver_server_db"]
      interval: 5s
      timeout: 3s
      retries: 5
      start_period: 5s
    restart: unless-stopped

  redis:
    image: redis:7-alpine
    command: ["redis-server", "--appendonly", "yes"]
    ports:
      - "6379:6379"
    volumes:
      - redis_data:/data
    healthcheck:
      test: ["CMD-SHELL", "redis-cli ping | grep PONG"]
      interval: 5s
      timeout: 3s
      retries: 5
      start_period: 3s
    restart: unless-stopped

  opsweaver-server:
    image: opsweaver-server:local
    build:
      context: ..
      dockerfile: Dockerfile
      args:
        SERVICE: opsweaver-server
    environment: *service-environment
    depends_on:
      postgres:
        condition: service_healthy
      redis:
        condition: service_healthy
    ports:
      - "8080:8080"
    healthcheck:
      test: ["CMD-SHELL", "wget -q -O - http://127.0.0.1:8080/readyz >/dev/null"]
      interval: 5s
      timeout: 3s
      retries: 10
      start_period: 5s
    restart: unless-stopped

  opsweaver-worker:
    image: opsweaver-worker:local
    build:
      context: ..
      dockerfile: Dockerfile
      args:
        SERVICE: opsweaver-worker
    environment: *service-environment
    depends_on:
      redis:
        condition: service_healthy
    ports:
      - "8081:8081"
    healthcheck:
      test: ["CMD-SHELL", "wget -q -O - http://127.0.0.1:8081/readyz >/dev/null"]
      interval: 5s
      timeout: 3s
      retries: 10
      start_period: 5s
    restart: unless-stopped

  opsweaver-gateway:
    image: opsweaver-gateway:local
    build:
      context: ..
      dockerfile: Dockerfile
      args:
        SERVICE: opsweaver-gateway
    environment: *service-environment
    depends_on:
      postgres:
        condition: service_healthy
    ports:
      - "8082:8082"
    healthcheck:
      test: ["CMD-SHELL", "wget -q -O - http://127.0.0.1:8082/readyz >/dev/null"]
      interval: 5s
      timeout: 3s
      retries: 10
      start_period: 5s
    restart: unless-stopped

volumes:
  postgres_data:
  redis_data:
```

- [ ] **Step 6: 验证配置和镜像构建**

Run:

```bash
docker compose -f deploy/docker-compose.yml config
services=$(docker compose -f deploy/docker-compose.yml config --services | sort)
test "$services" = $'opsweaver-gateway\nopsweaver-server\nopsweaver-worker\npostgres\nredis'
docker compose -f deploy/docker-compose.yml build
```

Expected: 全部命令退出码为 0，三个服务镜像构建成功。

- [ ] **Step 7: 提交 Compose 实现**

```bash
git add Dockerfile .dockerignore deploy/config/compose.yaml deploy/docker-compose.yml
git commit -m "feat: run services in docker compose"
```

### Task 2: 编写本地开发 README

**Files:**
- Create: `README.md`

- [ ] **Step 1: 确认 README 尚不存在**

Run: `test -f README.md`

Expected: FAIL。

- [ ] **Step 2: 创建中文 README**

````markdown
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
````

- [ ] **Step 3: 检查 README 必要内容**

Run:

```bash
rg -n "docker compose|OPSWEAVER_SERVER_DATABASE_DSN|cmd/migrate|healthz|readyz|metrics|down -v" README.md
```

Expected: 每类必要操作均有匹配。

- [ ] **Step 4: 提交 README**

```bash
git add README.md
git commit -m "docs: add local development guide"
```

### Task 3: 运行最终验收并完成 OpenSpec

**Files:**
- Modify: `openspec/changes/bootstrap-platform-foundation/tasks.md`

- [ ] **Step 1: 运行 Go 验证**

Run:

```bash
go test ./...
go vet ./...
```

Expected: 全部通过，无失败和 warning。

- [ ] **Step 2: 启动全栈并检查容器**

Run:

```bash
docker compose -f deploy/docker-compose.yml up -d --wait
docker compose -f deploy/docker-compose.yml ps
```

Expected: PostgreSQL、Redis 和三个服务均为 running/healthy。

- [ ] **Step 3: 验证两个逻辑数据库和迁移**

Run:

```bash
export OPSWEAVER_SERVER_DATABASE_DSN='postgres://opsweaver:opsweaver@localhost:5432/opsweaver_server_db?sslmode=disable'
export OPSWEAVER_GATEWAY_DATABASE_DSN='postgres://opsweaver:opsweaver@localhost:5432/opsweaver_gateway_db?sslmode=disable'
go run ./cmd/migrate -db server -command up
go run ./cmd/migrate -db gateway -command up
go run ./cmd/migrate -db server -command version
go run ./cmd/migrate -db gateway -command version
```

Expected: 两个数据库版本均为 `1` 且 `dirty=false`。

- [ ] **Step 4: 验证三个服务端点**

Run:

```bash
for port in 8080 8081 8082; do
  curl --fail --silent "http://127.0.0.1:${port}/healthz" >/dev/null
  curl --fail --silent "http://127.0.0.1:${port}/readyz" >/dev/null
  curl --fail --silent --dump-header /tmp/opsweaver-metrics.headers "http://127.0.0.1:${port}/metrics" >/dev/null
  rg -qi '^content-type: text/plain.*version=0.0.4' /tmp/opsweaver-metrics.headers
done
```

Expected: 所有请求成功，metrics 使用 Prometheus 文本格式。

- [ ] **Step 5: 严格验证 OpenSpec**

Run: `openspec validate bootstrap-platform-foundation --strict --no-interactive`

Expected: `Change 'bootstrap-platform-foundation' is valid`。

- [ ] **Step 6: 勾选 6.1–6.4 并再次验证**

将 `tasks.md` 中 6.1–6.4 从 `- [ ]` 改为 `- [x]`，然后重新运行：

```bash
openspec instructions apply --change bootstrap-platform-foundation --json
git diff --check
```

Expected: 进度为 28/28，diff 无格式错误。

- [ ] **Step 7: 提交验收状态**

```bash
git add openspec/changes/bootstrap-platform-foundation/tasks.md
git commit -m "chore: complete platform foundation verification"
```

### Task 4: 推送并创建 PR

**Files:**
- No file changes.

- [ ] **Step 1: 最终检查提交和工作区**

Run:

```bash
git status --short --branch
git log --oneline origin/main..HEAD
```

Expected: 仅 `.github/`、`.idea/` 保持未跟踪；功能提交完整。

- [ ] **Step 2: 推送分支**

Run: `git push -u origin feature/bootstrap-queue`

Expected: 推送成功并设置 upstream。

- [ ] **Step 3: 创建 PR**

Run:

```bash
gh pr create --base main --head feature/bootstrap-queue \
  --title "feat: complete platform foundation" \
  --body "$(cat <<'BODY'
## Summary
- add Redis/Asynq queue contracts, controlled retry, and service wiring
- run PostgreSQL, Redis, and all three services with Docker Compose
- document local startup, migrations, health checks, and shutdown
- complete the bootstrap-platform-foundation OpenSpec change

## Verification
- go test ./...
- go vet ./...
- Docker Compose build and five-container health verification
- server/worker/gateway health, ready, and metrics checks
- openspec validate bootstrap-platform-foundation --strict --no-interactive
BODY
)"
```

Expected: GitHub 返回新 PR URL。
