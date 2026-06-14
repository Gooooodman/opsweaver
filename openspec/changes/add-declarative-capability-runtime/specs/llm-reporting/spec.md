## ADDED Requirements

### Requirement: LLM 使用 OpenAI-compatible API
系统 SHALL 通过可配置的 Base URL 和 Model 调用 OpenAI-compatible API，并从加密凭证存储读取 API Key。

#### Scenario: 调用兼容 API
- **WHEN** Workflow 执行 LLM Step
- **THEN** Client 发送带超时的请求并返回结构化响应

### Requirement: LLM 输出必须通过 Schema
LLM Step 和 Report Step 的输出 MUST 通过声明的 JSON Schema；首次失败时系统最多执行一次修复重试。

#### Scenario: 首次输出非法
- **WHEN** 首次 LLM 输出无法解析或不符合 Schema
- **THEN** 系统携带校验错误执行一次修复请求

#### Scenario: 修复仍失败
- **WHEN** 第二次输出仍不符合 Schema
- **THEN** Step 失败且不再重试

### Requirement: 报告结论绑定 Evidence
报告生成输入 SHALL 仅包含当前任务的 Evidence 摘要和 ID，关键原因 MUST 引用至少一个有效 Evidence ID。

#### Scenario: LLM 返回未知 Evidence ID
- **WHEN** 报告引用不属于当前任务的 Evidence ID
- **THEN** 报告校验失败
