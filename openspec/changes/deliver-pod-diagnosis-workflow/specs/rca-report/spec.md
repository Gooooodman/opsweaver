## ADDED Requirements

### Requirement: RCA Report 结构完整
成功诊断 SHALL 生成包含 title、summary、impact、evidence、possible_causes、recommended_actions、next_steps 和 need_approval 的 Report。

#### Scenario: 生成完整报告
- **WHEN** Workflow 完成且存在有效 Evidence
- **THEN** Report 包含最多三个原因、置信度、只读建议和 Evidence 引用

### Requirement: 报告引用可验证 Evidence
每个关键原因 MUST 引用属于当前任务且可查询的 Evidence ID。

#### Scenario: 查询报告
- **WHEN** 用户通过 Task ID 查询 Report
- **THEN** API 返回 Report 及 Evidence 摘要和 ID

### Requirement: 证据不足必须显式表达
Workflow 完成但关键证据缺失时，Report MUST 标记不确定性，不得虚构根因。

#### Scenario: 非关键 Tool 失败
- **WHEN** 日志或指标 Evidence 缺失
- **THEN** Report 在 summary 或 next_steps 中明确说明证据不足
