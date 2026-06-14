## Why

代码审查发现多个 bug 和一致性问题：3 个严重问题（潜在 panic、除零、接口不符）、3 个中等问题（6 个命令的 `--value-type` flag 被静默忽略、info 命令未用共享校验、flame `--stats` 在非 text 格式静默忽略）、1 个低优先级问题（diagnose `--top` 未校验）。这些问题影响工具的可靠性和 AI 消费者对输出的信任度。

## What Changes

- 修复 `funcName()` 三个指针全 nil 时的 panic
- 修复 diff markdown 格式中 `BaseSamples=0` 时的除零问题
- 修复 `TrendMarkdownFormatter` 未实现 `TrendFormatter` 接口方法
- 修复 6 个命令（diff, tree, traces, list, flame, trend）注册了 `--value-type` 但实际未生效的问题
- 修复 info 命令未使用共享 `ValidateFormat` 的问题
- 修复 flame 命令 `--stats` 在非 text 格式时静默忽略的问题（改为输出警告）
- 修复 diagnose 命令 `--top` 未校验负数的问题

## Capabilities

### New Capabilities

（无新能力）

### Modified Capabilities

- `shared-validation`: 修复 info 命令未使用共享校验函数，补充 diagnose --top 范围校验
- `profile-analyze`: 修复 flame --stats 静默忽略、flame JSON 字段名不一致
- `profile-diff`: 修复 diff markdown 除零风险
- `trend-command`: 修复 TrendMarkdownFormatter 接口不符、funcName 潜在 panic
- `profile-diagnose`: 无行为变更（已在 shared-validation 中覆盖）

## Impact

- `pkg/output/formatter.go`: funcName panic 修复、diff 除零修复、TrendMarkdownFormatter 方法重命名、flame JSON 字段名统一
- `pkg/commands/info.go`: 使用共享 ValidateFormat
- `pkg/commands/flame.go`: --stats 非.text 格式加警告
- `pkg/commands/diagnose.go`: --top 范围校验
- `pkg/commands/{diff,tree,traces,list,flame,trend}.go`: --value-type 要么实现要么移除
