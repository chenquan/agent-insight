## Context

测试发现 `gobench.heap` 和 `java.heap` 在 `analyze --format json` 输出中 `type` 字段为空字符串。根因：`pkg/profile/analysis.go:153-156`：

```go
if p.PeriodType != nil {
    metadata.Type = p.PeriodType.Type
}
```

当 `PeriodType` 为 nil 时直接跳过赋值，`metadata.Type` 保持零值 `""`。

观察：`pprof/profile.Profile` 的 `PeriodType` 在某些 heap profile 上确实可能为 nil（profile metadata 缺失），但 `SampleType` 列表（value_types）通常存在，且包含足够的语义信息（`alloc_objects` / `inuse_space` 等）。

## Goals / Non-Goals

**Goals:**
- 当 `PeriodType` 缺失或 type 为空时，从 `SampleType` 推断 profile 类型
- 推断结果稳定且符合直觉：`heap`（检测到 `inuse_space`/`alloc_space`/`inuse_objects`/`alloc_objects`）/ `cpu`（检测到 `cpu`）/ `goroutine`（检测到 `goroutine`）/ 兜底 `unknown`
- 保持现有正常 profile（PeriodType 非空）行为不变

**Non-Goals:**
- 不重新设计 profile type 检测逻辑
- 不影响其他命令（list / traces / tree）的 metadata 提取（它们各自独立）

## Decisions

### Decision 1: 新增 `inferProfileType(p)` 纯函数

把推断逻辑独立成纯函数，理由：
- 易单测
- 与 `extractMetadata` 职责分离
- 未来增加新类型只需扩展该函数

### Decision 2: 优先级 PeriodType > SampleType

保持 `PeriodType` 优先（它是 profile 的"权威"元数据），仅在缺失或空时 fallback 到 SampleType 推断。理由：
- 不破坏现有逻辑
- 仅修复实际有问题的 case

### Decision 3: 推断规则保守

仅在 SampleType 包含明确关键字时返回特定类型，否则返回 `unknown`（而非空字符串）。理由：
- `unknown` 比 `""` 更友好（下游代码可以做 `if type != "unknown"` 判断）
- 避免误判

## Risks / Trade-offs

- [风险] 推断规则可能误判（如 `inuse_objects` 同时出现在 goroutine profile 上）→ 缓解：检测到 `goroutine` 关键字时优先返回 `goroutine`；多关键字匹配时按优先级排序
- [风险] 老用户脚本可能依赖 `type: ""` 做判断 → 缓解：低概率，且该行为本就是 bug

## Migration Plan

无数据迁移。纯代码修复。