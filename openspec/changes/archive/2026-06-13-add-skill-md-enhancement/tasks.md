# Tasks: add-skill-md-enhancement

## 1. 准备工作

- [x] 1.1 阅读 `pkg/skill/template.md` 现状(330 行,6 段),确认段顺序
- [x] 1.2 跑一次 `analyze testdata/cpu.pb.gz --format json` 抓真实 JSON 片段,准备给输出解读段做示例
- [x] 1.3 跑一次 `analyze testdata/heap.pb.gz --value-type alloc_space` 和 `inuse_space` 抓内存示例
- [x] 1.4 确认 5 个工作流名称(快速概览 / CPU 性能分析 / 内存分析 / 调用路径追踪 / 版本对比)

## 2. 改写 `pkg/skill/template.md`

### 2.1 A 段:强化"典型工作流"(加决策点 + 陷阱提示)

- [x] 2.1.1 工作流 1 (快速概览, 1 步):加决策点(看到占比高 → analyze top N)+ 陷阱提示(别直接看 JSON 给结论)
- [x] 2.1.2 工作流 2 (CPU 性能分析, 3 步):每步加决策点(leaf vs internal 分支)+ 陷阱提示(runtime.* 不一定是问题)
- [x] 2.1.3 工作流 3 (内存分析, 3 步):每步加决策点(alloc_space vs inuse_space 分支)+ 陷阱提示(inuse_objects 误用)
- [x] 2.1.4 工作流 4 (调用路径追踪, 2 步):每步加决策点(找到 callee 后的下一步)+ 陷阱提示(callee 链过长)
- [x] 2.1.5 工作流 5 (版本对比, 1 步):加决策点(delta_percent 异常值)+ 陷阱提示(new_functions 不一定是 regression)

### 2.2 B 段:新增"决策树"段

- [x] 2.2.1 在"典型工作流"和"输出解读"之间插入 `## 决策树: 看到 X → 跑 Y` 段
- [x] 2.2.2 表格包含 8-10 行:CPU 2-3 行 + 内存 3 行 + goroutine 1 行 + diff 1-2 行 + 错误恢复 3 行

### 2.3 C 段:强化"输出解读"段

- [x] 2.3.1 `analyze` 段加 `### 模式识别 (flat vs cum)` 子段,4 种模式 + 各 1 个 JSON 示例
- [x] 2.3.2 `diff` 段加 `### 模式识别` 子段,4 种 delta 解读(2 个正负 + new/deleted)
- [x] 2.3.3 保留原有字段说明(flat / cum / percent / sample type 等)

### 2.4 保持不变

- [x] 2.4.1 段顺序不变(描述 → 何时使用 → 命令速查 → 典型工作流 → 决策树 → 输出解读 → 注意事项)
- [x] 2.4.2 frontmatter description 不动
- [x] 2.4.3 注意事项段保留(可在尾部加 1-2 条关于决策树/工作流的提示)

## 3. 验证

- [x] 3.1 `make build` 通过(理论上无影响,仅 sanity check)
- [x] 3.2 `make lint` 通过
- [x] 3.3 跑 `./agent-insight init --force`,检查 `.claude/skills/agent-insight/SKILL.md` 内容
- [x] 3.4 diff 行数 < 250 行新增(避免 SKILL.md 过长)
- [x] 3.5 人工 review:
  - 决策点准确性(基于 testdata 跑过验证命令)
  - 陷阱提示合理性(无 AI 误导)
  - 模式识别 JSON 示例真实性(从实际输出剪贴)
  - 决策树表覆盖度(8-10 个场景完整)

## 4. 收尾

- [x] 4.1 git commit (message: `feat: 强化 SKILL.md 决策化文档`)
- [x] 4.2 跑 `/opsx:archive` 归档 change 到 `openspec/specs/`
- [x] 4.3 验证归档后 `openspec/changes/add-skill-md-enhancement/` 移入 `archive/`
- [x] 4.4 跑 `openspec validate --strict` 确认无 schema 违规
