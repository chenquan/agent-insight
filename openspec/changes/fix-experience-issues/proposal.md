## Why

现有 9 个命令的核心功能完整，但输出质量和一致性存在多处体验问题：diff 不校验 profile 类型导致无意义对比、summary 措辞不适配 heap/goroutine 类型、JSON 精度和字段命名不一致等。这些问题影响 AI 助手对输出结果的解读可靠性。

## What Changes

- **P1**: diff 命令增加 profile 类型一致性校验，拒绝混合类型对比
- **P2**: analyze 的 summary 措辞根据 profile 类型自适应（heap 说"内存热点"、goroutine 说"阻塞点"等）
- **P3**: analyze JSON 输出中 flat/cum 值附带单位信息（heap: bytes，goroutine: count）
- **P4**: info 命令对 goroutine profile 输出总 goroutine 数量
- **P5**: 各命令的 help text 和 examples 完善
- **P6**: JSON 输出中百分比字段统一保留 2 位小数
- **P7**: diff text 输出中 cum 列修正为有意义的值
- **P8**: JSON 输出字段命名统一为 snake_case

## Capabilities

### New Capabilities

### Modified Capabilities
- `profile-diff`: 新增类型校验要求；cum 值输出修正
- `profile-analyze`: summary 措辞自适应；值附带单位；百分比精度统一

## Impact

- **修改文件**: `pkg/profile/diff.go`、`pkg/profile/analysis.go`、`pkg/profile/info.go`、`pkg/output/formatter.go`、`pkg/commands/diff.go`、`pkg/commands/analyze.go`、`pkg/commands/info.go`
- **不新增文件**: 全部为现有代码修改
- **不新增依赖**
