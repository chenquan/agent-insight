## Why

`analyze` 命令在 `gobench.heap` 和 `java.heap` 等 profile 上输出 `type: ""`（空字符串），影响结果可读性与下游 AI 解析。根因：`pkg/profile/analysis.go` 的 `extractMetadata` 仅从 `p.PeriodType.Type` 读取 profile 类型；当 `PeriodType` 为 nil 时类型字段为空字符串且无 fallback。

## What Changes

- 在 `pkg/profile/analysis.go` 的 `extractMetadata` 增加 fallback 逻辑：若 `p.PeriodType` 为 nil 或 type 为空，则从 `p.SampleType` 列表推断 profile 类型（检测 `inuse_space` / `alloc_space` / `inuse_objects` 等关键字推断为 `heap`；检测 `cpu` 推断为 `cpu`）。
- 推断规则集中到一个新函数 `inferProfileType(p)` 中，便于单测。
- 增加单元测试覆盖空 PeriodType 场景。

## Capabilities

### New Capabilities

无。

### Modified Capabilities

- `profile-analyze`：补充 REQUIREMENT —— 当 profile 的 `PeriodType` 缺失时，`type` 字段 SHALL 从 `SampleType` 推断而非返回空字符串。

## Impact

- **代码**：`pkg/profile/analysis.go`（`extractMetadata` + 新增 `inferProfileType`）
- **输出**：JSON 输出 `type` 字段在原为空的 profile 上变为非空字符串（如 `heap`）
- **测试**：新增单元测试覆盖 PeriodType 为 nil 的场景