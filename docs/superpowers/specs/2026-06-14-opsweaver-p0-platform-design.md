# OpsWeaver P0 平台设计

## 1. 背景

OpsWeaver 是面向运维、研发和测试团队的工程智能平台。P0 的目标不是覆盖大量诊断场景，而是建立可插拔、可编排、可审计、证据可追溯的工程智能底座，并通过 Pod 重启诊断验证完整链路。

本设计覆盖 PRD 中 M0-M8 的完整 P0 范围。P0 不提供 Web UI，不实现自动修复、复杂权限、多集群、Agent Registry、Eino、RAG 或 Loki 集成。

## 2. 目标与非目标

### 2.1 目标

1. 内置 Tool、MCP Tool 和 Skill 可以统一注册、发现和启停。
2. Tool Gateway 是所有外部工具调用的唯一入口。
3. MCP Server 可以通过配置或 API 注册，无需修改核心代码。
4. Skill 可以通过声明式 Workflow 注册和执行，无需修改 Orchestrator。
5. 诊断任务通过独立 Worker 异步执行。
6. 所有工具调用都有 Tool Call、Evidence 和 Audit 记录。
7. RCA Report 的关键结论必须引用 Evidence。
8. P0 所有可执行工具均为 `read_only`。
9. Pod 重启诊断可以完成端到端验证。

### 2.2 非目标

1. 不提供 Vue Web UI，仅提供 API 和 OpenAPI 文档。
2. 不实现完整 Auth、RBAC、Approval 或 OPA。
3. 不执行 restart、scale、rollback、delete 或 apply。
4. 不支持多集群、Loki、Trace、RAG、ChatOps。
5. 不支持 Agent Registry、Eino Runtime 或多 Agent。
6. 不支持 stdio MCP、旧版 HTTP+SSE 或自定义 Transport。
7. 不支持脚本、二进制、Go 插件或任意代码执行。

## 3. 技术选择

| 领域 | 选择 |
|---|---|
| 后端语言 | Go |
| HTTP 框架 | Gin |
| 数据访问 | GORM |
| 数据库迁移 | 显式版本化 SQL |
| 数据库 | 两个逻辑独立的 PostgreSQL 数据库 |
| 缓存与队列 | Redis + Asynq |
| 本地基础设施 | Docker Compose |
| 服务通信 | HTTP/JSON |
| Kubernetes | 单集群，配置选择 kubeconfig 或 in-cluster |
| 日志查询 | Mock Logs API |
| LLM | OpenAI-compatible API |
| MCP | Streamable HTTP |
| 凭证加密 | AES-256-GCM，主密钥由环境变量注入 |

引入 Asynq 的范围仅限任务入队、消费、重试和超时，不使用定时任务等额外能力。

本地开发使用 Docker Compose 启动一个 PostgreSQL 容器和一个 Redis 容器。PostgreSQL 容器初始化 `opsweaver_server_db` 与 `opsweaver_gateway_db` 两个逻辑数据库；生产环境仍按两个数据库的数据所有权边界管理。

## 4. 总体架构

P0 部署三个 Go 服务：

```text
External Client
    |
    v
opsweaver-server --enqueue--> Redis / Asynq --> opsweaver-worker
    ^                                         |
    | internal HTTP                           | HTTP/JSON
    |                                         v
    +----------------------------------- opsweaver-gateway
                                              |
                    +-------------------------+----------------------+
                    |                         |                      |
                    v                         v                      v
              Built-in Tools             MCP Servers          Evidence/Audit
```

### 4.1 opsweaver-server：控制面

职责：

1. 对外提供 Task、Report、Tool、MCP 和 Skill API。
2. 管理 Tool、MCP、Skill、Workflow Registry。
3. 进行意图解析、参数抽取和 Skill 匹配。
4. 创建任务并提交 Asynq。
5. 为 Worker 提供版本快照和任务状态回写的内部 API。
6. 保存最终 Report。

数据所有权：

```text
opsweaver_server_db:
  tasks
  reports
  tools
  mcp_servers
  skills
  workflows
  encrypted_credentials
```

禁止：

1. 不直接访问 Kubernetes、Prometheus、Mock Logs 或 MCP Tool。
2. 不直接写入 Tool Call、Evidence 或 Audit 数据。

### 4.2 opsweaver-worker：执行面

职责：

1. 消费 Asynq 诊断和 MCP 同步任务。
2. 原子获取任务执行权。
3. 获取固定版本的 Skill/Workflow 快照。
4. 执行声明式 Workflow。
5. 通过 Tool Gateway 调用 Tool。
6. 调用 OpenAI-compatible LLM 生成结构化报告。
7. 通过内部 API 回写步骤状态、报告和任务结果。

Worker 不拥有业务数据库，也不直接访问 `opsweaver_server_db` 或 `opsweaver_gateway_db`。

### 4.3 opsweaver-gateway：工具面

职责：

1. 提供内部 Tool Invoke API。
2. 获取并缓存 ToolSpec。
3. 校验参数、Scope、Risk Level 和幂等键。
4. 分发到 Built-in Tool Adapter 或 MCP Adapter。
5. 对结果进行标准化和脱敏。
6. 写入 Tool Call、Evidence 和 Audit。
7. 提供 MCP `tools/list` 和健康检查能力。

数据所有权：

```text
opsweaver_gateway_db:
  tool_calls
  evidences
  audit_logs
```

Tool Gateway 不承担 Skill 编排、意图识别或报告生成。

### 4.4 Redis

Redis 用于：

1. Asynq 任务队列。
2. 任务状态缓存。
3. MCP 健康状态缓存。

P0 不使用 Redis 实现通用缓存框架、分布式锁或业务数据持久化。

## 5. 服务与数据边界

1. `opsweaver_server_db` 和 `opsweaver_gateway_db` 是逻辑独立的数据库，禁止跨库直接读写。
2. 本地 Docker Compose 复用同一个 PostgreSQL 实例，生产部署不依赖该拓扑。
3. 服务只能写入自己拥有的数据库。
4. 跨服务读取和写入必须通过 HTTP API。
5. Worker 不直接访问任何业务数据库。
6. Tool Gateway 通过内部 API 或版本化同步接口获取 ToolSpec，不直接读取 `opsweaver_server_db`。
7. 内部 API 使用静态服务令牌认证；完整服务身份治理延后。

## 6. 任务生命周期与一致性

### 6.1 状态模型

```text
pending -> queued -> running -> succeeded
                          \-> failed
                          \-> clarification_required
```

### 6.2 入队

1. `opsweaver-server` 先在 `opsweaver_server_db` 创建 `pending` 任务。
2. 成功提交 Asynq 后更新为 `queued`。
3. 入队失败时更新为 `failed`，记录英文错误信息。
4. Asynq 载荷仅包含 `task_id`、`attempt` 和 `trace_id`。

### 6.3 执行权与幂等

1. Worker 通过内部 API 原子获取任务执行权。
2. 已处于终态或已被其他 Worker 获取的任务不重复执行。
3. Workflow Step 幂等键为 `task_id + step_id + attempt`。
4. Tool Gateway 对相同幂等键返回已有 InvokeResult。
5. 幂等命中不重复访问外部系统，也不重复创建 Tool Call。

### 6.4 版本快照

1. Task 创建时记录 `skill_name` 和 `skill_version`。
2. Worker 按指定版本获取完整 Skill、Workflow、Prompt 和 Tool 引用快照。
3. 执行期间 Registry 的新版本或启停变化不影响已创建任务。

### 6.5 重试

| 错误类型 | 行为 |
|---|---|
| 网络超时、HTTP 5xx、MCP 临时不可用 | Asynq 最多重试 3 次，指数退避 |
| 参数错误、风险策略拒绝、Workflow 校验失败 | 不重试 |
| LLM 输出格式错误 | 同一次任务内最多修复重试 1 次 |
| 非关键 Tool Step 失败 | 记录 Error Evidence，继续执行 |
| 关键 Tool Step 失败 | 终止 Workflow |

Tool Step 是否关键由 Workflow 显式声明，默认关键。

最终结果必须区分：

1. `failed`：流程无法完成。
2. `succeeded` 但证据不足：流程完成，报告明确表达不确定性。

## 7. Capability 与 Registry

### 7.1 Capability

P0 支持：

```text
builtin_tool
mcp_tool
skill
```

`agent` 仅保留类型枚举和未来接口位置，不创建 Agent 表，不开放注册 API。

### 7.2 权威来源

1. `opsweaver_server_db` 是 Tool、MCP、Skill 和 Workflow 的权威来源。
2. `configs/` 文件只是导入源，不作为运行时权威来源。
3. 启动扫描和管理 API 使用相同的 Registry 校验服务。
4. 文件导入按 `name + version` 幂等执行。
5. 启动扫描不自动删除数据库记录。

### 7.3 版本规则

1. Tool、MCP、Skill 和 Workflow 使用不可变版本记录。
2. 更新定义时创建新版本，不原地覆盖已发布版本。
3. 启用和禁用作用于具体版本。
4. 同一 Capability 同一时刻最多有一个默认启用版本。

## 8. MCP 设计

1. P0 仅支持 Streamable HTTP。
2. MCP 注册信息包括 Endpoint、Auth、Allowed Tools、Scope、Risk、Timeout 和 Rate Limit。
3. 注册后由 `opsweaver-server` 提交 MCP 同步任务。
4. Worker 调用 Tool Gateway 的 MCP 管理接口执行 `tools/list`。
5. 同步结果通过 `opsweaver-server` 内部 API 写入 Tool Registry。
6. MCP Tool 统一命名为 `<mcp_server_name>.<remote_tool_name>`。
7. 远端 Tool 声明不能放宽本地 Allowed Tools、Scope 或 Risk 限制。
8. 所有 MCP `tools/call` 必须经过 Tool Gateway 并写入审计。

## 9. Skill 与 Workflow

### 9.1 Skill

Skill 由以下内容组成：

```text
skill.yaml
workflow.yaml
prompt.md
examples.yaml
```

注册时校验：

1. 引用文件存在且可解析。
2. Input/Output Schema 合法。
3. 引用的 Tool 版本存在且启用。
4. Skill 风险上限不低于引用 Tool 的实际风险。
5. Workflow Step ID 唯一，引用关系有效。
6. Prompt 引用存在。

### 9.2 Workflow Step

P0 支持：

```text
tool
condition
llm
report
```

执行模式仅支持顺序执行和简单条件分支，不支持复杂 DAG、循环、并行步骤或任意代码。

### 9.3 LLM

1. 使用 OpenAI-compatible API。
2. 配置包含 `base_url`、`model` 和加密后的 API Key。
3. LLM 输出必须通过 JSON Schema 校验。
4. 报告生成 Prompt 只能引用当前任务的 Evidence 摘要和 ID。
5. 结构化输出修复最多重试一次。

## 10. Tool Gateway 执行链

```text
Invoke Request
    |
    v
Load ToolSpec
    |
    v
Validate Arguments
    |
    v
Check Scope and Risk
    |
    v
Check Idempotency
    |
    v
Dispatch Adapter
    |
    v
Normalize and Mask Result
    |
    +--> Write Tool Call
    +--> Write Evidence or Error Evidence
    +--> Write Audit Log
    |
    v
InvokeResult
```

P0 风险策略：

```text
read_only: allow
low: deny
medium: deny
high: deny
critical: deny
```

无论调用成功、外部系统失败或策略拒绝，都必须记录 Tool Call 和 Audit；能够形成诊断上下文的失败必须生成 Error Evidence。

## 11. 内置 Tool

P0 实现：

1. `k8s.describe_pod`
2. `k8s.get_events`
3. `prometheus.query`
4. `logs.search`

实现约束：

1. Kubernetes 为单集群，通过配置选择 kubeconfig 或 in-cluster。
2. Prometheus 使用 HTTP API。
3. `logs.search` 只接入 Mock Logs API，不实现 Loki。
4. 所有 Tool 均为 `read_only`。

## 12. Evidence、Audit 与 Report

### 12.1 Evidence

1. 每个 Tool Call 至少生成一条 Evidence 或 Error Evidence。
2. Evidence 保存来源、类型、摘要、结构化内容、时间戳和 Tool Call ID。
3. Evidence 内容在落库前脱敏。
4. 关键结论必须引用一个或多个 Evidence ID。

### 12.2 Audit

Audit 至少包含：

```text
audit_id
task_id
tool_call_id
user_id
team
tool_name
arguments_masked
risk_level
decision
duration_ms
success
error
created_at
```

错误信息、日志消息和 API 错误响应使用英文。

### 12.3 Report

RCA Report 包含：

1. 摘要。
2. 影响。
3. Evidence 列表。
4. 最多三个可能原因及置信度。
5. 建议动作及风险等级。
6. 下一步只读排查命令。
7. Evidence ID 引用。
8. 证据不足声明。

## 13. API 边界

### 13.1 对外 API

```text
POST /api/v1/tasks
GET  /api/v1/tasks/{id}
GET  /api/v1/tasks/{id}/report

GET  /api/v1/tools
GET  /api/v1/tools/{name}
POST /api/v1/tools/{name}/enable
POST /api/v1/tools/{name}/disable

POST /api/v1/mcp-servers
GET  /api/v1/mcp-servers
GET  /api/v1/mcp-servers/{id}
POST /api/v1/mcp-servers/{id}/sync-tools
POST /api/v1/mcp-servers/{id}/enable
POST /api/v1/mcp-servers/{id}/disable

POST /api/v1/skills
GET  /api/v1/skills
GET  /api/v1/skills/{name}
POST /api/v1/skills/{name}/run
POST /api/v1/skills/{name}/enable
POST /api/v1/skills/{name}/disable
```

Tool Invoke 不作为普通用户 API 开放。

### 13.2 内部 API

`opsweaver-server` 提供：

1. 获取任务执行权。
2. 获取 Skill/Workflow 版本快照。
3. 回写步骤状态。
4. 回写 Report。
5. 更新任务终态。
6. 写入 MCP Tool 同步结果。

`opsweaver-gateway` 提供：

1. Tool Invoke。
2. MCP `tools/list`。
3. MCP 健康检查。

所有 API 都提供 OpenAPI 定义。

## 14. 凭证安全

1. MCP Auth 和 LLM API Key 使用 AES-256-GCM 加密后存储。
2. 主密钥通过环境变量注入，数据库和配置文件不保存主密钥。
3. 启动时主密钥缺失或长度非法，相关服务直接启动失败。
4. 每条密文使用独立随机 Nonce。
5. API 响应、日志、Audit 和错误信息不得包含明文凭证。
6. P0 不实现主密钥轮换；部署文档必须说明轮换前需要重新加密现有凭证。

## 15. Orchestrator 与 Pod 诊断

### 15.1 Orchestrator

采用规则优先、LLM fallback：

1. 规则匹配 Pod 重启、CrashLoopBackOff、OOMKilled 等触发词。
2. 规则无法确定时调用 LLM 进行结构化意图解析。
3. 缺少 namespace、pod 或 service 时返回 `clarification_required`。
4. Orchestrator 只选择 Skill 和提取参数，不调用 Tool。

### 15.2 Pod 重启诊断

```text
User Input
    |
    v
Intent and Entity Parsing
    |
    v
pod_restart_diagnosis
    |
    +--> k8s.describe_pod
    +--> k8s.get_events
    +--> logs.search
    +--> prometheus.query
    |
    v
Evidence Collection
    |
    v
LLM Structured Report
    |
    v
RCA Report
```

## 16. OpenSpec 规划

完整 P0 拆分为四个按顺序实施的 OpenSpec Change。

### 16.1 bootstrap-platform-foundation

范围：

1. Monorepo 和单 Go Module。
2. `opsweaver-server`、`opsweaver-worker`、`opsweaver-gateway` 骨架。
3. 配置、结构化日志、健康检查、Prometheus 指标。
4. Docker Compose 启动 PostgreSQL 和 Redis。
5. PostgreSQL 初始化 `opsweaver_server_db` 和 `opsweaver_gateway_db`。
6. GORM 和显式 SQL 迁移。
7. Redis 和 Asynq。
8. 核心模型、错误模型、静态内部服务令牌。
9. AES-256-GCM 凭证加密。

### 16.2 build-tool-execution-plane

范围：

1. Tool Registry。
2. Tool Gateway Invoke 链路。
3. Risk、Scope、Schema 和幂等校验。
4. Built-in Tool Adapter。
5. Tool Call、Evidence 和 Audit。
6. Kubernetes、Prometheus 和 Mock Logs 集成。

### 16.3 add-declarative-capability-runtime

范围：

1. MCP Registry 和 Streamable HTTP Adapter。
2. MCP Tool 同步和健康检查。
3. Skill Registry。
4. Workflow Runtime。
5. 版本快照。
6. OpenAI-compatible LLM。
7. 结构化 Report 生成。

### 16.4 deliver-pod-diagnosis-workflow

范围：

1. Orchestrator 规则和 LLM fallback。
2. 异步任务生命周期。
3. Worker 执行与状态回写。
4. Pod 重启诊断 Skill。
5. RCA Report API。
6. 完整端到端测试。

## 17. 测试策略

### 17.1 单元测试

覆盖：

1. Risk Policy。
2. Scope 与 JSON Schema 校验。
3. Workflow 条件和关键步骤行为。
4. AES-256-GCM 加解密和非法密钥。
5. 幂等键和重复任务处理。
6. LLM 结构化输出校验。

### 17.2 Repository 测试

使用独立测试数据库验证：

1. GORM 映射。
2. 显式 SQL 迁移。
3. 唯一约束和状态转换。
4. 不可变版本规则。

### 17.3 Contract 测试

覆盖三个服务间的 HTTP/JSON：

1. Worker 与 `opsweaver-server`。
2. Worker 与 `opsweaver-gateway`。
3. ToolSpec 同步。
4. 错误响应和超时。

### 17.4 Adapter 测试

1. Kubernetes fake client。
2. Prometheus `httptest.Server`。
3. Mock Logs API。
4. 模拟 Streamable HTTP MCP Server。
5. 模拟 OpenAI-compatible API。

### 17.5 集成与端到端测试

1. Redis + Asynq。
2. `opsweaver_server_db` + `opsweaver_gateway_db`。
3. 重复投递和 Worker 重启。
4. Tool Gateway 幂等。
5. 提交 Pod 重启问题并等待 Report。
6. 验证 Report 引用可查询的 Evidence。

### 17.6 验证命令

```text
go test ./...
go vet ./...
openspec validate --strict
```

实际命令以项目生成后的 Makefile 和 OpenSpec CLI 支持为准。

## 18. 风险与取舍

### 18.1 两个逻辑数据库和严格服务边界

代价是跨服务 API、契约测试和故障处理更多。收益是数据所有权清晰，Tool Gateway 的证据审计链可以独立演进。本地开发为降低资源成本，在一个 PostgreSQL 容器中承载两个逻辑数据库，但不放宽服务数据边界。

### 18.2 独立 Worker 与 Asynq

代价是引入队列语义、幂等和重试设计。收益是长耗时 Tool/LLM 调用不会阻塞外部 HTTP 请求，Worker 可以独立扩展。

### 18.3 GORM

GORM 提高模型开发效率，但迁移仍使用显式 SQL，关键查询需要可观测并避免隐式 N+1。

### 18.4 数据库加密凭证

环境变量主密钥实现简单，但 P0 不支持轮换。密钥治理和 KMS/Vault 集成留到后续阶段。

### 18.5 Mock Logs API

可以优先验证平台主链路，但不能代表真实日志后端的查询语义和性能。Loki Adapter 在 P1 或后续独立 Change 中实现。

## 19. 验收标准

1. 三个服务可以独立启动并通过健康检查。
2. Docker Compose 可以启动 PostgreSQL 和 Redis，并初始化两个逻辑数据库。
3. 两个数据库的迁移可以从空库执行。
4. Asynq 可以完成任务入队、消费、重试和终态回写。
5. 四个内置 Tool 可以通过 Tool Gateway 调用。
6. 所有成功和失败的 Tool 调用都可查询 Tool Call 和 Audit。
7. 每个 Tool Call 都有 Evidence 或 Error Evidence。
8. MCP Server 注册后可以同步并调用新 Tool，无需修改 Go 代码。
9. Skill 注册后可以执行，无需在 Orchestrator 增加分支。
10. 非 `read_only` Tool 无法执行。
11. Pod 重启诊断 Report 包含摘要、Evidence、可能原因、建议和下一步。
12. Report 的关键结论引用有效 Evidence ID。
13. `go test ./...`、`go vet ./...` 和 OpenSpec 严格校验通过。
