## Purpose

提供 `init` 子命令，用于生成 Claude Code skill 文件（SKILL.md），使 agent-insight 能被 Claude Code 自动识别和触发。

## Requirements

### Requirement: init 命令生成 Claude Code skill 文件
系统 SHALL 提供 `init` 子命令，在当前工作目录下生成 `.claude/skills/agent-insight/SKILL.md` 文件。

#### Scenario: 首次生成 skill 文件
- **WHEN** 用户执行 `agent-insight init` 且 `.claude/skills/agent-insight/` 目录不存在
- **THEN** 系统创建目录 `.claude/skills/agent-insight/` 并生成 `SKILL.md` 文件，输出成功信息

#### Scenario: skill 文件已存在时覆盖
- **WHEN** 用户执行 `agent-insight init` 且 `.claude/skills/agent-insight/SKILL.md` 已存在
- **THEN** 系统提示文件已存在并覆盖，输出覆盖信息

### Requirement: SKILL.md 包含完整的触发条件
SKILL.md 的 frontmatter description SHALL 包含明确的触发关键词，包括 pprof、CPU profiling、内存泄漏、heap 分析、性能热点、火焰图、.pb.gz 文件。

#### Scenario: Claude Code 通过关键词触发
- **WHEN** Claude Code 读取 SKILL.md 的 description 字段
- **THEN** description 包含触发关键词：pprof、性能分析、CPU profiling、内存泄漏、heap、火焰图、.pb.gz、性能热点

### Requirement: SKILL.md 包含完整命令使用指南
SKILL.md SHALL 内嵌所有 4 个子命令（analyze/list/flame/diff）的用法说明，包含 flags 和示例命令。

#### Scenario: Claude Code 获取命令用法
- **WHEN** Claude Code 阅读 SKILL.md 内容
- **THEN** 文档包含每个子命令的用法、常用 flags 说明和示例命令

### Requirement: SKILL.md 包含典型工作流
SKILL.md SHALL 包含典型性能分析工作流的描述，指导 Claude Code 在不同场景下组合使用命令。

#### Scenario: Claude Code 按工作流分析性能问题
- **WHEN** 用户提到"服务很慢"或类似性能问题
- **THEN** Claude Code 按照 SKILL.md 中的工作流指南，依次使用 analyze → list → flame 命令进行分析

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

### Requirement: SKILL.md 包含输出解读指南
SKILL.md SHALL 包含 JSON 输出字段的含义说明，帮助 Claude Code 正确解读分析结果。

#### Scenario: Claude Code 解读分析输出
- **WHEN** Claude Code 执行 analyze 命令获得 JSON 输出
- **THEN** Claude Code 根据 SKILL.md 中的字段说明正确解读 flat/cum/percent 等指标

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

### Requirement: init 命令支持 --force 标志
init 命令 SHALL 支持 `--force` / `-f` 标志，跳过已存在文件的确认提示直接覆盖。

#### Scenario: 使用 --force 覆盖
- **WHEN** 用户执行 `agent-insight init --force` 且 SKILL.md 已存在
- **THEN** 系统直接覆盖文件，不提示确认

#### Scenario: 不使用 --force 且文件存在
- **WHEN** 用户执行 `agent-insight init` 且 SKILL.md 已存在
- **THEN** 系统输出提示"文件已存在，使用 --force 覆盖"并以非零状态码退出
