---
name: agent-insight
description: "使用 agent-insight 分析 pprof 性能 profile 文件。TRIGGER 当: 用户提到 pprof、性能分析、CPU profiling、内存泄漏、heap 分析、性能热点、火焰图、.pb.gz 文件、goroutine 泄漏、延迟分析、概况、元信息、调用链、执行路径、调用树、层级结构、或想对比 profile 性能差异。"
---

agent-insight 是一个轻量级 pprof 分析 CLI 工具，专为 AI 编码助手设计。单二进制、零外部依赖，输出结构化 JSON 结果。

## 何时使用

| 触发场景 | 推荐命令 |
|----------|----------|
| 用户拿到 profile 想先了解概况 | `agent-insight info profile.pb.gz` |
| 用户提到 CPU 性能/慢/延迟 | `agent-insight analyze cpu.pb.gz` |
| 用户提到内存泄漏/OOM | `agent-insight analyze heap.pb.gz --value-type alloc_objects` |
| 用户问"谁调用了 X 函数" | `agent-insight list profile.pb.gz "pattern"` |
| 用户想看火焰图 | `agent-insight flame profile.pb.gz` |
| 用户想看调用链/执行路径 | `agent-insight traces profile.pb.gz "pattern"` |
| 用户想看调用树/层级结构 | `agent-insight tree profile.pb.gz` |
| 用户想对比版本间性能变化 | `agent-insight diff base.prof target.prof` |
| 用户想追踪多版本性能趋势 | `agent-insight trend profiles/ --format json` |
| 用户有多个同类 profile 需要合并 | `agent-insight merge profiles/ -o merged.pb.gz` |
| 用户提到 .pb.gz / .prof 文件 | 自动识别为 pprof 文件 |
| 用户说"帮我看看/诊断一下"但没有具体问题 | `agent-insight diagnose profile.pb.gz` |

## 命令速查

所有命令支持 `--format json` 输出 JSON 格式（推荐用于解析）。

### info — Profile 元信息概览

```bash
agent-insight info <profile.pb.gz> [flags]
```

| Flag | 默认值 | 说明 |
|------|--------|------|
| `--format` | text | 输出格式: text, json, markdown |

零计算开销，快速查看 profile 类型、时长、采样数、值类型、符号状态、映射信息。

示例:
```bash
agent-insight info cpu.pb.gz --format json
agent-insight info heap.pb.gz
```

### analyze — 分析性能热点

```bash
agent-insight analyze <profile.pb.gz> [flags]
```

| Flag | 默认值 | 说明 |
|------|--------|------|
| `--top N` | 15 | 返回前 N 个热点 |
| `--cum` | false | 按累计值排序（默认按 flat 排序） |
| `--focus pattern` | | 正则过滤关注的函数 |
| `--ignore pattern` | | 正则排除函数 |
| `--format` | text | 输出格式: text, json, markdown |
| `--call-depth N` | 5 | 调用栈深度 |
| `--collapse` | false | 附带折叠栈格式输出 |
| `--value-type` | | 指定分析的值类型（多值 profile） |

示例:
```bash
agent-insight analyze cpu.pb.gz --format json
agent-insight analyze heap.pb.gz --value-type alloc_objects --top 20
agent-insight analyze profile.pb.gz --focus "runtime\." --cum
```

### list — 查询函数调用关系

```bash
agent-insight list <profile.pb.gz> <function-pattern> [flags]
```

| Flag | 默认值 | 说明 |
|------|--------|------|
| `--depth N` | 5 | 调用者/被调用者最大深度 |
| `--callers-only` | false | 只显示调用者 |
| `--callees-only` | false | 只显示被调用者 |
| `--exclude pattern` | | 正则排除函数 |
| `--format` | text | 输出格式 |
| `--value-type` | | 指定值类型 |

示例:
```bash
agent-insight list cpu.pb.gz "runtime.mallocgc" --format json
agent-insight list heap.pb.gz "main\." --callers-only
```

### flame — 生成折叠栈格式

```bash
agent-insight flame <profile.pb.gz> [flags]
```

| Flag | 默认值 | 说明 |
|------|--------|------|
| `--depth N` | 0 (无限) | 最大栈深度 |
| `--focus pattern` | | 正则过滤函数 |
| `--ignore pattern` | | 正则排除函数 |
| `--stats` | false | 包含统计信息 |
| `--top N` | 0 (无限) | 限制前 N 个栈 |
| `--value-type` | | 指定值类型 |

输出格式: `func1;func2;func3 count`，可直接用于火焰图工具。

示例:
```bash
agent-insight flame cpu.pb.gz --stats
agent-insight flame heap.pb.gz --focus "main\." --top 50
```

### traces — 查看原始采样调用链

```bash
agent-insight traces <profile.pb.gz> [flags]
```

| Flag | 默认值 | 说明 |
|------|--------|------|
| `--focus pattern` | | 正则，只展示包含匹配函数的 trace |
| `--ignore pattern` | | 正则，排除包含匹配函数的 trace |
| `--top N` | 20 | 限制输出条数 |
| `--value-type` | | 指定值类型 |
| `--format` | text | 输出格式 |

展示每条采样的完整调用路径和值，与 flame（聚合视图）互补。

示例:
```bash
agent-insight traces cpu.pb.gz --format json
agent-insight traces cpu.pb.gz --focus "runtime.mallocgc" --top 10
```

### tree — 层级调用树

```bash
agent-insight tree <profile.pb.gz> [flags]
```

| Flag | 默认值 | 说明 |
|------|--------|------|
| `--focus pattern` | | 正则，只展示包含匹配函数的分支 |
| `--ignore pattern` | | 正则，排除包含匹配函数的分支 |
| `--depth N` | 5 | 最大展示深度 |
| `--top N` | 10 | 每层最多展示子节点数 |
| `--cum` | true | 按 cum 排序 |
| `--value-type` | | 指定值类型 |
| `--format` | text | 输出格式 |

从根到叶的全局调用结构视图。

示例:
```bash
agent-insight tree cpu.pb.gz --format json
agent-insight tree heap.pb.gz --depth 3 --top 5
```

### diff — 对比两个 profile

```bash
agent-insight diff <base.prof> <target.prof> [flags]
```

| Flag | 默认值 | 说明 |
|------|--------|------|
| `--min-delta N` | 0 | 最小变化百分比阈值 |
| `--focus pattern` | | 正则过滤函数 |
| `--ignore pattern` | | 正则排除函数 |
| `--format` | text | 输出格式 |
| `--hide-new` | false | 隐藏新增函数 |
| `--hide-deleted` | false | 隐藏删除函数 |
| `--top N` | 15 | 每类限制前 N 个 |
| `--value-type` | | 指定值类型 |

示例:
```bash
agent-insight diff base.pb.gz target.pb.gz --format json
agent-insight diff v1.pb.gz v2.pb.gz --min-delta 5 --top 10
```

### merge — 合并多个同类 profile

```bash
agent-insight merge <profile...> -o <output> [flags]
```

| Flag | 默认值 | 说明 |
|------|--------|------|
| `-o, --output` | (必需) | 输出文件路径 |

将多个同类型的 pprof profile 合并为一个 `.pb.gz` 文件。支持指定文件列表或目录（递归扫描）。合并后的文件可直接用于 analyze/diff/flame 等命令。

示例:
```bash
agent-insight merge cpu1.pb.gz cpu2.pb.gz cpu3.pb.gz -o merged.pb.gz
agent-insight merge ./profiles/ -o merged.pb.gz
```

### trend — 性能趋势分析

```bash
agent-insight trend <path...> [flags]
```

| Flag | 默认值 | 说明 |
|------|--------|------|
| `--format` | text | 输出格式: text, json, markdown |
| `--focus pattern` | | 正则过滤关注的函数 |
| `--ignore pattern` | | 正则排除函数 |
| `--min-impact N` | 1 | 最小影响百分比阈值（任意时间点 flat 占比） |
| `--threshold N` | 5 | 趋势分类阈值（平均值的百分比） |
| `--top N` | 10 | 每类输出数量限制 |
| `--sort-by` | mtime | 排序方式: mtime, name |
| `--value-type` | | 指定值类型 |
| `--include-new` | false | 包含新增热点 |
| `--include-volatile` | false | 包含高波动函数 |

分析 3 个以上同类型 profile 的性能趋势。支持目录扫描（递归发现 .pb/.pb.gz）或显式文件列表。输出 Top Regressions、Top Improvements，可选 New Hotspots 和 Volatile Functions。

需要至少 3 个 profile，否则建议使用 `diff`。

示例:
```bash
agent-insight trend ./cpu-profiles/ --format json
agent-insight trend p1.pb.gz p2.pb.gz p3.pb.gz --focus "pkg/server.*"
agent-insight trend ./cpu/ --include-new --include-volatile --top 5
```

### diagnose — 生成 AI 诊断提示词

```bash
agent-insight diagnose <profile.pb.gz> [flags]
```

| Flag | 默认值 | 说明 |
|------|--------|------|
| `--top N` | 10 | 热点函数数量 |
| `--context text` | | 用户提供的应用上下文 |
| `--format` | text | 输出格式: text, json, markdown |

分析 profile 后生成结构化诊断提示词，包含语言检测、热点数据、调用树、关键路径和针对性诊断引导。输出给 AI 助手后由其完成诊断。

**何时用**: 用户不确定问题在哪，说"帮我看看/诊断一下"但没有具体焦点 → diagnose 一键自动执行 analyze+tree+traces。
**何时不用**: 用户有具体问题（"X 函数为什么慢" → analyze+list、"谁调用了 X" → list、"对比版本" → diff、"内存泄漏" → analyze --value-type、"火焰图" → flame、"goroutine 泄漏" → traces --focus）。

示例:
```bash
agent-insight diagnose cpu.pb.gz
agent-insight diagnose cpu.pb.gz --context "HTTP microservice" --format json
agent-insight diagnose heap.pb.gz --top 5
```

## 典型工作流

### 快速概览

```
用户说"看看这个 profile" / "这是什么文件"
  │
  1. agent-insight info profile.pb.gz --format json
      → 快速了解 profile 类型、时长、采样数、值类型
  │
  2. （根据 info 结果决定下一步：analyze? tree? traces?）
```

**决策点**:
- `type=cpu` 且 `samples` 较多 → 工作流 2（CPU 性能分析）
- `type=heap` 且有 `alloc_space` 等多个值类型 → 工作流 3（内存分析）
- `type=goroutine` 且 `samples` 巨大 → 怀疑 goroutine 泄漏,用 `traces --focus` 看 top stack
- `value_types` 只有 1 个 → 该 profile 是单值,不需要 `--value-type` 切换
- 用户没有具体问题指向，说"帮我看看"之类 → `diagnose` 一键全面诊断

**陷阱提示**:
- 别只跑 info 就给结论,info 只看元数据,看不到热点分布
- `samples < 100` 时分析结果不稳定,需要更长采集周期
- `has_symbols=false` 是生产环境常见情况,工具会自动 fallback 到 address,不是错误

### CPU 性能分析

```
用户说"服务很慢" / "CPU 占用高"
  │
  1. agent-insight analyze cpu.pb.gz --format json --cum
  │   → 找到累计耗时最多的函数
  │
  2. agent-insight list cpu.pb.gz "<hotspot-func>" --format json
  │   → 查看热点函数的调用者和被调用者
  │
  3. agent-insight flame cpu.pb.gz --focus "<hotspot>"
      → 生成火焰图数据，查看完整调用路径
```

**决策点 (步骤 1)**:
- top 都是 leaf 函数 (`runtime.*` / `syscall.*`) → 该函数是 self-bottleneck,直接跳步骤 3 用 `flame` 看完整栈
- top 是 internal 函数 (业务代码) → 该函数是 caller,不是 self-bottleneck。用 `list <func> --callers-only` 找真正的触发者
- top flat_percent 都很分散 (< 10% 各自) → 没有明显热点,考虑 `--top` 加到 30-50 看更广

**决策点 (步骤 2)**:
- 看到该函数被 N 个不同 caller 调用 → 调用面广,优化此函数收益大
- 看到该函数只被 1-2 个 caller 调用 → 调用面窄,优先看 caller 而不是这个函数
- 看到 callee 是 `runtime.*` → 该函数在频繁触发 runtime,不是 runtime 本身慢

**决策点 (步骤 3)**:
- `flame` 输出中某栈深度突然变宽 → 找到"汇聚点",展开它
- `flame` 输出全是 1-2 层浅栈 → profile 采集在 native 代码或采样周期过粗

**陷阱提示**:
- 别只看 flat 排序,internal 函数的 cum 高但 flat 低,通常是 caller,不是真正的热点
- `runtime.*` 占比高不等于 `runtime` 是问题,它可能只是被某段代码高频触发
- `flame` 输出按 `;` 分隔的栈,深度用 `wc -l` 看采样数,不要逐行计数
- `--cum` 是诊断入口;真正优化要找 flat 高的 leaf

### 内存分析

```
用户说"内存泄漏" / "OOM"
  │
  1. agent-insight analyze heap.pb.gz --format json --value-type alloc_objects
  │   → 找到分配最多的对象
  │
  2. agent-insight analyze heap.pb.gz --format json --value-type inuse_space
  │   → 找到占用最多的内存
  │
  3. agent-insight list heap.pb.gz "<hotspot>" --callers-only --format json
      → 追踪谁在分配这些对象
```

**决策点 (步骤 1)**:
- `alloc_objects` top 是高频小对象 (bytes.Buffer / string concat) → 是分配速率问题,不是泄漏
- `alloc_objects` top 是 `main.*` 大块分配 → 是分配源问题,看调用者
- `alloc_objects` top 是 `runtime.slice` 之类 → 是 GC 压力,看 slice/map 增长

**决策点 (步骤 2)**:
- `inuse_space` 跟 `alloc_space` 排序差异大 → 当前在用的大对象 ≠ 分配最多的大对象,看分配源
- `inuse_space` top 是 `main.makeSlice` 之类 → 累积型大对象,看是否被全局引用
- `inuse_space` 跟 `alloc_space` 几乎一样 → 当前持有的就是分配出来的,看为什么没释放

**决策点 (步骤 3)**:
- `--callers-only` 看到是 `main.handler` → HTTP/GRPC 路径上的分配,看 QPS × 单次分配量
- `--callers-only` 看到是 `init` 函数 → 启动期分配,通常不是泄漏点
- `--callers-only` 看到 N 个 caller 都调 → 公共路径,优化收益大

**陷阱提示**:
- `inuse_objects` 几乎没用,大对象决定内存,不是对象个数
- 看到 `runtime.mallocgc` 占比高 ≠ 内存问题,它是所有分配的"出口"
- Heap profile 4 个值类型语义不同:`alloc_*` 是分配速率,`inuse_*` 是当前持有;不能混用
- 内存"泄漏"在 Go 里通常是全局引用未释放,不是 C 风格的 malloc 忘 free

### 调用路径追踪

```
用户说"哪些路径调到了 mallocgc" / "看看调用链"
  │
  1. agent-insight traces profile.pb.gz --focus "runtime.mallocgc" --format json
      → 展示所有经过 mallocgc 的原始调用链
  │
  2. agent-insight tree profile.pb.gz --focus "runtime.mallocgc"
      → 在调用树中定位该函数的层级位置
```

**决策点 (步骤 1)**:
- `traces` 输出的 `stack` 数组长度差异大 → 不同入口走不同路径,逐一展开
- `traces` 输出的 `percent` 集中在几条 → 优化这几条就能覆盖大部分开销
- `traces` 输出空 → `--focus` 正则太严,放宽或改用 `analyze --focus`

**决策点 (步骤 2)**:
- `tree` 中该函数在第 1-2 层 → 它靠近入口,优化它影响整个程序
- `tree` 中该函数在第 5+ 层 → 它在深路径,先看它的直接 caller
- `tree` 中该函数有多个 parent → 它是多入口共享的,优化收益大

**陷阱提示**:
- `traces` 跟 `flame` 是互补的:`traces` 看每条原始链,`flame` 看聚合;不要混
- `tree` 按 `--cum` 排序时,根节点 cum 总和等于 profile 总量,这是校验 baseline
- `stack` 数组是 `根 → 叶` 顺序(Sample.Location[0] 是 leaf),不是 `叶 → 根`

### 版本对比

```
用户说"升级后变慢了" / "想对比性能"
  │
  1. agent-insight diff base.pb.gz target.pb.gz --format json
      → 找到回归和改进的函数
```

**决策点**:
- `regressions[]` 非空且 `flat_delta_percent > 20` → 明显回归,优先优化
- `regressions[]` 非空但 `flat_delta_percent < 10` → 噪音,叠加更多样本再判
- `improvements[]` 非空 → 优化已生效,继续验证
- `new[]` 非空 → 这次新增的热点,先确认是不是计划内的新代码
- `deleted[]` 非空 → 这次消失的热点,可能是重构掉了或 dead code
- `overall.total_percent > 0` → 整体性能下降,系统级问题
- `overall.total_percent < 0` → 整体性能提升,系统级优化有效

**陷阱提示**:
- `new_functions[]` 非空不一定是 regression,可能只是新代码还没被旧 profile 采到
- `delta_percent` 极小但绝对值大时,小函数的高频调用也可能造成真实问题
- 一定要确认 base 和 target 是同类型 profile (cpu vs cpu),否则 diff 无意义
- 加 `--min-delta 5` 过滤掉 5% 以内的变化,关注显著回归

## 决策树: 看到 X → 跑 Y

| 你看到 | 下一步 |
|--------|--------|
| `info` 显示 `type=cpu` + 大量 samples | `analyze --cum` 找 top 热点 |
| `info` 显示 `type=heap` + 内存增长 | `analyze --value-type alloc_space` 看分配源 |
| `info` 显示 `type=heap` + 当前占用高 | `analyze --value-type inuse_space` 看大对象 |
| `info` 显示 `type=heap` + 大量小对象 | `analyze --value-type alloc_objects` 看分配速率 |
| `info` 显示 `type=goroutine` + count > 10000 | 怀疑泄漏,`traces --focus` 看 top stack |
| `analyze` top 是 leaf 函数 (`runtime.*`) | `flame` 看完整栈,确认是否 self-bottleneck |
| `analyze` top 是 internal 函数 (业务代码) | `list <func> --callers-only` 找真正的触发者 |
| `analyze` top 都很分散 (< 10% 各自) | `--top 50` 加宽,或换 `--cum` 排序 |
| `list` 输出 callee 是 `runtime.*` | runtime 是被高频触发,不是问题源,看 caller |
| `diff` 看到 `new_functions[]` 非空 | `list <new-func>` 验证是不是真"新",还是 base 漏采 |
| `diff` 看到 `regressions[]` 且 `delta_percent > 20` | 重点优化,定位到该函数 |
| 命令报 `no samples matched` | 放宽 `--focus` 正则,或去掉 `--focus` |
| 命令报 `unknown value-type` | `info` 看 `value_types` 列表,用其中之一 |
| 命令报 `failed to load profile` | `info <file>` 验证文件能解析,确认格式 |
| 用户说"帮我看看/诊断一下"但没有具体问题 | `diagnose` 生成诊断提示词,自动执行 analyze+tree+traces |

## 输出解读

### info JSON 输出字段

| 字段 | 说明 |
|------|------|
| `type` | Profile 类型（cpu、heap、goroutine 等） |
| `duration` | 采集时长 |
| `period` | 采样周期 |
| `samples` | 采样总数 |
| `value_types` | 值类型列表（含 type/unit） |
| `has_symbols` | 是否有函数符号 |
| `has_file_lines` | 是否有源码行号 |
| `mappings` | 二进制映射列表（含文件路径、BuildID） |

### traces JSON 输出字段

| 字段 | 说明 |
|------|------|
| `stack` | 调用路径（根→叶的函数名数组） |
| `value` | 该采样的值 |
| `percent` | 占总量百分比 |

### tree JSON 输出字段

| 字段 | 说明 |
|------|------|
| `name` | 函数名 |
| `flat` | 直接归到该函数的值 |
| `flat_percent` | flat 占总量百分比 |
| `cum` | 经过该函数的累计值 |
| `cum_percent` | cum 占总量百分比 |
| `children` | 子节点列表（递归结构） |

### analyze JSON 输出字段

| 字段 | 说明 |
|------|------|
| `flat` | 函数自身耗时/内存（不含子调用） |
| `cum` | 累计耗时/内存（含子调用） |
| `flat_percent` | flat 占总量百分比 |
| `cum_percent` | cum 占总量百分比 |
| `function` | 函数名 |
| `file` | 源文件路径 |
| `line` | 行号 |

### analyze 模式识别 (flat vs cum)

理解 flat 和 cum 的关系是分析的核心。4 种典型模式:

**1. `flat` 高 + `cum` 高 → self-bottleneck**
- 函数本身慢,需要优化函数实现
- 是优化首选目标

```json
{
  "function": "runtime.mallocgc",
  "file": "runtime/malloc.go:1020",
  "flat": 600,
  "flat_percent": 54.55,
  "cum": 600,
  "cum_percent": 54.55
}
```
↑ `flat == cum`,纯 leaf,这是 `runtime.mallocgc` 的 self-bottleneck。

**2. `flat` 低 + `cum` 高 → caller bottleneck**
- 函数被很多 caller 触发,优化此函数影响有限
- 要看 caller,优化触发它的入口

(本仓库 testdata 未覆盖此模式,实际中表现为如 `flat=50, cum=800` 的中间层函数。)

**3. `flat` 高 + `cum` = `flat` → leaf**
- 叶子函数,没被谁调,自己耗时
- 与 self-bottleneck 等价,优先优化

(`runtime.mallocgc` 示例即符合此模式。)

**4. `flat` = 0 + `cum` 高 → pure caller**
- 纯调用方,自己没耗时,但所有耗时都从它过
- 看它的 callee 找真正热点

```json
{
  "function": "main.handleRequest",
  "file": "main.go:42",
  "flat": 0,
  "cum": 950,
  "cum_percent": 86.36,
  "callees": [
    { "Function": "runtime.mallocgc", "FlatValue": 500 },
    { "Function": "encoding/json.Marshal", "FlatValue": 300 },
    { "Function": "io.ReadAll", "FlatValue": 150 }
  ]
}
```
↑ `flat=0, cum=950`,所有 950 个采样都从它的 callees 走,这是 `main.handleRequest` 的 pure caller 模式。

### diff JSON 输出字段

| 字段 | 说明 |
|------|------|
| `delta_flat` | flat 值变化量 |
| `delta_cum` | cum 值变化量 |
| `flat_delta_percent` | flat 变化百分比（正值=回归，负值=改进） |
| `new` | target 中新增的函数 |
| `deleted` | target 中消失的函数 |
| `regressions` | 回归的函数列表 |
| `improvements` | 改进的函数列表 |
| `overall` | 整体变化（base/target 总量、total_percent） |

### diff 模式识别

`diff` 的核心是判断"变化方向"和"变化量",4 种典型解读:

**1. `flat_delta_percent > 0` → regression**

```json
{
  "function": "runtime.mallocgc",
  "file": "runtime/malloc.go:1020",
  "base_flat": 600,
  "target_flat": 900,
  "flat_delta": 300,
  "flat_delta_percent": 50,
  "is_new": false,
  "is_deleted": false
}
```
↑ `flat_delta_percent=50` 表示 target 比 base 慢 50%,这是 regression。

**2. `flat_delta_percent < 0` → improvement**

```json
{
  "function": "runtime.mallocgc",
  "file": "runtime/malloc.go:1020",
  "base_flat": 600,
  "target_flat": 300,
  "flat_delta": -300,
  "flat_delta_percent": -50,
  "is_new": false,
  "is_deleted": false
}
```
↑ `flat_delta_percent=-50` 表示 target 比 base 快 50%,这是 improvement。

**3. `new[]` 非空 → new hotspots**

```json
{
  "function": "main.newHotFunc",
  "file": "main.go:99",
  "base_flat": 0,
  "target_flat": 200,
  "flat_delta": 200,
  "is_new": true,
  "is_deleted": false
}
```
↑ base 中没有,target 中出现 200 个采样,这是新出现的热点。

**4. `deleted[]` 非空 → gone hotspots**

```json
{
  "function": "encoding/json.Marshal",
  "file": "encoding/json/encode.go:160",
  "base_flat": 300,
  "target_flat": 0,
  "flat_delta": -300,
  "flat_delta_percent": -100,
  "is_new": false,
  "is_deleted": true
}
```
↑ base 中有 300 个采样,target 中消失,这是 gone hotspot (可能是重构或被替换)。

### list JSON 输出字段

| 字段 | 说明 |
|------|------|
| `callers` | 调用目标函数的函数列表（含 flat/cum） |
| `callees` | 目标函数调用的函数列表（含 flat/cum） |
| `self` | 目标函数自身的 flat/cum |

## 注意事项

- 始终优先使用 `--format json` 以获得结构化输出
- 分析前可先用 `info` 了解 profile 概况，再选择合适的分析命令和参数
- CPU profile 用 `--cum` 找到根因，`--cum=false`（默认）找到直接热点
- Heap profile 有多个值类型，用 `--value-type` 指定：`alloc_objects`、`alloc_space`、`inuse_objects`、`inuse_space`
- 生产环境可能缺少 debug symbols，工具会自动降级显示地址信息
- 面对典型工作流时,先按"决策点"分支选择命令;看到未知错误先看"决策树"段
- flat 高 + cum 高是 self-bottleneck (优化目标),flat=0 + cum 高是 pure caller (看 callee)
