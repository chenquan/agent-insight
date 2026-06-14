## MODIFIED Requirements

### Requirement: ValidateFormat 公共验证函数
系统 SHALL 提供 `ValidateFormat(format string) error` 函数，校验输出格式是否为 `text`、`json` 或 `markdown` 之一。非法格式 MUST 返回错误并列出合法值。所有命令（包括 info）SHALL 使用此函数，不 SHALL 手写格式校验逻辑。

#### Scenario: 合法格式
- **WHEN** 调用 `ValidateFormat("json")`
- **THEN** 返回 nil

#### Scenario: 非法格式
- **WHEN** 调用 `ValidateFormat("yaml")`
- **THEN** 返回 error，消息包含 "invalid format" 和合法值列表

#### Scenario: info 命令使用共享校验
- **WHEN** info 命令接收到非法格式参数
- **THEN** 通过 ValidateFormat 函数返回错误，错误消息与其他命令一致

## ADDED Requirements

### Requirement: ValidatePositiveInt 参数校验函数
系统 SHALL 提供 `ValidatePositiveInt(value int, name string) error` 函数，校验整数值是否为正数。非正数 MUST 返回错误并包含参数名作为上下文说明。

#### Scenario: 合法正整数
- **WHEN** 调用 `ValidatePositiveInt(10, "top")`
- **THEN** 返回 nil

#### Scenario: 零值
- **WHEN** 调用 `ValidatePositiveInt(0, "top")`
- **THEN** 返回 error，消息包含 "top" 和 "must be positive"

#### Scenario: 负数
- **WHEN** 调用 `ValidatePositiveInt(-1, "top")`
- **THEN** 返回 error，消息包含 "top" 和 "must be positive"
