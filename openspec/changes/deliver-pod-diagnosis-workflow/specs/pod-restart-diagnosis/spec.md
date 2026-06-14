## ADDED Requirements

### Requirement: Pod 重启 Skill 调用固定 Tool
`pod_restart_diagnosis` SHALL 调用 `k8s.describe_pod`、`k8s.get_events`、`logs.search` 和 `prometheus.query` 收集证据。

#### Scenario: 完整执行
- **WHEN** namespace 和 pod 参数完整且四个 Tool 可用
- **THEN** Workflow 执行四个 Tool Step 并收集各自 Evidence ID

### Requirement: 诊断保持只读
Skill 的 Risk Policy MUST 为 `read_only`，不得生成自动执行的变更步骤。

#### Scenario: Workflow 引用非只读 Tool
- **WHEN** Skill 新版本引用非只读 Tool
- **THEN** 注册校验失败

### Requirement: 部分证据失败可继续
日志和指标步骤 SHALL 为非关键步骤，Pod 描述步骤 SHALL 为关键步骤。

#### Scenario: 日志查询失败
- **WHEN** `logs.search` 返回临时或永久错误
- **THEN** Workflow 保留 Error Evidence 并继续生成证据不足的报告

#### Scenario: Pod 描述失败
- **WHEN** `k8s.describe_pod` 无法定位目标
- **THEN** Workflow 终止并将任务标记为失败
