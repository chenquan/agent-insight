# labels-support

## Purpose

定义 agent-insight 对 pprof profile 上 label（`Sample.Label`）的支持：发现、过滤、聚合三个层次，并提供跨命令一致的语义与数据结构。

## Requirements

### Requirement: Profile loader 暴露 label 摘要和推断类型

系统 SHALL 提供 `Profile` 包装类型，作为 loader 的返回值。`Profile` 内嵌 `*profile.Profile`，并附加两个一次性计算的字段：
- `LabelSummaries []LabelSummary`：profile 中所有 label 的概览
- `InferredType string`：profile 类型的推断结果

`LabelSummaries` 排序规则：按 distinct value 数降序（最丰富的 label 排前面）。

`InferredType` 与 `inferProfileType(p)` 返回值一致；当 `PeriodType` 缺失时通过 `SampleType` 关键字推断。

#### Scenario: CPU profile 加载后 InferredType 为 "cpu"
- **WHEN** 加载一个 PeriodType 为 cpu 的 profile
- **THEN** `p.InferredType == "cpu"`

#### Scenario: heap profile PeriodType 缺失时 InferredType 仍为 "heap"
- **WHEN** 加载一个 PeriodType 为 nil 但 SampleType 含 `inuse_space` 的 profile
- **THEN** `p.InferredType == "heap"`

#### Scenario: profile 含 label 时 LabelSummaries 非空
- **WHEN** 加载一个 sample 上含 `state` label 的 goroutine profile
- **THEN** `p.LabelSummaries` 含 key="state" 的 LabelSummary

#### Scenario: profile 不含 label 时 LabelSummaries 为空
- **WHEN** 加载一个 sample 上无任何 label 的 CPU profile
- **THEN** `len(p.LabelSummaries) == 0`

### Requirement: LabelSummary 区分 string 和 numeric label

每个 `LabelSummary` SHALL 包含 `Type` 字段，可选值为 `"string"` 或 `"numeric"`。数字 label SHALL 包含 `Unit` 字段（来自 `pprof.Label.NumUnit` 索引的 string table），string label 的 `Unit` 为 `nil`。

#### Scenario: 数字 label 包含 Unit
- **WHEN** profile 含 `cpu` 类型的数字 label，NumUnit 索引指向 "nanoseconds"
- **THEN** LabelSummary 的 `Type == "numeric"` 且 `Unit == "nanoseconds"`

#### Scenario: 字符串 label Unit 为 nil
- **WHEN** profile 含 `state` 类型的字符串 label
- **THEN** LabelSummary 的 `Type == "string"` 且 `Unit == nil`

### Requirement: LabelFilter 解析 --tag key=value flag

`LabelFilter` SHALL 从 `[]string` 类型 flag 值构造（每个元素形如 `key=value`），生成 `map[string][]string` 形式的 focus 和 ignore 集合。

- 同 key 多次出现 → 多个 value 合并（OR）
- value 缺失 `=` 或 key 为空或 value 为空 → 构造时返回错误
- 数字 label 的 value SHALL 是十进制整数字符串（如 `1500000`），不接受浮点（`1.5e6`）、十六进制（`0x1E`）、带分隔符（`1,500,000`）

#### Scenario: 同 key 多次合并为 OR
- **WHEN** 用户传 `--tag state=blocked --tag state=running`
- **THEN** LabelFilter.Focus["state"] == ["blocked", "running"]

#### Scenario: 跨 key 表示 AND
- **WHEN** 用户传 `--tag state=blocked --tag wait_reason=IO`
- **THEN** Focus 包含 state 和 wait_reason 两个 key，匹配时两个 key 的 value 都需满足

#### Scenario: key=value 缺等号报错
- **WHEN** 用户传 `--tag stateblocked`（缺 `=`）
- **THEN** `NewLabelFilter` 返回错误，提示格式应为 `key=value`

#### Scenario: 空 key 或空 value 报错
- **WHEN** 用户传 `--tag =blocked` 或 `--tag state=`
- **THEN** `NewLabelFilter` 返回错误，提示 key 和 value 都必须非空

#### Scenario: 数字 label value 非整数字符串报错
- **WHEN** 用户传 `--tag cpu=1.5e6`（浮点）
- **THEN** `NewLabelFilter` 返回错误，提示 numeric label value 必须是十进制整数

### Requirement: LabelFilter.Apply 过滤 sample 集合

`LabelFilter.Apply(p *Profile) (*Profile, error)` SHALL 按以下语义过滤 `p.Sample`：

- focus 条件：跨 key AND，**同 key OR**。sample 必须对每个 focus key 至少匹配一个 value，否则不通过
- ignore 条件：跨 key AND，**同 key OR**。sample 对每个 ignore key 必须不匹配任何 value，或者 sample 根本不包含该 key（vacuously true，sample 不含该 key 不算"命中 ignore"）
- 返回的新 Profile 是 p 的浅拷贝：`LabelSummaries` 和 `InferredType` 直接复用，**`Sample` 切片替换为新切片**（不修改原 p）
- matched == 0 时返回错误（见 0 样本错误要求）

#### Scenario: 单 focus 条件过滤
- **WHEN** 用户调 `analyze goroutine.pb.gz --tag state=blocked` 且 profile 中 1000 个 sample 有 300 个 state=blocked
- **THEN** filter 后 Profile 的 `len(Sample) == 300`，其余 700 个被排除

#### Scenario: focus key 在 sample 中不存在时不通过
- **WHEN** sample 不含 `state` label，但用户传 `--tag state=blocked`
- **THEN** 该 sample 不通过 filter（focus key 必须存在）

#### Scenario: focus + ignore 组合
- **WHEN** focus=state=blocked, ignore=wait_reason=IO
- **THEN** 保留 sample：state=blocked 且 wait_reason != IO

#### Scenario: ignore key 在 sample 中不存在时 vacuously true
- **WHEN** sample 不含 `wait_reason` label，用户传 `--tag-ignore wait_reason=IO`
- **THEN** 该 sample 通过 filter（sample 不含此 key 自然不可能"被 ignore 命中"）

#### Scenario: 数字 label 字符串匹配
- **WHEN** profile 中 sample 的 label `cpu=1500000`（Label.Num=1500000）
- **THEN** `labelValueToString` 返回字符串 `"1500000"`
- **AND** 用户传 `--tag cpu=1500000` 时该 sample 通过 filter

#### Scenario: 数字 label 字符串不匹配
- **WHEN** sample 的 label `cpu=1500000`，用户传 `--tag cpu=1500001`
- **THEN** 该 sample 不通过（字符串 `"1500000" != "1500001"`）

#### Scenario: 数字 label value 非整数不匹配
- **WHEN** sample 的 label `cpu=1500000`，用户传 `--tag cpu=1500000.0`（虽然 `NewLabelFilter` 在解析时已报错，所以该 scenario 实际不会发生；如绕过构造直接传入 filter，`labelValueToString` 仍产生 `"1500000"`，字符串比对不通过）

### Requirement: 0 样本匹配时报错并附 hint

`LabelFilter.Apply` 在 matched == 0 时 SHALL 返回错误，错误信息 SHALL 包含：
- 原始 sample 总数
- 未带任何 label 的 sample 数
- 提示检查 key/value 拼写或使用含 label 的 profile

错误信息 SHALL 在不同情况下用相同模板，仅替换数字部分。

#### Scenario: 全部无 label
- **WHEN** profile 1000 sample 全无 label，用户传 `--tag state=blocked`
- **THEN** 错误信息含 "tag filter matched 0 of 1000 samples" 和 "(1000 samples have no labels)"
- **AND** 含 "check --tag key=value spelling, or use a profile that has labels"

#### Scenario: 部分无 label
- **WHEN** profile 1000 sample 中 500 个无 label、500 个带 state label，但其中 0 个 state=blocked，用户传 `--tag state=blocked`
- **THEN** 错误信息含 "tag filter matched 0 of 1000 samples" 和 "(500 samples have no labels)"

#### Scenario: key 拼写错误
- **WHEN** profile 1000 sample 中 200 个带 state label，用户传 `--tag staet=blocked`（拼写错）
- **THEN** 错误信息含 "check --tag key=value spelling" 提示

### Requirement: Label breakdown 描述函数在各 label value 上的 flat 分布

`ComputeFunctionBreakdowns(p *Profile, hotspots []FunctionInfo, cfg BreakdownConfig) []FunctionLabelBreakdown` SHALL：
- 仅对 hotspots 中排名前 `cfg.Top` 的函数计算 breakdown
- 默认 `cfg.Top == 20`（调用方未指定时）
- 每个函数对 `cfg.Keys` 指定的 label key 计算分布
- 分布按 flat 降序排序
- flat 累加仅考虑直接归到该函数的 sample（不计算 cum）
- v1 breakdown **不计算 cum**：emit 的 entry 中 cum/cum_pct SHALL 为 0（保留字段供 v2 实现，不省略以保持 schema 稳定）
- `cfg.Keys` 为空时返回 `nil`

#### Scenario: 默认 top 20 展开
- **WHEN** 用户跑 `analyze goroutine.pb.gz --tag-breakdown-on state`（不指定 top）
- **THEN** 返回 20 个或更少（若 hotspots 不足 20）个函数的 breakdown

#### Scenario: 自定义 top N
- **WHEN** 用户跑 `analyze goroutine.pb.gz --tag-breakdown-on state --tag-breakdown-top 5`
- **THEN** 返回 5 个或更少函数的 breakdown

#### Scenario: 多个 breakdown key
- **WHEN** 用户跑 `analyze goroutine.pb.gz --tag-breakdown-on state,http.method`
- **THEN** 每个函数的 breakdown 包含 state 和 http.method 两个 key 的分布

#### Scenario: 不指定 breakdown key 时不计算
- **WHEN** 用户跑 `analyze goroutine.pb.gz`（不传 `--tag-breakdown-on`）
- **THEN** `analyze` 输出中 `label_breakdown` 字段为 null/空数组

#### Scenario: flat 累加正确
- **WHEN** 一个函数 10 个 sample 中 6 个 state=blocked（flat=60），4 个 state=running（flat=40）
- **THEN** breakdown 中 state=blocked 的 flat=60, flat_pct=60.00（占该函数 total flat 的 60%）

#### Scenario: breakdown v1 cum 字段为 0
- **WHEN** 用户跑 `analyze goroutine.pb.gz --tag-breakdown-on state`
- **THEN** breakdown entry 中 `cum` 字段存在但恒为 0，`cum_pct` 恒为 0.00

### Requirement: 跨命令 filter 语义一致

`--tag` 和 `--tag-ignore` filter SHALL 在所有支持它的命令（`analyze` / `list` / `traces` / `diff`）中行为完全一致：
- 同 key OR，跨 key AND
- 0 样本报错行为一致
- 数字 label 字符串比对
- 都仅作用于 label 维度；函数名正则排除是 `list` 独有的 `--ignore-function` flag

`diff` 命令对 base 和 target 两个 profile 应用**同一 filter**，再进行 diff 计算。

#### Scenario: diff 两边都 filter
- **WHEN** 用户跑 `diff base.pb.gz target.pb.gz --tag http.status=500`
- **THEN** base 和 target 都先按 `http.status=500` 过滤，再 diff

#### Scenario: diff 中一边 0 样本报错
- **WHEN** base 1000 sample、target 1000 sample 都含 http.status=500，user 改传 `--tag http.status=600`（不存在）
- **THEN** diff 报错（filter 在 base 上就匹配 0）

#### Scenario: 数字 label 语义跨命令一致
- **WHEN** 用户在 `analyze` 和 `traces` 中都用 `--tag cpu=1500000`
- **THEN** 两个命令的过滤结果完全一致

### Requirement: 数字 label value 序列化

`labelValueToString(*profile.Label) string` SHALL 按以下规则：
- 当 `Label.Str != 0` 时返回 `p.StringTable[Label.Str]`（string label）
- 当 `Label.Str == 0 && Label.Num != 0` 时返回 `strconv.FormatInt(Label.Num, 10)`（numeric label）
- 当 `Label.Str == 0 && Label.Num == 0` 时返回空字符串 `""`（边界情况，理论上不会发生）

`p.StringTable` 来自父 `*profile.Profile`。

#### Scenario: string label value
- **WHEN** Label.Str=5, p.StringTable[5]="blocked"
- **THEN** 返回 "blocked"

#### Scenario: numeric label value
- **WHEN** Label.Num=1500000, Label.NumUnit=2（p.StringTable[2]="nanoseconds"）
- **THEN** 返回 "1500000"（不包含 unit，unit 是 LabelSummary 级别的元信息）

#### Scenario: 零值 label
- **WHEN** Label.Str=0, Label.Num=0
- **THEN** 返回 ""（理论上 pprof 不会出现这种情况，实现层做防御性处理）
