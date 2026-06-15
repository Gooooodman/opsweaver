## ADDED Requirements

### Requirement: Asynq 职责分离
`opsweaver-server` SHALL 使用 Asynq Client 入队，`opsweaver-worker` SHALL 运行 Asynq Server 消费任务，`opsweaver-gateway` MUST NOT 依赖 Asynq。

#### Scenario: 提交任务
- **WHEN** `opsweaver-server` 提交有效任务
- **THEN** 任务进入配置队列并可被 Worker 消费

### Requirement: 队列与缓存隔离
Asynq SHALL 使用 Redis DB 0，任务状态缓存和 MCP 健康缓存 SHALL 使用 Redis DB 1。

#### Scenario: 初始化 Redis Client
- **WHEN** 三个服务加载 Redis 配置
- **THEN** 队列 Client 与缓存 Client 使用不同逻辑 DB

### Requirement: 任务具有受控重试
异步任务 SHALL 支持超时、最多三次重试和指数退避，并允许处理器将永久错误标记为不可重试。

#### Scenario: 临时错误
- **WHEN** 处理器返回可重试的临时错误
- **THEN** Asynq 按指数退避重新投递且总尝试次数不超过三次

#### Scenario: 永久错误
- **WHEN** 处理器返回参数错误或策略拒绝等永久错误
- **THEN** 任务不再重试
