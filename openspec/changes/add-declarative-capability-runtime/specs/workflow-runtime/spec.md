## ADDED Requirements

### Requirement: P0 Workflow 支持受限步骤
Workflow Runtime SHALL 仅支持顺序执行的 `tool`、`condition`、`llm` 和 `report` Step，不得执行脚本、二进制或任意代码。

#### Scenario: 执行合法 Workflow
- **WHEN** Worker 收到通过校验的版本快照
- **THEN** Runtime 按声明顺序执行步骤并记录每步状态

#### Scenario: 出现不支持的 Step
- **WHEN** Workflow 包含循环、并行、脚本或未知类型
- **THEN** 注册或执行前校验失败

### Requirement: Tool Step 通过 Gateway 执行
每个 Tool Step MUST 通过 Tool Gateway，并使用 `task_id + step_id + attempt` 幂等键。

#### Scenario: 重试相同步骤
- **WHEN** Worker 因临时错误重试相同 Tool Step
- **THEN** Gateway 幂等结果防止重复外部调用

### Requirement: Workflow 区分关键步骤
Tool Step SHALL 支持 `critical` 声明且默认值为 true。

#### Scenario: 非关键步骤失败
- **WHEN** `critical: false` 的步骤失败
- **THEN** Runtime 保存失败上下文并继续后续步骤

#### Scenario: 关键步骤失败
- **WHEN** 关键步骤失败
- **THEN** Runtime 终止执行并返回不可完成状态
