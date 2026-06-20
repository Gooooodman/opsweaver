# Compose 三服务运行设计

## 目标

让 `docker compose -f deploy/docker-compose.yml up -d --wait` 同时启动 PostgreSQL、Redis、`opsweaver-server`、`opsweaver-worker` 和 `opsweaver-gateway`，并通过三个服务的健康、就绪和指标端点验证本地基础环境。

## 架构

仓库根目录新增一个多阶段 `Dockerfile`。构建阶段通过 `SERVICE` 参数选择 `cmd/` 下的入口，并将结果统一输出为 `/app/service`；运行阶段只包含静态 Go 二进制和本地运行配置。三个 Compose 服务复用该 Dockerfile，仅传入不同构建参数。

`deploy/config.yaml` 保存端口、数据库 DSN 和 Redis 地址等非敏感本地配置。容器通过 Compose 服务名访问 `postgres:5432` 和 `redis:6379`。内部服务令牌和主密钥仍由环境变量提供，不写入配置文件；Compose 仅提供明确标注的本地开发默认值，部署环境可覆盖。

## 服务编排

- `opsweaver-server` 依赖 PostgreSQL 和 Redis 健康，暴露端口 8080。
- `opsweaver-worker` 依赖 Redis 健康，暴露健康端口 8081。
- `opsweaver-gateway` 依赖 PostgreSQL 健康，暴露端口 8082。
- 三个服务均使用 `/readyz` 作为容器健康检查。
- PostgreSQL 和 Redis 保持现有数据卷、端口和健康检查，不改变数据初始化方式。

## 错误处理

镜像构建失败、依赖健康检查失败或服务就绪检查失败时，Compose 应返回非零状态。服务继续沿用现有英文结构化启动错误，不增加容器专用错误格式。

本地开发默认令牌和主密钥只用于 Compose 启动，不作为生产默认值。README 将明确说明生产环境必须覆盖这些变量。

## 验证

1. 运行 `docker compose -f deploy/docker-compose.yml config` 检查配置。
2. 运行 `docker compose -f deploy/docker-compose.yml build` 验证三个镜像可构建。
3. 运行 `docker compose -f deploy/docker-compose.yml up -d --wait`，确认五个容器健康。
4. 请求三个服务的 `/healthz`、`/readyz` 和 `/metrics`，检查状态码和 Prometheus Content-Type。
5. 运行 `go test ./...`、`go vet ./...` 和严格 OpenSpec 校验。

## 范围限制

本设计只覆盖本地 Compose 运行，不增加 Kubernetes 清单、生产镜像发布流程、密钥管理系统或多环境配置机制。
