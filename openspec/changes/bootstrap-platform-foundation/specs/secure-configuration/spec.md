## ADDED Requirements

### Requirement: 内部接口使用服务令牌
所有内部 HTTP API MUST 校验约定 Header 中的静态服务令牌，并使用常量时间比较。

#### Scenario: 有效服务令牌
- **WHEN** 内部请求携带正确令牌
- **THEN** 请求进入业务处理链

#### Scenario: 缺失或错误令牌
- **WHEN** 内部请求未携带令牌或令牌不匹配
- **THEN** 系统返回 HTTP 401 和英文错误响应

### Requirement: 敏感凭证加密存储
MCP 凭证和 LLM API Key MUST 使用 AES-256-GCM 加密，每条密文 MUST 使用独立随机 Nonce。

#### Scenario: 加密并解密凭证
- **WHEN** 服务使用有效 32 字节主密钥保存并读取凭证
- **THEN** 数据库仅保存密文且读取结果等于原始明文

#### Scenario: 主密钥非法
- **WHEN** 主密钥缺失或解码后不是 32 字节
- **THEN** 需要加解密能力的服务启动失败并输出英文错误

### Requirement: 敏感数据不得泄露
日志、API 响应和错误信息 MUST NOT 包含主密钥、API Key、Bearer Token 或解密后的 MCP 凭证。

#### Scenario: 外部依赖认证失败
- **WHEN** 使用敏感凭证调用外部服务失败
- **THEN** 错误记录仅包含脱敏信息
