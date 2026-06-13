## Context

agent-insight 当前有三层架构（commands / profile / output），支持 analyze、list、flame、diff、merge 等命令。`diff` 命令能对比两个 profile 的函数级差异，但无法处理多 profile 时间序列。用户需要手动两两 diff 来追踪趋势，缺乏全局视角和趋势量化能力。

当前 `merge` 命令已有 `discoverProfiles` 函数用于目录扫描，可复用。

## Goals / Non-Goals

**Goals:**
- 新增 `trend` 命令，接受 3 个以上同类型 pprof 文件，输出函数级性能趋势
- 通过线性回归量化趋势方向和速率
- 四层过滤机制控制输出量，默认输出精简
- 三种输出格式（text/json/markdown），JSON 对 AI 解析最优

**Non-Goals:**
- 不做交互式可视化（如 Web UI、实时图表）
- 不做自动按子目录分组 merge（多 profile 同时间点场景需先手动 merge）
- 不做文件名时间戳解析（依赖 mtime + `--sort-by` flag）
- 不做统计检验（如 Mann-Kendall），线性回归足够
- N < 3 时直接报错，不退化成 diff

## Decisions

### D1: 输入方式 — 目录扫描 + 显式文件列表

复用 merge 的 `discoverProfiles` 模式处理目录，同时支持显式列出文件。时间排序默认使用文件 mtime，提供 `--sort-by name|mtime` flag 覆盖。

**替代方案**：glob 模式（如 `builds/*/cpu.prof`）——需要 shell 展开，增加复杂性，MVP 不需要。

### D2: 最少 3 个 profile

N < 3 时报错并提示使用 `diff`。2 个点的"趋势"没有统计意义，与 diff 职责重叠。

**替代方案**：N >= 2 也运行，N < 3 不算 slope 只算 delta——输出语义不清晰，不如严格分界。

### D3: 趋势计算 — 线性回归

对每个函数的 flat/cumulative 值序列做最小二乘线性回归，得到 slope（变化速率）和 trend 方向（regressing/improving/stable）。分类公式：`|slope / avg| * 100 > threshold` 时判定为 regressing 或 improving，否则为 stable。`--threshold` 默认值 5（即平均值的 5%）。

**替代方案**：Mann-Kendall 非参数检验——对小样本（N < 10）优势不明显，增加实现复杂性。

### D4: 缺失值处理 — null 标记

函数在某些时间点不存在时，JSON 输出用 `null`，text 输出用 `-`。线性回归时跳过 null 点。

**替代方案**：填 0——会扭曲趋势（"不存在" ≠ "消耗为 0"）；排除缺失函数——会漏掉新增热点。

### D5: 四层过滤架构

- L0: `--focus` / `--ignore`（复用 diff 的 regex 模式）
- L1: `--min-impact`（默认 1，函数在任意时间点的 `flat_value / total_samples_at_that_point * 100` 需超过阈值；设为 0 时不过滤）
- L2: `--threshold`（趋势斜率阈值，决定 stable 的分界）
- L3: `--top N`（默认 10，每类输出数量限制）

min-impact 是新增的关键过滤层：过滤掉始终占比极小的函数噪音。

### D6: 五段式输出结构

1. Summary：全局走势概览
2. Top Regressions：slope 最大的回归函数
3. Top Improvements：slope 最大的改善函数
4. New Hotspots（`--include-new`）：中途出现且增长的函数
5. Volatile（`--include-volatile`）：slope ≈ 0 但波动大的函数

Section 4/5 默认不输出，按需开启。

### D7: 新增热点检测逻辑

首次出现时间点 > 总时间跨度 30%，且最终占比 > `min-impact`，按最终占比排序。简单启发式规则，对 AI 定位问题足够。

### D8: 数据结构

`TrendResult` 包含：时间点列表、全局走势、函数趋势数组（含完整序列和统计指标）、预排序的回归/改善列表。每个函数趋势包含：flat/cum 序列、slope、trend 方向、峰谷值、变异系数。

### D9: Trend 函数签名

遵循三层分离，`profile.Trend` 只接收已加载的数据，不做文件 I/O。`TimePoint` 携带标签和时间戳信息。

```go
type TimePoint struct {
    Label string  // 文件名或用户指定标签
    Time  int64   // Unix timestamp (mtime)，用于排序
}

func Trend(profiles []*profile.Profile, timePoints []TimePoint, config TrendConfig) (*TrendResult, error)
```

commands 层负责：发现文件、读取 mtime、排序、加载 profile、构造 `TimePoint` 列表，然后调用 `profile.Trend`。

## Risks / Trade-offs

- **[大量 profile 时内存占用]** → 每个 profile 需加载到内存计算函数值 map，10+ profile 可能占用较多内存。MVP 阶段可接受，后续可考虑流式处理。
- **[mtime 不准确]** → 文件被拷贝或 tar 打包后 mtime 可能改变。提供 `--sort-by name` 作为 fallback。
- **[线性回归对小样本敏感]** → N=3 时，1 个异常点就能显著改变 slope。接受此限制，在输出中附带原始序列让 AI 自行判断。
- **[函数匹配跨 profile 不精确]** → pprof 的 Location ID 在不同编译产物间不一致（无符号时尤其明显）。和 diff 一致，用函数名/地址匹配。
