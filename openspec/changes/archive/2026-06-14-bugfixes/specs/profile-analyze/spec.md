## MODIFIED Requirements

### Requirement: Flame command format consistency
flame 命令的 JSON 输出 SHALL 使用 `count` 字段名表示采样值，与 text 格式输出和 profile 层的 `FoldedStack.Count` 字段保持一致。当 `--stats` flag 与非 text 格式同时使用时，系统 SHALL 向 stderr 输出警告信息但不中断执行。

#### Scenario: JSON 输出字段名一致
- **WHEN** 用户以 `--format json` 运行 flame 命令
- **THEN** JSON 输出使用 `count` 字段而非 `value`

#### Scenario: --stats 与 json 格式
- **WHEN** 用户使用 `--stats --format json` 运行 flame 命令
- **THEN** 系统向 stderr 输出警告 "stats output is only supported in text format" 但继续正常输出 JSON

#### Scenario: --stats 与 text 格式
- **WHEN** 用户使用 `--stats --format text` 运行 flame 命令
- **THEN** 系统正常输出 stats 信息，无警告
