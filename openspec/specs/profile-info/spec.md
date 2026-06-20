## Purpose

Report metadata from a pprof profile file, including profile type, sample count, duration, and label summary, with lightweight (zero-computation) output for AI-assisted inspection.

## Requirements

### Requirement: Output includes label summary

`info` 命令 SHALL 在输出中包含 label 摘要行，描述 profile 中 label 的总览：label key 数量、distinct value 总数。

摘要 SHALL 从 `*Profile.LabelSummaries` 读取（loader 一次性计算的缓存），不重复遍历 sample。

- text 格式：`Labels: N keys, M unique values`
- JSON 格式：`label_summary: { "key_count": N, "distinct_values": M }`
- markdown 格式：单行 `**Labels**: N keys, M unique values`
- 当 profile 不含 label 时：`Labels: 0 keys, 0 unique values`（或 JSON `key_count: 0, distinct_values: 0`）

#### Scenario: goroutine profile 有 label
- **WHEN** 用户跑 `agent-insight info goroutine.pb.gz` 且 profile 含 3 个 label key 共 247 个 distinct value
- **THEN** 输出含 `Labels: 3 keys, 247 unique values`

#### Scenario: cpu profile 无 label
- **WHEN** 用户跑 `agent-insight info cpu.pb.gz` 且 profile 不含任何 label
- **THEN** 输出含 `Labels: 0 keys, 0 unique values`

#### Scenario: JSON 格式
- **WHEN** 用户跑 `agent-insight info goroutine.pb.gz --format json`
- **THEN** JSON 输出含 `label_summary: { "key_count": 3, "distinct_values": 247 }`

#### Scenario: 摘要不影响 info 轻量定位
- **WHEN** 用户跑 `info` 命令
- **THEN** label 摘要仅为一行 / 一字段，不展开 value 列表（详细 value 列表用 `tags` 命令）

#### Scenario: distinct 值算法
- **WHEN** profile 中 `state` 有 4 个 value（blocked/running/syscall/preempted），`http.path` 有 243 个 value
- **THEN** `info` 输出的 distinct_values = 247（即 4 + 243，跨 label 累加）
