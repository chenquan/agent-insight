## Context

当前项目为全新 Go CLI 工具，目标是提供一个轻量级、零配置的 pprof 分析工具。现有解决方案如 `go tool pprof` 提供交互式界面但不适合 AI 直接解析，而 Python 工具链需要复杂的环境配置。本工具专为 Claude Code 等 AI 编码助手设计，输出结构化数据以减少解析复杂度。

**约束条件**：
- 必须是单二进制文件，无需外部依赖
- 核心依赖仅 `github.com/google/pprof/profile`
- 支持大文件流式解析，内存占用可控
- 输出格式必须 AI 友好（JSON 结构清晰）

## Goals / Non-Goals

**Goals:**
- 创建单二进制 CLI 工具 `agent-insight`
- 支持 pprof protobuf 格式文件的解析和分析
- 提供 analyze、list、flame、diff 四个核心命令
- 支持多种输出格式（text、json、markdown）
- 实现流式解析以支持大文件

**Non-Goals:**
- 不提供交互式界面或 Web UI
- 不支持图形化火焰图生成（仅输出折叠栈格式）
- 不实现 pprof 的所有可视化功能
- 不支持实时 profile 监控（仅分析已采集的文件）

## Decisions

### 1. 语言选择：Go

**理由**：
- `github.com/google/pprof/profile` 是 Go 官方库，原生支持最佳
- Go 编译为单静态二进制，符合零配置要求
- 优秀的并发和流式处理能力，适合大文件解析

### 2. 命令结构：子命令模式

采用 `agent-insight <command> [flags]` 结构：
- `analyze`: 主要分析命令，输出热点和摘要
- `list`: 函数级查询，查看调用关系
- `flame`: 生成折叠栈格式
- `diff`: 对比两个 profile

**替代方案考虑**：单一命令加模式参数（如 `--mode analyze`）。 rejected 因为子命令更符合 CLI 惯例，参数更清晰。

### 3. 输出格式设计

**JSON 格式关键决策**：
```json
{
  "type": "cpu|heap|goroutine...",
  "top": [{"function": "...", "flat": N, "cum": N}],
  "paths": ["stack;trace;here 12.1%"],
  "summary": "自然语言描述"
}
```

- 使用 `flat`/`cum` 而非 `flat_value`/`cumulative` 以保持简洁
- 路径使用分号分隔，紧凑且易解析
- 包含 `summary` 字段，AI 可直接展示给用户

### 4. 流式解析策略

使用 `profile.Parse` 的 io.Reader 接口：
- 不一次性加载整个文件到内存
- 对大文件（>1GB）保持可控内存占用
- 在解析过程中即时计算聚合数据

### 5. 过滤和排序

- 支持 `--focus`（正则包含）和 `--ignore`（正则排除）
- 默认按 `flat` 值排序，`--cum` 切换为累积排序
- `--top N` 限制输出条数，避免上下文溢出

### 6. 符号信息缺失处理策略

**背景**：实际 profile 文件经常缺少函数符号信息，仅有 Location IDs 和内存地址。

**策略**：
- **优先使用符号信息**：当 Function 信息存在时，显示函数名和文件名
- **降级到 Location ID**：无符号信息时，显示 Location ID 和内存地址
- **Module 信息推断**：从 Mapping 中提取二进制文件名作为上下文
- **优雅降级**：JSON 输出结构保持一致，仅填充可用字段

**JSON 输出适配**：
```json
// 有符号信息时
{
  "function": "runtime.mallocgc",
  "file": "runtime/malloc.go:1020",
  "module": null,
  "location_id": null
}

// 无符号信息时  
{
  "function": null,
  "file": null, 
  "module": "/lib/libc-2.15.so",
  "location_id": 19,
  "address": "0x430bac"
}
```

### 7. 多值类型处理

**背景**：不同 profile 类型有不同的值结构：
- CPU: 单值 `[cpu_nanoseconds]`
- Heap: 双值 `[objects, bytes]`  
- Go Heap: 四值 `[alloc_objects, alloc_space, inuse_objects, inuse_space]`

**策略**：
- **自动检测值类型**：从 Profile.SampleType 推断可用值
- **默认值选择**：
  - CPU profile: 使用 cpu_nanoseconds
  - Heap profile: 默认使用 inuse_bytes（关注内存占用）
  - Goroutine profile: 使用 count
- **用户可覆盖**：通过 `--value-type` 参数明确指定使用哪个值
- **完整信息保留**：JSON 中包含所有可用值，供深度分析

## Risks / Trade-offs

### [Risk] 大文件内存占用

**风险**：profile 文件可能达到数 GB，解析时内存溢出。

**缓解措施**：
- 使用 io.Reader 流式读取
- 在解析过程中即时过滤和聚合，不保留完整 Sample 列表
- 对超大文件提供 `--sample-limit` 参数强制降采样

### [Risk] pprof 格式兼容性

**风险**：未来 pprof 格式变化导致解析失败。

**缓解措施**：
- 依赖官方 `github.com/google/pprof/profile` 库，跟随官方更新
- 在 CI 中测试不同 Go 版本生成的 profile 文件
- 提供清晰的错误信息，提示用户使用兼容的 `go tool pprof` 版本

### [Trade-off] 功能简化

**权衡**：简化功能意味着无法处理某些边缘情况（如罕见 profile 类型）。

**接受理由**：
- 目标用户是 AI 编码助手，主要关注常见瓶颈（CPU、内存、goroutine）
- 复杂分析仍可通过 `go tool pprof` 完成
- 保持工具轻量级，避免过度工程化

### [Risk] JSON 输出稳定性

**风险**：JSON 结构变化可能破坏依赖它的 AI 流程。

**缓解措施**：
- 在正式发布前冻结 JSON schema
- 提供 `--version` 和 schema 版本字段
- 文档中明确承诺向后兼容性策略

### [Risk] 符号信息缺失影响可用性

**风险**：生产环境 profile 经常缺少符号信息，导致输出仅有 Location ID，降低工具的实用性。

**缓解措施**：
- 在 JSON 中同时提供符号信息和 Location 信息（如果都可用）
- 输出 Module/Mapping 信息提供额外上下文
- 在 summary 中说明符号信息的可用性
- 提供清晰的建议："如需符号信息，请确保在采集时包含调试符号"

### [Risk] 多值类型复杂度

**风险**：不同 profile 类型的多值结构增加了解析和展示的复杂度，可能导致用户混淆。

**缓解措施**：
- 提供智能默认值选择（如 heap 默认使用 inuse_bytes）
- 在 JSON 中包含所有可用值，供高级用户使用
- 在输出中明确标注当前使用的值类型
- 文档中详细说明不同 profile 类型的值结构
