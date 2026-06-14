## ADDED Requirements

### Requirement: Docker Compose 提供本地依赖
项目 SHALL 提供 Docker Compose 配置启动一个 PostgreSQL 容器和一个 Redis 容器，并为两者提供健康检查。

#### Scenario: 启动本地基础设施
- **WHEN** 开发者执行项目约定的 Compose 启动命令
- **THEN** PostgreSQL 和 Redis 容器进入健康状态

### Requirement: PostgreSQL 初始化两个逻辑数据库
PostgreSQL 容器 MUST 初始化 `opsweaver_server_db` 和 `opsweaver_gateway_db`，两个服务 SHALL 使用不同连接配置。

#### Scenario: 从空数据卷初始化
- **WHEN** PostgreSQL 使用空数据卷首次启动
- **THEN** 两个数据库均被创建且各自迁移可独立执行

### Requirement: 数据库使用显式迁移
系统 MUST 使用版本化 SQL 管理数据库结构，不得以 GORM AutoMigrate 作为部署迁移机制。

#### Scenario: 应用空库迁移
- **WHEN** 对空的 `opsweaver_server_db` 或 `opsweaver_gateway_db` 执行迁移命令
- **THEN** 所有迁移按版本顺序成功应用并记录版本
