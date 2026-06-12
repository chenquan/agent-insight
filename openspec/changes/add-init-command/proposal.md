## Why

Claude Code 不知道 agent-insight 工具的存在，无法在用户提到 pprof、性能分析等话题时主动使用该工具。需要一个 `init` 命令让工具自己生成 Claude Code skill 文件，实现"自举"——安装后 Claude Code 就能自动识别何时使用 agent-insight 以及如何正确调用。

## What Changes

- 新增 `init` 子命令，生成 `.claude/skills/agent-insight/SKILL.md`
- SKILL.md 内嵌完整的命令使用指南（analyze/list/flame/diff 的用法、flags、示例）
- 包含触发场景表，教 Claude Code 何时主动使用该工具
- 包含典型工作流和输出解读指南

## Capabilities

### New Capabilities
- `init-command`: `init` 子命令，生成项目级 Claude Code skill 文件，包含触发条件、命令速查、工作流和输出解读

### Modified Capabilities

（无现有 capability 的需求变更）

## Impact

- 新增 `pkg/commands/init.go` 和 `cmd/` 中的注册代码
- 新增 `pkg/skill/` 包用于生成 SKILL.md 内容
- 无外部依赖变更
- 不影响现有命令（analyze/list/flame/diff）
