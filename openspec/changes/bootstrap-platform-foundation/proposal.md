## Why

OpsWeaver 当前只有 PRD 和架构设计，缺少可运行、可测试的工程基础。必须先建立稳定的三服务骨架、数据边界、异步任务基础和安全配置，后续 Tool、MCP、Skill 与诊断场景才能在一致约束下实现。

## What Changes

- 初始化单 Go Module Monorepo，提供 `opsweaver-server`、`opsweaver-worker`、`opsweaver-gateway` 三个可独立启动的服务。
- 提供统一配置、结构化日志、健康检查、Prometheus 指标和英文错误响应。
- 提供 Docker Compose，在一个 PostgreSQL 容器中初始化 `opsweaver_server_db`、`opsweaver_gateway_db`，并启动 Redis。
- 使用 GORM 访问数据库，使用显式版本化 SQL 管理两个数据库的迁移。
- 集成 Asynq：`opsweaver-server` 负责入队，`opsweaver-worker` 负责消费，Redis DB 0 用于队列。
- 建立任务状态缓存和 MCP 健康缓存的 Redis DB 1 连接边界。
- 提供静态内部服务令牌认证和 AES-256-GCM 凭证加密基础。

## Capabilities

### New Capabilities

- `service-bootstrap`: 三个 Go 服务的启动、配置、日志、健康检查和指标行为。
- `local-infrastructure`: Docker Compose、双逻辑数据库、Redis 和版本化迁移行为。
- `secure-configuration`: 内部服务令牌和敏感凭证加密存储行为。
- `async-task-queue`: Asynq 入队、消费、重试配置和 Redis 隔离行为。

### Modified Capabilities

无。

## Impact

- 新增 Go Module、服务入口、基础包、配置、迁移、Docker Compose 和测试基础设施。
- 新增 Gin、GORM、PostgreSQL Driver、Asynq、Redis Client、Prometheus Client 等依赖。
- 后续 Change 依赖本 Change 提供的服务启动、数据库、Redis、认证和错误模型。
