## Why

平台底座必须通过一个真实但范围受控的诊断场景证明 Registry、异步任务、Workflow、Tool Gateway、Evidence 和 Report 能完整协作。Pod 重启诊断是 P0 的首个端到端验收场景。

## What Changes

- 提供诊断 Task 创建、查询和状态生命周期 API。
- 实现规则优先、LLM fallback 的意图解析和 Skill 匹配。
- 缺少 namespace、pod 或 service 时返回 `clarification_required`。
- 使用 Asynq 异步执行任务，Worker 通过内部 API 获取执行权和回写状态。
- 注册 `pod_restart_diagnosis` Skill，调用四个 P0 内置 Tool。
- 生成包含摘要、影响、Evidence、最多三个原因、建议和下一步的 RCA Report。
- 提供完整 Contract、Integration 和 End-to-End 测试。

## Capabilities

### New Capabilities

- `diagnosis-task`: 诊断任务创建、入队、执行权、状态回写和查询行为。
- `intent-routing`: 规则优先、LLM fallback、参数抽取和澄清行为。
- `pod-restart-diagnosis`: Pod 重启 Skill 的工具步骤、失败策略和证据收集行为。
- `rca-report`: RCA Report 持久化、查询、Evidence 引用和证据不足表达行为。

### Modified Capabilities

无。

## Impact

- 扩展 `aiops-server` 的 Task、Report 和 Orchestrator API。
- 扩展 `worker` 的诊断任务处理器。
- 新增 Pod 重启 Skill、Workflow、Prompt 和示例配置。
- 增加 Redis、双数据库、模拟外部服务参与的端到端测试。
- 依赖前三个 Change 的全部基础能力。
