## MODIFIED Requirements

### Requirement: Filter by pprof labels

`traces` 命令 SHALL 支持 `--tag key=value` 和 `--tag-ignore key=value` flag，对 profile 的 sample 做 label 维度过滤，再在过滤后的 sample 上做调用链展示。

- `--tag` 可重复多次
- 同 key 多次 → OR
- 跨 key → AND
- 0 样本匹配时退出并报错

#### Scenario: 单 tag 过滤
- **WHEN** 用户跑 `agent-insight traces goroutine.pb.gz --tag state=blocked`
- **THEN** 输出只展示 state=blocked 的 trace

#### Scenario: 同 key OR
- **WHEN** 用户跑 `agent-insight traces goroutine.pb.gz --tag state=blocked --tag state=running`
- **THEN** 输出展示 state 是 blocked 或 running 的 trace

#### Scenario: 跨 key AND
- **WHEN** 用户跑 `agent-insight traces goroutine.pb.gz --tag state=blocked --tag wait_reason=IO`
- **THEN** 输出仅展示同时满足两个条件的 trace

#### Scenario: --tag-ignore 排除
- **WHEN** 用户跑 `agent-insight traces goroutine.pb.gz --tag-ignore state=running`
- **THEN** 输出排除 state=running 的 trace

#### Scenario: --tag 与 --focus 组合
- **WHEN** 用户跑 `agent-insight traces goroutine.pb.gz --tag state=blocked --focus "runtime.mallocgc"`
- **THEN** 先按 label 过滤 sample，再按 --focus 过滤 trace，输出是两者的交集

#### Scenario: 0 样本退出
- **WHEN** 用户跑 `agent-insight traces cpu.pb.gz --tag state=blocked` 且 cpu.pb.gz 无 state label
- **THEN** 命令以非零状态退出，错误信息含 "tag filter matched 0 of N samples"
