## Why

agent-insight 当前有 5 个命令（analyze、list、flame、diff、init），但缺少三个常见分析场景：快速查看 profile 元信息、查看原始采样调用链、全局调用树视图。AI 编码助手在做性能分析时，需要先了解 profile 概况，再逐步深入具体路径和调用结构，当前工具链在这三个环节存在断点。

## What Changes

- 新增 `info` 命令：轻量级 profile 元信息概览，输出类型、时长、采样数、值类型列表、符号状态、映射信息等，无需执行重量级分析
- 新增 `traces` 命令：展示匹配 pattern 的原始采样调用链，每条链显示完整路径和对应值，与 flame 的聚合视图互补
- 新增 `tree` 命令：层级调用树视图，从根到叶逐层聚合展示全局调用结构，补全 analyze（平面排名）和 list（单函数上下文）之外的视角
- 三个命令均支持 `--format text|json|markdown` 多格式输出
- 注册到 rootCmd，更新 skill template 和 README

## Capabilities

### New Capabilities
- `info-command`: profile 元信息概览命令，展示 profile 类型、时长、采样统计、值类型、符号状态、映射信息
- `traces-command`: 原始采样调用链查询命令，按 pattern 过滤并展示完整调用路径和值
- `tree-command`: 层级调用树命令，从根到叶逐层聚合的全局调用结构视图

### Modified Capabilities
（无现有 spec 需要修改）

## Impact

- 新增文件：`pkg/commands/info.go`、`pkg/commands/traces.go`、`pkg/commands/tree.go`
- 新增文件：`pkg/profile/info.go`、`pkg/profile/traces.go`、`pkg/profile/tree.go`
- 修改文件：`cmd/root.go`（注册新命令）
- 修改文件：`pkg/output/formatter.go`（新增输出格式化器）
- 修改文件：`pkg/skill/template.md`（更新 skill 文档）
- 修改文件：`README.md`（更新使用文档）
- 依赖不变：仅使用已有的 `github.com/google/pprof/profile` 和 `github.com/spf13/cobra`
