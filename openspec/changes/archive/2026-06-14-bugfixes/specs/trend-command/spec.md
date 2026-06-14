## ADDED Requirements

### Requirement: funcName handles all-nil pointers safely
trend 输出中的 `funcName` 函数 SHALL 在 Function、Address、LocationID 三个指针全为 nil 时返回 "unknown" 字符串，不 SHALL panic。

#### Scenario: 三个指针全为 nil
- **WHEN** FunctionTrend 的 Function、Address、LocationID 均为 nil
- **THEN** funcName 返回 "unknown"

#### Scenario: 只有 Function 为 nil
- **WHEN** Function 为 nil 但 Address 不为 nil
- **THEN** funcName 返回 Address 的值

### Requirement: TrendMarkdownFormatter implements TrendFormatter interface
TrendMarkdownFormatter SHALL 实现 `FormatTrendResult(result *profile.TrendResult) error` 方法，满足 TrendFormatter 接口定义。

#### Scenario: Markdown 格式输出
- **WHEN** 使用 trend 命令的 `--format markdown` 输出
- **THEN** 通过 TrendFormatter 接口调用 FormatTrendResult 正常渲染 markdown 格式
