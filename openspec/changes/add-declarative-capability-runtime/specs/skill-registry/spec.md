## ADDED Requirements

### Requirement: Skill 使用数据库权威版本
Skill Registry SHALL 按 `name + version` 保存不可变 Skill、Workflow 和 Prompt 定义，运行时不得直接读取配置文件作为权威来源。

#### Scenario: 导入 Skill 文件
- **WHEN** 启动扫描发现合法的 `skill.yaml`、`workflow.yaml` 和 `prompt.md`
- **THEN** 系统通过与 API 注册相同的校验服务幂等导入数据库

### Requirement: Skill 注册执行完整校验
Registry MUST 校验文件引用、Input/Output Schema、Tool 版本、Risk Policy、Step ID 和 Prompt。

#### Scenario: 引用不存在的 Tool
- **WHEN** Skill requiredTools 引用了不存在或禁用的 Tool 版本
- **THEN** 注册失败并返回英文校验错误

### Requirement: 已创建任务绑定 Skill 快照
系统 SHALL 为任务保存明确的 Skill 版本，并向 Worker 返回该版本的完整快照。

#### Scenario: 注册新版本后执行旧任务
- **WHEN** 任务绑定 v1 后 Registry 发布 v2
- **THEN** Worker 仍使用 v1 的 Workflow、Prompt 和 Tool 引用
