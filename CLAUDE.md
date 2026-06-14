# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## 要求
- 使用中文交流
- 所有变更必须使用 openspec 流程完成，至少包含：propose -> apply -> archive
- 使用 github.com/google/pprof/profile 解析 pprof 性能文件
- 该工具是给AI使用的，以AI 大模型的使用体验优先，可不考虑人类使用

## 常用命令

```bash
make build          # 构建二进制到 ./
make test           # 运行所有测试
make lint           # 运行 golangci-lint
make clean          # 清理构建产物
./agent-insight info testdata/cpu.pb.gz         # 快速验证命令
./agent-insight analyze testdata/cpu.pb.gz --format json
```

## 架构（三层分离）

```
main.go
  └─ cmd/root.go          注册所有 cobra 子命令
       └─ pkg/commands/   命令层：参数解析、参数校验、调用 profile 层、调用 output 层
            ├─ pkg/profile/   核心计算层：读取 pprof、做分析/对比/树构建
            ├─ pkg/output/    输出层：text/json/markdown 三种 formatter（单文件 formatter.go）
            └─ pkg/skill/     init 命令依赖：嵌入的 SKILL.md 模板
```

**分层规则：**
- `commands/` 不做计算，只解析 flag、调 profile 层、选 output formatter
- `profile/` 不做格式化，只返回结构化结果（`XxxResult` struct）
- `output/` 不知道命令行存在，只接收 `*profile.XxxResult` 渲染到 writer
- 新增子命令 = 在 `commands/` 加 cobra 命令 + `profile/` 加核心逻辑 + `output/` 加 formatter + `cmd/root.go` 注册

## 命令清单（11 个）

| 命令 | 用途 | 关键约束 |
|---|---|---|
| `info` | profile 元信息（零计算） | 单文件 |
| `analyze` | 热点 flat/cum 分析 | 单文件，支持 `--value-type` |
| `list` | 函数调用关系查询 | 单文件 + 正则 |
| `flame` | 折叠栈（folded stack） | 输出给 flamegraph.pl |
| `traces` | 单 sample 调用链 | 与 flame 互补 |
| `tree` | 层级调用树 | 可控深度 |
| `diff` | 双 profile 对比 | 严格 2 参数 |
| `trend` | 多 profile 线性回归趋势 | **至少 3 个 profile**；支持目录递归发现 `.pb` / `.pb.gz` |
| `merge` | 多 profile 合并 | **输出 `.pb.gz` 文件**（`-o`），不是格式化文本 |
| `init` | 生成 `.claude/skills/agent-insight/SKILL.md` | 用 `--force` 覆盖 |
| `diagnose` | 生成 AI 诊断提示 | 调用 `language.go` 检测编程语言；嵌入 `Analysis` + `TreeResult` + `TracesResult` 三段数据 + profile 类型专属 guidance |

## 关键约定

**pprof 数据访问：** 通过 `github.com/google/pprof/profile` 库，不要自己解析 protobuf。`Sample.Location` 索引 0 是叶子、最后一个是根。`pkg/profile/loader.go` 同时支持 `.pb.gz`（gzip 压缩）和 `.pb`（原始 protobuf）。

**共享校验逻辑：** `pkg/commands/validate.go` 提供 `ValidateFormat(format)` 和 `ValidateRegex(pattern, name)`。新命令一律复用这两个函数，不要在校验代码里重写 format 分支或裸 `regexp.Compile`。

**多值 profile：** heap profile 有 `alloc_objects`、`alloc_space`、`inuse_objects`、`inuse_space` 四种值类型。`profile/analysis.go` 的 `selectDefaultValueType` 处理智能默认。`--value-type` flag 允许用户覆盖。

**符号缺失降级：** 测试和生产环境的 profile 经常没有函数符号。Hotspot/FunctionInfo 等结构用 `*string` 指针字段，无符号时 fallback 到 `LocationID` + `Address` + `Module`。

**Profile 类型推断：** `PeriodType` 缺失时调用 `inferProfileType(p)` 兜底，`diagnose` 命令依赖这个类型字符串选择 `baseGuidance` 中的专属 guidance 模板（cpu/heap/goroutine/contentions/thread 等）。`diagnose` 还调用 `pkg/profile/language.go` 从函数文件路径推断编程语言，决定 guidance 的语言版本（Go / C++ / Rust 等）。

**AI 友好输出：** 所有 Result 结构都支持 `--format text|json|markdown`。JSON 是默认推荐（LLM 解析最稳定）。每个 Result 类型对应 `output/formatter.go` 中的三组 formatter。

**OpenSpec 流程：** 任何代码变更都先 `/opsx:explore` → `/opsx:propose` → `/opsx:apply` → `/opsx:archive`。不要直接写代码。每个 change 放在 `openspec/changes/<name>/`，archive 时同步到 `openspec/specs/`。

## 测试

- 测试数据：`testdata/cpu.pb.gz` 和 `testdata/heap.pb.gz`（由 `testdata/generate.go` 生成）
- 单测试：`go test -v -run TestName ./pkg/profile/`
- benchmark：`go test -bench=. ./pkg/profile/`
- **测试数据源**：`github.com/google/pprof@v0.0.0-20260604005048-7023385849c0/profile/testdata`
## 技能集成

`pkg/skill/template.md` 是 `init` 命令嵌入的模板，生成 `.claude/skills/agent-insight/SKILL.md`。**新增子命令时必须同步更新这个模板**，否则 Claude Code 不会知道新命令的存在。
