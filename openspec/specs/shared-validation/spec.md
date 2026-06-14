## ADDED Requirements

### Requirement: ValidateFormat 公共验证函数
系统 SHALL 提供 `ValidateFormat(format string) error` 函数，校验输出格式是否为 `text`、`json` 或 `markdown` 之一。非法格式 MUST 返回错误并列出合法值。

#### Scenario: 合法格式
- **WHEN** 调用 `ValidateFormat("json")`
- **THEN** 返回 nil

#### Scenario: 非法格式
- **WHEN** 调用 `ValidateFormat("yaml")`
- **THEN** 返回 error，消息包含 "invalid format" 和合法值列表

### Requirement: ValidateRegex 公共验证函数
系统 SHALL 提供 `ValidateRegex(pattern, name string) error` 函数，校验正则表达式是否可编译。空字符串 MUST 视为合法（不过滤）。非法正则 MUST 返回错误并包含 `name` 作为上下文说明。

#### Scenario: 合法正则
- **WHEN** 调用 `ValidateRegex("runtime.*", "focus")`
- **THEN** 返回 nil

#### Scenario: 空字符串
- **WHEN** 调用 `ValidateRegex("", "focus")`
- **THEN** 返回 nil

#### Scenario: 非法正则
- **WHEN** 调用 `ValidateRegex("[invalid", "focus")`
- **THEN** 返回 error，消息包含 "invalid focus pattern"

### Requirement: ValidateFormat 和 ValidateRegex 有单元测试
系统 SHALL 为 ValidateFormat 和 ValidateRegex 提供单元测试，覆盖合法值、非法值和空值场景。

#### Scenario: 测试覆盖
- **WHEN** 运行 `go test ./pkg/commands/ -run TestValidate`
- **THEN** 所有测试通过
