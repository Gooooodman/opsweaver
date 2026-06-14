## ADDED Requirements

### Requirement: MCP 注册仅支持 Streamable HTTP
MCP Registry SHALL 接受 Streamable HTTP Server 配置，并 MUST 拒绝 stdio、旧版 HTTP+SSE 和自定义 Transport。

#### Scenario: 注册支持的 MCP
- **WHEN** 管理员提交合法 Streamable HTTP MCP 配置
- **THEN** Registry 创建不可变版本记录并加密保存认证信息

#### Scenario: 注册不支持的 Transport
- **WHEN** 配置声明 stdio 或其他非支持 Transport
- **THEN** Registry 返回英文校验错误

### Requirement: MCP Tool 同步受本地约束
系统 SHALL 异步调用 `tools/list`，将远端 Tool 命名为 `<server>.<tool>`，并以本地 Allowed Tools、Scope 和 Risk 收窄远端声明。

#### Scenario: 同步允许的 Tool
- **WHEN** `tools/list` 返回 Allowed Tools 中的合法定义
- **THEN** Tool Registry 创建对应 MCP ToolSpec 版本

#### Scenario: 远端声明更高权限
- **WHEN** 远端 Tool 声明超出本地 Scope 或 Risk
- **THEN** 同步结果使用更严格的本地限制

### Requirement: MCP 调用经过 Tool Gateway
所有 `tools/call` MUST 由 Tool Gateway 执行并产生 Tool Call、Evidence 和 Audit。

#### Scenario: MCP Tool 调用成功
- **WHEN** Worker 调用已同步且启用的 MCP Tool
- **THEN** Tool Gateway 使用解密凭证调用 MCP Server 并返回标准 InvokeResult
