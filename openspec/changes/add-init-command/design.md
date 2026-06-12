## Context

agent-insight 是一个 pprof 分析 CLI 工具，有 4 个子命令（analyze/list/flame/diff），当前 Claude Code 不知道它的存在。用户希望运行 `agent-insight init` 后，在当前项目下生成 `.claude/skills/agent-insight/SKILL.md`，让 Claude Code 自动学会何时以及如何使用该工具。

## Goals / Non-Goals

**Goals:**
- `agent-insight init` 生成项目级 `.claude/skills/agent-insight/SKILL.md`
- SKILL.md 内嵌完整使用指南（命令、flags、示例、工作流、输出解读）
- 触发条件清晰，Claude Code 能自动识别使用时机

**Non-Goals:**
- 不生成用户级 skill（~/.claude/skills/）
- 不生成 CLAUDE.md 片段
- 不做交互式选择（一键生成）
- 不自动检测或注册到 plugin 系统

## Decisions

### 1. Skill 内容内嵌 vs 运行时获取

**选择：内嵌完整文档**

SKILL.md 直接包含所有命令说明、flags、示例。agent-insight 命令接口稳定，skill 文件体积小（约 200 行），不需要运行时调用 `--help`。

### 2. 代码组织

**选择：`pkg/skill/` 包 + `pkg/commands/init.go` 命令**

- `pkg/skill/generator.go` — 负责 SKILL.md 模板生成
- `pkg/commands/init.go` — init 命令的 cobra 注册和执行逻辑

这样职责分离：skill 内容生成与命令注册解耦。

### 3. 生成策略

**选择：Go embed 嵌入模板**

使用 `//go:embed` 将 SKILL.md 模板嵌入二进制，运行时直接写出。好处：
- 无需外部文件依赖
- 模板和代码一起版本控制
- 安装后只需一个二进制文件

### 4. 目标路径

**选择：`.claude/skills/agent-insight/SKILL.md`**

如果目录不存在则自动创建。如果文件已存在，提示用户并覆盖。

## Risks / Trade-offs

- **Skill 内容过时** → agent-insight 命令接口变化后需要手动更新模板。缓解：模板内容保持与 README 同步
- **目录结构假设** → 假设 `.claude/skills/` 是 skill 存放路径。缓解：这是 Claude Code 的标准约定
- **覆盖已有文件** → 如果用户已自定义过 skill，init 会覆盖。缓解：执行时提示文件已存在，要求确认
