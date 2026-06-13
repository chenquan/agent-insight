## Context

agent-insight 当前支持 info/analyze/list/flame/diff/traces/tree/init 共 8 个子命令，但缺少 profile 合并能力。生产环境（尤其 Go 服务）通常通过多次短时间采样采集 CPU/heap profile，需要合并后才能获得有代表性的分析数据。

pprof 库（`github.com/google/pprof/profile`）原生提供 `profile.Merge()` 函数，可直接利用。

## Goals / Non-Goals

**Goals:**
- 提供纯合并能力，输出标准 `.pb.gz` 文件供其他命令使用
- 支持多文件路径和目录自动发现两种输入方式
- 校验输入 profile 类型一致性，给出清晰错误信息

**Non-Goals:**
- 不在合并过程中做过滤或分析
- 不支持时间切片（pprof Sample 无时间戳，技术上不可行）
- 不做趋势分析（留给后续 change）
- 不支持输出到 stdout 二进制流（合并结果必须写文件）

## Decisions

**D1: 输出方式 — 仅写文件，不输出分析结果**

纯合并工具。合并后的 `.pb.gz` 可被现有 analyze/diff/flame 等命令消费。不提供 `--analyze` 一步到位模式，保持职责单一。

替代方案：合并后直接输出分析结果（两步合一）。放弃原因：merge 是基础操作，组合其他命令更灵活。

**D2: 输入方式 — 文件列表 + 目录自动发现**

- 显式指定文件：`agent-insight merge a.pb.gz b.pb.gz c.pb.gz -o out.pb.gz`
- 目录模式：`agent-insight merge ./profiles/ -o out.pb.gz`（递归发现所有 `.pb` 和 `.pb.gz`）

目录模式递归扫描子目录，按完整路径排序保证确定性。如果目录下有多种类型的 profile，报错提示用户指定文件。

**D3: 类型校验 — 基于 PeriodType 比对**

使用 `profile.PeriodType.Type` 判断 profile 类型。所有输入 profile 的 PeriodType 必须一致。如果某个 profile 没有 PeriodType，跳过校验（兼容不完整 profile）。

**D4: 三层架构遵循**

```
pkg/commands/merge.go  → 参数解析（cobra）、输入校验、调用 profile 层
pkg/profile/merge.go   → 核心合并逻辑、类型校验、调用 profile.Merge()
cmd/root.go            → 注册 MergeCmd
```

不需要新增 output formatter — merge 不输出分析结果，只写文件。

## Risks / Trade-offs

- **[内存]** 多个大 profile 合并时内存占用较高 → 这是 pprof.Merge 的固有限制，暂不优化
- **[目录发现]** 目录下混合类型 profile 会报错 → 设计上要求类型一致，报错信息需清晰
- **[输出格式]** 固定输出 `.pb.gz` → 如果输入是 `.pb` 也输出 `.pb.gz`，保持一致性
