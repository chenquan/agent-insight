## Context

agent-insight 现有 10 个命令，每个命令在 `pkg/commands/` 中独立实现参数解析、验证、输出路由。当前存在两类可靠性问题：(1) `analysis.go` 的 `calculateHotspots` 对 totalValue == 0 无防护——当所有 sample 被 focus/ignore 过滤掉时，百分比计算产生 NaN；(2) commands 层 10 个命令各自重复实现 format 和 regex 验证逻辑，约 20 段重复代码且措辞不一致。

代码审查确认：diff/trend 的类型校验已通过 `ValidateTypeConsistency` 实现；list/traces/tree/diff 的百分比计算已有 `if totalValue > 0` 防护；flame 不做百分比计算；loader.go 错误消息已包含路径。这些不需要处理。

## Goals / Non-Goals

**Goals:**
- 修复 analysis.go 中唯一的除零/NaN 风险
- 消除 commands 层验证逻辑重复，统一措辞
- 为 commands 层补充关键测试

**Non-Goals:**
- flag 重命名、输出函数重构、新增功能、性能优化

## Decisions

**1. 除零防护在 calculateHotspots 内部而非 loader 层**

选项 A：在 loader.go 加载后校验零样本
选项 B：在 calculateHotspots 中对 totalValue == 0 做防护

选择 B。原因：(1) 只有 analysis.go 有真实除零风险，其他函数已有防护；(2) loader 层校验会影响 merge 命令——merge 加载多个文件时个别空文件不应报错；(3) totalValue == 0 也可能是 focus/ignore 过滤后的结果，不是 profile 本身的问题。防护方式：totalValue == 0 时直接返回空 hotspots 切片，不报错。

**2. 公共验证函数放在 commands 包内**

`pkg/commands/validate.go`。这些函数仅在 commands 层使用，不需要独立包。

**3. ValidateRegex 统一措辞为 "invalid {name} pattern"**

现有代码混用 "regex"、"pattern"、无前缀三种格式。统一为 "invalid focus pattern"、"invalid ignore pattern"、"invalid exclude pattern"。

## Risks / Trade-offs

- **[公共验证函数微调错误消息]** → 措辞统一，不影响退出码或行为。
- **[commands 测试需要 testdata]** → 复用现有 testdata/cpu.pb.gz 和 heap.pb.gz。
