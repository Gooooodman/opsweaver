## Context

仓库目前只有 PRD 与 P0 总体设计。本 Change 必须建立后三个 Change 共同依赖的运行、配置、数据、队列和安全基础，同时保持三个服务和两个数据库的所有权边界。

## Goals / Non-Goals

**Goals:**

- 三个 Go 服务可独立启动、观测和测试。
- 本地通过 Docker Compose 一键启动 PostgreSQL 与 Redis。
- `opsweaver_server_db` 与 `opsweaver_gateway_db` 使用独立连接和迁移。
- Asynq Client/Server 职责清晰，队列与缓存逻辑隔离。
- 内部 HTTP API 和敏感凭证具备 P0 安全基线。

**Non-Goals:**

- 不实现业务 Registry、Tool Invoke 或诊断 Workflow。
- 不实现生产级服务身份、KMS/Vault 或密钥轮换。
- 不实现 Kubernetes 部署清单。

## Decisions

### 决策 1：单 Go Module Monorepo

使用 `github.com/Gooooodman/opsweaver`，服务入口位于 `cmd/`，共享基础包位于 `internal/platform/`。相比多 Module，该结构降低 P0 依赖与版本治理成本。

### 决策 2：一个本地 PostgreSQL 实例、两个逻辑数据库

Docker Compose 只启动一个 PostgreSQL 容器，通过初始化脚本创建两个数据库。服务使用不同 DSN，禁止跨库 Repository。相比两个容器，本地资源更少；相比共享数据库，所有权更明确。

### 决策 3：GORM + 显式 SQL 迁移

GORM 用于 Repository 映射和查询，迁移由独立命令按目录执行，不调用 AutoMigrate。这样保留开发效率，同时使结构变更可审阅和回滚。

### 决策 4：Asynq 部署职责

`opsweaver-server` 持有 Asynq Client，`opsweaver-worker` 运行 Asynq Server，`opsweaver-gateway` 不引入 Asynq。Redis DB 0 用于 Asynq，DB 1 用于状态与健康缓存。生产允许配置独立 Endpoint。

### 决策 5：配置与错误模型

配置由 YAML 默认值和环境变量覆盖组成，启动时集中校验。日志使用 `slog` JSON Handler；API 错误结构统一为 `code`、`message`、`trace_id`，message 使用英文。

### 决策 6：凭证加密

使用标准库 AES-256-GCM。环境变量提供 Base64 编码的 32 字节主密钥；密文封装版本、Nonce 和 Ciphertext，便于后续迁移。

## Risks / Trade-offs

- [Risk] 本地共享 PostgreSQL 实例可能掩盖生产网络隔离问题 → 使用独立 DSN、迁移目录和集成测试强制边界。
- [Risk] 静态服务令牌泄露后可访问内部 API → 令牌只从环境变量读取，日志统一脱敏，后续升级服务身份。
- [Risk] Asynq 与缓存共享 Redis 实例可能相互影响 → 使用不同逻辑 DB，文档要求生产队列使用无淘汰策略或独立 Endpoint。
- [Risk] P0 不支持主密钥轮换 → 密文包含版本字段，部署文档明确轮换前需离线重加密。

## Migration Plan

1. 启动 Docker Compose 并确认依赖健康。
2. 对两个空数据库分别执行迁移。
3. 配置主密钥和内部服务令牌。
4. 启动三个服务并检查 health、ready、metrics。
5. 回滚时先停止服务，再按迁移工具执行向下迁移；本地环境可删除 Compose 数据卷。

## Open Questions

无阻塞问题。具体依赖版本在实施时选用当时稳定且兼容的版本，并由 `go.mod` 锁定。
