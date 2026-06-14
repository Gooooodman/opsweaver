## ADDED Requirements

### Requirement: 所有调用均有 Tool Call 与 Audit
Tool Gateway MUST 为成功、外部失败、校验失败和策略拒绝记录 Tool Call 与 Audit。

#### Scenario: Adapter 调用失败
- **WHEN** 外部系统返回错误或超时
- **THEN** Tool Call 和 Audit 记录失败状态、英文错误和耗时

### Requirement: 每个调用产生 Evidence
每个 Tool Call MUST 关联至少一条 Evidence 或 Error Evidence，并保存来源、类型、摘要、时间戳和 Tool Call ID。

#### Scenario: Tool 调用成功
- **WHEN** Adapter 返回合法结果
- **THEN** 系统保存脱敏后的 Evidence 并在 InvokeResult 返回 Evidence ID

#### Scenario: Tool 调用失败
- **WHEN** 调用失败且错误可形成诊断上下文
- **THEN** 系统保存 Error Evidence

### Requirement: 审计数据脱敏
Arguments、Tool Result 和 Error 在持久化前 MUST 经过统一脱敏器。

#### Scenario: 参数包含令牌
- **WHEN** Arguments 包含 token、password、authorization 或 secret 字段
- **THEN** Tool Call、Evidence 和 Audit 中对应值被替换为掩码
