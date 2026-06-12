## ADDED Requirements

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

### Requirement: SKILL.md 包含输出解读指南
SKILL.md SHALL 包含 JSON 输出字段的含义说明，帮助 Claude Code 正确解读分析结果。

#### Scenario: Claude Code 解读分析输出
- **WHEN** Claude Code 执行 analyze 命令获得 JSON 输出
- **THEN** Claude Code 根据 SKILL.md 中的字段说明正确解读 flat/cum/percent 等指标

### Requirement: init 命令支持 --force 标志
init 命令 SHALL 支持 `--force` / `-f` 标志，跳过已存在文件的确认提示直接覆盖。

#### Scenario: 使用 --force 覆盖
- **WHEN** 用户执行 `agent-insight init --force` 且 SKILL.md 已存在
- **THEN** 系统直接覆盖文件，不提示确认

#### Scenario: 不使用 --force 且文件存在
- **WHEN** 用户执行 `agent-insight init` 且 SKILL.md 已存在
- **THEN** 系统输出提示"文件已存在，使用 --force 覆盖"并以非零状态码退出
