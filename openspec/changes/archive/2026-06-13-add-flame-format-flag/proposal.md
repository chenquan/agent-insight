## Why

`flame` 命令当前只输出折叠栈 text 格式（`func1;func2;func3 count`），不支持 `--format` flag。对 AI 解析、Markdown 报告集成等场景不友好。其他命令（`analyze` / `list` / `tree` / `traces` / `diff` / `info`）均支持 `--format text|json|markdown`，能力不一致。

## What Changes

- 在 `pkg/commands/flame.go` 注册 `--format` flag，默认 `text`（保持现有行为）。
- 在 `pkg/output/formatter.go` 新增 `FormatFlameResultJSON(result)` 函数，输出 JSON 结构 `{total_stacks, filtered_stacks, unique_stacks, stacks: [{stack: [...], value: N}, ...]}`。
- 在 `pkg/output/formatter.go` 新增 `FormatFlameResultMarkdown(result)` 函数，输出 Markdown 表格 + 折叠栈代码块。
- 在 `pkg/commands/flame.go` 根据 `--format` 值选择对应 formatter。
- 默认 `--format text` 行为不变，保证现有管道用户不受影响。

## Capabilities

### New Capabilities

无。

### Modified Capabilities

- `profile-flame`：补充 REQUIREMENT —— `flame` 命令 SHALL 支持 `--format` flag，值域 `text|json|markdown`，默认 `text`。

## Impact

- **代码**：`pkg/commands/flame.go`（注册 flag + 分发）、`pkg/output/formatter.go`（新增两个 format 函数）
- **行为**：默认行为零变化；新增 json/markdown 输出
- **依赖**：可能需要 `encoding/json`（项目已用，无需新增）
- **测试**：新增单元测试验证 `--format json` 与 `--format markdown` 输出