## ADDED Requirements

### Requirement: Diff compares two profiles
系统 SHALL 比较两个 pprof profile 文件并识别性能变化。

#### Scenario: Compare two CPU profiles
- **WHEN** 用户对比两个相同 PeriodType 的 profile
- **THEN** 系统正常输出对比结果

#### Scenario: Compare mixed profile types
- **WHEN** 用户对比两个不同 PeriodType 的 profile（如 cpu 和 heap）
- **THEN** 系统报错并提示不能对比不同类型的 profile

## ADDED Requirements

### Requirement: Diff text output hides zero cum values
系统 SHALL 在 text 格式输出中隐藏 cum 值为 0 的列，减少视觉噪声。

#### Scenario: Text output with zero cum
- **WHEN** diff 结果中某函数的 cum 值为 0
- **THEN** text 格式输出中不显示该函数的 cum 列
