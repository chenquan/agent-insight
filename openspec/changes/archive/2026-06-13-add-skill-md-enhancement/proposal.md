## Why

`agent-insight` 的核心使用场景是 Claude Code 通过 SKILL.md 触发后调用 CLI 命令。当前 `pkg/skill/template.md` 已包含命令速查、典型工作流、输出解读、注意事项四大部分,但**典型工作流仅列步骤,缺少"决策点"和"陷阱提示";输出解读仅列字段,缺少"模式识别"启发式;整体缺少"决策树"快速参考**。

AI agent 拿到 SKILL.md 后,需要自己推断"看到 X 应该跑 Y",反复试错消耗 token。强化 SKILL.md 的工作流决策化、决策树段、输出解读启发式,可让 AI 直接按文档执行,无需推断。

本 change 是**纯文档改动**(`pkg/skill/template.md` 改写),0 代码改动,直接提升 AI agent 的实际使用体验。

## What Changes

- **强化"典型工作流"段**:每个工作流的每一步加"决策点"段(看到 X 改用 Y)和"陷阱提示"段(常见误用提醒)。覆盖快速概览、CPU 性能分析、内存分析、调用路径追踪、版本对比 5 个工作流。
- **新增"决策树"段**:表格 "看到 X → 跑 Y",覆盖 8-10 个常见诊断场景的快速参考,放在典型工作流和输出解读之间。
- **强化"输出解读"段**:为 analyze / diff / flame / traces 命令的输出加"模式识别"启发式(如 flat vs cum 的 4 种模式、delta 的正负含义),用实际 JSON 示例辅助说明。
- **保留现有结构**:不删除"何时使用" / "命令速查" / "注意事项" 段,只在合适位置插入/改写。
- **0 代码改动**:不修改 Go 代码,不修改 `init` 命令行为,只更新 `pkg/skill/template.md`。
- **Claude Code 重新加载生效**:用户运行 `agent-insight init --force` 重新生成 SKILL.md 即可获得增强版。

不涉及 BREAKING 变更。`init` 命令生成的 SKILL.md 内容更丰富,但所有现有命令、flag、行为不变。

## Capabilities

### New Capabilities

无

### Modified Capabilities

- `init-command`: SKILL.md 模板内容强化。`典型工作流` 段增加决策点和陷阱提示;新增 `决策树` 段;`输出解读` 段增加模式识别启发式。

## Impact

- **代码**:无 Go 代码改动。
- **文档**:`pkg/skill/template.md` 改写,约 150-200 行新增/修改,纯 markdown。
- **测试**:无代码测试。人工 review 增强后的 SKILL.md 完整性、可读性、决策树准确性。
- **依赖**:无新依赖。
- **用户**:Claude Code 重新加载 skill 后,体验直接提升;人类用户阅读 SKILL.md 也更清楚。
- **回滚**:单文件改动,git revert 单 commit 即可。