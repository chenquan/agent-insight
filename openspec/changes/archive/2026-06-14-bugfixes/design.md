## Context

代码审查发现 agent-insight 存在多个 bug 和一致性问题。这些问题涉及 pkg/output、pkg/commands、pkg/profile 三个包。所有修复都是局部的代码修改，不涉及架构变更或新依赖。

现有代码模式：
- `--value-type` flag 只在 `analyze` 命令中完整实现（注册 + 读取 + 传递到 profile 层），其他 6 个命令只注册了 flag 但未读取
- `ValidateFormat` 是 `pkg/commands/validate.go` 中的共享函数，已被 analyze/diff/diagnose/flame/list/trend/tree/traces 使用，唯独 info 命令手写了校验逻辑

## Goals / Non-Goals

**Goals:**
- 消除所有已知 bug（panic、除零、接口不符）
- 消除所有沉默的错误（flag 被忽略但不报错）
- 保持现有行为不变，只修复不正确的地方

**Non-Goals:**
- 不重构代码结构（如抽取通用格式分发函数）
- 不新增功能
- 不修改 JSON 输出格式（字段增删留到 docs change）
- 不补充测试（留到 output-tests change）

## Decisions

### --value-type: 移除而非实现

**选择：** 从 diff、tree、traces、list、flame、trend 命令中移除 `--value-type` flag 注册。

**理由：** 这 6 个命令的实现中，profile 层的各函数（Tree、Traces、List、Flame、Diff、Trend）内部已经通过 `selectDefaultValueType` 自动选择了合适的值类型。手动覆盖值类型只在 analyze 命令（heap profile 有 4 种值类型，用户可能需要切换）有实际价值。在其他命令中实现 --value-type 的收益极低，且与自动选择逻辑可能冲突。如果未来确实需要，可以再加回来。

### flame --stats: 警告而非报错

**选择：** 当 `--stats` 与 `--format json/markdown` 同时使用时，输出 stderr 警告但仍继续执行。

**理由：** --stats 是个辅助信息，不影响主输出。因为警告而中断执行过于严格。

### TrendMarkdownFormatter: 重命名方法

**选择：** 将 `FormatTrendMarkdownResult` 重命名为 `FormatTrendResult`，使其满足 `TrendFormatter` 接口。

**理由：** 其他 Markdown Formatter（如 DiffMarkdownFormatter）已经使用统一的 `FormatDiffResult` 方法名。这是命名一致性的修复，不改变行为。

## Risks / Trade-offs

- **[移除 --value-type]** → 如果有用户正在使用这些 flag，会收到 "unknown flag" 错误。但这比当前"flag 被静默忽略"的行为更好，因为静默忽略更难发现。
- **[diff 除零修复]** → BaseSamples=0 意味着空 profile，修复后输出 "N/A" 而非 "+Inf%"，这是更合理的行为。
