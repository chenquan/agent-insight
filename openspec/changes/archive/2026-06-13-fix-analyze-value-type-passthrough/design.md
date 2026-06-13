## Context

`analyze` 命令的 `--value-type` flag 在 PR/issue 阶段被设计用于让用户指定多值 profile 的分析维度，但实际实现中：
- `pkg/commands/analyze.go:97-102` 读取 flag 后赋值给 `_ = analyzeValueType`（直接丢弃）
- `pkg/profile/analysis.go:100-102` 的 `if config.ValueType == nil` 总是 true，于是 fallback 到 `selectDefaultValueType`

导致无论用户指定 `alloc_objects` / `alloc_space` / `inuse_objects` / `inuse_space` 哪一个，输出都基于默认的 `inuse_space`。这是一个用户可见的"flag 形同虚设"的 bug。

## Goals / Non-Goals

**Goals:**
- `--value-type` flag 真正控制 profile 分析使用的值类型
- 用户指定无效值类型时给出明确错误（列出 profile 中可用的 value types）
- 现有 spec 中的 "User-specified value type" scenario 真正通过测试验证

**Non-Goals:**
- 不修改 `selectDefaultValueType` 的默认选择逻辑
- 不影响其他命令（list / traces / tree 等）的 value-type 处理
- 不重构 AnalysisConfig 结构

## Decisions

### Decision 1: 在 commands 层完成 string → ValueTypeConfig 转换

把 `analyzeValueType` 字符串解析逻辑放在 `pkg/commands/analyze.go` 而非 `pkg/profile/analysis.go`，理由：
- commands 层本就是 flag 解析边界
- profile 层只接收已构造好的 `ValueTypeConfig` 结构体，职责单一
- 与 `pkg/profile/analysis.go` 的 `NewAnalysis(p, config AnalysisConfig)` 签名一致

### Decision 2: 校验逻辑放在 commands 层

无效 value-type 的报错在 commands 层完成（参数校验），通过 `pprof` 的 `profile.SampleType` 列表查找。理由：
- 避免在 profile 层引入额外的错误路径
- 错误信息可以结合 cobra 的 ExitCode 正确传递

### Decision 3: 用 ADDED Requirements 而非 MODIFIED

虽然现有 spec 已有 "User-specified value type" scenario（第91-94行），但因 B-1 bug 该 scenario 未真正生效。**选择 ADDED 一个新的 verification scenario** 而非 MODIFIED 旧的，理由：
- 旧 scenario 描述意图正确，问题是实现未对齐
- ADDED 一个 "value-type flag actually changes output" scenario 更清晰地表达"修复实现以满足 spec"
- MODIFIED 会触发"完整复制旧内容"的工作流，对当前改动不必要

## Risks / Trade-offs

- [风险] 现有 spec scenario 91-94 与新 ADDED scenario 看似重复 → 缓解：旧 scenario 描述"行为意图"，新 scenario 描述"可验证的正确性"，archive 时合并到主 spec
- [风险] 命令行解析顺序变化可能影响现有用户脚本 → 缓解：行为变化仅在 `--value-type` 非空时生效，无 flag 用户不受影响

## Migration Plan

无需数据迁移。仅在 `pkg/commands/analyze.go` 内部修复行为。