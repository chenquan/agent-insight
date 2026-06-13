## Why

`diagnose` 命令在 skill 模板中的触发条件写得太模糊（"用户需要 AI 辅助性能诊断"），导致 AI 编码助手无法准确判断何时该用 `diagnose`（一键全面诊断）vs 手动组合命令（analyze/list/traces 精准分析）。`diagnose` 的核心价值是**探索性场景**——用户不确定问题在哪时一键获得全面概览，而不是所有性能分析场景的默认入口。

## What Changes

- 精确化 `pkg/skill/template.md` 中 `diagnose` 的触发条件：从模糊的"用户需要 AI 辅助性能诊断"改为明确的"用户不确定问题在哪，需要全面概览"
- 在 diagnose 命令说明中添加"何时用 vs 何时不该用"的对比指引
- 更新决策树中 diagnose 的触发行，使用具体场景描述
- 在快速概览工作流的决策点中增加"不确定问题"分支指向 diagnose

## Capabilities

### New Capabilities

（无）

### Modified Capabilities

- `profile-diagnose`: 修改 skill 模板中 diagnose 的触发条件和使用指引（命令行为不变，只改 skill 文档）

## Impact

- 修改文件：`pkg/skill/template.md`（纯文案调整，不涉及代码逻辑）
- 不影响命令行行为、API、依赖
