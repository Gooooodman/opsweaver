## ADDED Requirements

### Requirement: Tool 使用不可变版本
Tool Registry SHALL 按 `name + version` 保存不可变 ToolSpec，并确保同一 Tool 同一时刻最多一个默认启用版本。

#### Scenario: 注册新版本
- **WHEN** 使用已有名称和新版本注册合法 ToolSpec
- **THEN** Registry 创建新记录且不修改旧版本

#### Scenario: 重复注册同一版本
- **WHEN** 使用相同名称和版本重复注册
- **THEN** Registry 返回冲突错误且不覆盖已有定义

### Requirement: Tool 可查询和启停
Registry SHALL 提供列表、详情、启用和禁用行为，运行时只返回启用版本。

#### Scenario: 禁用默认版本
- **WHEN** 管理员禁用 Tool 的默认版本
- **THEN** 后续运行时解析不再返回该版本
