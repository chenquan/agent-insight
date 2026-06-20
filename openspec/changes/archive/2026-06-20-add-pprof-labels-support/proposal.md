# Add pprof Labels Support

## Why

`agent-insight` 当前完全没暴露 pprof profile 上的 label（`Sample.Label`）信息。`go tool pprof` 早就有 `tagfocus` / `tagignore` / `tagshow` 这套机制，对 AI 工作流而言这是巨大缺失：

- **goroutine profile** 上 `state` / `wait_reason` label 能直接区分阻塞态 vs 运行态采样；现在 AI 拿到的 samples 是混在一起的
- **service profile**（持续剖析）上 `http.method` / `http.status` / `http.path` / `thread.id` 等 label 决定"哪个端点慢"；现在根本看不到
- **block / mutex profile** 上的 `wait_type` / `block_reason` 同样是关键维度

更关键的是，**label breakdown**（一个函数在每个 label value 上分别占多少耗时）是 `go tool pprof` 都没原生提供的能力。这正是 agent-insight 面向 AI 优化的差异化点——直接告诉 LLM "`Query` 在 `http.method=POST` 上占了 50% CPU"，比"query 总耗时高"有用得多。

加上**允许破坏性更新**的许可以后，趁机把 `pkg/profile/loader.go` 重构成返回包装类型，让 label 摘要和 profile 类型推断在加载时一次性计算，所有命令共享——而不是每个命令自己抽一遍。

## What Changes

### 新增命令
- **`tags` 命令**：发现层，列出 profile 中所有 pprof label 及其 value 分布。支持 `text|json|markdown`，支持 `--top N` 截断 value 数量，数字 label 自动带单位。

### 新增过滤能力
- **`--tag key=value` 过滤器**：可重复使用，支持 string 和 numeric label
  - **同 key 多次 → OR**（`--tag state=blocked --tag state=running` 表示 state 是 blocked 或 running）
  - **跨 key → AND**（`--tag state=blocked --tag wait_reason=IO` 表示两个条件都满足）
- **`--tag-ignore key=value` 过滤器**：与 `--tag` 同语义，反向
- 0 样本时报错并提示 profile 中是否完全无 label
- 不提供 `--include-unlabeled` flag（v1 严格语义，错误信息兜底）

### 新增分析能力
- **`analyze` 输出包含 label breakdown**：默认对 top 20 函数展开 `label_breakdown` 字段，展示每个函数在每个指定 label value 上的 flat/cum 分布
  - `--tag-breakdown-on key1,key2` 指定要展开的 label key
  - `--tag-breakdown-top N` 指定展开几个函数（默认 20）
  - 函数内的 label value 不截断（label value 通常很少，截断会反直觉）

### 修改现有命令
- **`info` 输出加 label 摘要**：`Labels: 3 keys, 247 unique values`（一行，不展开）
- **`list` 改名 `--exclude` → `--ignore-function`**：函数名正则排除改名为 `--ignore-function`（破坏性变更）。`--tag-ignore` 是新加的 label filter flag，不兼做函数名排除
- **`analyze` / `list` / `traces` / `diff`**：全部支持 `--tag` / `--tag-ignore` filter（diff 中两个 profile 同 filter）

### 架构重构
- **`Loader.LoadFromFile` 重构**：返回新的 `*Profile` 包装类型，内嵌 `*profile.Profile`，附加 `LabelSummaries []LabelSummary` 和 `InferredType string`，加载时一次性计算
- 所有命令改为消费 `*Profile`，统一访问 `p.LabelSummaries` 和 `p.InferredType`
- 抽 `pkg/profile/labels.go`：`LabelFilter`、`ApplyLabelFilter`、`ExtractLabelSummaries`、`ComputeFunctionBreakdowns`

### 文档与模板
- `pkg/skill/template.md` 同步：触发词加"标签"、何时使用表加一行、命令速查加 `tags` 段、典型工作流加"按标签过滤"段、决策树补"label"行、输出解读补字段表
- `README.md` 命令清单加 `tags`，更新 `analyze` / `list` / `traces` / `diff` 的 flag 说明

## Capabilities

### New Capabilities
- `labels-support`：顶层规范，定义 `LabelFilter` 语义、breakdown 行为、跨命令 filter 一致性、loader 包装类型契约
- `tags-command`：`tags` 子命令规范

### Modified Capabilities
- `profile-analyze`：加 `--tag` / `--tag-ignore` / `--tag-breakdown-on` / `--tag-breakdown-top` flag，输出加 `label_breakdown` 字段
- `profile-list`：加 `--tag` / `--tag-ignore` flag，`--exclude` 改名为 `--ignore-function`
- `profile-traces`：加 `--tag` / `--tag-ignore` flag
- `profile-diff`：加 `--tag` / `--tag-ignore` flag（两个 profile 同 filter）
- `profile-info`：输出加 `label_summary` 字段（一行摘要）

## Impact

### 新增文件
- `pkg/profile/profile.go`：`Profile` 包装类型
- `pkg/profile/labels.go`：`LabelFilter` / `ApplyLabelFilter` / `ExtractLabelSummaries` / `ComputeFunctionBreakdowns`
- `pkg/profile/labels_test.go`：filter / summary 单元测试
- `pkg/profile/tags.go`：`Tags` 函数 + `TagsResult` 结构
- `pkg/profile/tags_test.go`
- `pkg/commands/tags.go`：`TagsCmd` cobra 命令
- `pkg/commands/tags_test.go`
- `testdata/goroutine.pb.gz`：goroutine 测试数据（含 label）
- `testdata/generate.go`：扩展以生成 goroutine profile

### 修改文件
- `pkg/profile/loader.go`：`LoadFromFile` 返回 `*Profile` 而不是 `*profile.Profile`
- `pkg/profile/analysis.go`：消费 `*Profile`，加 breakdown 计算
- `pkg/profile/diff.go`：消费 `*Profile`，两边都做 label filter
- `pkg/profile/info.go`：消费 `*Profile.LabelSummaries`
- `pkg/profile/traces.go`：消费 `*Profile`
- `pkg/profile/list.go`：消费 `*Profile`
- `pkg/profile/diagnose.go`：消费 `*Profile`
- `pkg/commands/analyze.go`：加 4 个新 flag
- `pkg/commands/list.go`：加 2 个新 label flag + `--exclude` 改名为 `--ignore-function`
- `pkg/commands/traces.go`：加 2 个新 flag
- `pkg/commands/diff.go`：加 2 个新 flag
- `pkg/commands/info.go`：无需新 flag（自动用 `p.LabelSummaries`）
- `pkg/output/formatter.go`：5 个 Result 都补新字段渲染（tags / analyze 的 breakdown / info 的 label_summary）
- `cmd/root.go`：注册 `TagsCmd`
- `pkg/skill/template.md`：同步
- `README.md`：同步

### 破坏性变更
1. **Loader API**：`LoadFromFile` 返回类型从 `*profile.Profile` 改为 `*Profile`（包装类型，内嵌原类型，访问字段不变）。所有调用方需更新——这是机械改动。
2. **`list` flag 改名**：`--exclude pattern` → `--ignore-function pattern`。语义不变（正则排除函数），只是名字改了。`--tag-ignore` 是新加的 label filter flag，与改名前的 `--exclude` 完全独立。

### 依赖
- 不变：仅使用 `github.com/google/pprof/profile` 和 `github.com/spf13/cobra`

### 性能
- `LabelSummaries` 加载时一次遍历所有 sample 计算，O(n)
- `ApplyLabelFilter` 一次遍历，O(n)
- `ComputeFunctionBreakdowns` 复用已计算的 hotspot map，再加一次遍历，O(n)
- 总开销对 100k 样本应 < 50ms
