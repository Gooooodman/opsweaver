## 1. MCP Registry 与凭证

- [ ] 1.1 在 `internal/domain/mcp/server_test.go` 编写 Streamable HTTP、拒绝 stdio、Allowed Tools、Scope、Risk 校验测试，再实现 MCPServerSpec
- [ ] 1.2 创建 `migrations/opsweaver_server/0003_mcp_servers.up.sql` 与 down migration，认证字段仅保存版本化密文
- [ ] 1.3 在 `internal/registry/mcp/service_test.go` 编写注册、版本冲突、启停和凭证脱敏响应测试，再实现 Repository/Service
- [ ] 1.4 创建 MCP Registry Handler、路由和 OpenAPI 定义，注册成功后提交 `mcp.sync-tools` Asynq 任务
- [ ] 1.5 提交 MCP Registry，提交信息 `feat: add mcp server registry`

## 2. MCP Streamable HTTP Adapter

- [ ] 2.1 定义 `internal/adapter/mcp/client.go` 的 `ListTools`、`CallTool`、`Ping` 平台接口，并创建模拟 MCP Server Contract 测试
- [ ] 2.2 通过兼容性测试选择并封装 Go MCP SDK，实现 Streamable HTTP、Bearer 认证、超时和错误映射
- [ ] 2.3 在 `internal/adapter/mcp/mapper_test.go` 编写 `<server>.<tool>` 命名、Allowed Tools 过滤、Scope/Risk 收窄测试，再实现 Mapper
- [ ] 2.4 在 Tool Gateway 增加 MCP list/call/health 内部接口，确保 call 复用现有 Tool Invoke、Evidence 和 Audit 链
- [ ] 2.5 在 Worker 实现 `mcp.sync-tools` Handler，通过 Gateway 获取列表并回写 opsweaver-server 内部同步接口
- [ ] 2.6 使用模拟 MCP Server 运行注册、同步、调用和健康缓存集成测试
- [ ] 2.7 提交 MCP Adapter，提交信息 `feat: integrate streamable http mcp`

## 3. Skill Registry 与版本快照

- [ ] 3.1 定义 `internal/domain/skill/skill.go`、`workflow.go` 的判别联合类型，在测试中覆盖未知 Step、重复 ID、非法引用和大小限制
- [ ] 3.2 创建 `migrations/opsweaver_server/0004_skills_workflows.up.sql` 与 down migration，保存不可变 Skill、Workflow、Prompt 内容
- [ ] 3.3 在 `internal/registry/skill/validator_test.go` 编写文件、Schema、Tool 版本、Risk Policy、Prompt 校验测试，再实现 Validator
- [ ] 3.4 实现 API 注册和 `configs/skills/` 启动扫描共用的导入 Service，测试 `name/version` 幂等且不自动删除数据库版本
- [ ] 3.5 创建 Skill 列表、详情、运行、启停和内部版本快照 API，补充 HTTP Contract 测试
- [ ] 3.6 提交 Skill Registry，提交信息 `feat: add declarative skill registry`

## 4. Workflow Runtime

- [ ] 4.1 在 `internal/workflow/condition_test.go` 编写固定比较操作符和受限 JSON Path 测试，再实现 Condition Evaluator
- [ ] 4.2 在 `internal/workflow/runtime_test.go` 使用 Fake Tool Invoker 编写顺序执行、关键失败中断、非关键失败继续和幂等键测试
- [ ] 4.3 实现 `tool`、`condition`、`llm`、`report` Step Runner，禁止循环、并行和任意代码
- [ ] 4.4 将 Runtime 装配到 Worker，并通过 opsweaver-server 内部 API 获取固定版本快照和回写步骤状态
- [ ] 4.5 提交 Workflow Runtime，提交信息 `feat: add constrained workflow runtime`

## 5. OpenAI-compatible LLM 与报告草稿

- [ ] 5.1 定义 `internal/adapter/llm/client.go` 的结构化 Generate 接口，在测试中覆盖 Base URL、Model、超时和加密 API Key
- [ ] 5.2 使用 `httptest.Server` 实现 OpenAI-compatible Client，支持原生 Schema 路径和 JSON fallback 路径
- [ ] 5.3 在 `internal/adapter/llm/schema_test.go` 编写非法 JSON、Schema 失败、一次修复成功和二次失败测试，再实现校验与修复
- [ ] 5.4 在 `internal/report/validator_test.go` 编写未知 Evidence ID、跨任务 Evidence 和缺失引用测试，再实现 Report Draft Validator
- [ ] 5.5 运行 MCP + Skill + Workflow + LLM 集成测试并提交，提交信息 `feat: add evidence-grounded llm reporting`

## 6. Change 验证

- [ ] 6.1 运行 `go test ./...` 和 `go vet ./...`
- [ ] 6.2 运行模拟 MCP 与 OpenAI-compatible Provider 的 Contract 测试
- [ ] 6.3 运行 `openspec validate add-declarative-capability-runtime --strict --no-interactive`
