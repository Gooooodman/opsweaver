## ADDED Requirements

### Requirement: 意图解析规则优先
Orchestrator SHALL 优先使用启用 Skill 的 Trigger 规则匹配输入，仅在规则无法确定时使用 LLM fallback。

#### Scenario: 命中 Pod 重启规则
- **WHEN** 输入包含 Pod 重启、CrashLoopBackOff 或 OOMKilled 等触发词
- **THEN** Orchestrator 匹配 `pod_restart_diagnosis` 且不调用 LLM fallback

### Requirement: 参数抽取和澄清
Orchestrator SHALL 抽取 env、cluster、namespace、service、pod 和 time range；缺少可定位目标的参数时 MUST 返回澄清问题。

#### Scenario: 缺少定位参数
- **WHEN** 输入没有 namespace、service 或 pod 等可定位信息
- **THEN** 任务进入 `clarification_required` 并返回中文澄清问题

### Requirement: LLM fallback 受 Schema 约束
LLM fallback MUST 返回符合 Intent Schema 的结果，且不得直接触发 Tool。

#### Scenario: 规则无法匹配
- **WHEN** 规则无法确定 Skill
- **THEN** LLM 仅返回 intent、entities、confidence 和 clarification 状态
