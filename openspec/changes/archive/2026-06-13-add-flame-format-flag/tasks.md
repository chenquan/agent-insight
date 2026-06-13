## 1. Add format flag in flame command

- [x] 1.1 在 `pkg/commands/flame.go` 注册 `--format` flag，默认 `text`，可选值 `text|json|markdown`
- [x] 1.2 在 `runFlame` 函数中根据 `flameFormat` 值分发到对应 formatter

## 2. Implement JSON and Markdown formatters

- [x] 2.1 在 `pkg/output/formatter.go` 新增 `FormatFlameResultJSON(w io.Writer, result *profile.FlameResult) error`，使用 `encoding/json` 序列化 FlameResult
- [x] 2.2 在 `pkg/output/formatter.go` 新增 `FormatFlameResultMarkdown(w io.Writer, result *profile.FlameResult) error`，输出元信息表格 + 前 20 栈表格 + 完整折叠栈代码块
- [x] 2.3 在 `pkg/output/formatter.go` 给现有 `FlameFormatter` 增加 `Format(result, format string)` 统一入口（避免在 commands 层做 if-else）

## 3. Validate format value

- [x] 3.1 在 `runFlame` 中校验 `flameFormat` 必须是 `text|json|markdown` 之一，否则返回明确错误

## 4. Add unit tests

- [x] 4.1 在 `pkg/commands/flame_test.go`（或新建）添加测试：`--format json` 输出包含 `total_stacks` 字段
- [x] 4.2 添加测试：`--format markdown` 输出包含 `|` 表格分隔符和 ``` 代码块
- [x] 4.3 添加测试：`--format yaml` 返回非零退出码
- [x] 4.4 添加测试：默认（无 `--format`）输出与之前 text 行为一致

## 5. Manual verification

- [x] 5.1 `make build && ./agent-insight flame testdata/cpu.pb.gz --format json | jq` 验证 JSON 可解析
- [x] 5.2 `./agent-insight flame testdata/cpu.pb.gz | head -3` 验证默认 text 输出未变
- [x] 5.3 `./agent-insight flame testdata/cpu.pb.gz --format markdown > /tmp/flame.md && head -20 /tmp/flame.md` 验证 markdown 输出可读