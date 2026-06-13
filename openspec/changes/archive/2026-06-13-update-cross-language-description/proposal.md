## Why

skill template 和 README 将 agent-insight 描述为 "Go 性能 profile 分析工具"，但实际上 pprof 是跨语言协议，工具已支持 Go、C++、Java 等多种语言的 profile（之前的 `normalize-cpp-mappings-path` change 和 TEST-REPORT 已验证）。当前描述会误导用户认为只能分析 Go profile。

## What Changes

- `pkg/skill/template.md`：description 从 "分析 Go 性能 profile 文件" 改为 "分析 pprof 性能 profile 文件"
- `README.md`：副标题重写，明确说明支持多种语言的 pprof profile
- `README.md`：Features 中 "(Go heap, etc.)" 改为 "(Go, C++, etc.)"

## Capabilities

### New Capabilities

无。

### Modified Capabilities

无（纯文档修改，不影响 spec 层面的行为定义）。

## Impact

- `pkg/skill/template.md`（skill 描述文本）
- `README.md`（副标题 + Features 一行）
- 无代码、无 API、无行为变更
