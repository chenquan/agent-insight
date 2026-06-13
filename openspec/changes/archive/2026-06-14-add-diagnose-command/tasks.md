## 1. 语言检测

- [x] 1.1 在 `pkg/profile/` 新增语言检测函数 `DetectLanguage(p *profile.Profile) string`，遍历 Function.Name/Filename 用正则匹配 Go/C++/Rust/Java/C 模式，取匹配数最多的语言返回
- [x] 1.2 编写语言检测单元测试，覆盖 Go/C++/Rust/Java/C/Unknown 和无符号场景

## 2. 诊断引导模板

- [x] 2.1 在 `pkg/profile/` 新增引导模板数据结构，定义 5 种 profile 类型（CPU/Heap/Goroutine/Contentions/Unknown）的基础引导文本
- [x] 2.2 定义 5 种语言（Go/C++/Rust/Java/C）的语言追加文本，Unknown 语言返回空追加
- [x] 2.3 编写引导模板拼装逻辑，根据 profile 类型和语言组合生成最终引导文本

## 3. Prompt 构建核心逻辑

- [x] 3.1 在 `pkg/profile/diagnose.go` 新增 `DiagnosePrompt` 结构体，包含角色、概况、分析数据、引导、用户上下文等字段
- [x] 3.2 实现 `BuildDiagnosePrompt(p *profile.Profile, topN int, context string) (*DiagnosePrompt, error)` 函数，复用 Analyze/BuildTree/GetTopTraces 提取数据，拼装完整 prompt
- [x] 3.3 实现 `DiagnosePrompt.Text() string` 方法，将结构化数据渲染为纯文本 prompt
- [x] 3.4 编写 prompt 构建单元测试

## 4. 命令层

- [x] 4.1 在 `pkg/commands/diagnose.go` 新增 cobra 命令 `DiagnoseCmd`，支持 `--top N`（默认 10）、`--context string`、`--format text|markdown|json` flags
- [x] 4.2 在 `cmd/root.go` 注册 `DiagnoseCmd` 子命令

## 5. 输出层

- [x] 5.1 在 `pkg/output/formatter.go` 新增 `FormatDiagnose(*profile.DiagnosePrompt) error` 方法，支持 text/markdown/json 三种格式
- [x] 5.2 JSON 格式输出包含 `prompt` 字段和 `data` 字段（原始分析数据）

## 6. 测试与验证

- [x] 6.1 使用 `testdata/cpu.pb.gz` 和 `testdata/heap.pb.gz` 端到端验证 diagnose 命令输出
- [x] 6.2 验证 --context 和 --top N flag 行为

## 7. 文档更新

- [x] 7.1 更新 `pkg/skill/template.md`，新增 diagnose 命令的用法说明
