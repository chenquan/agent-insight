## Context

`flame` 命令的输出格式是高度专门化的"折叠栈"格式（`func1;func2;func3 count`），专为 flame graph 工具链设计。但该格式对 AI/LLM 解析不友好——无法直接用结构化查询定位"最热的栈"。

其他命令（`analyze` / `list` / `tree` / `traces` / `diff` / `info`）均通过 `--format` flag 提供 `text|json|markdown` 三种输出。`flame` 是唯一缺失 `--format` 的命令，导致：

1. AI 助手分析 profile 时无法用 JSON 解析火焰图数据
2. 生成的报告无法嵌入 markdown 表格形式的"前 N 热点栈"

## Goals / Non-Goals

**Goals:**
- `flame` 命令支持 `--format text|json|markdown`，默认 `text`
- 现有 text 输出**字节级兼容**（保证管道到 `flamegraph.pl` 的用户不受影响）
- JSON 输出结构稳定，适合 AI 解析
- Markdown 输出包含表格 + 折叠栈代码块，便于报告嵌入

**Non-Goals:**
- 不改变折叠栈的聚合/排序逻辑
- 不引入新的值类型
- 不修改 FlameResult 数据结构

## Decisions

### Decision 1: 默认保持 text，flag 仅扩展输出

不强制要求 `--format` 必传，默认值 `text` 保证向后兼容。理由：
- 现有 `agent-insight flame profile.pb.gz | flamegraph.pl > graph.svg` 用法零变化
- 用户主动指定 `--format json` 时启用新能力

### Decision 2: JSON 输出复用 FlameResult 结构

JSON 输出结构对齐 FlameResult：`{total_stacks, filtered_stacks, unique_stacks, stacks: [...], config}`。理由：
- 避免引入新的数据模型
- 与 profile 层单一数据源保持一致

### Decision 3: Markdown 输出包含表格 + 折叠栈块

Markdown 输出包含：
1. 元信息表格（total / filtered / unique / value_type）
2. 前 20 个栈的表格（栈路径 + 值）
3. 完整折叠栈的代码块（保证火焰图工具仍可消费）

理由：兼顾报告嵌入和工具链兼容。

## Risks / Trade-offs

- [风险] Markdown 输出中的折叠栈代码块会让报告变长 → 缓解：仅在 `--format markdown` 时输出，text/JSON 不受影响
- [风险] JSON schema 变更会影响未来下游解析 → 缓解：在 spec 中固化结构，archive 时同步到主 specs

## Migration Plan

无需数据迁移。新增能力，不破坏现有行为。