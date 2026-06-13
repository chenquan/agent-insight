## Context

agent-insight 使用 `github.com/google/pprof/profile` 解析 pprof 格式文件，这是一个跨语言的开放协议。工具已经通过实际测试验证了对 Go、C++、Java profile 的支持，但 skill template 和 README 的面向用户描述仍将其定位为 "Go 性能分析工具"。

## Goals / Non-Goals

**Goals:**
- 去掉面向用户文档中的 "Go 专用" 误导
- 明确说明支持多种语言的 pprof profile

**Non-Goals:**
- 不修改代码内部的 Go 示例（runtime.mallocgc 等）
- 不修改 CLAUDE.md、testdata、Go 源码注释
- 不添加新的语言支持功能

## Decisions

1. **skill template description 用 "pprof" 替代 "Go"**：pprof 本身就是跨语言协议名称，无需枚举语言。
2. **README 副标题加入跨语言说明**：在工具定位层面明确说明，而非仅改括号里的列举。
3. **保留 Go 示例函数名不改**：示例中的 `runtime.mallocgc` 是合理的示例数据，不会造成误导。

## Risks / Trade-offs

- 无风险：纯文本修改，不影响任何行为。
