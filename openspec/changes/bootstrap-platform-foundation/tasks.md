## 1. 工程骨架与配置

- [x] 1.1 创建 `go.mod`（module `github.com/Gooooodman/opsweaver`）、`Makefile`、`.env.example` 和 `cmd/{opsweaver-server,opsweaver-worker,opsweaver-gateway}/main.go`，运行 `go test ./...` 验证三个入口可编译
- [x] 1.2 先在 `internal/platform/config/config_test.go` 编写有效配置、缺失必填项和环境变量覆盖测试，再实现 `internal/platform/config/config.go`
- [x] 1.3 先在 `internal/platform/logging/logging_test.go` 验证 JSON 日志和敏感字段掩码，再实现基于 `slog` 的 `internal/platform/logging/logging.go`
- [x] 1.4 创建 `internal/platform/apperror/error.go`、`internal/platform/httpx/response.go` 及测试，统一英文 API 错误结构 `code/message/trace_id`
- [x] 1.5 ~~提交工程骨架，提交信息 `chore: bootstrap Go service structure`~~（已拆分为 1.1–1.4 各自独立提交：`chore: bootstrap Go service structure` / `feat: add validated application config` / `feat: validate master key length and redact secrets` / `feat: add structured logging` / `feat: add unified error response`）

## 2. 本地基础设施与迁移

- [x] 2.1 创建 `deploy/docker-compose.yml`、`deploy/postgres/init/01-create-databases.sql`，运行 `docker compose -f deploy/docker-compose.yml config` 验证 Compose
- [x] 2.2 创建 `migrations/opsweaver_server/`、`migrations/opsweaver_gateway/` 首批版本 SQL 和 `cmd/migrate/main.go`，禁止使用 GORM AutoMigrate
- [x] 2.3 在 `internal/platform/database/database_test.go` 编写两个 DSN 独立连接测试，再实现 `internal/platform/database/database.go`
- [x] 2.4 启动 Compose，分别执行两个数据库迁移并查询迁移版本表，确认空库初始化成功
- [ ] 2.5 提交基础设施，提交信息 `feat: add local postgres redis infrastructure`

## 3. 健康检查与可观测性

- [ ] 3.1 在 `internal/platform/health/health_test.go` 编写 liveness 与 dependency readiness 测试，再实现 `internal/platform/health/health.go`
- [ ] 3.2 创建 `internal/platform/metrics/metrics.go`，注册 HTTP 请求、依赖状态和 Asynq 处理指标，并补充重复注册测试
- [ ] 3.3 在三个服务中挂载 `/healthz`、`/readyz`、`/metrics`，运行针对性 HTTP 测试验证状态码和 Prometheus Content-Type
- [ ] 3.4 提交可观测性，提交信息 `feat: add service health and metrics`

## 4. 内部认证与凭证加密

- [ ] 4.1 在 `internal/platform/auth/service_token_test.go` 编写正确、缺失、错误令牌测试，再实现常量时间比较中间件
- [ ] 4.2 在 `internal/platform/crypto/aesgcm_test.go` 编写 round-trip、随机 Nonce、非法密钥和篡改密文测试，再实现版本化 AES-256-GCM 密文封装
- [ ] 4.3 在 `internal/platform/mask/mask_test.go` 编写 token/password/authorization/secret 递归脱敏测试，再实现 `internal/platform/mask/mask.go`
- [ ] 4.4 将主密钥和内部令牌接入配置校验，验证启动错误和日志均不包含秘密
- [ ] 4.5 提交安全基线，提交信息 `feat: add internal auth and credential encryption`

## 5. Redis 与 Asynq

- [ ] 5.1 在 `internal/platform/redisx/client_test.go` 验证 Asynq 使用 DB 0、缓存使用 DB 1，再实现 Redis Client 工厂
- [ ] 5.2 创建 `internal/queue/types.go`、`internal/queue/client.go`、`internal/queue/server.go`，定义诊断与 MCP 同步任务类型
- [ ] 5.3 在 `internal/queue/retry_test.go` 编写临时错误三次指数退避和永久错误不重试测试，再实现错误分类与 Asynq Options
- [ ] 5.4 将 Asynq Client 装配到 `opsweaver-server`，将 Asynq Server 与健康端口装配到 `opsweaver-worker`，确认 `opsweaver-gateway` 无 Asynq 依赖
- [ ] 5.5 使用 Compose 运行最小入队/消费集成测试并提交，提交信息 `feat: add asynq task queue foundation`

## 6. Change 验证

- [ ] 6.1 运行 `go test ./...` 和 `go vet ./...`，修复所有失败
- [ ] 6.2 运行 `docker compose -f deploy/docker-compose.yml up -d --wait`，验证两个数据库、Redis 和三个服务的 health/ready/metrics
- [ ] 6.3 更新 README 的本地启动、迁移、环境变量和停止命令
- [ ] 6.4 运行 `openspec validate bootstrap-platform-foundation --strict --no-interactive`
