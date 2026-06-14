## ADDED Requirements

### Requirement: Tool Gateway 是唯一执行入口
Worker、Skill Runtime 和未来 Agent Runtime MUST 通过 Tool Gateway 的内部 HTTP API 调用 Tool。

#### Scenario: 调用已启用 Tool
- **WHEN** Worker 携带有效服务令牌提交 InvokeRequest
- **THEN** Tool Gateway 完成校验、分发并返回标准 InvokeResult

### Requirement: 调用前执行强制校验
Tool Gateway MUST 校验 Tool 启用状态、JSON Schema、Scope 和 Risk Level；P0 仅允许 `read_only`。

#### Scenario: 非只读 Tool
- **WHEN** InvokeRequest 指向 `low` 或更高风险 Tool
- **THEN** Tool Gateway 拒绝执行并返回英文策略错误

#### Scenario: 参数不符合 Schema
- **WHEN** Arguments 不符合 Tool Input Schema
- **THEN** Tool Gateway 返回参数错误且不调用 Adapter

### Requirement: Tool 调用幂等
Tool Gateway SHALL 以 `task_id + step_id + attempt` 作为幂等键，并对重复请求返回首次持久化结果。

#### Scenario: 重复 InvokeRequest
- **WHEN** 相同幂等键被再次提交
- **THEN** Tool Gateway 不访问外部系统且返回已有 InvokeResult
