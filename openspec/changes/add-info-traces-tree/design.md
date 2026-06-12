## Context

agent-insight 是一个面向 AI 编码助手的 pprof 分析 CLI，当前有 analyze、list、flame、diff、init 五个命令。核心分析能力（热点排名、函数关系、火焰图、profile 对比）已经具备，但缺少三个常见分析环节：轻量概况查看、原始调用链查看、全局调用树视图。

项目架构遵循 `commands → profile → output` 三层分层：commands 解析参数，profile 做核心计算，output 做格式化。新增三个命令延续这一模式。

## Goals / Non-Goals

**Goals:**
- 新增 `info` 命令：零计算开销的 profile 元信息概览
- 新增 `traces` 命令：按 pattern 过滤的原始采样调用链展示
- 新增 `tree` 命令：层级调用树，从根到叶逐层聚合
- 三个命令均支持 text/json/markdown 输出格式
- 延续现有代码风格和分层架构

**Non-Goals:**
- 不做 HTTP 拉取 profile
- 不做 profile 合并
- 不做源码级标注（后续独立 change）
- 不做自动诊断/洞察（后续独立 change）

## Decisions

### 1. info 命令只读 Profile 字段，不执行采样遍历

**选择：** 直接读取 `profile.Profile` 的元信息字段（SampleType、TimeNanos、DurationNanos、Mapping 等），不遍历 Sample 切片。

**理由：** info 的定位是"快速概览"，应在毫秒级完成。遍历 Sample 属于 analyze 的职责。统计采样总数（`len(p.Sample)`）不算遍历。

**替代方案：** 在 analyze 结果中加一个 `--info-only` flag — 但这会导致加载不必要的计算逻辑。

### 2. traces 命令输出原始 Sample 粒度

**选择：** 每个 Sample 输出一条 trace，不做聚合。

**理由：** traces 与 flame 互补。flame 聚合同路径，traces 保留原始粒度。用户可以看到"同一个函数被哪些不同的调用路径触发"。

**替代方案：** 合并相似 trace — 但这会失去原始信息，与 flame 功能重叠。

### 3. tree 命令使用递归树结构

**选择：** 构建 `CallTreeNode` 递归结构，每个节点存储函数名、flat/cum 值、子节点列表。从所有 Sample 的 Location 链构建，根节点是调用栈最底部函数。

**数据结构：**
```
CallTreeNode {
    Name     string
    Flat     int64   // 直接归到该节点的值
    Cum      int64   // 经过该节点的累计值
    Children []CallTreeNode
}
```

**构建算法：** 遍历所有 Sample，对每条 Location 链（反转后从根到叶），沿路径逐层查找或创建子节点，累加 cum；仅叶子节点累加 flat。

**替代方案：** 用 map 存储再转换为树 — 递归结构更直观，性能足够（profile 通常不超过数万条采样）。

### 4. traces 和 tree 支持 focus/ignore 过滤

**选择：** 与 analyze/flame 保持一致，traces 和 tree 也支持 `--focus` 和 `--ignore` 正则过滤。

**理由：** 用户交互模式一致，降低认知成本。

### 5. 输出格式化器延续现有模式

**选择：** 每个命令在 `pkg/output/formatter.go` 中新增对应的 Formatter 结构体，实现 `FormatXxxResult` 方法。

**理由：** 与 analyze/diff/flame/list 保持一致，方便维护。

### 6. info 不嵌入现有工作流的第一步

**选择：** info 作为独立可选步骤，不插入到 CPU/内存/对比等现有工作流中。

**理由：** 当用户已知 profile 类型（如"CPU 很慢"），直接 analyze 更高效。info 适用于"不确定这是什么文件"的场景。强行插入会让简单场景多一步无用操作。

### 7. traces 的 pattern 为可选参数

**选择：** pattern 可选，不传时输出全部采样调用链（通过 `--top 20` 限制条数）。

**理由：** traces 的典型场景是"先看看有哪些调用路径"，用户不一定有明确目标。与 list 的必传 pattern 不同——list 是"我已知函数名，查关系"。

### 8. tree 默认按 cum 排序

**选择：** `--cum` 默认为 true（与 analyze 默认 flat 不同）。

**理由：** 调用树的本质是理解执行流，cum 值展示"多少资源流经这个分支"，天然适合树形结构。flat 排序更适合 analyze 的平面排名场景。

### 9. Skill 模板更新策略

**选择：** 更新 `pkg/skill/template.md`，涵盖以下变更：

1. **frontmatter 触发词**：新增 概况、元信息、调用链、执行路径、调用树、层级结构
2. **何时使用表**：新增 3 行触发场景（概况→info，调用链→traces，调用树→tree）
3. **命令速查**：新增 info/traces/tree 三个命令段（含 flags 表和示例）
4. **典型工作流**：新增"快速概览"和"调用路径追踪"两个工作流
5. **输出解读**：新增 info/traces/tree 三个 JSON 字段说明表
6. **注意事项**：新增"分析前可先用 info 了解概况"

**理由：** skill template 是 AI 使用 agent-insight 的唯一指引，必须与实际命令能力同步。

## Risks / Trade-offs

- **tree 性能** → 大量 Sample 时构建树可能较慢。缓解：设置 `--top` 限制输出节点数，默认只展示 cum 最高的分支。
- **traces 输出量大** → 匹配大量采样时输出可能很多。缓解：默认 `--top 20` 限制条数，用户可调整。
- **info 中 Mapping 路径可能很长** → text 格式中截断显示，json 格式完整输出。
