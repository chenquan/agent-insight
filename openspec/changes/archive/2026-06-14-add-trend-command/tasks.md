## 1. 公共代码提取

- [x] 1.1 将 `discoverProfiles` 从 `pkg/commands/merge.go` 提取到 `pkg/profile/loader.go` 的 `Loader.DiscoverProfiles(dir string) ([]string, error)` 方法
- [x] 1.2 更新 `pkg/commands/merge.go` 调用 `loader.DiscoverProfiles` 替代本地 `discoverProfiles`

## 2. 核心计算层 (pkg/profile/trend.go)

- [x] 2.1 定义 TimePoint 结构（Label string, Time int64）
- [x] 2.2 定义 TrendResult、FunctionTrend、OverallTrend 等数据结构（FunctionTrend 含 FlatSeries/CumSeries 为 `[]*int64`，支持 null）
- [x] 2.3 实现 TrendConfig 结构（MinImpact、Threshold、TopN、FocusPattern、IgnorePattern、ValueType、IncludeNew、IncludeVolatile）
- [x] 2.4 实现 Trend 入口函数签名：`func Trend(profiles []*profile.Profile, timePoints []TimePoint, config TrendConfig) (*TrendResult, error)`
- [x] 2.5 实现 profile 数量校验（>= 3，否则返回错误）
- [x] 2.6 实现函数值提取：遍历每个 profile 构建 flat/cumulative map，缺失值为 nil
- [x] 2.7 实现线性回归计算（最小二乘法，跳过 nil 值，avg 为 0 时直接返回 slope=0）
- [x] 2.8 实现趋势分类逻辑：`|slope/avg| * 100 > threshold` → regressing/improving，否则 stable
- [x] 2.9 实现变异系数（coefficient of variation）计算
- [x] 2.10 实现四层过滤：focus/ignore → min-impact（flat_i/total_i*100 取 max） → threshold → top N
- [x] 2.11 实现新增热点检测（首次出现索引 > 总点数*0.3，最终占比 > min-impact）
- [x] 2.12 实现全局走势计算（总采样数序列 + overall slope）
- [x] 2.13 编写单元测试

## 3. 命令层 (pkg/commands/trend.go)

- [x] 3.1 创建 TrendCmd cobra 命令，支持目录和文件列表输入（`cobra.MinimumNArgs(1)`）
- [x] 3.2 添加 flags：--format、--focus、--ignore、--min-impact、--threshold、--top、--sort-by、--value-type、--include-new、--include-volatile
- [x] 3.3 实现 runTrend 函数：参数校验、调用 loader.DiscoverProfiles 或使用显式路径、读取 mtime、排序、加载 profile、构造 TimePoint 列表、调用 profile.Trend、选择 formatter

## 4. 输出层 (pkg/output/formatter.go)

- [x] 4.1 定义 TrendResult 的 JSON formatter（TrendJSONFormatter）
- [x] 4.2 定义 TrendResult 的 Text formatter（TrendTextFormatter）
- [x] 4.3 定义 TrendResult 的 Markdown formatter（TrendMarkdownFormatter）
- [x] 4.4 确保 JSON 输出中 nil 值序列化为 null、Text/Markdown 用 "-"

## 5. 集成与注册

- [x] 5.1 在 cmd/root.go 注册 TrendCmd
- [x] 5.2 更新 pkg/skill/template.md 添加 trend 命令说明
- [x] 5.3 更新 README.md 添加 trend 使用示例

## 6. 测试

- [x] 6.1 生成 trend 测试数据（testdata/ 下多个时间序列 profile）
- [x] 6.2 编写 Trend 函数的集成测试（验证完整流程）
- [x] 6.3 运行 `make test` 确保所有测试通过
- [x] 6.4 运行 `make lint` 确保代码质量
