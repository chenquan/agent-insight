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
| 用户提到 .pb.gz / .prof 文件 | 自动识别为 pprof 文件 |

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

### 调用路径追踪
```
用户说"哪些路径调到了 mallocgc" / "看看调用链"
  │
  1. agent-insight traces profile.pb.gz "runtime.mallocgc" --format json
      → 展示所有经过 mallocgc 的原始调用链
  │
  2. agent-insight tree profile.pb.gz --focus "runtime.mallocgc"
      → 在调用树中定位该函数的层级位置
```

### 版本对比
```
用户说"升级后变慢了" / "想对比性能"
  │
  1. agent-insight diff base.pb.gz target.pb.gz --format json
      → 找到回归和改进的函数
```

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
| `function_name` | 函数名 |
| `file` | 源文件路径 |
| `line` | 行号 |

### diff JSON 输出字段

| 字段 | 说明 |
|------|------|
| `delta_flat` | flat 值变化量 |
| `delta_cum` | cum 值变化量 |
| `delta_percent` | 变化百分比（正值=回归，负值=改进） |
| `new_functions` | target 中新增的函数 |
| `deleted_functions` | target 中消失的函数 |

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
