## Why

agent-insight 当前能对比两个 profile（`diff` 命令），但无法追踪性能随时间的变化趋势。用户在持续优化、CI 监控、生产环境退化发现等场景下，需要分析 3 个以上 profile 组成的序列，回答"哪些函数在持续恶化？"、"整体性能走势如何？"等趋势性问题。当前只能手动两两 diff，效率低且缺乏全局视角。

## What Changes

- 新增 `trend` 命令：接收 3 个以上 pprof 文件，分析函数级性能随时间的变化趋势
- 支持目录扫描（自动发现 .pb/.pb.gz）和显式文件列表两种输入方式
- 对每个函数计算 flat/cumulative 时间序列，通过线性回归检测趋势方向（regressing/improving/stable）
- 四层过滤机制（focus/ignore、min-impact、threshold、top N）控制输出量
- 五段式输出结构：Summary、Top Regressions、Top Improvements、New Hotspots（可选）、Volatile（可选）
- 支持 `--format text|json|markdown` 三种输出格式

## Capabilities

### New Capabilities
- `trend-command`: 时间序列趋势分析命令，加载多个 profile、计算函数级时间序列、线性回归趋势检测、多层过滤、结构化输出

### Modified Capabilities
（无现有 spec 需要修改）

## Impact

- 新增文件：`pkg/commands/trend.go`、`pkg/profile/trend.go`
- 修改文件：`cmd/root.go`（注册 trend 子命令）
- 修改文件：`pkg/output/formatter.go`（新增 TrendResult 的 text/json/markdown formatter）
- 修改文件：`pkg/skill/template.md`（更新 skill 文档，添加 trend 命令说明）
- 修改文件：`README.md`（更新使用文档）
- 依赖不变：仅使用已有的 `github.com/google/pprof/profile` 和 `github.com/spf13/cobra`
