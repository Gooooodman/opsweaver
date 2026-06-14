## ADDED Requirements

### Requirement: 三个服务可独立启动
系统 SHALL 提供 `aiops-server`、`worker`、`tool-gateway` 三个独立入口，并从统一配置结构加载各自所需配置。

#### Scenario: 使用有效配置启动
- **WHEN** 操作者使用完整有效的配置启动任一服务
- **THEN** 服务监听配置端口并输出英文结构化启动日志

#### Scenario: 使用无效配置启动
- **WHEN** 必填配置缺失或格式非法
- **THEN** 服务启动失败并输出不包含敏感值的英文错误信息

### Requirement: 服务暴露运行状态
HTTP 服务 SHALL 暴露 `/healthz`、`/readyz` 和 `/metrics`；Worker SHALL 暴露独立健康检查 HTTP 端口。

#### Scenario: 依赖可用
- **WHEN** 服务及其必需依赖均可用
- **THEN** `/healthz` 和 `/readyz` 返回成功，`/metrics` 返回 Prometheus 文本格式

#### Scenario: 必需依赖不可用
- **WHEN** 数据库或 Redis 等必需依赖不可用
- **THEN** `/healthz` 仍表示进程存活而 `/readyz` 返回非成功状态
