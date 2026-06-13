## Context

`agent-insight` 的 `init` 子命令生成 `pkg/skill/template.md` 的内容到 `.claude/skills/agent-insight/SKILL.md`,供 Claude Code 自动加载和触发。

当前 `pkg/skill/template.md` (330 行) 结构:
- 描述 + 触发关键词 (frontmatter)
- 何时使用 (10 场景的 table)
- 命令速查 (8 个命令的 flags 列表)
- 典型工作流 (5 个场景: 快速概览/CPU/内存/调用路径/版本对比)
- 输出解读 (各命令 JSON 字段说明)
- 注意事项 (5 条)

本 change 之前讨论过"过滤/分类/insights"等多个方向,最终用户决定**纯文档强化 SKILL.md**,0 代码改动,直接提升 AI agent 的实际使用体验。

核心改动:
- 强化"典型工作流"段,加"决策点"和"陷阱提示"
- 新增"决策树"段
- 强化"输出解读"段,加"模式识别"启发式

## Goals / Non-Goals

**Goals:**
- AI agent 拿到 SKILL.md 后,能直接按"决策树"和"工作流"执行,无需自行推断
- 现有"何时使用" / "命令速查" / "注意事项"段保留并适当增强
- 单文件改动 (`pkg/skill/template.md`),git diff 清晰可 review
- 与现有 Go 代码无任何关联,纯 markdown 改动

**Non-Goals:**
- 不修改 Go 代码
- 不修改 `init` 命令的行为
- 不改变 `init` 命令的 flag、输出路径、覆盖逻辑
- 不增加新的命令或 flag
- 不修改其他文档 (README 等)

## Decisions

### 1. SKILL.md 段顺序保持稳定

**决策**: 维持现有的段顺序,只在合适位置插入/改写。

```
1. 描述 + 触发关键词 (frontmatter)         ← 不变
2. 何时使用 (10 场景的 table)               ← 不变
3. 命令速查 (8 个命令的 flags 列表)         ← 不变
4. 典型工作流 (5 个场景)                    ← 强化 (A)
5. 决策树 (NEW)                             ← 新增 (B)
6. 输出解读 (各命令 JSON 字段说明)          ← 强化 (C)
7. 注意事项 (5 条简短提示)                  ← 不变
```

**理由**: Claude Code 已经按这个顺序理解 SKILL.md,打乱顺序会破坏向后兼容。

### 2. A 段:典型工作流加 "决策点" 和 "陷阱提示"

**决策**: 每个工作流的每一步,加两段辅助说明:
- **决策点**: "看到 X 改用 Y" —— 给出基于步骤结果的分支选择
- **陷阱提示**: 常见误用和反模式

示例 (CPU 性能分析 步骤 2):
```
2. agent-insight list cpu.pb.gz "<hotspot-func>" --format json
   → 查看热点函数的调用者和被调用者
   
   决策点:
   • top 是 leaf 函数 (runtime.*) → 跳到步骤 3 用 flame 看完整栈
   • top 是 internal 函数 (业务代码) → 它是 caller, 不是 self-bottleneck
     用 list 找它的 caller, 找到真正的触发者
   
   陷阱提示:
   • 别只看 flat 排序, internal 函数的 cum 高但 flat 低
     这通常是 caller, 不是真正的热点
   • runtime.* 占比高不等于 runtime 是问题, 它可能只是
     被某段代码高频触发
```

**5 个工作流全部强化**:
- 快速概览 (1 步)
- CPU 性能分析 (3 步)
- 内存分析 (3 步)
- 调用路径追踪 (2 步)
- 版本对比 (1 步)

### 3. B 段:决策树表格 (新增)

**决策**: 表格形式,8-10 个常见诊断场景的快速参考,放在 A 段后、C 段前。

格式:
```markdown
## 决策树: 看到 X → 跑 Y

| 你看到 | 下一步 |
|--------|--------|
| cpu profile + leaf bottleneck | analyze --format json --cum → flame |
| analyze top 都是 internal (caller) | list <func> --callers-only |
| heap + 内存增长 | alloc_space (看分配源) |
| heap + 当前占用高 | inuse_space (看大对象) |
| heap + 大量小对象 | alloc_objects |
| goroutine + count > 10000 | 怀疑泄漏, traces 看 top stack |
| diff + 看到新函数 | list 验证是不是真新 |
| 命令报 "no samples matched" | 放宽 --focus, 或去掉 |
| 命令报 "unknown value-type" | info 看 available types |
| 命令报 "failed to load profile" | info 验证文件能解析 |
```

**8-10 个场景**: 覆盖 CPU/内存/goroutine/diff 四类核心场景 + 3 个常见错误恢复。

### 4. C 段:输出解读加 "模式识别" 启发式

**决策**: 为 analyze / diff 命令的输出加"模式识别"段,用实际 JSON 示例说明。

**analyze flat vs cum 4 种模式**:
```markdown
### analyze 模式识别 (flat vs cum)

- `flat` 高 + `cum` 高 → **self-bottleneck**: 函数本身慢,需要优化函数实现
- `flat` 低 + `cum` 高 → **caller bottleneck**: 函数被很多 caller 触发,优化此函数影响有限,要看 caller
- `flat` 高 + `cum` = `flat` → **leaf**: 叶子函数,没被谁调,自己耗时
- `flat` = 0 + `cum` 高 → **pure caller**: 纯调用方,自己没耗时,但所有耗时都从它过
```

**diff delta 解读**:
```markdown
### diff 模式识别

- `delta_percent > 0` → **regression**: target 比 base 慢
- `delta_percent < 0` → **improvement**: target 比 base 快
- `new_functions[]` 非空 → **new hotspots**: base 中没有的函数
- `deleted_functions[]` 非空 → **gone hotspots**: base 中有但 target 中消失
```

**实际 JSON 示例**: 给一个真实的 cpu.pb.gz 输出片段,标注 "这是 self-bottleneck" / "这是 caller"。

### 5. 触发关键词 (frontmatter description) 不变

**决策**: 不修改 frontmatter 的 description 字段。

**理由**: 当前关键词已覆盖核心场景 (pprof、CPU、内存、火焰图等),本次增强是"SKILL.md 内部",不影响外部触发。

### 6. 注意事项段保持

**决策**: 现有 5 条注意事项保留,可在尾部加 1-2 条关于决策树/工作流的提示。

## Risks / Trade-offs

**[Risk] SKILL.md 变得过长,Claude Code 加载开销增大** → Mitigation: 控制在 200 行新增,总长度 ~500 行;Claude Code 加载 markdown 文件是流式,影响小。

**[Risk] 决策点和陷阱提示不准确,误导 AI** → Mitigation: 人工 review 准确性,基于现有命令真实行为;基于 testdata 跑过验证。

**[Risk] 模式识别启发式过度简化** → Mitigation: 给出"通常情况"而非"绝对规则",文中明确 "通常"、"大多数情况" 等限定词。

**[Risk] 决策树表不完整,新增场景覆盖不到** → Mitigation: 决策树是 8-10 个高频场景,非穷举;典型工作流段补足具体场景。

**[Trade-off] 文档质量依赖作者经验** → 接受: 作者对 agent-insight 命令族和典型 AI 工作流有充分理解;Phase archive 时 review。

## Migration Plan

无破坏性变更:
- 单文件改动 (`pkg/skill/template.md`)
- 用户跑 `agent-insight init --force` 即可获得增强版 SKILL.md
- 不影响 Go 代码、`init` 命令行为、其他命令
- git revert 单 commit 即可回滚

**部署步骤**:
1. 改写 `pkg/skill/template.md`
2. 跑 `agent-insight init --force` 验证生成的文件
3. 人工 review 生成的 `.claude/skills/agent-insight/SKILL.md`
4. 提交 PR,archive change

**回滚策略**:
- git revert 单 commit 即可
- 无需数据迁移
- 用户本地 `.claude/skills/agent-insight/SKILL.md` 可重新 init 恢复

## Open Questions

- **决策树表应放多少场景**: v1 选 8-10 个高频,后续根据反馈增删
- **"陷阱提示"是否要给具体命令示例**: v1 给定性提示,不给具体命令 (避免冗长)
- **JSON 示例应放多少**: 每个模式 1 个真实示例 (~5-10 行),不过度
- **是否要英文版**: v1 维持中文 (与现有 SKILL.md 一致);v2 考虑 i18n