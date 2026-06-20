## MODIFIED Requirements

### Requirement: Filter by pprof labels

`diff` 命令 SHALL 支持 `--tag key=value` 和 `--tag-ignore key=value` flag。系统 SHALL 对 base 和 target 两个 profile **应用同一 filter**，再进行 diff 计算。

- `--tag` 可重复多次
- 同 key 多次 → OR
- 跨 key → AND
- filter 在 base 上先执行；若 base 匹配 0 样本则报错（不进入 target）

#### Scenario: 单 tag 过滤
- **WHEN** 用户跑 `agent-insight diff v1.pb.gz v2.pb.gz --tag http.status=500`
- **THEN** base 和 target 都先按 `http.status=500` 过滤，再 diff

#### Scenario: 同 key OR
- **WHEN** 用户跑 `agent-insight diff v1.pb.gz v2.pb.gz --tag http.status=500 --tag http.status=503`
- **THEN** base 和 target 都按 status=500 OR status=503 过滤

#### Scenario: 跨 key AND
- **WHEN** 用户跑 `agent-insight diff v1.pb.gz v2.pb.gz --tag http.method=POST --tag http.status=500`
- **THEN** base 和 target 都按两个条件 AND 过滤

#### Scenario: --tag-ignore 排除
- **WHEN** 用户跑 `agent-insight diff v1.pb.gz v2.pb.gz --tag-ignore http.status=200`
- **THEN** base 和 target 都排除 http.status=200 的样本

#### Scenario: base 0 样本报错
- **WHEN** 用户跑 `agent-insight diff v1.pb.gz v2.pb.gz --tag http.status=600`（不存在的 status）
- **THEN** 命令在 base 上就匹配 0，退出并报错

#### Scenario: --tag 与 --focus 组合
- **WHEN** 用户跑 `agent-insight diff v1.pb.gz v2.pb.gz --tag http.status=500 --focus "main.*"`
- **THEN** 先按 label 过滤两边 sample，再按 --focus 过滤函数，diff 是两者的交集
