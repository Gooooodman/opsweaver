## ADDED Requirements

### Requirement: 创建诊断任务并异步入队
`aiops-server` SHALL 先创建 `pending` 任务，成功入队后更新为 `queued`，并立即向调用方返回 Task ID。

#### Scenario: 入队成功
- **WHEN** 用户提交合法诊断请求且 Asynq 可用
- **THEN** API 返回 Task ID，任务状态为 `queued`

#### Scenario: 入队失败
- **WHEN** Asynq 入队失败
- **THEN** 任务更新为 `failed` 并记录英文错误

### Requirement: Worker 原子获取执行权
Worker MUST 通过内部 API 原子获取任务执行权，重复投递不得导致并行重复执行。

#### Scenario: 两个 Worker 获取同一任务
- **WHEN** 两个 Worker 同时尝试获取同一 Task ID
- **THEN** 只有一个获得执行权并将状态更新为 `running`

### Requirement: 任务状态可查询
系统 SHALL 支持查询 `pending`、`queued`、`running`、`succeeded`、`failed` 和 `clarification_required` 状态及时间戳。

#### Scenario: 查询已完成任务
- **WHEN** 用户查询成功任务
- **THEN** API 返回终态、Report ID 和完成时间
