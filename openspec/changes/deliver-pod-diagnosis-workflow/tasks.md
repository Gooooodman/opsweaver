## 1. Task 数据模型与 API

- [ ] 1.1 定义 `internal/domain/task/task.go` 状态机，在测试中覆盖合法转换和非法回退
- [ ] 1.2 创建 `migrations/opsweaver_server/0005_tasks_reports.up.sql` 与 down migration，增加 Task、Report、步骤状态和必要唯一约束
- [ ] 1.3 在 `internal/task/service_test.go` 编写先建 pending、入队成功 queued、入队失败 failed 测试，再实现创建服务
- [ ] 1.4 创建 Task 创建、详情、Report 查询 Handler 和 OpenAPI 定义，补充异步 202 响应 Contract 测试
- [ ] 1.5 提交 Task API，提交信息 `feat: add asynchronous diagnosis tasks`

## 2. Worker 执行权与状态回写

- [ ] 2.1 在 `internal/task/execution_test.go` 编写数据库 CAS 仅一个 Worker 获权、终态拒绝和重复投递测试，再实现内部执行权 API
- [ ] 2.2 创建步骤开始、步骤完成、Report 回写、终态更新内部 API，全部挂载服务令牌认证
- [ ] 2.3 在 Worker 实现诊断 Task Handler，Asynq Payload 仅解析 `task_id`、`attempt`、`trace_id`
- [ ] 2.4 编写 Worker 临时错误重试、永久错误不重试和崩溃后重复消费集成测试
- [ ] 2.5 提交 Worker 执行协调，提交信息 `feat: coordinate diagnosis worker execution`

## 3. Orchestrator 与澄清

- [ ] 3.1 在 `internal/orchestrator/matcher_test.go` 编写从启用 Skill Trigger 匹配 Pod 重启、CrashLoopBackOff、OOMKilled 测试，再实现规则索引
- [ ] 3.2 在 `internal/orchestrator/entities_test.go` 编写 env、cluster、namespace、service、pod、time range 抽取测试，再实现参数解析
- [ ] 3.3 编写缺少目标信息进入 `clarification_required` 和中文澄清问题测试
- [ ] 3.4 实现仅在规则不确定时调用 LLM fallback，并测试 fallback 只返回 Intent Schema、不调用 Tool
- [ ] 3.5 提交 Orchestrator，提交信息 `feat: add rule-first intent routing`

## 4. Pod 重启诊断 Skill

- [ ] 4.1 创建 `configs/skills/pod_restart_diagnosis/{skill.yaml,workflow.yaml,prompt.md,examples.yaml}`
- [ ] 4.2 编写 Skill 导入测试，验证固定 Tool 版本、`read_only` Risk Policy、describe 为关键、events/logs/metrics 为非关键
- [ ] 4.3 使用 Fake Tool Gateway 编写完整四步成功、日志失败继续、Pod 描述失败终止测试
- [ ] 4.4 使用模拟 Kubernetes、Prometheus、Mock Logs 运行 Worker Workflow 集成测试
- [ ] 4.5 提交 Pod Skill，提交信息 `feat: add pod restart diagnosis skill`

## 5. RCA Report

- [ ] 5.1 在 `internal/report/rca_test.go` 编写完整字段、最多三个原因、只读建议、Evidence 所属校验和证据不足测试
- [ ] 5.2 实现 Report Builder/Validator 和 GORM Repository，未知 Evidence ID 或非只读动作必须拒绝
- [ ] 5.3 实现按 Task ID 查询 Report，并验证 Evidence 摘要与 ID 可回溯到 Tool Gateway
- [ ] 5.4 提交 RCA Report，提交信息 `feat: add evidence-backed rca reports`

## 6. 端到端验收

- [ ] 6.1 创建 `tests/e2e/pod_diagnosis_test.go`，通过 Docker Compose 启动 Redis、PostgreSQL 和模拟外部服务
- [ ] 6.2 提交 Pod 重启自然语言请求，轮询任务直到终态，验证 Skill 匹配和四类 Tool Call
- [ ] 6.3 验证 Report 包含摘要、影响、Evidence、最多三个原因、建议、下一步和有效 Evidence ID
- [ ] 6.4 编写日志/指标不可用的 E2E，验证任务成功但报告明确证据不足
- [ ] 6.5 编写重复 Asynq 投递 E2E，验证外部 Tool 不重复调用
- [ ] 6.6 运行 `go test ./...`、`go vet ./...` 和 `openspec validate deliver-pod-diagnosis-workflow --strict --no-interactive`
- [ ] 6.7 更新 README 的 P0 端到端演示命令并提交，提交信息 `test: verify pod diagnosis workflow end to end`
