## Why

`analyze --value-type` flag 对多值 profile（heap 类）不生效。无论用户指定 `alloc_objects`、`alloc_space`、`inuse_objects`、`inuse_space` 中哪一个，工具都返回相同的默认 `inuse_space` 结果，导致用户无法区分内存分配热点与存活对象热点。

## What Changes

- 在 `pkg/commands/analyze.go` 中将 `--value-type` 字符串参数解析为 `ValueTypeConfig`，并赋值给 `AnalysisConfig.ValueType`，使该 flag 真正透传到 `pkg/profile/analysis.go` 的分析逻辑。
- 在 `pkg/profile/analysis.go` 中校验用户指定的 value-type 是否存在于 profile 的 `ValueTypes` 列表中，无效值返回明确错误。
- 在 `pkg/commands/analyze.go` 添加单元测试，验证 flag 透传后输出与默认不同。

## Capabilities

### New Capabilities

无。

### Modified Capabilities

- `profile-analyze`：补充 `--value-type` flag 的 REQUIREMENT —— 必须按用户指定值类型输出，而非总是 fallback 到默认值。

## Impact

- **代码**：`pkg/commands/analyze.go`（修复 flag 处理）、`pkg/profile/analysis.go`（新增 value-type 校验）
- **API/输出**：行为变更。对同一 heap profile 用不同 `--value-type` 调用的输出将出现预期差异（这是修复而非回归）
- **测试**：新增单元测试验证 flag 透传