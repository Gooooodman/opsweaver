## Context

前三个 Change 提供服务、队列、Tool、Evidence、MCP、Skill 和 Workflow。本 Change 用 Pod 重启诊断验证完整链路，并补齐面向用户的 Task 与 Report API。

## Goals / Non-Goals

**Goals:**

- 异步诊断任务具有明确状态和幂等执行权。
- Orchestrator 规则优先，必要时使用 LLM fallback。
- Pod 重启 Skill 收集四类证据并生成可验证 RCA Report。
- 端到端测试覆盖队列、双数据库和服务调用。

**Non-Goals:**

- 不提供 Web UI。
- 不提供 5xx、CI/CD 或其他诊断 Skill。
- 不执行修复动作。

## Decisions

### 决策 1：任务创建与入队非事务补偿

先创建 `pending` 任务，再入队并更新 `queued`。跨 PostgreSQL 与 Redis 不使用分布式事务；入队失败立即标记 `failed`。后续可增加 Outbox，但 P0 通过补偿满足可靠性。

### 决策 2：数据库 CAS 获取执行权

Worker 调用内部 API，通过条件更新 `queued -> running` 获取执行权。终态和已运行任务拒绝重复获取，避免 Asynq 至少一次投递造成重复执行。

### 决策 3：规则来自 Skill Trigger

Orchestrator 不硬编码 Skill 名称分支，而是索引启用 Skill 的 Trigger。规则匹配明确时不调用 LLM；LLM fallback 只做结构化路由和参数抽取。

### 决策 4：Pod Workflow 关键性

`describe_pod` 为关键步骤；events、logs、metrics 为非关键步骤。这样目标不存在时快速失败，外围证据缺失时仍能输出“不确定”的报告。

### 决策 5：Report 二次校验

Worker 回写 Report 前校验结构、Evidence 所属 Task、原因数量和风险等级。未知 Evidence ID、非只读动作或超过三个原因均拒绝持久化。

## Risks / Trade-offs

- [Risk] 创建任务后进程在入队前崩溃，任务停留 pending → 提供超时 pending 扫描命令作为运维恢复手段，但 P0 不启用 Scheduler。
- [Risk] 自然语言参数抽取不稳定 → 规则优先、明确 Schema、低置信度进入澄清。
- [Risk] Mock Logs 降低场景真实性 → E2E 重点验证平台链路，不宣称验证真实日志性能。
- [Risk] LLM 产生过度确定结论 → 强制 Evidence 引用和证据不足字段。

## Migration Plan

1. 应用 Task 与 Report 迁移。
2. 导入 `pod_restart_diagnosis` Skill 及固定版本。
3. 配置 Kubernetes、Prometheus、Mock Logs 与 LLM 测试 Endpoint。
4. 执行端到端测试后开放 Task API。
5. 回滚时禁用 Skill，停止接受新任务，等待运行中任务结束。

## Open Questions

无阻塞问题。P0 的 pending 恢复使用手动 CLI/管理命令，不引入 Asynq Scheduler。
