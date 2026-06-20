## MODIFIED Requirements

### Requirement: Filter by pprof labels

`analyze` 命令 SHALL 支持 `--tag key=value` 和 `--tag-ignore key=value` flag，对 profile 的 sample 做 label 维度过滤。

- `--tag` 可重复多次
- 同 key 多次 → OR（`--tag state=blocked --tag state=running` 等于 state 是 blocked 或 running）
- 跨 key → AND（`--tag state=blocked --tag wait_reason=IO` 等于两个条件都满足）
- `--tag-ignore` 反向，同语义
- 数字 label value 必须是十进制整数字符串
- 0 样本匹配时退出并报错（见 labels-support 0 样本报错要求）

#### Scenario: 单 tag 过滤
- **WHEN** 用户跑 `agent-insight analyze goroutine.pb.gz --tag state=blocked`
- **THEN** 输出只含 state=blocked 的 sample 贡献，hotspots 数量减少

#### Scenario: 同 key OR
- **WHEN** 用户跑 `agent-insight analyze goroutine.pb.gz --tag state=blocked --tag state=running`
- **THEN** 输出含 state=blocked 或 state=running 的 sample

#### Scenario: 跨 key AND
- **WHEN** 用户跑 `agent-insight analyze goroutine.pb.gz --tag state=blocked --tag wait_reason=IO`
- **THEN** 输出仅含同时满足两个条件的 sample

#### Scenario: --tag-ignore 排除
- **WHEN** 用户跑 `agent-insight analyze goroutine.pb.gz --tag-ignore state=running`
- **THEN** 输出排除所有 state=running 的 sample

#### Scenario: 0 样本退出
- **WHEN** 用户跑 `agent-insight analyze cpu.pb.gz --tag state=blocked` 且 cpu.pb.gz 不含 state label
- **THEN** 命令以非零状态退出，错误信息含 "tag filter matched 0 of N samples"

#### Scenario: filter 与 --focus 组合
- **WHEN** 用户跑 `agent-insight analyze goroutine.pb.gz --tag state=blocked --focus "main.*"`
- **THEN** 先按 label 过滤 sample，再按 --focus 过滤函数，输出是两者的交集

### Requirement: Output label breakdown for top hotspots

`analyze` 命令 SHALL 在输出中包含 top hotspots 的 label breakdown 字段。`analyze` 命令 SHALL 支持：
- `--tag-breakdown-on key1,key2` 指定要展开的 label key（CSV）
- `--tag-breakdown-top N` 指定展开几个函数（默认 20，缺省值在调用方固定为 20）

breakdown 字段位置（JSON 格式）：

```json
{
  "functions": [
    {
      "function": "...",
      "flat": 1234,
      "label_breakdown": [...]   // ← 挂在 function entry 下面
    }
  ]
}
```

每个 function entry 的 `label_breakdown` 字段结构：
- 数组形式：按用户指定的 `cfg.Keys` 顺序，每个 key 一个 entry
- 每个 entry 含 `key` 和 `values` 数组
- `values` 数组按 flat 降序排序
- 数字 label 的 percentage 保留 2 位小数
- v1 不计算 cum：`cum` 和 `cum_pct` 字段 SHALL 存在但恒为 0/0.00（schema 稳定）

#### Scenario: 不指定 breakdown 时不输出
- **WHEN** 用户跑 `agent-insight analyze profile.pb.gz` 不传 `--tag-breakdown-on`
- **THEN** 输出中 `functions[i].label_breakdown` 字段为 `null`（JSON）或省略（text/markdown）

#### Scenario: 指定 breakdown key
- **WHEN** 用户跑 `agent-insight analyze goroutine.pb.gz --tag-breakdown-on state`
- **THEN** 输出中前 20 个 function entry 的 `label_breakdown` 字段含 `state` 维度

#### Scenario: 多个 breakdown key
- **WHEN** 用户跑 `agent-insight analyze goroutine.pb.gz --tag-breakdown-on state,http.method`
- **THEN** 每个函数 entry 的 `label_breakdown` 数组依次含 `state` 和 `http.method` 两个 entry

#### Scenario: 自定义 top 数量
- **WHEN** 用户跑 `agent-insight analyze profile.pb.gz --tag-breakdown-on state --tag-breakdown-top 5`
- **THEN** 输出中前 5 个 function entry 有 `label_breakdown` 数组，第 6 个及之后 `label_breakdown` 为 `null`

#### Scenario: breakdown 默认 top 20
- **WHEN** 用户跑 `agent-insight analyze profile.pb.gz --tag-breakdown-on state`（不传 top）
- **THEN** 前 20 个 function entry 有 `label_breakdown`，第 21 个及之后为 `null`

#### Scenario: breakdown flat 百分比为占该函数 total flat 的比例
- **WHEN** 函数 total flat=1000，breakdown 显示 state=blocked flat=600
- **THEN** 该 entry 的 flat=600, flat_pct=60.00

#### Scenario: breakdown v1 cum 字段为 0
- **WHEN** 用户跑 `analyze --tag-breakdown-on state`
- **THEN** breakdown entry 中 `cum` 字段存在但恒为 0，`cum_pct` 恒为 0.00

#### Scenario: breakdown 在三种格式中均存在
- **WHEN** 用户跑 `analyze --format text|json|markdown --tag-breakdown-on state`
- **THEN** 三种格式输出都含 label breakdown（text 缩进展示，JSON 嵌套，markdown 表格）

### Requirement: Filter combined with breakdown

`analyze` 命令 SHALL 同时支持 `--tag` 过滤和 `--tag-breakdown-on` breakdown，两者组合使用：先按 label 过滤 sample，再对过滤后的 sample 计算 breakdown。

#### Scenario: filter 后的 breakdown
- **WHEN** 用户跑 `agent-insight analyze goroutine.pb.gz --tag state=blocked --tag-breakdown-on wait_reason`
- **THEN** 输出中 hotspots 只含 state=blocked 的 sample 贡献，每个 top 函数的 breakdown 仅含 wait_reason 维度（数据是过滤后的）

#### Scenario: filter + breakdown 都不传时不变
- **WHEN** 用户跑 `agent-insight analyze profile.pb.gz`（不传任何新 flag）
- **THEN** 命令行为与 v0.X 之前的 `analyze` 完全一致：仅 top N 函数，无 label_breakdown 字段
