# init-command Change Spec Delta

## ADDED Requirements

### Requirement: 典型工作流段包含决策点和陷阱提示
SKILL.md 的"典型工作流"段 SHALL 在每个工作流的每个步骤下补充"决策点"和"陷阱提示"两个子段。

- **决策点**:基于该步骤的输出结果,告诉 AI 看到 X 应该改用 Y 命令或调整参数(例如 "看到 leaf 函数改用 flame")
- **陷阱提示**:列出该步骤常见的误用和反模式(例如 "别只看 flat 排序")

5 个工作流(快速概览 / CPU 性能分析 / 内存分析 / 调用路径追踪 / 版本对比) MUST 全部包含这两个子段。

#### Scenario: 典型工作流包含决策点
- **WHEN** Claude Code 阅读 SKILL.md 的"典型工作流"段
- **THEN** 每个步骤下 MUST 有"决策点"子段,描述看到特定输出后应该改用的命令或参数

#### Scenario: 典型工作流包含陷阱提示
- **WHEN** Claude Code 阅读 SKILL.md 的"典型工作流"段
- **THEN** 每个步骤下 MUST 有"陷阱提示"子段,描述常见误用和反模式

#### Scenario: 5 个工作流全部覆盖
- **WHEN** 审计 SKILL.md 的"典型工作流"段
- **THEN** 快速概览 / CPU 性能分析 / 内存分析 / 调用路径追踪 / 版本对比 5 个工作流 MUST 都包含决策点和陷阱提示

### Requirement: SKILL.md 包含决策树段
SKILL.md SHALL 在"典型工作流"和"输出解读"之间新增"决策树:看到 X → 跑 Y"段,以表格形式列出 8-10 个常见诊断场景的快速参考。

每个表格行 MUST 描述一个"输入症状 → 推荐命令"映射,覆盖:
- CPU 性能场景(leaf / internal bottleneck)
- 内存分析场景(alloc_space / inuse_space / alloc_objects)
- goroutine 场景(count 异常)
- diff 场景(新函数 / 消失函数)
- 常见错误恢复(no samples matched / unknown value-type / failed to load profile)

#### Scenario: 决策树覆盖 CPU 性能场景
- **WHEN** Claude Code 阅读决策树段
- **THEN** 表格 MUST 包含"cpu profile + leaf bottleneck → analyze + flame"和"analyze top 都是 internal → list --callers-only"至少 2 个 CPU 相关行

#### Scenario: 决策树覆盖内存分析场景
- **WHEN** Claude Code 阅读决策树段
- **THEN** 表格 MUST 包含"heap + 内存增长 → alloc_space"和"heap + 当前占用高 → inuse_space"和"heap + 大量小对象 → alloc_objects"3 个内存相关行

#### Scenario: 决策树覆盖常见错误恢复
- **WHEN** Claude Code 阅读决策树段
- **THEN** 表格 MUST 包含至少 2 行错误恢复指引(如"no samples matched"和"unknown value-type"对应的处置)

### Requirement: 输出解读段包含模式识别启发式
SKILL.md 的"输出解读"段 SHALL 为 `analyze` 命令补充 `flat` vs `cum` 4 种模式识别,为 `diff` 命令补充 `delta_percent` / `new_functions` / `deleted_functions` 的解读启发式。

`analyze` MUST 包含以下 4 种 flat/cum 模式:
- `flat` 高 + `cum` 高 → self-bottleneck
- `flat` 低 + `cum` 高 → caller bottleneck
- `flat` 高 + `cum` = `flat` → leaf
- `flat` = 0 + `cum` 高 → pure caller

`diff` MUST 包含:
- `delta_percent > 0` → regression
- `delta_percent < 0` → improvement
- `new_functions[]` 非空 → new hotspots
- `deleted_functions[]` 非空 → gone hotspots

每个模式 MUST 给出实际 JSON 示例(从 testdata 提取的 ~5-10 行片段)辅助说明。

#### Scenario: analyze 输出包含 4 种 flat/cum 模式说明
- **WHEN** Claude Code 阅读"输出解读"段
- **THEN** 文档 MUST 列出 flat/cum 的 4 种模式及各自对应的诊断含义(self-bottleneck / caller bottleneck / leaf / pure caller)

#### Scenario: diff 输出包含 delta_percent 解读
- **WHEN** Claude Code 阅读"输出解读"段
- **THEN** 文档 MUST 解释 `delta_percent > 0` 为 regression,`delta_percent < 0` 为 improvement

#### Scenario: 每个模式配 JSON 示例
- **WHEN** Claude Code 阅读"输出解读"段
- **THEN** flat/cum 4 种模式 MUST 各配 1 个真实 JSON 示例(~5-10 行),且 MUST 标注"这是 X 模式"
