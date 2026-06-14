## Context

本 Change 将第三方 MCP Tool 与声明式 Skill 接入既有 Tool Gateway。Registry 仍由 `aiops-server` 管理，Worker 负责异步同步和 Workflow 执行，Tool Gateway 负责 MCP 网络调用和证据审计。

## Goals / Non-Goals

**Goals:**

- Streamable HTTP MCP 注册、同步、健康检查和调用。
- 文件导入与 API 注册统一的 Skill 校验链。
- 固定版本快照和受限顺序 Workflow。
- OpenAI-compatible 结构化报告草稿。

**Non-Goals:**

- 不实现 stdio、旧 SSE、多 Agent、复杂 DAG、循环或并行步骤。
- 不执行任意代码。
- 不实现最终 Pod 场景路由。

## Decisions

### 决策 1：最小 MCP Client 封装

定义平台自有 MCP Client 接口并在 Adapter 内封装选定 Go SDK；核心层只依赖 `ListTools`、`CallTool`、`Ping`。这避免 SDK 类型扩散，也允许在 SDK 不满足时替换实现。

### 决策 2：MCP 同步走异步任务

注册只保存并校验配置，随后提交 `mcp.sync-tools` Asynq 任务。Worker 请求 Tool Gateway 执行远程 `tools/list`，再将映射结果回写 `aiops-server`。

### 决策 3：Skill 内容入库

配置文件仅为导入源。Skill、Workflow 与 Prompt 内容在注册时规范化并保存不可变版本，Worker 获取完整快照，执行时不依赖共享文件系统。

### 决策 4：小型解释器而非通用工作流引擎

Workflow 使用判别联合类型表示四种 Step，顺序执行；Condition 仅支持对前序结构化输出的有限比较操作。相比引入工作流引擎，该方案满足 P0 且更易审计。

### 决策 5：结构化 LLM 边界

LLM Client 仅暴露基于 JSON Schema 的 Generate 方法。若 Provider 不支持原生 Structured Outputs，则使用 JSON 模式与本地校验；修复请求最多一次。

## Risks / Trade-offs

- [Risk] MCP SDK 与协议版本变化 → 封装 Client 接口并增加模拟 Server Contract 测试。
- [Risk] Condition 表达式演化为脚本语言 → P0 只允许固定操作符和 JSON Path 子集。
- [Risk] Prompt 或快照体积增长 → 设置注册大小上限并在数据库使用文本/JSONB。
- [Risk] 不同 OpenAI-compatible Provider 行为不一致 → Contract 测试覆盖原生 Schema 和 JSON fallback 两种路径。

## Migration Plan

1. 应用 MCP、Skill、Workflow 和凭证迁移。
2. 配置 LLM 凭证并验证加解密。
3. 注册模拟 MCP 与示例 Skill。
4. 运行同步、Workflow 和 LLM Contract 测试。
5. 回滚时禁用相关版本，保留历史定义和审计记录。

## Open Questions

无阻塞问题。具体 MCP Go SDK 在实施前通过小型兼容性测试确定，但不得改变平台 Client 接口。
