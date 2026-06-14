## Context

本 Change 在平台基础上建立 Tool 的控制面与执行面。Tool 定义属于 `opsweaver-server`，调用记录、Evidence 与 Audit 属于 `opsweaver-gateway`，两者不得跨库访问。

## Goals / Non-Goals

**Goals:**

- 数据库权威的不可变 Tool Registry。
- Tool Gateway 唯一调用入口及 P0 只读策略。
- 成功和失败均可审计、可追溯。
- 四个内置 Tool 通过统一 Adapter 接口运行。

**Non-Goals:**

- 不实现 MCP、Skill、LLM 或最终 RCA Report。
- 不实现多集群、Loki 或写操作 Tool。
- 不向普通用户开放 Tool Invoke。

## Decisions

### 决策 1：ToolSpec 同步副本

`opsweaver_server_db` 保存权威 ToolSpec；`opsweaver-gateway` 通过版本化内部同步接口维护只读本地缓存。相比每次远程查询，调用路径更稳定；同步记录携带版本和校验和，避免静默漂移。

### 决策 2：统一 Adapter 接口

Adapter 接收已经校验的 `InvokeRequest`，返回原始数据与 Evidence 类型提示。Policy、脱敏、持久化不进入 Adapter，避免外部系统实现污染平台规则。

### 决策 3：幂等与持久化

`tool_calls` 对幂等键建立唯一约束。首次调用创建 running 记录，完成后在单个 `opsweaver_gateway_db` 事务内写 Tool Call 终态、Evidence 和 Audit。重复请求读取已完成结果；处理中请求返回可重试冲突。

### 决策 4：风险策略固定

P0 Policy Checker 使用固定映射：仅 `read_only` allow，其他 deny。保留接口但不引入 OPA 或 Approval 状态机。

### 决策 5：内置客户端可替换

Kubernetes Adapter 依赖最小 Client 接口并在测试使用 fake client；Prometheus 和 Mock Logs 依赖标准 HTTP Client 与 `httptest.Server`。

## Risks / Trade-offs

- [Risk] ToolSpec 同步延迟导致 Gateway 使用旧版本 → InvokeRequest 指定版本，Gateway 缺失版本时拒绝而非回退。
- [Risk] 大 Evidence 增加 PostgreSQL 压力 → P0 直接存 JSONB，但设置请求和结果大小上限。
- [Risk] 外部超时占用连接 → 每个 ToolSpec 有超时上限，Client 使用 Context 取消。
- [Risk] 脱敏规则遗漏 → 提供递归键名和 Header 脱敏测试，原始凭证禁止进入结构化结果。

## Migration Plan

1. 应用 Tool Registry 与 Tool Gateway 数据库迁移。
2. 注册并同步四个内置 ToolSpec。
3. 启动 Tool Gateway，使用内部 Contract 测试验证 Invoke。
4. 回滚时先禁用 Tool，再回滚服务；保留审计数据。

## Open Questions

无阻塞问题。P0 Evidence 单条序列化上限在实施任务中固定为 1 MiB。
