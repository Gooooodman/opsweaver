# Engineering Copilot Platform PRD

> 面向运维、研发、测试团队的自然语言工程提效平台。  
> 本版本采用 **底座优先、插件化优先、场景后置** 的设计原则。  
> MVP 第一阶段控制在 3～5 个服务，重点建设 Orchestrator、Tool Gateway、Registry、MCP Adapter、Skill Runtime、Evidence、Audit 等底层能力。  
> Auth / RBAC / Approval 只做字段和接口预留，不做复杂权限系统。

---

## 1. 文档信息

| 字段 | 内容 |
|---|---|
| 文档名称 | Engineering Copilot Platform PRD |
| 文件名 | `prd.md` |
| 当前版本 | v0.2 |
| 阶段 | MVP 底座优先版 |
| 目标用户 | 运维 / SRE / DevOps / 研发 / 测试 |
| 核心定位 | DevOps + AIOps + TestOps Engineering Copilot |
| 技术策略 | Go First / Go-only for MVP |
| 后端框架 | Gin 优先 |
| 前端框架 | Vue 优先 |
| 主要目标 | 先建设可插拔工程智能底座，再逐步叠加诊断场景 |
| 服务规模 | MVP 3～5 个服务 |
| 权限策略 | Auth / RBAC / Approval 预留，不作为 P0 主目标 |

---

## 2. 背景与问题

当前工程团队在排障、发布、测试、研发协作中存在以下问题：

1. 故障排查需要频繁切换 Kubernetes、Prometheus、日志平台、Trace、GitLab CI、发布平台、CMDB、测试平台。
2. 排查过程依赖个人经验，标准化不足，Runbook 难以复用。
3. 告警、日志、指标、发布、代码、测试报告之间缺少统一证据链。
4. 运维、研发、测试之间缺少统一问题上下文，问题定位容易变成低效沟通。
5. LLM / Agent 能力可以提升效率，但如果直接访问生产系统，会带来误操作、权限、审计和数据泄露风险。
6. 各团队可能会开发自己的 MCP Server、Skill、Agent。如果每接入一种能力都要修改核心平台代码，平台会迅速失控。

因此平台需要优先解决：

```text
1. 外部能力如何注册。
2. 工具能力如何统一调用。
3. Skill 如何声明式编排。
4. Agent 如何受控执行。
5. 所有工具调用如何审计。
6. LLM 输出如何绑定证据。
7. 新能力接入如何不改核心代码。
```

---

## 3. 产品定位

### 3.1 一句话定位

Engineering Copilot Platform 是一个面向运维、研发、测试的自然语言工程提效平台。平台通过统一的 Capability 抽象、Registry 注册中心、Tool Gateway、MCP Adapter、Skill Runtime、轻量 Orchestrator、Evidence Store 和 Audit Log，实现可插拔、可编排、可审计的工程智能底座。

### 3.2 设计原则

| 原则 | 说明 |
|---|---|
| 底座优先 | 先做 Registry、Tool Gateway、Adapter、Evidence、Audit |
| 插件化优先 | MCP / Skill / Agent 通过注册接入，不改核心代码 |
| 场景后置 | Pod 诊断、5xx 诊断、CI 分析放在底座之后 |
| Go First | MVP 主链路全部使用 Go |
| 只读优先 | P0 阶段所有工具默认 read_only |
| 证据优先 | RCA 报告必须引用 evidence_id |
| 工作流优先 | 常见问题走固定 Skill Workflow，不依赖自由 Agent |
| Agent 受控 | Agent 只能调用 allowedTools / allowedSkills |
| 审计优先 | 所有工具调用必须记录 tool_call 和 audit_log |
| 核心稳定 | 新增 MCP / Skill / Agent 不能修改核心平台代码 |

### 3.3 不做什么

MVP 阶段不做：

1. 不做完整多租户。
2. 不做复杂 RBAC。
3. 不做生产自动修复。
4. 不做自动回滚、自动重启、自动扩容。
5. 不做复杂 Agent Marketplace。
6. 不做完整 Evidence Graph 可视化。
7. 不做完整测试平台。
8. 不做复杂发布风险系统。
9. 不做多 Agent 协作。
10. 不让 Agent 绕过 Tool Gateway。
11. 不支持任意代码插件。
12. 不支持上传任意脚本执行。
13. 不把 stdio MCP 作为 P0 目标。
14. 不把 Trace / RAG / ChatOps 作为 P0 强依赖。

---

## 4. MVP 范围重新定义

### 4.1 MVP 的核心目标

MVP 的目标不是一次性完成 AIOps / DevOps / TestOps 的所有场景，而是先完成一个 **可扩展的工程智能底座**。

P0 验收标准：

```text
1. Tool 能注册。
2. MCP Server 能注册，并同步 tools。
3. Skill 能注册，并执行声明式 Workflow。
4. Tool Gateway 能统一调用内置 Tool 和 MCP Tool。
5. Orchestrator 能根据用户输入选择 Skill。
6. 所有工具调用有审计。
7. 所有关键结果能生成 Evidence。
8. RCA 报告能引用 Evidence。
9. 新增 MCP / Skill 不改核心代码。
10. P0 所有工具只读。
```

---

## 5. 优先级规划

## 5.1 P0：平台底座

P0 只做底层能力，不追求场景数量。

```text
P0-0：项目骨架
P0-1：核心数据模型
P0-2：Tool Registry
P0-3：Tool Gateway
P0-4：Built-in Tool Adapter
P0-5：MCP Registry + MCP Adapter
P0-6：Skill Registry + Skill Runtime
P0-7：Evidence Store + Audit Log
P0-8：轻量 Orchestrator
P0-9：第一个最小 Pod 诊断 Skill
```

### P0 不做

```text
1. Agent Registry
2. 多 Agent
3. 完整 RAG
4. Trace Tool
5. 发布风险分析
6. 测试失败归因
7. Evidence Graph 可视化
8. 自动修复
9. 完整 RBAC
10. ChatOps 机器人
```

---

## 5.2 P1：核心场景

底座跑通后，再做核心诊断场景。

```text
P1-1：Pod 异常诊断
P1-2：服务 5xx / RT 升高诊断
P1-3：CI/CD 失败分析
P1-4：GitLab Tool
P1-5：Vue Web UI 简版
P1-6：RCA Report 页面
P1-7：Tool Call Audit 页面
P1-8：Registry 管理页面
```

---

## 5.3 P2：增强智能

```text
P2-1：Runbook 简单检索
P2-2：Incident Timeline
P2-3：Trace Tool
P2-4：ChatOps 机器人
P2-5：Agent Registry
P2-6：Agent Runtime
P2-7：Agent Planner
P2-8：Agent Execution Trace
```

---

## 5.4 P3：平台治理与高级场景

```text
P3-1：完整 RBAC
P3-2：Approval 审批流
P3-3：OPA Policy
P3-4：多租户
P3-5：Evidence Graph 可视化
P3-6：自动生成 Runbook
P3-7：自动生成 Postmortem
P3-8：发布风险分析
P3-9：测试失败归因
P3-10：多 Agent 协作
```

---

## 6. 默认技术栈

| 层级 | 默认选择 | 说明 |
|---|---|---|
| 后端语言 | Go | MVP 主链路统一 Go |
| Web/API 框架 | Gin | 默认使用 Gin |
| 前端框架 | Vue | Web UI 默认 Vue |
| 前端构建 | Vite | 配合 Vue 使用 |
| UI 组件 | Element Plus | MVP 阶段优先 |
| 数据库 | PostgreSQL | 存任务、报告、Registry、审计 |
| 缓存 | Redis | 存 session、任务状态、限流计数 |
| K8s SDK | client-go | Kubernetes 只读查询 |
| 日志 | zap / slog | 结构化日志 |
| 可观测 | OpenTelemetry Go SDK + Prometheus metrics | 服务自身可观测 |
| Workflow | 自研轻量状态机 | P0 不引入复杂工作流引擎 |
| Agent 编排 | P0 自研轻量 Agentic Workflow；P2 接入 EinoAgentRuntime | Eino 是增强层，不是平台地基 |
| MCP | Go 实现 MCP Adapter / Proxy | P0 只支持 Streamable HTTP |
| RAG | Go + pgvector / Qdrant / Milvus | P2 再做，P0 不强依赖 |

---

## 7. 服务拆分

### 7.1 MVP 推荐服务

P0 / P1 阶段建议控制在 3～5 个服务。

| 服务 | P0 是否必须 | 语言 | 职责 |
|---|---|---|---|
| aiops-server | 必须 | Go + Gin | API、Orchestrator、Registry、Skill Runtime、Report |
| tool-gateway | 必须 | Go | Tool 调用、MCP Adapter、Policy 预留、Audit、Evidence |
| worker | 可选 | Go | 异步任务、长耗时诊断、定时同步 MCP tools |
| web-ui | P1 | Vue | Chat、任务列表、报告、Registry、Audit 页面 |
| knowledge-service | P2 | Go | Runbook / SOP / 历史故障检索 |

### 7.2 P0 实际部署建议

P0 可以先部署为：

```text
aiops-server
tool-gateway
PostgreSQL
Redis
```

其中：

```text
aiops-server 内置：
  - API
  - Orchestrator
  - Registry
  - Skill Runtime
  - Report Builder

tool-gateway 内置：
  - Tool Registry Cache
  - Built-in Tool Adapter
  - MCP Adapter
  - Tool Invoke
  - Evidence Writer
  - Audit Writer
```

### 7.3 暂不拆出的模块

P0 不单独拆：

```text
orchestrator-service
registry-service
skill-runtime-service
agent-runtime-service
knowledge-service
audit-service
policy-service
```

这些先作为模块存在，后续再拆服务。

---

## 8. 总体架构

```text
                 ┌────────────────────┐
                 │     Vue Web UI      │
                 │       P1            │
                 └─────────┬──────────┘
                           │
                 ┌─────────▼──────────┐
                 │    aiops-server     │
                 │ Gin API             │
                 │ Orchestrator        │
                 │ Registry            │
                 │ Skill Runtime       │
                 │ Report Builder      │
                 └─────────┬──────────┘
                           │
                 ┌─────────▼──────────┐
                 │    tool-gateway     │
                 │ Tool Invoke         │
                 │ Built-in Adapter    │
                 │ MCP Adapter         │
                 │ Audit / Evidence    │
                 └─────┬────────┬─────┘
                       │        │
          ┌────────────▼───┐ ┌──▼────────────────┐
          │ Built-in Tools │ │ Third-party MCPs   │
          │ K8s / Prom /   │ │ Streamable HTTP    │
          │ Logs / GitLab  │ │ MCP Servers        │
          └────────────────┘ └───────────────────┘

旁路数据：
  ├── PostgreSQL：tasks / reports / tools / mcps / skills / evidences / audit_logs
  ├── Redis：session / task status / rate limit
  └── Object Storage，可选：大日志、报告附件、trace json
```

---

## 9. 核心抽象：Capability

### 9.1 设计目标

MCP、Skill、Agent 表面不同，但平台内部需要统一看成：

```text
Capability
```

类型包括：

```text
builtin_tool
mcp_tool
skill
agent
```

平台核心只认识 Capability，不直接绑定某个具体系统。

### 9.2 Capability 模型

```yaml
kind: Capability
metadata:
  name: k8s.describe_pod
  type: builtin_tool
  owner: sre-team
  version: v1
spec:
  description: Describe Kubernetes pod
  inputSchema: {}
  outputSchema: {}
  riskLevel: read_only
  scopes:
    envs:
      - dev
      - staging
      - prod
    namespaces:
      - "*"
  timeoutSeconds: 10
  enabled: true
```

### 9.3 Capability 统一接口，Go 示例

```go
type Capability interface {
    Name() string
    Type() CapabilityType
    Describe(ctx context.Context) CapabilitySpec
    Invoke(ctx context.Context, req InvokeRequest) (InvokeResult, error)
}

type CapabilityType string

const (
    CapabilityBuiltinTool CapabilityType = "builtin_tool"
    CapabilityMCPTool     CapabilityType = "mcp_tool"
    CapabilitySkill       CapabilityType = "skill"
    CapabilityAgent       CapabilityType = "agent"
)
```

### 9.4 统一调用请求

```go
type InvokeRequest struct {
    TaskID    string         `json:"task_id"`
    UserID    string         `json:"user_id"`
    Name      string         `json:"name"`
    Arguments map[string]any `json:"arguments"`
    Context   InvokeContext  `json:"context"`
}

type InvokeContext struct {
    Env       string `json:"env"`
    Cluster   string `json:"cluster"`
    Namespace string `json:"namespace"`
    Service   string `json:"service"`
    RiskLevel string `json:"risk_level"`
    TraceID   string `json:"trace_id"`
}
```

### 9.5 统一调用结果

```go
type InvokeResult struct {
    Success    bool   `json:"success"`
    Data       any    `json:"data"`
    EvidenceID string `json:"evidence_id"`
    RiskLevel  string `json:"risk_level"`
    DurationMs int64  `json:"duration_ms"`
    Error      string `json:"error,omitempty"`
}
```

---

## 10. Registry 设计

Registry 是平台控制平面，P0 必须优先做好。

### 10.1 Registry 类型

```text
Tool Registry
MCP Registry
Skill Registry
Agent Registry，P2
```

### 10.2 Registry 职责

```text
1. 注册
2. 校验
3. 发现
4. 启用 / 禁用
5. 版本管理
6. Owner 管理
7. Scope 管理
8. Risk 管理
9. Schema 管理
10. Health Check
11. Audit
```

---

## 11. Tool Registry

### 11.1 目标

Tool Registry 存储所有可调用工具，包括：

```text
1. 内置工具
2. MCP 同步出来的工具
```

### 11.2 ToolSpec 字段

```text
name
type
owner
source
description
input_schema
output_schema
risk_level
scope
timeout_seconds
enabled
created_at
updated_at
```

### 11.3 P0 API

```http
GET  /api/v1/tools
GET  /api/v1/tools/{name}
POST /api/v1/tools/{name}/enable
POST /api/v1/tools/{name}/disable
```

### 11.4 P0 内置工具

P0 先实现：

```text
k8s.describe_pod
k8s.get_events
prometheus.query
logs.search
```

P1 再加：

```text
gitlab.get_pipeline
gitlab.get_failed_jobs
gitlab.get_job_log
```

P2 再加：

```text
trace.query
release.get_recent_deployments
test.get_failed_cases
```

---

## 12. Tool Gateway

### 12.1 目标

Tool Gateway 是所有 Tool 的唯一执行入口。

Orchestrator、Skill、Agent 都不能直接访问外部系统，必须通过 Tool Gateway。

### 12.2 执行流程

```text
Orchestrator / Skill / Agent
    ↓
Tool Gateway
    ↓
查询 ToolSpec
    ↓
参数校验
    ↓
Policy Check，P0 简化
    ↓
Adapter Dispatch
    ├── BuiltinToolAdapter
    └── MCPToolAdapter
    ↓
外部系统
    ↓
结果标准化
    ↓
脱敏
    ↓
写 tool_calls
    ↓
写 evidences
    ↓
返回 InvokeResult
```

### 12.3 Tool Invoke API

```http
POST /api/v1/tools/invoke
```

请求：

```json
{
  "tool_name": "k8s.describe_pod",
  "arguments": {
    "cluster": "default",
    "namespace": "order",
    "pod": "order-service-xxx"
  },
  "context": {
    "user_id": "u_001",
    "task_id": "task_001",
    "risk_level": "read_only"
  }
}
```

响应：

```json
{
  "success": true,
  "tool_name": "k8s.describe_pod",
  "data": {},
  "evidence_id": "ev_001",
  "duration_ms": 321,
  "risk_level": "read_only",
  "masked": true
}
```

### 12.4 风险等级

| 等级 | 示例 | P0 行为 |
|---|---|---|
| read_only | 查询日志、describe pod、查询指标 | 允许 |
| low | 重新执行测试、重试 CI | P0 不执行 |
| medium | 重启非生产环境、扩容测试环境 | P0 不执行 |
| high | 生产回滚、生产重启 | P0 禁止 |
| critical | 删除资源、修改权限、清理数据 | P0 禁止 |

P0 明确约束：

```text
所有自动执行工具必须 read_only。
非 read_only 工具只能注册，不能执行。
```

---

## 13. MCP Registry + MCP Adapter

### 13.1 目标

第三方团队可以开发 MCP Server，并通过配置注册到平台。平台无需修改核心代码即可发现并调用 MCP 工具。

### 13.2 范围限制

P0 只支持：

```text
Streamable HTTP MCP Server
```

P0 不支持：

```text
stdio MCP
本地进程 MCP
旧版 HTTP+SSE
自定义 transport
```

### 13.3 MCP Registry 说明

MCP Registry 是平台自定义注册中心，不是 MCP 官方标准。

它管理：

```text
endpoint
transport
owner
auth
allowedTools
riskLevel
scope
timeout
rateLimit
enabled
health
```

### 13.4 MCP 配置示例

```yaml
kind: MCPServer
metadata:
  name: payment-prometheus-mcp
  owner: payment-team
  description: Payment team Prometheus MCP server
spec:
  transport: streamable_http
  endpoint: http://payment-prometheus-mcp.default.svc:8080/mcp
  auth:
    type: bearer
    secretRef: payment-prometheus-token
  allowedTools:
    - query_promql
    - list_alerts
  scopes:
    envs:
      - prod
      - staging
    namespaces:
      - payment
  riskLevel: read_only
  timeoutSeconds: 10
  rateLimit:
    qps: 5
  enabled: true
```

### 13.5 MCP 注册流程

```text
团队提交 mcp.yaml
    ↓
MCP Registry 校验配置
    ↓
MCP Adapter 调用 tools/list
    ↓
转换为 ToolSpec
    ↓
写入 Tool Registry
    ↓
Tool Gateway 可以调用
```

### 13.6 MCP P0 API

```http
POST /api/v1/mcp-servers
GET  /api/v1/mcp-servers
GET  /api/v1/mcp-servers/{id}
POST /api/v1/mcp-servers/{id}/sync-tools
POST /api/v1/mcp-servers/{id}/enable
POST /api/v1/mcp-servers/{id}/disable
```

### 13.7 MCP Adapter 职责

```text
1. 初始化连接
2. 调用 tools/list
3. 调用 tools/call
4. 处理 timeout
5. 处理认证
6. 处理 schema
7. 处理错误码
8. 结果标准化
9. health check
10. 写入 Tool Registry
```

---

## 14. Skill Registry + Skill Runtime

### 14.1 Skill 定义

Skill 是可复用任务流程，用于封装某类问题的排查经验。

Skill 不是任意代码插件。

P0 只支持：

```text
YAML Workflow
Prompt 模板
固定 Tool 调用
简单条件分支
LLM 总结
```

P0 不支持：

```text
上传 Go 插件
上传 Shell 脚本
上传任意二进制
执行任意代码
```

### 14.2 Skill 目录结构

```text
skills/
  pod_restart_diagnosis/
    skill.yaml
    workflow.yaml
    prompt.md
    examples.yaml
```

### 14.3 Skill 配置示例

```yaml
kind: Skill
metadata:
  name: pod_restart_diagnosis
  owner: sre-team
  description: Diagnose Kubernetes Pod restart issues
spec:
  triggers:
    - pod restart
    - pod 重启
    - crashloopbackoff
    - oomkilled
  requiredTools:
    - k8s.describe_pod
    - k8s.get_events
    - logs.search
    - prometheus.query
  workflowRef: workflow.yaml
  promptRef: prompt.md
  inputSchema:
    required:
      - namespace
      - pod
  outputSchema:
    fields:
      - summary
      - evidence
      - possible_causes
      - recommendations
      - risk
  riskPolicy:
    maxRiskLevel: read_only
  enabled: true
```

### 14.4 Skill Workflow 示例

```yaml
steps:
  - id: describe_pod
    type: tool
    tool: k8s.describe_pod

  - id: get_events
    type: tool
    tool: k8s.get_events

  - id: search_logs
    type: tool
    tool: logs.search

  - id: query_memory
    type: tool
    tool: prometheus.query
    args:
      promql_template: "container_memory_working_set_bytes{namespace='{{namespace}}', pod=~'{{pod}}.*'}"

  - id: generate_report
    type: llm
    prompt: prompt.md
```

### 14.5 Skill Runtime 执行流程

```text
加载 Skill
    ↓
校验 requiredTools 是否存在
    ↓
校验 riskPolicy
    ↓
校验 workflow steps
    ↓
按步骤调用 Tool Gateway
    ↓
每步生成 Evidence
    ↓
调用 LLM 生成报告
    ↓
输出 Report
```

### 14.6 Skill P0 API

```http
POST /api/v1/skills
GET  /api/v1/skills
GET  /api/v1/skills/{name}
POST /api/v1/skills/{name}/run
POST /api/v1/skills/{name}/enable
POST /api/v1/skills/{name}/disable
```

### 14.7 Skill 注册校验

注册 Skill 时必须校验：

```text
1. requiredTools 是否存在。
2. requiredTools 是否 enabled。
3. riskPolicy 是否超过工具风险。
4. workflow.yaml 是否合法。
5. prompt.md 是否存在。
6. inputSchema 是否完整。
7. outputSchema 是否合法。
```

---

---

## 15. Agent 编排选型策略

### 15.1 结论

MVP 阶段采用：

```text
P0：自研轻量 Agentic Workflow
P1：继续增强 Skill Runtime 与 Workflow 能力
P2：引入 Eino 作为 Agent Runtime 的一种实现
P3：再考虑多 Agent 协作、AgentAsTool、Supervisor Agent
```

核心原则：

> 平台核心自研，Agent 能力可插拔。  
> Eino 是增强层，不是平台地基。

---

### 15.2 为什么 P0 不直接使用 Eino 作为核心

P0 的核心目标不是构建最强 Agent，而是先建设可控的工程智能底座：

```text
1. Tool Registry
2. Tool Gateway
3. MCP Adapter
4. Skill Registry
5. Skill Runtime
6. Evidence Store
7. Audit Log
8. Orchestrator
```

这些能力无论是否使用 Eino 都需要平台自己设计。

如果 P0 阶段直接把 Orchestrator、Skill Runtime、Tool 调用链全部绑定到 Eino，可能带来以下问题：

```text
1. Registry 数据模型被框架抽象牵引。
2. Tool Gateway / Audit / Evidence 插入点不够清晰。
3. MCP Tool 与 Built-in Tool 的统一抽象容易混乱。
4. Skill YAML 的声明式 Workflow 不一定能完全按平台设计演进。
5. 后续如果替换 Agent 框架，核心链路改造成本高。
```

因此 P0 阶段采用自研轻量 Agentic Workflow，先保证平台底座稳定、可控、可审计、可扩展。

---

### 15.3 自研轻量 Agentic Workflow 的范围

这里的“自研”不是重写一个复杂 Agent 框架，而是实现一个最小可用的声明式 Workflow Runtime。

P0 仅支持：

```text
1. tool step
2. llm step
3. condition step，简单条件
4. report step
5. 顺序执行
6. 失败中断
7. 超时控制
8. Evidence 写入
9. Audit 写入
```

P0 不支持：

```text
1. 复杂 DAG
2. 多 Agent 协作
3. 反思 / 自我纠错
4. 长期记忆
5. 自动规划无限步
6. 任意代码执行
7. 自动生产变更
```

示例 Workflow：

```yaml
steps:
  - id: describe_pod
    type: tool
    tool: k8s.describe_pod

  - id: get_events
    type: tool
    tool: k8s.get_events

  - id: search_logs
    type: tool
    tool: logs.search

  - id: query_memory
    type: tool
    tool: prometheus.query

  - id: generate_report
    type: llm
    prompt: prompts/pod_restart.md
```

执行链路：

```text
Skill Runtime
    ↓
SimpleWorkflowRuntime
    ↓
Tool Gateway
    ↓
Adapter
    ↓
External System
    ↓
Evidence / Audit
    ↓
Report
```

---

### 15.4 Eino 的定位

Eino 作为 Go 生态下的 LLM / Agent 应用框架，可以在 P2 阶段作为 Agent Runtime 的一种实现接入。

Eino 适合用于：

```text
1. 开放式问题规划
2. Tool-use Agent
3. RAG Agent
4. 多步推理
5. AgentAsTool
6. 多 Agent 协作
7. 更复杂的 Agent Runtime
```

Eino 不负责：

```text
1. 平台 Registry 控制面
2. Tool Gateway 执行入口
3. Evidence 主链路
4. Audit 主链路
5. Policy / Approval 主链路
6. MCP Server 注册管理
```

正确关系：

```text
Orchestrator
    ↓
AgentRuntime interface
    ├── SimpleWorkflowRuntime，P0 自研
    └── EinoAgentRuntime，P2 引入
```

---

### 15.5 Runtime 接口设计

平台核心必须先定义自己的 Runtime 接口，避免被具体框架绑定。

```go
type WorkflowRuntime interface {
    Run(ctx context.Context, req WorkflowRunRequest) (*WorkflowRunResult, error)
}

type AgentRuntime interface {
    Run(ctx context.Context, req AgentRunRequest) (*AgentRunResult, error)
}

type ToolInvoker interface {
    Invoke(ctx context.Context, req ToolInvokeRequest) (*ToolInvokeResult, error)
}
```

P0 实现：

```text
WorkflowRuntime = SimpleWorkflowRuntime
ToolInvoker     = ToolGatewayClient
AgentRuntime    = 暂不开放第三方 Agent
```

P2 实现：

```text
AgentRuntime = EinoAgentRuntime
```

这样平台可以在不修改 Orchestrator 核心逻辑的情况下，引入 Eino 或其他 Agent 框架。

---

### 15.6 Eino 引入条件

只有满足以下条件后，才引入 Eino：

```text
1. Tool Gateway 已稳定。
2. MCP Adapter 已跑通。
3. Skill Runtime 已跑通。
4. Evidence / Audit 已稳定。
5. 至少 2～3 个 Skill 已跑通。
6. 出现固定 Workflow 难以覆盖的开放式问题。
```

适合 Eino 的开放式问题示例：

```text
1. 今天支付链路是不是整体不稳定？
2. 这次发布后是否影响上下游？
3. 这批测试失败是环境问题还是代码问题？
4. 最近一周哪个服务风险最高？
5. 某个团队最近的发布质量如何？
```

---

### 15.7 Eino 接入后的边界

即使引入 Eino，也必须遵守平台安全边界：

```text
Eino Agent
    ↓
Tool Gateway
    ↓
Policy Check
    ↓
Adapter
    ↓
External System
```

禁止：

```text
Eino Agent
    ↓
直接访问 Kubernetes / Prometheus / Logs / GitLab
```

Eino Agent 只能调用平台暴露的 Tool / Skill，且必须受以下约束：

```text
1. allowedTools
2. allowedSkills
3. maxSteps
4. maxTokens
5. maxDuration
6. maxRiskLevel
7. memoryScope
8. outputSchema
```

---

### 15.8 分阶段落地计划

```text
P0：
  - 自研 SimpleWorkflowRuntime
  - 自研 Skill Runtime
  - 自研 Tool Gateway
  - 自研 MCP Adapter
  - Evidence / Audit 主链路

P1：
  - 增强 Workflow 能力
  - 完善 2～3 个诊断 Skill
  - 完善 Vue 页面

P2：
  - 引入 EinoAgentRuntime
  - 支持 Agent Registry
  - 支持 Planner Agent
  - 支持复杂开放问题

P3：
  - 多 Agent 协作
  - AgentAsTool
  - Supervisor Agent
  - 更复杂的 RAG / Planning / Reflection
```

---

### 15.9 选型结论

最终选型为：

```text
P0：自研轻量 Agentic Workflow，保证平台底座可控。
P2：引入 Eino，增强 Agent 能力。
```

一句话：

> Eino 是 Agent 能力增强层，不是平台基础设施层。  
> 平台核心必须掌握 Registry、Tool Gateway、Skill Runtime、Evidence、Audit 的主动权。

## 16. Agent 设计，P2

### 15.1 为什么 Agent 不放 P0

Agent 比 Skill 风险更高，涉及：

```text
1. 自主规划
2. 多步工具调用
3. allowedTools
4. allowedSkills
5. maxSteps
6. memoryScope
7. model config
8. prompt
9. riskPolicy
10. execution trace
11. 输出校验
```

因此 P0 不做第三方 Agent Registry。

### 15.2 P0 内置 Agent 能力

P0 只保留内置能力：

```text
1. Intent Parser
2. Report Generator
3. RCA Summarizer
```

这些不作为第三方 Agent 注册。

### 15.3 P2 Agent Registry 示例

```yaml
kind: Agent
metadata:
  name: sre-rca-agent
  owner: sre-team
  description: SRE root cause analysis agent
spec:
  model: gpt-4.1
  maxSteps: 8
  maxDurationSeconds: 60
  memoryScope: session
  allowedTools:
    - k8s.describe_pod
    - k8s.get_events
    - logs.search
    - prometheus.query
  allowedSkills:
    - pod_restart_diagnosis
  riskPolicy:
    maxAutoRiskLevel: read_only
    requireApproval:
      - release.rollback
      - k8s.restart_deployment
  outputSchema:
    fields:
      - summary
      - evidence
      - root_causes
      - recommendations
      - next_steps
  enabled: true
```

### 15.4 Agent 强约束

```text
1. Agent 只能调用 allowedTools。
2. Agent 只能调用 allowedSkills。
3. Agent 不能绕过 Tool Gateway。
4. Agent 不能直接访问外部系统。
5. Agent 不能直接执行 Shell。
6. Agent 不能直接访问生产凭证。
7. Agent 输出必须引用 evidence_id。
8. Agent 失败必须返回失败原因。
```

---

## 17. Orchestrator 轻量版

### 16.1 P0 目标

P0 Orchestrator 不做复杂 Agent Planner，只做轻量路由：

```text
用户输入
    ↓
Intent Parser
    ↓
参数抽取
    ↓
匹配 Skill
    ↓
执行 Skill Runtime
    ↓
收集 Evidence
    ↓
生成 Report
```

### 16.2 Intent Parser 策略

P0 采用：

```text
规则优先 + LLM fallback
```

规则示例：

```text
包含 “pod 重启 / crashloop / oom / oomkilled” → pod_restart_diagnosis
包含 “5xx / 错误率 / rt / 延迟” → service_5xx_diagnosis，P1
包含 “pipeline / ci / job failed / 构建失败” → cicd_failure_analysis，P1
```

### 16.3 Orchestrator 输出

```json
{
  "intent": "pod_restart_diagnosis",
  "entities": {
    "env": "prod",
    "cluster": "default",
    "namespace": "order",
    "service": "order-service",
    "pod": "order-service-xxx",
    "time_range": "30m"
  },
  "matched_skill": "pod_restart_diagnosis",
  "confidence": 0.91,
  "need_clarification": false
}
```

### 16.4 多轮对话

P0 只做最小多轮：

```text
如果 namespace / pod / service 缺失，返回 clarification_required。
```

例如：

```json
{
  "status": "clarification_required",
  "question": "请提供 namespace 或服务名。"
}
```

---

## 18. Evidence Store

### 17.1 目标

Evidence 是 RCA 报告的基础。  
LLM 不能空口判断，必须基于 Evidence 总结。

### 17.2 Evidence 类型

```text
k8s_pod_status
k8s_event
k8s_log
metric_series
git_pipeline
git_job_log
mcp_result
runbook_chunk，P2
manual_note，P2
```

### 17.3 Evidence 数据结构

```json
{
  "evidence_id": "ev_001",
  "task_id": "task_001",
  "source": "k8s.describe_pod",
  "type": "k8s_pod_status",
  "title": "Pod last state is OOMKilled",
  "content": {},
  "summary": "Pod exited with OOMKilled reason.",
  "timestamp": "2026-06-14T10:20:00Z",
  "confidence": 0.95,
  "metadata": {
    "namespace": "order",
    "pod": "order-service-xxx"
  }
}
```

### 17.4 P0 约束

```text
1. 每个 tool_call 至少生成一条 evidence 或 error evidence。
2. 报告中的关键结论必须引用 evidence_id。
3. 如果证据不足，报告必须明确说明“不确定”。
```

---

## 19. Audit Log

### 18.1 目标

所有工具调用必须可追踪、可复盘。

### 18.2 Audit 字段

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

### 18.3 P0 审计要求

```text
1. 所有 Tool Gateway 调用必须写 audit_log。
2. 所有 MCP tools/call 必须写 audit_log。
3. 所有失败调用也必须写 audit_log。
4. 参数需要脱敏后入库。
```

---

## 20. Report 设计

### 19.1 RCA Report 结构

```json
{
  "report_id": "rpt_001",
  "task_id": "task_001",
  "title": "order-service Pod restart diagnosis",
  "summary": "order-service is likely restarting due to OOMKilled.",
  "impact": "order-service availability may be affected.",
  "evidence": [
    {
      "evidence_id": "ev_001",
      "summary": "Pod last state is OOMKilled."
    }
  ],
  "possible_causes": [
    {
      "cause": "Memory limit too low or memory leak",
      "confidence": 0.86,
      "evidence_ids": ["ev_001", "ev_002"]
    }
  ],
  "recommended_actions": [
    {
      "action": "Check memory limit and recent release.",
      "risk_level": "read_only",
      "requires_approval": false
    }
  ],
  "next_steps": [],
  "need_approval": false
}
```

### 19.2 P0 Report 要求

```text
1. 有摘要。
2. 有证据列表。
3. 有可能原因 Top 3。
4. 有建议动作。
5. 有风险等级。
6. 有下一步排查命令。
7. 有 evidence_id 引用。
8. 证据不足时必须说明。
```

---

## 21. 数据模型

### 20.1 PostgreSQL 表

P0 表：

```text
tasks
reports
tools
mcp_servers
skills
workflows
tool_calls
evidences
audit_logs
```

P1 表：

```text
gitlab_pipelines_cache
diagnosis_templates
```

P2 表：

```text
agents
agent_runs
agent_steps
runbook_docs
runbook_chunks
timeline_events
```

P3 表：

```text
users
teams
roles
permissions
approvals
policies
```

### 20.2 Redis

```text
session context
task status cache
rate limit counter
tool result cache
mcp health status cache
```

---

## 22. Auth / RBAC / Approval 预留

### 21.1 P0 预留字段

所有请求上下文保留：

```json
{
  "user_id": "u_001",
  "team": "sre",
  "roles": ["admin"],
  "env_scope": ["dev", "staging", "prod"]
}
```

### 21.2 Policy Check 预留

Tool Gateway 内部保留接口：

```go
type PolicyDecision string

const (
    PolicyAllow            PolicyDecision = "allow"
    PolicyDeny             PolicyDecision = "deny"
    PolicyApprovalRequired PolicyDecision = "approval_required"
)

type PolicyChecker interface {
    Check(ctx context.Context, req PolicyRequest) (PolicyDecision, error)
}
```

P0 默认策略：

```text
read_only：allow
low：deny
medium：deny
high：deny
critical：deny
```

### 21.3 Approval 预留

P0 只保留状态，不做审批流。

```text
approval_required
approved
rejected
expired
```

---

## 23. P0 详细里程碑

### M0：项目骨架

目标：

```text
1. Go module 初始化。
2. Gin aiops-server。
3. tool-gateway。
4. PostgreSQL。
5. Redis。
6. Docker Compose。
7. /healthz。
8. /metrics。
9. 基础配置管理。
```

验收：

```text
服务可启动。
数据库可连接。
healthz 正常。
metrics 可访问。
```

---

### M1：核心数据模型

目标：

```text
1. tasks 表。
2. reports 表。
3. tools 表。
4. mcp_servers 表。
5. skills 表。
6. workflows 表。
7. tool_calls 表。
8. evidences 表。
9. audit_logs 表。
```

验收：

```text
可以创建 task。
可以写入 tool_call。
可以写入 evidence。
可以写入 audit_log。
```

---

### M2：Tool Registry

目标：

```text
1. 注册内置 Tool。
2. 查询 Tool。
3. 启用 / 禁用 Tool。
4. ToolSpec schema。
5. risk_level。
6. scope。
```

验收：

```text
GET /api/v1/tools 返回内置工具列表。
k8s.describe_pod、k8s.get_events、prometheus.query、logs.search 可被注册。
```

---

### M3：Tool Gateway

目标：

```text
1. POST /api/v1/tools/invoke。
2. 根据 tool_name 找 ToolSpec。
3. 参数校验。
4. risk_level 校验。
5. Adapter Dispatch。
6. tool_call 记录。
7. evidence 记录。
8. audit_log 记录。
```

验收：

```text
调用 k8s.describe_pod 可得到标准 InvokeResult。
调用失败也能记录 tool_call 和 audit_log。
```

---

### M4：Built-in Tool Adapter

目标：

```text
1. k8s.describe_pod。
2. k8s.get_events。
3. prometheus.query。
4. logs.search，先用 mock 或单一后端。
```

验收：

```text
4 个内置工具可以通过 Tool Gateway 调用。
```

---

### M5：MCP Registry + MCP Adapter

目标：

```text
1. 注册 MCP Server。
2. 同步 tools/list。
3. 转换 ToolSpec。
4. tools/call 调用。
5. health check 简版。
```

验收：

```text
新增 mcp.yaml 后，不改 Go 代码即可出现新的 MCP Tool。
MCP Tool 可以通过 Tool Gateway 调用。
```

---

### M6：Skill Registry + Skill Runtime

目标：

```text
1. 注册 Skill。
2. 校验 requiredTools。
3. 加载 workflow.yaml。
4. 执行 tool step。
5. 执行 llm step。
6. 生成报告。
```

验收：

```text
新增 pod_restart_diagnosis skill 后，不改 Go 代码即可执行。
Skill 执行过程会生成 tool_call、evidence、report。
```

---

### M7：轻量 Orchestrator

目标：

```text
1. 用户输入。
2. Intent Parser。
3. 参数抽取。
4. Skill 匹配。
5. Skill Runtime 执行。
6. Report Builder。
```

验收：

```text
用户输入“帮我看 order-service 为什么重启”，系统能匹配 pod_restart_diagnosis Skill。
```

---

### M8：第一个端到端场景

目标：

```text
Pod 重启诊断端到端跑通。
```

流程：

```text
用户输入
    ↓
Intent Parser
    ↓
pod_restart_diagnosis Skill
    ↓
Tool Gateway
    ├── k8s.describe_pod
    ├── k8s.get_events
    ├── logs.search
    └── prometheus.query
    ↓
Evidence Store
    ↓
Report Builder
    ↓
RCA Report
```

验收：

```text
报告包含 summary、evidence、possible_causes、recommendations、next_steps。
```

---

## 24. P1 里程碑

### P1-M1：服务 5xx 诊断

```text
1. service_5xx_diagnosis Skill。
2. prometheus.query 支持错误率和 RT 查询模板。
3. logs.search 支持错误日志 TopN。
4. RCA 报告输出。
```

### P1-M2：CI/CD 失败分析

```text
1. GitLab Tool。
2. get pipeline。
3. get failed jobs。
4. get job log。
5. cicd_failure_analysis Skill。
```

### P1-M3：Vue Web UI

页面：

```text
1. Chat 输入页。
2. Diagnosis Task 列表。
3. Report 详情页。
4. Tool Call Audit 页。
5. Registry 管理页。
```

---

## 25. 插件化设计验收标准

### 24.1 新增 MCP 不改核心代码

理想流程：

```text
新增 mcp.yaml
    ↓
Registry 注册
    ↓
MCP Adapter sync tools
    ↓
Tool Registry 出现新工具
    ↓
Tool Gateway 可调用
```

验收标准：

```text
新增 MCP Server 不需要修改 aiops-server / tool-gateway 代码。
```

---

### 24.2 新增 Skill 不改核心代码

理想流程：

```text
新增 skill.yaml + workflow.yaml + prompt.md
    ↓
Skill Registry 注册
    ↓
Skill Runtime 校验
    ↓
Orchestrator 可匹配
    ↓
用户自然语言可触发
```

验收标准：

```text
新增 Skill 不需要在 Orchestrator 中新增 if else。
```

---

### 24.3 Orchestrator 只依赖抽象

Orchestrator 只能依赖：

```text
Capability
Tool
Skill
Evidence
Report
```

不应该依赖：

```text
Prometheus HTTP API 细节
Kubernetes client-go 细节
GitLab API 细节
MCP JSON-RPC 细节
```

---

### 24.4 Agent 不能绕过 Tool Gateway

即使 P2 加 Agent，也必须满足：

```text
Agent
    ↓
Tool Gateway
    ↓
Policy Check
    ↓
Adapter
    ↓
External System
```

---

### 24.5 报告必须能回溯证据

报告中每个关键结论必须关联：

```text
evidence_id
tool_call_id
source
timestamp
```

---

## 26. 需要进一步核对的问题

开工前必须确认：

### 25.1 日志后端

必须确认 P0 使用哪种日志后端：

```text
Loki
Elasticsearch
阿里云 SLS
Mock Logs API
```

建议：

```text
P0 先用 Loki 或 Mock Logs API。
```

### 25.2 Prometheus 指标标准

需要确认是否存在：

```text
http_requests_total
http_request_duration_seconds_bucket
container_memory_working_set_bytes
kube_pod_container_status_restarts_total
```

以及 label：

```text
namespace
pod
service
status_code
path
method
```

### 25.3 K8s 访问方式

需要确认：

```text
kubeconfig
in-cluster service account
单集群
多集群
```

建议：

```text
P0 单集群 + kubeconfig / in-cluster 二选一。
```

### 25.4 LLM 接入方式

需要确认：

```text
OpenAI-compatible API
Claude-compatible API
Gemini-compatible API
内部 new-api / LiteLLM / one-api
```

建议配置：

```yaml
model:
  provider: openai_compatible
  baseURL: http://model-gateway.local/v1
  model: gpt-4.1-mini
```

### 25.5 MCP 协议范围

P0 明确：

```text
只支持 Streamable HTTP MCP。
```

### 25.6 是否完全只读

P0 建议：

```text
只读。
不执行 restart。
不执行 scale。
不执行 rollback。
不执行 delete。
不执行 apply。
只生成建议命令。
```

### 25.7 前端是否 P1

如果目标是尽快验证底座，P0 可以先 API + 简单页面。  
如果需要展示效果，Vue Web UI 放 P1 早期。

---

## 27. 推荐代码结构

项目采用 **单仓库 Monorepo + 单 Go Module**。  
第一阶段不拆多个 repo，避免服务边界尚未稳定时增加治理成本。

结构目标：

```text
1. 目录清晰。
2. 抽象明确。
3. 便于扩展。
4. 适合 Go + Gin 项目习惯。
5. 支持 MCP / Skill / Agent 插件化扩展。
6. 避免把核心逻辑写成普通 CRUD。
```

---

### 27.1 总体结构

```text
engineering-copilot/
├── go.mod
├── go.sum
├── Makefile
├── README.md
├── .env.example
├── .gitignore
│
├── cmd/
│   ├── aiops-server/
│   │   └── main.go
│   ├── tool-gateway/
│   │   └── main.go
│   └── worker/
│       └── main.go
│
├── internal/
│   ├── bootstrap/
│   ├── constants/
│   ├── utils/
│   ├── model/
│   ├── repository/
│   ├── router/
│   ├── handler/
│   ├── service/
│   ├── adapter/
│   └── infra/
│
├── configs/
│   ├── app.yaml
│   ├── tools/
│   ├── mcps/
│   ├── skills/
│   ├── workflows/
│   └── prompts/
│
├── migrations/
├── api/
├── docs/
├── deploy/
├── web/
└── scripts/
```

---

### 27.2 internal 目录结构

```text
internal/
├── bootstrap/
│   ├── app.go
│   ├── config.go
│   ├── logger.go
│   ├── database.go
│   ├── redis.go
│   └── telemetry.go
│
├── constants/
│   ├── risk.go
│   ├── capability.go
│   ├── task.go
│   ├── tool.go
│   └── error_code.go
│
├── utils/
│   ├── idgen/
│   │   └── idgen.go
│   ├── mask/
│   │   └── mask.go
│   ├── timeutil/
│   │   └── time.go
│   ├── jsonschema/
│   │   └── validate.go
│   └── errors/
│       └── errors.go
│
├── model/
│   ├── entity/
│   │   ├── task.go
│   │   ├── report.go
│   │   ├── tool.go
│   │   ├── mcp_server.go
│   │   ├── skill.go
│   │   ├── workflow.go
│   │   ├── evidence.go
│   │   └── audit_log.go
│   ├── dto/
│   │   ├── chat.go
│   │   ├── task.go
│   │   ├── tool.go
│   │   ├── mcp.go
│   │   ├── skill.go
│   │   ├── report.go
│   │   └── audit.go
│   └── vo/
│       ├── capability.go
│       ├── risk.go
│       ├── user_context.go
│       └── invoke.go
│
├── repository/
│   ├── task_repository.go
│   ├── report_repository.go
│   ├── tool_repository.go
│   ├── mcp_repository.go
│   ├── skill_repository.go
│   ├── workflow_repository.go
│   ├── evidence_repository.go
│   └── audit_repository.go
│
├── router/
│   ├── router.go
│   ├── middleware.go
│   └── routes/
│       ├── health.go
│       ├── chat.go
│       ├── task.go
│       ├── tool.go
│       ├── mcp.go
│       ├── skill.go
│       ├── report.go
│       └── audit.go
│
├── handler/
│   ├── chat_handler.go
│   ├── task_handler.go
│   ├── tool_handler.go
│   ├── mcp_handler.go
│   ├── skill_handler.go
│   ├── report_handler.go
│   └── audit_handler.go
│
├── service/
│   ├── orchestrator/
│   │   ├── service.go
│   │   ├── intent_parser.go
│   │   ├── skill_matcher.go
│   │   ├── context_resolver.go
│   │   └── report_builder.go
│   ├── registry/
│   │   ├── tool_registry.go
│   │   ├── mcp_registry.go
│   │   └── skill_registry.go
│   ├── toolgateway/
│   │   ├── service.go
│   │   ├── invoker.go
│   │   ├── dispatcher.go
│   │   ├── policy.go
│   │   ├── sanitizer.go
│   │   └── evidence_writer.go
│   ├── skillruntime/
│   │   ├── runtime.go
│   │   ├── workflow_executor.go
│   │   ├── step_runner.go
│   │   └── validator.go
│   ├── workflow/
│   │   ├── parser.go
│   │   ├── executor.go
│   │   └── condition.go
│   ├── evidence/
│   │   ├── service.go
│   │   └── builder.go
│   ├── audit/
│   │   └── service.go
│   └── report/
│       ├── service.go
│       └── formatter.go
│
├── adapter/
│   ├── builtin/
│   │   ├── k8s/
│   │   │   ├── adapter.go
│   │   │   ├── describe_pod.go
│   │   │   └── get_events.go
│   │   ├── prometheus/
│   │   │   ├── adapter.go
│   │   │   └── query.go
│   │   ├── logs/
│   │   │   ├── adapter.go
│   │   │   ├── mock.go
│   │   │   └── loki.go
│   │   └── gitlab/
│   │       ├── adapter.go
│   │       └── pipeline.go
│   ├── mcp/
│   │   ├── adapter.go
│   │   ├── client.go
│   │   ├── transport.go
│   │   ├── tools_list.go
│   │   ├── tools_call.go
│   │   └── mapper.go
│   ├── llm/
│   │   ├── client.go
│   │   ├── openai_compatible.go
│   │   └── prompt.go
│   └── agent/
│       └── eino/
│           └── README.md
│
└── infra/
    ├── db/
    │   └── postgres.go
    ├── redis/
    │   └── redis.go
    ├── k8s/
    │   └── client.go
    ├── prometheus/
    │   └── client.go
    ├── logs/
    │   └── loki_client.go
    └── gitlab/
        └── client.go
```

---

### 27.3 目录职责说明

| 目录 | 职责 | 注意事项 |
|---|---|---|
| `cmd/` | 服务启动入口 | 只放 `main.go`，不写业务逻辑 |
| `bootstrap/` | 初始化配置、日志、DB、Redis、Telemetry | 负责装配依赖 |
| `constants/` | 常量、枚举、错误码 | 不放业务配置 |
| `utils/` | 纯工具函数 | 禁止放业务逻辑，避免变成垃圾桶 |
| `model/entity` | 数据库实体 | 对应数据库表 |
| `model/dto` | API 请求/响应结构 | 面向 HTTP 接口 |
| `model/vo` | 业务值对象 | 如 RiskLevel、CapabilityType、UserContext |
| `repository/` | 数据访问层 | 只做 CRUD，不写业务判断 |
| `router/` | Gin 路由注册 | 只做路由分组和中间件挂载 |
| `handler/` | HTTP Handler | 负责参数绑定、调用 service、返回响应 |
| `service/` | 核心业务逻辑 | Orchestrator、Registry、ToolGateway、SkillRuntime 都放这里 |
| `adapter/` | 外部能力适配层 | Built-in Tool、MCP、LLM、未来 Eino 都在这里 |
| `infra/` | 基础设施 Client | DB、Redis、K8s、Prometheus、Loki、GitLab client |
| `configs/` | 声明式配置 | MCP、Skill、Workflow、Prompt |
| `migrations/` | 数据库迁移 | 版本化管理表结构 |
| `web/` | Vue Web UI | P1 开始完善 |

---

### 27.4 核心边界原则

项目虽然采用常见 Go Web 分层：

```text
model
repository
service
handler
router
adapter
infra
```

但它不是普通 CRUD 项目。

必须保留以下核心边界：

```text
service/orchestrator
service/registry
service/toolgateway
service/skillruntime
adapter/mcp
adapter/builtin
adapter/llm
```

原因：

```text
1. Orchestrator 负责意图识别与任务编排。
2. Registry 负责 MCP / Skill / Tool 注册发现。
3. ToolGateway 是唯一工具执行入口。
4. SkillRuntime 负责声明式 Workflow 执行。
5. Adapter 负责隔离外部系统。
6. Evidence / Audit 负责可追溯性。
```

---

### 27.5 依赖方向

推荐依赖方向：

```text
handler
  ↓
service
  ↓
repository / adapter
  ↓
infra
```

禁止反向依赖：

```text
repository 不依赖 service
adapter 不依赖 handler
infra 不依赖 service
model 不依赖 Gin / DB client / K8s client
```

核心规则：

```text
1. Handler 不写业务逻辑。
2. Repository 不写业务判断。
3. Adapter 不绕过 Tool Gateway。
4. Orchestrator 不直接调用 K8s / Prometheus / Logs / GitLab。
5. SkillRuntime 调工具必须通过 ToolGateway。
6. AgentRuntime 调工具也必须通过 ToolGateway。
```

---

### 27.6 P0 最小结构

P0 阶段可以先只保留：

```text
internal/
├── bootstrap/
├── constants/
├── utils/
├── model/
├── repository/
├── router/
├── handler/
├── service/
│   ├── orchestrator/
│   ├── registry/
│   ├── toolgateway/
│   ├── skillruntime/
│   ├── evidence/
│   ├── audit/
│   └── report/
├── adapter/
│   ├── builtin/
│   ├── mcp/
│   └── llm/
└── infra/
```

P0 暂不展开：

```text
adapter/agent/eino
service/agentruntime
knowledge-service
完整 RBAC
approval
```

---

### 27.7 判断结构是否合格的标准

合格标准：

```text
1. 新增 MCP 不需要改 Orchestrator。
2. 新增 Skill 不需要改 Orchestrator。
3. Orchestrator 不直接依赖外部系统 SDK。
4. ToolGateway 是所有工具调用唯一入口。
5. Adapter 只负责外部系统适配，不写平台业务。
6. Repository 只负责数据访问。
7. Handler 只负责 HTTP 入参和出参。
8. Evidence / Audit 在工具调用链路中天然产生。
9. P2 接入 Eino 时，只新增 adapter/agent/eino 和 AgentRuntime 实现，不重构平台核心。
```



## 28. 成功标准

这个平台的 P0 成功，不是看诊断场景有多少，而是看底座是否成立。

P0 成功标准：

```text
1. 内置 Tool 可以统一注册和调用。
2. MCP Tool 可以通过注册接入，不改代码。
3. Skill 可以通过 YAML 注册和执行，不改代码。
4. Orchestrator 不需要知道具体外部系统细节。
5. Tool Gateway 是唯一工具执行入口。
6. 所有调用都有 Audit。
7. 所有关键结果都有 Evidence。
8. Report 能引用 Evidence。
9. P0 全部 read_only。
10. 新增能力不会破坏核心架构。
```

---

## 29. 结论

当前项目应该按如下路线推进：

```text
先做平台底座：
  Tool Registry
  Tool Gateway
  MCP Adapter
  Skill Registry
  Skill Runtime
  Evidence
  Audit
  Orchestrator

再做诊断场景：
  Pod 诊断
  服务 5xx
  CI/CD 失败

最后做智能增强：
  Runbook RAG
  Timeline
  Agent Registry
  Agent Planner
  ChatOps
  RBAC / Approval
```

一句话：

> MVP 不要先做“大而全 AIOps”，而要先做“可插拔的工程智能底座”。  
> 只要 MCP / Skill / Agent 的适配层设计优秀，后续运维、研发、测试场景都可以自然扩展。
