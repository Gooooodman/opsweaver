## 1. Tool Registry 数据与契约

- [ ] 1.1 在 `internal/domain/tool/spec_test.go` 编写 ToolSpec、RiskLevel、Scope 和不可变版本校验测试，再实现 `internal/domain/tool/spec.go`
- [ ] 1.2 创建 `migrations/opsweaver_server/0002_tools.up.sql` 与 down migration，加入 `name/version` 唯一约束和默认版本约束
- [ ] 1.3 在 `internal/registry/tool/repository_test.go` 编写注册、冲突、启停和默认版本测试，再实现 GORM Repository
- [ ] 1.4 在 `internal/registry/tool/service_test.go` 编写业务校验测试，再实现注册、列表、详情、启停服务
- [ ] 1.5 创建 `internal/api/tool/handler.go`、路由和 OpenAPI 定义，运行 Handler 测试验证英文错误响应
- [ ] 1.6 提交 Tool Registry，提交信息 `feat: add versioned tool registry`

## 2. Tool Gateway 核心调用链

- [ ] 2.1 创建 `internal/toolgateway/contract.go`，定义 InvokeRequest、InvokeContext、InvokeResult、Adapter 和 ToolSpecProvider 接口及 JSON Contract 测试
- [ ] 2.2 在 `internal/toolgateway/validator_test.go` 编写 Tool 启用、JSON Schema、Scope 和只读风险测试，再实现 Validator
- [ ] 2.3 创建 `migrations/opsweaver_gateway/0001_tool_execution.up.sql`，定义 tool_calls、evidences、audit_logs 和幂等唯一索引
- [ ] 2.4 在 `internal/toolgateway/store/store_test.go` 编写 running 创建、成功提交、失败提交和重复幂等测试，再实现事务 Store
- [ ] 2.5 在 `internal/toolgateway/service_test.go` 使用 Fake Adapter 编写成功、策略拒绝、参数失败、超时和重复调用测试，再实现调用编排
- [ ] 2.6 创建 `internal/api/internal/toolinvoke/handler.go`，挂载服务令牌中间件并补充 HTTP Contract 测试
- [ ] 2.7 提交核心调用链，提交信息 `feat: add tool gateway invocation pipeline`

## 3. Evidence、Audit 与脱敏

- [ ] 3.1 在 `internal/toolgateway/evidence/builder_test.go` 编写成功 Evidence、Error Evidence、1 MiB 上限测试，再实现 Builder
- [ ] 3.2 在 `internal/toolgateway/audit/builder_test.go` 编写成功、失败、拒绝记录测试，再实现 Audit Builder
- [ ] 3.3 将递归脱敏接入 Arguments、Result 和 Error 持久化路径，添加凭证不落库的集成测试
- [ ] 3.4 验证 Tool Call 终态、Evidence、Audit 在同一数据库事务中提交，失败事务不留半成品
- [ ] 3.5 提交证据审计链，提交信息 `feat: persist tool evidence and audit trail`

## 4. 内置 Tool Adapter

- [ ] 4.1 在 `internal/adapter/kubernetes/client_test.go` 编写 kubeconfig/in-cluster 配置选择测试，再实现单集群 Client 工厂
- [ ] 4.2 使用 client-go fake 在 `internal/adapter/kubernetes/describe_pod_test.go` 编写 Pod 状态、容器状态、OOMKilled 测试，再实现 `k8s.describe_pod`
- [ ] 4.3 使用 client-go fake 在 `internal/adapter/kubernetes/get_events_test.go` 编写相关事件筛选和排序测试，再实现 `k8s.get_events`
- [ ] 4.4 使用 `httptest.Server` 在 `internal/adapter/prometheus/query_test.go` 编写成功、上游错误、超时测试，再实现 `prometheus.query`
- [ ] 4.5 定义 `docs/contracts/mock-logs-api.yaml`，使用 `httptest.Server` 在 `internal/adapter/logs/search_test.go` 编写过滤、排序、限制测试，再实现 `logs.search`
- [ ] 4.6 注册四个 Adapter 与内置 ToolSpec，运行统一 Adapter Contract 测试
- [ ] 4.7 提交内置 Tool，提交信息 `feat: add builtin read-only tools`

## 5. ToolSpec 同步与集成

- [ ] 5.1 创建 `opsweaver-server` 的版本化内部 ToolSpec 导出接口与 `opsweaver-gateway` 同步 Client，补充校验和与缺失版本测试
- [ ] 5.2 在 Tool Gateway 启动时加载 ToolSpec 缓存，Invoke 指定未知版本时拒绝而不回退
- [ ] 5.3 使用双数据库和模拟外部服务运行 Registry-to-Gateway 集成测试
- [ ] 5.4 运行 `go test ./...`、`go vet ./...` 和 `openspec validate build-tool-execution-plane --strict --no-interactive`
