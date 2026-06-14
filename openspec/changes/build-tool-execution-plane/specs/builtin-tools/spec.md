## ADDED Requirements

### Requirement: Kubernetes 内置 Tool
系统 SHALL 提供 `k8s.describe_pod` 和 `k8s.get_events`，并通过配置选择 kubeconfig 或 in-cluster 单集群客户端。

#### Scenario: 查询存在的 Pod
- **WHEN** 调用 `k8s.describe_pod` 且 namespace 与 pod 有效
- **THEN** Tool 返回结构化 Pod 状态、容器状态和最近终止原因

#### Scenario: 查询相关事件
- **WHEN** 调用 `k8s.get_events`
- **THEN** Tool 返回按时间排序的相关 Kubernetes Event

### Requirement: Prometheus 内置 Tool
系统 SHALL 提供 `prometheus.query`，仅允许调用配置的 Prometheus HTTP API。

#### Scenario: 执行合法 PromQL
- **WHEN** 调用方提交合法查询和时间参数
- **THEN** Tool 返回标准化 Metric Series

### Requirement: Mock Logs 内置 Tool
系统 SHALL 提供 `logs.search` 并仅连接 Mock Logs API。

#### Scenario: 查询日志
- **WHEN** 调用方提交 service、namespace、time range 和 query
- **THEN** Tool 返回数量受限、按时间排序的结构化日志条目
