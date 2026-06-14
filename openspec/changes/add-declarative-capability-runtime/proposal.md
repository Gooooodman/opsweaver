## Why

Tool 执行面建立后，平台还需要让第三方 MCP Server 和声明式 Skill 无需修改核心代码即可注册和运行。本 Change 提供插件化能力接入、受控 Workflow 执行和基于 Evidence 的结构化报告生成。

## What Changes

- 实现数据库权威的 MCP Registry，P0 仅支持 Streamable HTTP。
- 通过异步同步任务调用 MCP `tools/list`，将远端 Tool 映射为本地 ToolSpec。
- 实现 MCP `tools/call`、健康检查、认证、超时和本地约束收窄。
- 实现 Skill Registry，支持文件导入和 API 注册共用同一校验链。
- 实现不可变版本快照和 `tool`、`condition`、`llm`、`report` 顺序 Workflow。
- 实现 OpenAI-compatible LLM Client、结构化输出校验和一次修复重试。
- 生成必须引用当前任务 Evidence ID 的报告草稿。

## Capabilities

### New Capabilities

- `mcp-integration`: MCP 注册、Streamable HTTP 同步、调用和健康检查行为。
- `skill-registry`: Skill/Workflow 文件导入、API 注册、校验和不可变版本行为。
- `workflow-runtime`: 声明式 Workflow 快照、步骤执行、关键步骤和失败策略行为。
- `llm-reporting`: OpenAI-compatible LLM 调用、结构化输出和 Evidence 引用行为。

### Modified Capabilities

无。

## Impact

- 扩展 `aiops-server` 的 MCP、Skill、Workflow Registry API 和内部快照 API。
- 扩展 `worker` 的 MCP 同步和 Workflow 执行能力。
- 扩展 `tool-gateway` 的 MCP Adapter 和管理接口。
- 新增 YAML、MCP Client 与 OpenAI-compatible Client 相关依赖。
- 依赖前两个 Change 提供的异步队列、Tool Registry、Tool Invoke、Evidence 和 Audit。
