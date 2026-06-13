## Context

agent-insight 现有 9 个子命令功能完整，但输出质量和一致性存在 8 个体验问题。这些问题通过 code review 和实际运行测试数据发现。

## Goals / Non-Goals

**Goals:**
- 修复 8 个体验问题（P1-P8），提升输出质量和一致性
- 不改变命令的核心逻辑，只改进输出层和校验层

**Non-Goals:**
- 不新增子命令或核心分析功能
- 不改变 JSON 输出结构（只改字段名和值格式）

## Decisions

**P1: diff 类型校验 — 复用 merge 的 validateTypeConsistency 逻辑**

将 `pkg/profile/merge.go` 中的 `validateTypeConsistency` 提取为包级函数，diff 命令在调用前校验。不引入新依赖。

**P2: summary 措辞 — 按 profile 类型分支**

在 `pkg/output/formatter.go` 的 summary 生成逻辑中，根据 profile type 选择不同措辞：
- cpu → "performance bottleneck" / "CPU 热点"
- heap → "memory hotspot" / "内存热点"
- goroutine → "blocking point" / "阻塞点"
- 其他 → 通用措辞

**P3: 值附带单位 — 在 JSON 输出中添加 unit 字段**

在 analyze 的 JSON formatter 中，flat/cum 值旁附带 `unit` 字段（来自 ValueTypeConfig）。

**P4: goroutine 总数 — 在 info 输出中聚合 value**

goroutine profile 的 value 含义就是 goroutine 数量。在 InfoResult 中添加 `TotalValue int64`，对 goroutine profile 聚合所有 sample 的第一个 value。

**P5: help text — 逐命令审查补全**

为每个命令补充 Long 描述中的 Example 部分。

**P6: 百分比精度 — 统一为 math.Round(x*100)/100**

在 formatter 中输出百分比前统一 rounding。

**P7: diff cum 值 — 修正 buildLocationValueMap 的 cum 计算逻辑**

当前 cum 计算遍历 sample.Location 并给所有位置累加 value，但只累加了 leaf location 以外的 value。需要修正为：对每个 location 在所有 sample 中查找并累加。

实际上 diff 的 cum 语义是正确的（cum = 经过该函数的总量），问题出在测试数据中 sample 的 value 被分配给 leaf location，导致非 leaf 的 cum 计算为零。这是 pprof profile 数据结构本身的特点，不是 bug。改为：在 text 输出中不显示 cum 为 0 的列。

**P8: JSON 字段命名 — 统一 snake_case**

扫描所有 JSON formatter 输出的字段名，统一为 snake_case。

## Risks / Trade-offs

- **P1 向后兼容**: diff 原本允许混合类型对比，添加校验后不可用。但这个行为本身就是错误的。
- **P8 字段名变更**: 如果已有用户代码依赖现有 camelCase 字段名会 break。作为内部工具可接受。
- **P7 diff cum 显示**: 隐藏 cum=0 的列是 UX 折衷，可能丢失有效信息。仅在 text 格式中隐藏。
