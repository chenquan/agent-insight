## MODIFIED Requirements

### Requirement: Analysis summary adapts to profile type
系统 SHALL 在 summary 中使用与 profile 类型匹配的措辞。

#### Scenario: CPU profile summary
- **WHEN** 分析一个 cpu 类型的 profile
- **THEN** summary 使用 CPU 相关措辞（如 "CPU hotspot"、"性能热点"）

#### Scenario: Heap profile summary
- **WHEN** 分析一个 heap 类型的 profile
- **THEN** summary 使用内存相关措辞（如 "memory hotspot"、"内存热点"）

#### Scenario: Goroutine profile summary
- **WHEN** 分析一个 goroutine 类型的 profile
- **THEN** summary 使用 goroutine 相关措辞（如 "blocking point"、"阻塞点"）

### Requirement: JSON output includes value unit
系统 SHALL 在 JSON 输出的 flat/cum 值旁附带单位信息。

#### Scenario: Heap profile JSON output
- **WHEN** 分析 heap profile 并输出 JSON
- **THEN** 输出中包含 unit 字段（如 "bytes"）

#### Scenario: CPU profile JSON output
- **WHEN** 分析 cpu profile 并输出 JSON
- **THEN** 输出中包含 unit 字段（如 "nanoseconds" 或 "count"）

### Requirement: JSON percentage precision
系统 SHALL 在 JSON 输出中统一百分比字段保留 2 位小数。

#### Scenario: Percentage formatting
- **WHEN** 输出 flat_percent 和 cum_percent 字段
- **THEN** 值统一为 2 位小数（如 66.67 而非 66.66666666666666）

### Requirement: JSON field naming convention
系统 SHALL 在 JSON 输出中统一使用 snake_case 字段命名。

#### Scenario: Field naming consistency
- **WHEN** 输出 JSON 格式
- **THEN** 所有字段名使用 snake_case（如 flat_percent 而非 flatPercent）
