## ADDED Requirements

### Requirement: Diagnose --top 参数范围校验
diagnose 命令 SHALL 在 --top 参数值小于等于 0 时返回错误，提示参数必须为正数。

#### Scenario: --top 为负数
- **WHEN** 用户运行 `agent-insight diagnose profile.pb.gz --top -5`
- **THEN** 系统返回错误，消息包含 "top" 和 "must be positive"

#### Scenario: --top 为零
- **WHEN** 用户运行 `agent-insight diagnose profile.pb.gz --top 0`
- **THEN** 系统返回错误，消息包含 "top" 和 "must be positive"

#### Scenario: --top 为正数
- **WHEN** 用户运行 `agent-insight diagnose profile.pb.gz --top 10`
- **THEN** 系统正常执行
