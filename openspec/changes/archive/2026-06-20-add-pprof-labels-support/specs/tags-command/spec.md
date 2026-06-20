# tags-command

## Purpose

`tags` 命令列出 pprof profile 中所有 label 及其 value 分布，供 AI 在做标签过滤或 breakdown 分析前了解 profile 的 label 维度全貌。
## ADDED Requirements
### Requirement: tags 命令展示 profile 中所有 label 概览

系统 SHALL 提供 `tags` 命令，接受一个 profile 文件路径，输出该 profile 的全部 label 概览（key + value 分布 + 计数）。

#### Scenario: goroutine profile 输出
- **WHEN** 用户执行 `agent-insight tags goroutine.pb.gz`
- **THEN** 输出包含 profile 类型、采样数、以及所有 label（如 `state`、`wait_reason`）的 value 分布

#### Scenario: 无 label profile 输出空 label 列表
- **WHEN** 用户执行 `agent-insight tags cpu.pb.gz` 且 profile 不含 label
- **THEN** 输出包含提示 "no labels found"（或在 JSON 中 `labels: []`）

#### Scenario: 数字 label 输出带单位
- **WHEN** profile 含 `cpu` 数字 label，unit 为 "nanoseconds"
- **THEN** text 输出显示 `cpu (numeric, nanoseconds, N values)`
- **AND** JSON 输出 `unit == "nanoseconds"`

### Requirement: tags 命令支持多格式输出

系统 SHALL 支持 `--format` flag，可选值为 text（默认）、json、markdown。

#### Scenario: JSON 格式输出
- **WHEN** 用户执行 `agent-insight tags goroutine.pb.gz --format json`
- **THEN** 输出为合法 JSON，含 `profile_path` / `type` / `total_samples` / `labels` 字段

#### Scenario: Markdown 格式输出
- **WHEN** 用户执行 `agent-insight tags goroutine.pb.gz --format markdown`
- **THEN** 输出为 Markdown 表格

### Requirement: tags 命令支持 --top 限制 value 数量

数字 label 在持续剖析 profile 中可能有成千上万个不同 value（如 `cpu` 时间分布）。系统 SHALL 提供 `--top N` flag 限制每个 label value 列表的输出数量，默认 **50**（数字依据：分析场景下 50 个 value 已足够覆盖主要耗时聚集点；`analyze --tag-breakdown-top` 默认 20 是因为按函数粒度，50 按 label value 粒度仍属合理上界）。

string label 通常 value 数 ≤ 20，--top 不影响 string label（不截断）。

当数字 label 被截断时，输出 SHALL 标注 `values_truncated: true`（JSON）或 "showing top N of M"（text）。

#### Scenario: 默认 50 value 截断
- **WHEN** 某数字 label 有 8521 个不同 value，用户跑 `tags` 不传 `--top`
- **THEN** 输出展示 top 50 value，并标注 "8521 unique values"

#### Scenario: 自定义 top 数量
- **WHEN** 用户传 `--top 20`
- **THEN** 输出展示 top 20 value

#### Scenario: string label 不截断
- **WHEN** `state` label 只有 4 个不同 value（blocked/running/syscall/preempted）
- **THEN** `--top` 不影响该 label，4 个全部展示

### Requirement: tags 命令参数验证

系统 SHALL 验证输入参数：恰好一个 profile 文件路径，format 必须为 text/json/markdown 之一。

#### Scenario: 缺少文件参数
- **WHEN** 用户执行 `agent-insight tags` 不带文件路径
- **THEN** 返回错误提示

#### Scenario: 无效格式参数
- **WHEN** 用户执行 `agent-insight tags profile.pb.gz --format xml`
- **THEN** 返回错误提示有效格式列表（`ValidateFormat` 复用）

### Requirement: tags 不接受 label filter

`tags` 命令 SHALL **不接受** `--tag` / `--tag-ignore` flag。`tags` 命令反映 profile 原始 label 分布，filter 后的 label 分布属于 v2 范围。

#### Scenario: --tag 在 tags 命令中不被识别
- **WHEN** 用户执行 `agent-insight tags profile.pb.gz --tag state=blocked`
- **THEN** 报 "unknown flag: --tag"

#### Scenario: --tag-ignore 在 tags 命令中不被识别
- **WHEN** 用户执行 `agent-insight tags profile.pb.gz --tag-ignore state=blocked`
- **THEN** 报 "unknown flag: --tag-ignore"

### Requirement: tags 命令输出按 distinct value 数排序

多个 label 同时存在时，系统 SHALL 按 distinct value 数降序排序展示，让信息量最高的 label 排在前面。同一 label 内的 value 列表按 count 降序排序。

v1 SHALL NOT 提供控制 value 排序的 flag（不支持按 value 字典序、按 percentage 等排序）；用户用 `--top N` 控制数量即可。

#### Scenario: 多 label 排序
- **WHEN** profile 同时含 `state`（4 distinct）和 `http.path`（200 distinct）
- **THEN** `http.path` 排在 `state` 前面

#### Scenario: 同 label 内 value 按 count 排序
- **WHEN** `state` label 中 syscall 有 4891 个 sample、blocked 有 3200 个 sample
- **THEN** `syscall` 排在 `blocked` 前面
