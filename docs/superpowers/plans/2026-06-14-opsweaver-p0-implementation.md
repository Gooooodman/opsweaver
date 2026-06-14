# OpsWeaver P0 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 按四个可独立验收的 OpenSpec Change 交付 OpsWeaver P0 平台底座和 Pod 重启诊断闭环。

**Architecture:** `opsweaver-server` 负责控制面与 Registry，`opsweaver-worker` 通过 Asynq 执行异步 Workflow，`opsweaver-gateway` 是唯一 Tool 执行入口。开发环境使用 Docker Compose 启动一个 PostgreSQL 实例中的两个逻辑数据库和一个 Redis 实例。

**Tech Stack:** Go、Gin、GORM、PostgreSQL、Redis、Asynq、Docker Compose、client-go、Prometheus HTTP API、Streamable HTTP MCP、OpenAI-compatible API。

---

## 文件结构

实施后主要目录：

```text
cmd/
  opsweaver-server/main.go
  opsweaver-worker/main.go
  opsweaver-gateway/main.go
  migrate/main.go
internal/
  platform/       # config、logging、database、redis、auth、crypto、metrics
  domain/         # Tool、MCP、Skill、Workflow、Task、Report 值对象
  registry/       # Tool、MCP、Skill 控制面
  toolgateway/    # Invoke、Policy、幂等、Evidence、Audit
  adapter/        # Kubernetes、Prometheus、Logs、MCP、LLM
  workflow/       # 受限声明式 Workflow Runtime
  orchestrator/   # 规则优先路由和参数抽取
  task/           # Task 生命周期和执行权
  report/         # RCA Report 校验与持久化
configs/skills/
migrations/opsweaver_server/
migrations/opsweaver_gateway/
deploy/
tests/e2e/
```

## 1. Change：bootstrap-platform-foundation

**OpenSpec：** `openspec/changes/bootstrap-platform-foundation/`

- [ ] 1.1 按 `tasks.md` 以 TDD 实现单 Go Module、三个服务入口和统一配置
- [ ] 1.2 实现 Docker Compose、双逻辑数据库、显式迁移和 Redis
- [ ] 1.3 实现 health、ready、metrics、英文错误响应和结构化日志
- [ ] 1.4 实现静态内部服务令牌与 AES-256-GCM 凭证加密
- [ ] 1.5 实现 Asynq Client/Server、DB 0 队列与 DB 1 缓存隔离
- [ ] 1.6 运行 `go test ./...`、`go vet ./...` 和 Change 严格校验
- [ ] 1.7 仅在本 Change 验收通过后开始 Change 2

## 2. Change：build-tool-execution-plane

**OpenSpec：** `openspec/changes/build-tool-execution-plane/`

- [ ] 2.1 按 `tasks.md` 以 TDD 实现不可变 Tool Registry 与管理 API
- [ ] 2.2 实现 Tool Gateway 的 Schema、Scope、Risk、幂等和 Adapter Dispatch
- [ ] 2.3 实现 Tool Call、Evidence、Error Evidence、Audit 和统一脱敏
- [ ] 2.4 实现 Kubernetes、Prometheus 和 Mock Logs 四个内置 Tool
- [ ] 2.5 实现 ToolSpec 版本同步和跨服务 Contract 测试
- [ ] 2.6 运行 `go test ./...`、`go vet ./...` 和 Change 严格校验
- [ ] 2.7 仅在本 Change 验收通过后开始 Change 3

## 3. Change：add-declarative-capability-runtime

**OpenSpec：** `openspec/changes/add-declarative-capability-runtime/`

- [ ] 3.1 按 `tasks.md` 以 TDD 实现 MCP Registry 和加密凭证
- [ ] 3.2 封装 Streamable HTTP MCP Client 并实现同步、调用和健康检查
- [ ] 3.3 实现 Skill/Workflow 数据库权威版本、文件导入和快照 API
- [ ] 3.4 实现仅支持 tool、condition、llm、report 的顺序 Workflow Runtime
- [ ] 3.5 实现 OpenAI-compatible LLM、JSON Schema 校验和一次修复重试
- [ ] 3.6 运行 `go test ./...`、`go vet ./...` 和 Change 严格校验
- [ ] 3.7 仅在本 Change 验收通过后开始 Change 4

## 4. Change：deliver-pod-diagnosis-workflow

**OpenSpec：** `openspec/changes/deliver-pod-diagnosis-workflow/`

- [ ] 4.1 按 `tasks.md` 以 TDD 实现 Task 状态机、异步创建和查询 API
- [ ] 4.2 实现 Worker 数据库 CAS 执行权和内部状态回写 API
- [ ] 4.3 实现 Skill Trigger 规则优先、LLM fallback 和澄清流程
- [ ] 4.4 实现 `pod_restart_diagnosis` Skill 和四个固定 Tool Step
- [ ] 4.5 实现 Evidence 约束的 RCA Report 校验、持久化和查询
- [ ] 4.6 运行 Redis、双数据库、模拟外部服务参与的端到端测试
- [ ] 4.7 运行 `go test ./...`、`go vet ./...` 和 Change 严格校验

## 5. 完整 P0 验收

- [ ] 5.1 运行 `openspec validate --changes --strict --no-interactive`
- [ ] 5.2 从空数据卷运行 Compose 并执行两个数据库迁移
- [ ] 5.3 验证三个服务 health、ready 和 metrics
- [ ] 5.4 注册模拟 MCP 后同步并调用 Tool，确认无需修改核心代码
- [ ] 5.5 注册示例 Skill 后执行 Workflow，确认无需修改 Orchestrator 分支
- [ ] 5.6 提交 Pod 重启诊断请求，验证 Report 的关键原因引用有效 Evidence ID
- [ ] 5.7 确认非 `read_only` Tool 被拒绝且成功、失败、拒绝均有 Audit

## 执行约束

1. 每个 OpenSpec Change 独立实施、验证和提交，不并行修改共享核心文件。
2. 每项功能先写失败测试，再写最小实现，再运行针对性测试。
3. 实施细节以各 Change 的 `proposal.md`、`specs/`、`design.md`、`tasks.md` 为准。
4. 发现设计冲突时先更新 OpenSpec Artifact 并重新严格校验，不在代码中静默偏离。
5. 完成 Change 后再归档；未完成任务不得标记为已完成。
