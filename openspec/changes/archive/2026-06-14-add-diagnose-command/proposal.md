## Why

agent-insight 的现有命令（analyze、tree、traces 等）输出结构化数据，由上层 AI 助手自行理解并给出诊断建议。但不同 profile 类型的诊断关注点差异很大（CPU 关注计算热点和 GC 压力，Heap 关注内存分配和泄漏，Goroutine 关注阻塞和泄漏），且不同编程语言的优化手段不同。新增 `diagnose` 命令，将 profile 分析数据组装成针对特定 profile 类型和编程语言定制的高质量诊断 prompt，让 Claude Code 等 AI 助手能更准确、更高效地完成性能诊断。

## What Changes

- 新增 `diagnose` 子命令，接受 pprof 文件路径，输出结构化的诊断 prompt
- 新增语言检测能力，从 Function.Name/Filename/Mapping 推断编程语言（Go/C++/Rust/Java/C/Unknown）
- 新增诊断引导模板系统，按 profile 类型（CPU/Heap/Goroutine/Contentions/Unknown）和语言两个维度拼装引导内容
- 支持用户通过 `--context` flag 传入应用背景信息，嵌入 prompt
- 支持 `--top N` 控制热点函数数量，避免 prompt 过长
- 支持 `--format text|markdown|json` 三种输出格式
- 更新 SKILL.md 模板，包含 diagnose 命令的用法说明

## Capabilities

### New Capabilities
- `profile-diagnose`: diagnose 命令的核心能力，包括语言检测、引导模板拼装、prompt 组装输出

### Modified Capabilities

## Impact

- 新增文件：`pkg/commands/diagnose.go`、`pkg/profile/diagnose.go`
- 修改文件：`pkg/output/formatter.go`（新增 diagnose 格式化）、`cmd/root.go`（注册子命令）、`pkg/skill/template.md`（更新模板）
- 无新增外部依赖，完全复用现有 profile 层能力
