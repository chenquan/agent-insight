## Context

agent-insight 是为 AI 编码助手设计的 pprof 分析 CLI 工具，现有 analyze/tree/traces 等命令输出结构化数据（热点、调用树、调用路径等）。这些数据由上层 AI 助手（如 Claude Code）读取并自行理解给出诊断建议。

当前的问题：
1. AI 助手拿到原始结构化数据后需要自行判断诊断方向，不同 profile 类型的关注点不同但数据格式统一
2. 不同编程语言的优化手段差异很大（Go 的 sync.Pool vs C++ 的 custom allocator vs Rust 的所有权优化）
3. AI 助手可能遗漏某些 profile 类型特有的诊断维度

现有架构（三层分离）：
- `commands/` — cobra 命令层，参数解析、调用 profile 层、选 output formatter
- `profile/` — 核心计算层，读取 pprof、做分析/对比/树构建
- `output/` — 输出层，text/json/markdown 三种 formatter

## Goals / Non-Goals

**Goals:**
- 新增 `diagnose` 命令，将 profile 分析数据组装成高质量诊断 prompt
- 自动检测编程语言（Go/C++/Rust/Java/C/Unknown）
- 按 profile 类型 × 语言维度生成针对性诊断引导
- 保持三层分离架构，零额外外部依赖
- 支持 `--context` 传入用户上下文，`--top N` 控制数据量，`--format` 控制输出格式

**Non-Goals:**
- 不调用任何 LLM API（最终由 Claude Code 与大模型交互）
- 不做交互式诊断（单次命令输出 prompt）
- 不生成具体的代码修复建议（那是 AI 助手的事）
- 不支持除 Go/C++/Rust/Java/C 之外的语言检测

## Decisions

### 1. Prompt 构建在 profile 层而非 commands 层

**选择**: 在 `pkg/profile/diagnose.go` 实现 prompt 构建逻辑

**理由**: prompt 构建是核心计算逻辑（语言检测 + 引导模板拼装 + 数据组装），属于 profile 层职责。commands 层只做参数解析和输出格式选择，符合三层分离原则。

**备选**: 在 commands 层直接拼装 → 违反分层规则，commands 不做计算。

### 2. 引导模板采用两层拼装策略

**选择**: 基础引导（profile 类型维度）+ 语言追加（语言维度）

**理由**: 如果为每种 profile 类型 × 语言组合硬编码模板，需要 5×5=25 套模板，维护成本高。两层拼装只需 5 套基础引导 + 5 套语言模板 = 10 个组件，扩展语言只需加一份语言模板。

### 3. 语言检测基于函数名和文件名模式匹配

**选择**: 遍历 profile 中的 Function.Name 和 Function.Filename，用正则匹配已知语言模式，取匹配数量最多的语言作为结果。

**理由**: pprof profile 本身不包含语言标识，函数命名规范是最可靠的信号。统计匹配数量取众数可以处理混合语言程序（如 Go 调用 CGO）。

**检测规则**:
- Go: `runtime.`、`main.`、`encoding/` 等标准库前缀，或 `*.go` 文件
- C++: `std::`、`::` 双冒号模式，或 `*.cpp`/`*.cc`/`*.h` 文件
- Rust: `core::`、`alloc::`、`<T as Trait>::` 模式，或 `*.rs` 文件
- Java: `java.lang.`、`com.`、`org.` 包名模式，或 `*.java` 文件
- C: 无命名空间但 `*.c`/`*.h` 文件（在排除 C++ 后）
- Unknown: 无法匹配任何模式

### 4. 数据量控制策略

**选择**:
- 热点函数: `--top N` 控制，默认 10
- 调用树: 只保留 cum >= 1% 的路径，深度限制 5 层
- Traces: 取 value 最高的 5 条

**理由**: prompt 不能过长浪费 context window，但也不能太短丢失关键信息。这个裁剪策略覆盖绝大部分性能问题。

### 5. 输出格式

**选择**: text（默认，prompt 纯文本）、markdown（结构化 prompt）、json（prompt 字段 + 原始分析数据字段）

**理由**: text 最简单直接，Claude Code 把 stdout 当上下文阅读。markdown 适合需要结构化展示。json 方便程序化消费和 AI 深挖。

## Risks / Trade-offs

- **语言检测不准确** → 对于符号缺失的 profile，可能检测为 Unknown，此时使用通用引导，不影响基本诊断
- **引导模板不够精准** → 模板是静态的，可能跟不上语言/运行时的演进。缓解：模板作为 Go const 字符串，易于更新
- **prompt 过长** → 大型 profile 的数据部分可能很长。缓解：--top N 和深度限制控制数据量，默认值经过调优
- **两层拼装的引导可能不够连贯** → 基础引导和语言追加拼接可能不如手工编写的整体引导流畅。缓解：在实现时仔细调校拼接方式和衔接词
