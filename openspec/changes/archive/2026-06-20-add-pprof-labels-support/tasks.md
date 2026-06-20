# Tasks: pprof Labels Support

## 1. Loader 重构（破坏性基础设施）

- [x] 1.1 新建 `pkg/profile/profile.go`：`Profile` 包装类型，内嵌 `*profile.Profile`，附加 `LabelSummaries []LabelSummary` 和 `InferredType string` 字段；`NewProfile(p *profile.Profile) *Profile` 构造器
- [x] 1.2 修改 `pkg/profile/loader.go`：`LoadFromFile` 返回 `(*Profile, error)` 而不是 `(*profile.Profile, error)`，构造时调 `NewProfile`
- [x] 1.3 同步修改所有 `*profile.Profile` 消费者（5 个文件）：`pkg/profile/{analysis,diff,info,traces,list,diagnose}.go` 改为消费 `*Profile`
- [x] 1.4 跑 `go build ./...` 确认编译通过；`make test` 确认所有现有测试通过
- [x] 1.5 验证：`./agent-insight info testdata/cpu.pb.gz` 输出与重构前一致；`analyze` / `list` / `traces` / `diff` / `tree` / `merge` / `trend` / `diagnose` / `flame` 全部行为不变

## 2. Label 数据模型与算法

- [x] 2.1 新建 `pkg/profile/labels.go`：`LabelSummary` / `LabelValueSummary` / `LabelFilter` / `FunctionLabelBreakdown` / `LabelBreakdown` / `LabelValueContribution` / `BreakdownConfig` 结构体定义
- [x] 2.2 实现 `ExtractLabelSummaries(p *profile.Profile) []LabelSummary`：遍历 sample，区分 string/numeric label，收集 value 分布；按 distinct value 数降序排序；默认截断 value 到 top 50
- [x] 2.3 实现 `NewLabelFilter(focusFlags, ignoreFlags []string) (*LabelFilter, error)`：解析 `--tag key=value` 格式；同 key 多次合并到同一 key 的 value 列表（OR）
- [x] 2.4 实现 `(f *LabelFilter) Apply(p *Profile) (*Profile, error)`：过滤 sample；matched == 0 时报错并提示；matched > 0 时复制 profile 替换 sample，复用 `p.LabelSummaries` 和 `p.InferredType`
- [x] 2.5 实现 `(f *LabelFilter) matches(labels []*profile.Label) bool`：单 sample 判定
- [x] 2.6 实现 `ComputeFunctionBreakdowns(p *Profile, hotspots []FunctionInfo, cfg BreakdownConfig) []FunctionLabelBreakdown`：top-N 函数 + 累加 label value 分布
- [x] 2.7 实现 `labelValueToString(*profile.Label) string`：把 `Label.Str` / `Label.Num` / `Label.NumUnit` 序列化为字符串

## 3. Label 单元测试

- [x] 3.1 `pkg/profile/labels_test.go`：`TestExtractLabelSummaries_NoLabels`（无 label profile 返回空）
- [x] 3.2 `pkg/profile/labels_test.go`：`TestExtractLabelSummaries_StringLabels`（多 value 排序）
- [x] 3.3 `pkg/profile/labels_test.go`：`TestExtractLabelSummaries_NumericLabels`（带 unit）
- [x] 3.4 `pkg/profile/labels_test.go`：`TestExtractLabelSummaries_Mixed`（string + numeric 混合）
- [x] 3.5 `pkg/profile/labels_test.go`：`TestLabelFilter_Focus_WithinKeyOR`（同 key 多次 OR）
- [x] 3.6 `pkg/profile/labels_test.go`：`TestLabelFilter_Focus_AcrossKeyAND`（跨 key AND）
- [x] 3.7 `pkg/profile/labels_test.go`：`TestLabelFilter_Ignore`（ignore 行为）
- [x] 3.8 `pkg/profile/labels_test.go`：`TestLabelFilter_Combined`（focus + ignore 组合）
- [x] 3.9 `pkg/profile/labels_test.go`：`TestLabelFilter_ZeroSamples`（0 样本报错）
- [x] 3.10 `pkg/profile/labels_test.go`：`TestLabelFilter_EmptyFilter`（空 filter 是 no-op）
- [x] 3.11 `pkg/profile/labels_test.go`：`TestLabelFilter_NumericLabel`（数字 label 比对）
- [x] 3.12 `pkg/profile/labels_test.go`：`TestNewLabelFilter_InvalidFormat`（缺 `=` / 空 key / 空 value 报错）
- [x] 3.13 `pkg/profile/labels_test.go`：`TestComputeFunctionBreakdowns_Empty`（无 key / 无 hotspots）
- [x] 3.14 `pkg/profile/labels_test.go`：`TestComputeFunctionBreakdowns_TopN`（cfg.Top 生效）
- [x] 3.15 `pkg/profile/labels_test.go`：`TestComputeFunctionBreakdowns_FlatDistribution`（flat 累加正确）

## 4. 测试数据：goroutine.pb.gz

- [x] 4.1 检查 `github.com/google/pprof@v0.0.0-20260604005048-7023385849c0/profile/testdata` 是否有现成 goroutine profile 含 label
- [x] 4.2 若有：复制到 `testdata/goroutine.pb.gz` 并写注释说明数据来源（不适用：4.1 确认无现成含 label 的 goroutine profile，实际走 4.3 在 generate.go 中生成）
- [x] 4.3 若无：扩展 `testdata/generate.go` 加一个 case，构造 `profile.Sample` 时设置 `Label: []profile.Label{{Key: "state", Str: ...}, ...}`
- [x] 4.4 验证：`./agent-insight info testdata/goroutine.pb.gz` 输出含 label 摘要

## 5. tags 命令

- [x] 5.1 新建 `pkg/profile/tags.go`：`Tags` 函数 + `TagsResult` 结构（含 `ProfilePath` / `Type` / `TotalSamples` / `Labels []LabelSummary`）
- [x] 5.2 新建 `pkg/commands/tags.go`：`TagsCmd` cobra 命令，支持 `--top N` 和 `--format text|json|markdown` flag
- [x] 5.3 `pkg/output/formatter.go` 加 `TagsResult` 的三种 formatter
- [x] 5.4 `cmd/root.go` 注册 `TagsCmd`
- [x] 5.5 验证：`./agent-insight tags testdata/goroutine.pb.gz` 输出符合 design §3.1
- [x] 5.6 验证：`./agent-insight tags testdata/cpu.pb.gz` 无 label profile 输出 "no labels found" 提示
- [x] 5.7 验证：JSON / markdown 格式输出正确
- [x] 5.8 `pkg/profile/tags_test.go` + `pkg/commands/tags_test.go`：单元测试

## 6. analyze 支持 --tag 和 breakdown

- [x] 6.1 `pkg/commands/analyze.go` 加 flag：`--tag key=value`（StringSliceP，可重复）、`--tag-ignore key=value`、`--tag-breakdown-on key1,key2`、`--tag-breakdown-top N`（默认 20）
- [x] 6.2 `pkg/commands/analyze.go` 在调 `profile.Analyze` 之前用 `LabelFilter.Apply` 过滤
- [x] 6.3 `pkg/profile/analysis.go` `Analyze` 函数签名加 `breakdownConfig` 参数，函数内部调 `ComputeFunctionBreakdowns` 并把结果附加到 `AnalyzeResult`
- [x] 6.4 `AnalyzeResult` 新增 `LabelBreakdowns []FunctionLabelBreakdown` 字段
- [x] 6.5 `pkg/output/formatter.go` 三个 formatter 渲染 `label_breakdown` 字段
- [x] 6.6 验证：`./agent-insight analyze testdata/goroutine.pb.gz --tag state=blocked` 返回 filtered 结果
- [x] 6.7 验证：`./agent-insight analyze testdata/goroutine.pb.gz --tag-breakdown-on state --tag-breakdown-top 5` 输出含 `label_breakdown`
- [x] 6.8 验证：`./agent-insight analyze testdata/cpu.pb.gz --tag state=blocked` 报错 "tag filter matched 0 of N samples"
- [x] 6.9 `pkg/commands/analyze_test.go` 加用例：filter / breakdown 渲染

## 7. list 支持 --tag 和 --ignore-function（破坏性：--exclude 改名）

- [x] 7.1 `pkg/commands/list.go` 把 `--exclude pattern` 改名为 `--ignore-function pattern`（变量 `listExclude` → `listIgnoreFunction`）。这是函数名正则排除专用 flag，**与 `--tag-ignore` 完全独立**
- [x] 7.2 `pkg/commands/list.go` 加 `--tag key=value` 和 `--tag-ignore key=value` flag（StringSliceP，两个 label 维度 flag）
- [x] 7.3 `pkg/commands/list.go` 在调 `profile.List` 之前用 `LabelFilter.Apply` 过滤
- [x] 7.4 验证：`./agent-insight list cpu.pb.gz "main.*" --ignore-function "database.*"` 等价于旧 `--exclude "database.*"`
- [x] 7.5 验证：`./agent-insight list goroutine.pb.gz "Query" --tag state=blocked` 返回 filtered
- [x] 7.6 验证：旧 `--exclude` 不再被识别
- [x] 7.7 验证：`--tag-ignore "database.*"`（无 `=`）报错 "missing '=' format"（因为这是 label flag，不接受正则）
- [x] 7.8 验证：`--ignore-function` 和 `--tag-ignore` 共存正常工作
- [x] 7.9 `pkg/commands/list_test.go` 更新所有 `--exclude` 用例为 `--ignore-function`

## 8. traces 支持 --tag

- [x] 8.1 `pkg/commands/traces.go` 加 `--tag key=value` 和 `--tag-ignore key=value` flag
- [x] 8.2 `pkg/commands/traces.go` 在调 `profile.Traces` 之前过滤
- [x] 8.3 验证：`./agent-insight traces goroutine.pb.gz --tag state=blocked --format json` 返回 filtered
- [x] 8.4 `pkg/commands/traces_test.go` 加 filter 用例

## 9. diff 支持 --tag

- [x] 9.1 `pkg/commands/diff.go` 加 `--tag key=value` 和 `--tag-ignore key=value` flag
- [x] 9.2 `pkg/commands/diff.go` 在调 `profile.Diff` 之前对 base 和 target 都用 `LabelFilter.Apply` 过滤
- [x] 9.3 验证：`./agent-insight diff v1.pb.gz v2.pb.gz --tag http.status=500` 对 5xx 请求做 diff
- [x] 9.4 `pkg/commands/diff_test.go` 加 filter 用例

## 10. info 加 label 摘要

- [x] 10.1 `pkg/profile/info.go` `Info` 函数输出加 `LabelSummary` 字段
- [x] 10.2 `InfoResult` 加 `LabelSummary *LabelSummaryBrief` 字段（含 `KeyCount` / `DistinctValues`）
- [x] 10.3 `pkg/output/formatter.go` 三个 formatter 渲染 `Labels: N keys, M unique values` 行
- [x] 10.4 验证：`./agent-insight info testdata/goroutine.pb.gz` 输出含 label 摘要行
- [x] 10.5 验证：`./agent-insight info testdata/cpu.pb.gz` 输出含 `Labels: 0 keys`

## 11. Skill 模板同步

- [x] 11.1 `pkg/skill/template.md` frontmatter `description` 加 "pprof 标签" / "label" 触发词
- [x] 11.2 "何时使用"表加一行："用户提到 pprof label / goroutine state / http.method / 想知道样本按什么分组" → `agent-insight tags profile.pb.gz`
- [x] 11.3 "何时使用"表加一行："用户想按 pprof label 过滤热点" → `analyze/list/traces --tag key=value`
- [x] 11.4 "命令速查"加 `tags` 子段（usage / flags / 示例）
- [x] 11.5 "命令速查"更新 `analyze` / `list` / `traces` / `diff` 段，加 `--tag` / `--tag-ignore` flag 说明
- [x] 11.6 "典型工作流"加 "按 label 过滤" 段（goroutine state、http method 两个场景）
- [x] 11.7 "决策树"加："看到 `info` 输出 `Labels: 3 keys`" → 跑 `tags` 详细看
- [x] 11.8 "决策树"加："goroutine profile 想看阻塞态" → `analyze --tag state=blocked`
- [x] 11.9 "输出解读"加 `tags` JSON 字段表 + `analyze` 的 `label_breakdown` 字段表
- [x] 11.10 "注意事项"加："`list` 命令中 `--exclude` 已改名为 `--ignore-function`（仅作用于函数名正则排除），新增 `--tag-ignore key=value` 用于 label 过滤"

## 12. README 同步

- [x] 12.1 `README.md` 命令清单加 `tags`
- [x] 12.2 `README.md` 更新 `analyze` / `list` / `traces` / `diff` 段的 flag 列表
- [x] 12.3 `README.md` 加 "pprof Labels" 章节，说明 `--tag` 用法
- [x] 12.4 `README.md` 加 **BREAKING** 标注：v0.X 起 `--exclude` 改名为 `--ignore-function`；同时新增 `--tag-ignore key=value` 用于 label 过滤
- [x] 12.5 `cmd/root.go` 的 Long 描述更新命令列表

## 13. 集成验证

- [x] 13.1 端到端：goroutine.pb.gz 跑 `info` → `tags` → `analyze --tag state=blocked --tag-breakdown-on state` 全链路
- [x] 13.2 端到端：cpu.pb.gz（无 label）跑 `analyze --tag state=blocked` 验证 0 样本报错
- [x] 13.3 端到端：list 命令 `--ignore-function` 等价于旧 `--exclude`
- [x] 13.4 跑 `make test` 全部测试通过
- [x] 13.5 跑 `make lint` 无新增 lint 错误
- [x] 13.6 跑 `./agent-insight diagnose testdata/goroutine.pb.gz --format json` 验证 diagnose 命令未被破坏
- [x] 13.7 跑 `./agent-insight init --force` 生成 SKILL.md 验证包含新 commands 和 flags
