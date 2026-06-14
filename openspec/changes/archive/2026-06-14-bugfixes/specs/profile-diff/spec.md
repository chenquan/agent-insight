## ADDED Requirements

### Requirement: Diff handles zero-base profile safely
当 base profile 的 sample 数为 0 时，markdown 格式输出中的 sample 变化百分比 SHALL 显示 "N/A" 而非数值。text 和 JSON 格式 SHALL 同样安全处理此边界情况。

#### Scenario: Markdown 格式 BaseSamples 为 0
- **WHEN** base profile 包含 0 个 sample，target profile 包含 100 个 sample
- **THEN** markdown 输出中 sample 变化百分比显示为 "N/A"

#### Scenario: Text 格式 BaseSamples 为 0
- **WHEN** base profile 包含 0 个 sample
- **THEN** text 输出中 sample 变化百分比显示为 "N/A"

#### Scenario: 正常情况不受影响
- **WHEN** base profile 包含 50 个 sample，target 包含 100 个
- **THEN** sample 变化百分比正常计算为 "+100.00%"
