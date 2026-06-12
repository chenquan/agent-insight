## ADDED Requirements

### Requirement: info 命令展示 profile 元信息概览

系统 SHALL 提供 `info` 命令，接受一个 profile 文件路径，输出该 profile 的元信息概览，不执行采样级计算。

输出信息 SHALL 包含：
- Profile 类型（通过 PeriodType 推断，如 cpu、heap、goroutine）
- 采集时长（DurationNanos 转为人类可读格式）
- 采样周期（Period + PeriodType）
- 采样总数（len(Sample)）
- 值类型列表（SampleType 的 Type/Unit）
- 函数数量、位置数量
- 符号状态（HasFunctions、HasFileLines）
- 映射信息列表（文件路径、BuildID、符号可用性）
- 时间范围（TimeNanos 转为可读时间）
- 注释（Comments）

#### Scenario: 查看 CPU profile 元信息
- **WHEN** 用户执行 `agent-insight info cpu.pb.gz`
- **THEN** 输出包含 profile 类型为 cpu、时长、采样数、值类型列表、符号状态等信息

#### Scenario: 查看缺少符号的 profile
- **WHEN** 用户执行 `agent-insight info nosym.pb.gz` 且 profile 缺少函数符号
- **THEN** 输出显示符号状态为不可用，并展示映射文件路径和地址范围

#### Scenario: JSON 格式输出
- **WHEN** 用户执行 `agent-insight info profile.pb.gz --format json`
- **THEN** 输出为合法 JSON，包含所有元信息字段

### Requirement: info 命令支持多格式输出

系统 SHALL 支持 `--format` flag，可选值为 text（默认）、json、markdown。

#### Scenario: Markdown 格式输出
- **WHEN** 用户执行 `agent-insight info profile.pb.gz --format markdown`
- **THEN** 输出为 Markdown 格式的元信息概览

### Requirement: info 命令参数验证

系统 SHALL 验证输入参数：恰好一个 profile 文件路径，format 必须为 text/json/markdown 之一。

#### Scenario: 缺少文件参数
- **WHEN** 用户执行 `agent-insight info` 不带文件路径
- **THEN** 返回错误提示

#### Scenario: 无效格式参数
- **WHEN** 用户执行 `agent-insight info profile.pb.gz --format xml`
- **THEN** 返回错误提示有效格式列表
