# Design: pprof Labels Support

## 1. 架构概览

### 1.1 现状（重构前）

```
loader.LoadFromFile(path) -> *profile.Profile
                                 │
                                 ▼
   ┌──────────┬──────────┬──────────┬──────────┬──────────┐
   ▼          ▼          ▼          ▼          ▼          ▼
 analyze   list     traces      diff       info      diagnose
   │
   └── 每个命令自己调 inferProfileType(p) + 自己抽 labels
```

**问题**：
- label 抽取逻辑在多处重复（计划 5 个命令各一份）
- `InferredType` 也是 `analyze` 和 `diagnose` 各自调 `inferProfileType`
- label 是"附加概念"，不是"一等公民"

### 1.2 重构后

```
loader.LoadFromFile(path) -> *Profile
                              │
                              ├── Raw       *profile.Profile  (内嵌)
                              ├── Labels    []LabelSummary     (一次性计算)
                              └── Type      string             (一次性推断)
                                 │
                                 ▼
   ┌──────────┬──────────┬──────────┬──────────┬──────────┐
   ▼          ▼          ▼          ▼          ▼          ▼
 analyze   list     traces      diff       info      diagnose
   │          │          │          │          │          │
   └── p.Labels / p.Type 直接读，不重复计算 ──────────┘
```

**收益**：
- label 抽取 1 份（在 loader）
- Type 推断 1 份（在 loader）
- `info` 自动显示 label 摘要
- 所有命令都能在加 filter / breakdown 时直接消费

### 1.3 Profile 包装类型

```go
// pkg/profile/profile.go
type Profile struct {
    *profile.Profile             // 内嵌，命令仍可访问 .Sample / .Location 等
    LabelSummaries []LabelSummary // 所有 label 的概览
    InferredType   string         // 推断的 profile 类型
}

// 构造
func NewProfile(p *profile.Profile) *Profile {
    return &Profile{
        Profile:        p,
        LabelSummaries: ExtractLabelSummaries(p),
        InferredType:   inferProfileType(p),
    }
}
```

`LabelSummaries` 排序规则：按 distinct value 数降序（最丰富的 label 排前面，对 AI 最有信息量）。

## 2. 数据模型

### 2.1 Label Summary（profile 级别）

```go
// pkg/profile/labels.go
type LabelSummary struct {
    Key      string  // e.g. "state", "http.method", "wait_reason"
    Type     string  // "string" | "numeric"
    Unit     *string // 数字 label 的单位（如 "nanoseconds"），string label 为 nil
    Values   []LabelValueSummary // 排序：按 Count 降序
    Distinct int     // 不同 value 总数（截断时 > len(Values)）
}

type LabelValueSummary struct {
    Value   string  // 数字 label 序列化为 "1234" 或 "1.5k"
    Count   int64   // 出现次数（sample 数）
    Percent float64 // 占总样本百分比
}
```

### 2.2 Label Filter（命令级别）

```go
// pkg/profile/labels.go
type LabelFilter struct {
    Focus  map[string][]string  // key -> []value (同 key OR)
    Ignore map[string][]string  // key -> []value (同 key OR)
}

// 构造（从 cobra flag 解析）
func NewLabelFilter(focusFlags, ignoreFlags []string) (*LabelFilter, error) {
    // --tag key=value 解析成 map[key] = append(map[key], value)
    // 同 key 多次 → 多个 value（OR）
    // value 缺失 '=' 或 key 空 → 报错
}

// 应用（对 profile.Sample 做过滤）
func (f *LabelFilter) Apply(p *Profile) (*Profile, error) {
    if f == nil || (len(f.Focus) == 0 && len(f.Ignore) == 0) {
        return p, nil
    }
    filtered := make([]*profile.Sample, 0, len(p.Sample))
    matched := 0
    unlabeled := 0
    for _, s := range p.Sample {
        if !f.matches(s.Label) {
            unlabeled++
            continue
        }
        filtered = append(filtered, s)
        matched++
    }
    if matched == 0 {
        return nil, fmt.Errorf(
            "tag filter matched 0 of %d samples (%d samples have no labels); "+
            "check --tag key=value spelling, or use a profile that has labels",
            len(p.Sample), unlabeled)
    }
    // 复制 profile 结构但替换 sample
    newP := *p.Profile
    newP.Sample = filtered
    return &Profile{
        Profile:        &newP,
        LabelSummaries: p.LabelSummaries, // 复用，不重算
        InferredType:   p.InferredType,
    }, nil
}

// matches 单个 sample 是否通过 filter
func (f *LabelFilter) matches(labels []*profile.Label) bool {
    // 提取 sample 上的 label map
    m := make(map[string][]string, len(labels))
    for _, l := range labels {
        m[l.Key] = append(m[l.Key], labelValueToString(l))
    }
    // Focus 全部满足（AND across keys）
    for k, vs := range f.Focus {
        sampleVals, ok := m[k]
        if !ok {
            return false
        }
        // 至少一个 value 匹配（OR within key）
        if !anyMatch(vs, sampleVals) {
            return false
        }
    }
    // Ignore 全部不满足
    for k, vs := range f.Ignore {
        sampleVals, ok := m[k]
        if !ok {
            continue
        }
        if anyMatch(vs, sampleVals) {
            return false
        }
    }
    return true
}
```

### 2.3 Label Breakdown（函数级别）

```go
// pkg/profile/labels.go
type FunctionLabelBreakdown struct {
    Function FunctionInfo             // 复用 FunctionInfo
    Labels   []LabelBreakdown         // 只包含用户指定的 key
}

type LabelBreakdown struct {
    Key    string
    Values []LabelValueContribution
}

type LabelValueContribution struct {
    Value   string
    Flat    int64
    FlatPct float64   // 占该函数 total flat 的百分比
    Cum     int64
    CumPct  float64
}

type BreakdownConfig struct {
    Keys []string // 用户指定的 key（--tag-breakdown-on）
    Top  int      // 展开几个函数（--tag-breakdown-top），默认 20
}

// 在 Analyze 流程中计算
func ComputeFunctionBreakdowns(p *Profile, hotspots []FunctionInfo, cfg BreakdownConfig) []FunctionLabelBreakdown {
    if len(cfg.Keys) == 0 {
        return nil
    }
    // 取 hotspots 的 top cfg.Top
    n := cfg.Top
    if n > len(hotspots) {
        n = len(hotspots)
    }
    top := hotspots[:n]
    
    // 为每个 top 函数遍历所有 sample 累加 label value 分布
    result := make([]FunctionLabelBreakdown, 0, n)
    for _, fn := range top {
        bd := FunctionLabelBreakdown{Function: fn, Labels: make([]LabelBreakdown, 0, len(cfg.Keys))}
        // ... 计算每个 key 的 value 分布
        result = append(result, bd)
    }
    return result
}
```

## 3. 命令级设计

### 3.1 tags 命令

```
Usage: agent-insight tags <profile.pb.gz> [flags]

Flags:
  --top N           限制每个 label value 显示数量（默认 50）
  --format FORMAT   text|json|markdown（默认 text）
```

**输出结构**（text）：

```
Profile: cpu.pb.gz
Type: cpu
Total samples: 10,000

Labels (3 keys, 247 unique values):

  state (string, 4 values)
    blocked        3,200  32.00%
    syscall        4,891  48.91%
    running        1,247  12.47%
    preempted        730   7.30%

  wait_reason (string, 2 values)
    IO             3,000  30.00%
    channel        2,621  26.21%

  cpu (numeric, nanoseconds, 8,521 unique values)
    [showing top 50 of 8,521]
    1500000         12     0.12%
    1200000         11     0.11%
    ...
```

**JSON 输出**：

```json
{
  "profile_path": "cpu.pb.gz",
  "type": "cpu",
  "total_samples": 10000,
  "labels": [
    {
      "key": "state",
      "type": "string",
      "unit": null,
      "distinct": 4,
      "values": [
        {"value": "syscall", "count": 4891, "percent": 48.91},
        {"value": "blocked", "count": 3200, "percent": 32.00}
      ]
    },
    {
      "key": "cpu",
      "type": "numeric",
      "unit": "nanoseconds",
      "distinct": 8521,
      "values_truncated": true,
      "values": [...]
    }
  ]
}
```

### 3.2 --tag / --tag-ignore 统一语义

**所有支持的命令**：`analyze` / `list` / `traces` / `diff`

**解析规则**：
- `key=value` 必填，无 `=` 报错
- 同 key 多次 → OR
- 跨 key → AND
- 数字 label 的 value 字符串比对（`--tag cpu=1500000` 匹配 `Label.Num == 1500000`）

**应用时机**：
- `analyze` / `traces` / `list`：在 loader 之后立即 filter，结果作为 analyze 的输入
- `diff`：先 filter base，再 filter target，再 diff（同一 filter 应用于两边）

**0 样本处理**：
- 报错 `tag filter matched 0 of N samples`
- 错误信息附 hint：检查 key/value 拼写、profile 是否含 label

### 3.3 analyze 的 breakdown 行为

```
analyze cpu.pb.gz --tag-breakdown-on state,http.method --tag-breakdown-top 10
```

**输出（JSON）**：

```json
{
  "top": 10,
  "functions": [
    {
      "function": "database/sql.(*DB).QueryContext",
      "flat": 1234,
      "cum": 5678,
      "flat_percent": 12.34,
      "cum_percent": 56.78,
      "label_breakdown": [
        {
          "key": "state",
          "values": [
            {"value": "blocked", "flat": 740, "flat_pct": 60.0, "cum": 3400, "cum_pct": 59.9},
            {"value": "running", "flat": 494, "flat_pct": 40.0, "cum": 2278, "cum_pct": 40.1}
          ]
        },
        {
          "key": "http.method",
          "values": [
            {"value": "GET",  "flat": 370, "flat_pct": 30.0, ...},
            {"value": "POST", "flat": 617, "flat_pct": 50.0, ...},
            {"value": "PUT",  "flat": 247, "flat_pct": 20.0, ...}
          ]
        }
      ]
    }
  ]
}
```

**text 格式**：

```
database/sql.(*DB).QueryContext  flat: 1234 (12.34%)  cum: 5678 (56.78%)
  state:
    blocked    flat:  740 (60.00%)  cum: 3400 (59.86%)
    running    flat:  494 (40.00%)  cum: 2278 (40.11%)
  http.method:
    GET        flat:  370 (30.00%)  cum: 1700 (29.94%)
    POST       flat:  617 (50.00%)  cum: 2850 (50.18%)
    PUT        flat:  247 (20.00%)  cum: 1128 (19.86%)
```

### 3.4 info 的 label 摘要

**输出（text 现有格式追加）**：

```
Profile: cpu.pb.gz
Type: cpu
Duration: 30.12s
Period: 100ms
Samples: 10000
Value types: cpu (nanoseconds)
Has symbols: true
Has file lines: false
Labels: 3 keys, 247 unique values   ← 新增
Mappings:
  ...
```

**JSON 新字段**：

```json
{
  "type": "cpu",
  "duration": 30120000000,
  "samples": 10000,
  "value_types": [...],
  "has_symbols": true,
  "label_summary": {
    "key_count": 3,
    "distinct_values": 247
  },
  "mappings": [...]
}
```

### 3.5 list 的 --exclude → --ignore-function 重命名

`list` 命令独有 `--ignore-function` flag 用于函数名正则排除。新加的 `--tag-ignore key=value` 是 label 维度过滤，**不** 兼做函数名排除。两个 flag 完全独立，靠 cobra 自动识别。

**Before**：
```bash
agent-insight list cpu.pb.gz "main.*" --exclude "database.*"
```

**After**（两个独立 flag）：
```bash
# 函数名正则排除（仅 list）
agent-insight list cpu.pb.gz "main.*" --ignore-function "database.*"

# label 过滤（与 analyze / traces / diff 一致）
agent-insight list cpu.pb.gz "main.*" --tag state=blocked
agent-insight list cpu.pb.gz "main.*" --tag-ignore state=running
```

**注意**：原 `--exclude` flag 直接删除，迁移路径在 README 注明。

## 4. 关键算法

### 4.1 LabelSummaries 提取

```go
func ExtractLabelSummaries(p *profile.Profile) []LabelSummary {
    // 1. 遍历所有 sample，收集 label key 集合 + 统计
    //    map[key] -> map[value] -> count
    // 2. 区分 string label 和 numeric label：
    //    profile.Label.Str != 0 → string
    //    profile.Label.Num != 0 → numeric
    // 3. 数字 label 提取 unit（profile.Label.NumUnit 索引到 p.StringTable）
    // 4. 排序：先 string 后 numeric，每个 label 内 value 按 count 降序
    // 5. LabelSummary.Values 默认截断到 top 50，Distinct 字段记总数
}
```

### 4.2 ApplyLabelFilter

```go
func (f *LabelFilter) Apply(p *Profile) (*Profile, error) {
    // 1. 空 filter 直接返回
    // 2. 遍历 p.Sample，调用 f.matches 过滤
    // 3. matched == 0 → 报错
    // 4. matched > 0 → 复制 profile 结构 + 替换 Sample
    // 5. 复用 p.LabelSummaries（filter 不改变 label 集合的存在性，只改变 sample 计数）
    //    但 DisplayCounts 变了 —— 决策：filter 后的 LabelSummary 仍反映原 profile 的分布
    //    若用户想看 filter 后的分布，单独跑 tags 命令不带 filter
}
```

**关于 LabelSummary 是否重算**：

- filter 后再算 LabelSummary 的开销不大（O(n)）
- 但对"tags"命令 + filter 的组合可能有用
- 决策：**v1 不支持 filter 后的 tags 命令**——`tags` 命令反映 profile 原始 label 分布，不接受 filter flag
- filter 后的 label 分布用户可手算：跑 `tags` 看全集，跑 `analyze --tag X=Y` 做过滤

### 4.3 ComputeFunctionBreakdowns

```go
func ComputeFunctionBreakdowns(p *Profile, hotspots []FunctionInfo, cfg BreakdownConfig) []FunctionLabelBreakdown {
    if len(cfg.Keys) == 0 {
        return nil
    }
    n := cfg.Top
    if n <= 0 { n = 20 }
    if n > len(hotspots) { n = len(hotspots) }
    top := hotspots[:n]
    
    // 1. 准备 FunctionKey → idx 的 map
    keyToIdx := make(map[string]int, len(top))
    for i, fn := range top {
        keyToIdx[functionKey(fn)] = i
    }
    
    // 2. 遍历所有 sample：找到该 sample 落到的 top 函数 + 累加 label value
    accum := make(map[int]map[string]map[string]*valueAcc, n) // function idx -> key -> value -> acc
    for _, s := range p.Sample {
        // flat 函数：s.Location[0]
        flatFn := functionAt(s, 0)
        if idx, ok := keyToIdx[flatFn.key]; ok {
            addContribution(accum[idx], s.Label, s.Value[0], cfg.Keys)
        }
    }
    
    // 3. 构造 FunctionLabelBreakdown
    result := make([]FunctionLabelBreakdown, 0, n)
    for i, fn := range top {
        bd := FunctionLabelBreakdown{Function: fn}
        for _, key := range cfg.Keys {
            labelBD := LabelBreakdown{Key: key}
            // 把 accum[i][key] 转成 sorted []LabelValueContribution
            result = append(result, bd)
        }
    }
    return result
}
```

**关于 flat vs cum**：

- v1 breakdown 只算 flat（直接归到该函数的 sample）
- cum breakdown 需要遍历整条 stack 计算哪些 stack 路径经过该函数，对 top-N 函数成本高
- AI 工作流中 flat 分布是更直接的"这个函数时间花在哪"
- 决策：v1 只算 flat，cum 留待 v2

### 4.4 functionKey 唯一标识

```go
// 一个函数可能被多个 location 引用（同名但行号不同）
// 用 Function.Name + Function.File + Function.StartLine 作 key
func functionKey(fn FunctionInfo) string {
    if fn.Function != nil {
        return *fn.Function + ":" + stringOrEmpty(fn.File) + ":" + strconv.FormatInt(fn.StartLine, 10)
    }
    return fmt.Sprintf("loc:%d:%s", *fn.LocationID, *fn.Address)
}
```

## 5. 测试策略

### 5.1 单元测试

| 文件 | 覆盖 |
|------|------|
| `pkg/profile/labels_test.go` | `LabelFilter.Apply` 各种组合（focus / ignore / mixed / 0 样本 / 数字 label）|
| `pkg/profile/labels_test.go` | `ExtractLabelSummaries` 各种 profile 类型（无 label / string / numeric / 混合）|
| `pkg/profile/labels_test.go` | `ComputeFunctionBreakdowns` 边界（0 keys / 0 hotspots / 超大 N）|
| `pkg/profile/tags_test.go` | `Tags` 在各类型 profile 上的输出 |
| `pkg/commands/tags_test.go` | cobra 参数解析 + 集成测试 |

### 5.2 测试数据

`testdata/cpu.pb.gz` 和 `testdata/heap.pb.gz` 来自 `testdata/generate.go`，**不带 label**。

需要新增 `testdata/goroutine.pb.gz`：含 `state` / `wait_reason` 标签的 goroutine profile。

两种生成方式：
1. **从 `github.com/google/pprof` 仓库拿现成的**（推荐，官方数据稳定）
2. **在 `testdata/generate.go` 里加 case**（用 `profile.Sample{Label: ...}` 构造）

### 5.3 集成测试

```go
// pkg/profile/labels_test.go
func TestIntegrationGoroutineLabels(t *testing.T) {
    p, _ := loader.LoadFromFile("testdata/goroutine.pb.gz")
    
    // 1. 提取 label summary
    summaries := p.LabelSummaries
    require.Contains(t, summaryKeys(summaries), "state")
    
    // 2. filter state=blocked
    filter, _ := NewLabelFilter([]string{"state=blocked"}, nil)
    filtered, _ := filter.Apply(p)
    assert.Greater(t, len(filtered.Sample), 0)
    assert.Less(t, len(filtered.Sample), len(p.Sample))
    
    // 3. 同 key OR
    filter, _ = NewLabelFilter([]string{"state=blocked", "state=running"}, nil)
    filtered, _ = filter.Apply(p)
    assert.Greater(t, len(filtered.Sample), 0)
    
    // 4. 跨 key AND
    filter, _ = NewLabelFilter([]string{"state=blocked", "wait_reason=IO"}, nil)
    filtered, err := filter.Apply(p)
    require.NoError(t, err)
    assert.Greater(t, len(filtered.Sample), 0)
    
    // 5. 0 样本
    filter, _ = NewLabelFilter([]string{"state=nonexistent"}, nil)
    _, err = filter.Apply(p)
    require.Error(t, err)
}
```

## 6. 迁移与破坏性变更

### 6.1 Loader API 变更

```go
// Before
func (l *Loader) LoadFromFile(path string) (*profile.Profile, error)

// After
func (l *Loader) LoadFromFile(path string) (*Profile, error)
```

调用方 6 个文件需改：`pkg/profile/{analysis,diff,info,traces,list,diagnose}.go`

**机械改动**：
- `*profile.Profile` → `*Profile`
- `p.Sample` 等字段访问不变（内嵌）
- `inferProfileType(p)` → `p.InferredType`
- label 抽取 → `p.LabelSummaries`

### 6.2 list flag 改名

`pkg/commands/list.go` 中 `listExclude` 变量和 `--exclude` flag 改为 `listIgnoreFunction` 和 `--ignore-function`。**不**复用到 `--tag-ignore`，那是另一个独立 flag（label ignore）。

README 注明迁移：
> **BREAKING**: `--exclude` 在 v0.X 中被 `--ignore-function` 替代。语义不变（正则排除函数），仅改名。需更新脚本中的 flag。同时新增 `--tag-ignore key=value` 用于 label 过滤，与 analyze / traces / diff 中同名 flag 行为一致。

### 6.3 风险评估

- **Loader 重构**：影响所有命令，单元测试需要同步更新。机械改动，单测覆盖良好，风险可控。
- **flag 改名**：只有 `list` 命令的 `--exclude` 改为 `--ignore-function`，文档同步。`--tag-ignore` 是新加 flag（label 维度），与改名前的 `--exclude` 完全独立。
- **新数据字段**：所有 Result 增字段，向后兼容（旧消费者忽略未知字段）。
- **JSON schema**：tags / analyze 的 breakdown / info 的 label_summary 都是新字段，AI 解析 schema 升级是消费方的工作。
