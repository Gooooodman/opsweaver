## Why

平台需要一个唯一、受控、可审计的工具执行入口，避免 Orchestrator、Skill 或未来 Agent 直接访问 Kubernetes、Prometheus、日志系统或 MCP Server。本 Change 建立 P0 的核心工具执行面和证据链。

## What Changes

- 实现数据库权威的 Tool Registry，支持查询、启用、禁用和版本选择。
- 实现 Tool Gateway 内部 HTTP/JSON Invoke API。
- 在调用前执行 JSON Schema、Scope、Risk 和幂等校验。
- 仅允许 `read_only` Tool 自动执行，其他风险等级统一拒绝。
- 实现 Tool Call、Evidence、Error Evidence 和 Audit 的原子持久化。
- 实现 `k8s.describe_pod`、`k8s.get_events`、`prometheus.query`、`logs.search` 四个内置 Tool。
- Kubernetes 支持单集群 kubeconfig/in-cluster 二选一；日志后端仅实现 Mock Logs API。

## Capabilities

### New Capabilities

- `tool-registry`: Tool 定义、不可变版本、默认启用版本和启停查询行为。
- `tool-invocation`: Tool Gateway 的校验、分发、幂等和标准化响应行为。
- `evidence-audit`: Tool Call、Evidence 与 Audit 的完整记录和脱敏行为。
- `builtin-tools`: Kubernetes、Prometheus 和 Mock Logs 内置只读 Tool 行为。

### Modified Capabilities

无。

## Impact

- 扩展 `aiops-server` 的 Tool Registry API 和内部 ToolSpec 同步接口。
- 扩展 `tool-gateway` 的 Invoke API、数据库模型、Repository 和 Adapter。
- 新增 client-go、JSON Schema 校验器及外部 HTTP Client 依赖。
- 依赖 `bootstrap-platform-foundation` 提供的服务、数据库、配置和内部认证。
