## ADDED Requirements

### Requirement: Merge multiple profile files
系统 SHALL 提供 merge 子命令，支持将多个同类 pprof profile 文件合并为一个输出文件。

#### Scenario: Merge two profile files
- **WHEN** 用户执行 `agent-insight merge cpu1.pb.gz cpu2.pb.gz -o merged.pb.gz`
- **THEN** 系统将两个 profile 合并并写入 `merged.pb.gz`

#### Scenario: Merge more than two profile files
- **WHEN** 用户执行 `agent-insight merge a.pb.gz b.pb.gz c.pb.gz d.pb.gz -o out.pb.gz`
- **THEN** 系统将所有 profile 合并并写入 `out.pb.gz`

### Requirement: Input profile count validation
系统 SHALL 要求至少 2 个输入 profile 文件。

#### Scenario: Insufficient input files
- **WHEN** 用户只提供 1 个输入文件
- **THEN** 系统报错并提示至少需要 2 个 profile 文件

#### Scenario: Single input from directory with only one profile
- **WHEN** 用户指定目录且目录下只有 1 个 profile 文件
- **THEN** 系统报错并提示至少需要 2 个 profile 文件

### Requirement: Profile type consistency validation
系统 SHALL 校验所有输入 profile 的类型一致。

#### Scenario: Matching profile types
- **WHEN** 所有输入 profile 的 PeriodType 相同
- **THEN** 系统正常执行合并

#### Scenario: Mixed profile types
- **WHEN** 输入 profile 中存在不同的 PeriodType（如 cpu 混合 heap）
- **THEN** 系统报错并列出冲突的类型信息

#### Scenario: Profiles without PeriodType
- **WHEN** 某个输入 profile 没有 PeriodType
- **THEN** 系统跳过该 profile 的类型校验，不报错

### Requirement: Directory input mode
系统 SHALL 支持指定目录作为输入，自动发现目录下的 pprof 文件。

#### Scenario: Directory with multiple profiles
- **WHEN** 用户指定一个包含多个 `.pb.gz` 和 `.pb` 文件的目录
- **THEN** 系统递归扫描子目录，自动发现所有 pprof 文件并执行合并

#### Scenario: Directory with mixed profile types
- **WHEN** 目录下存在多种类型的 profile
- **THEN** 系统报错提示类型不一致

#### Scenario: Empty directory
- **WHEN** 用户指定一个不包含 pprof 文件的目录
- **THEN** 系统报错提示目录下没有找到 profile 文件

### Requirement: Output file format
系统 SHALL 将合并结果输出为 gzip 压缩的 pprof protobuf 格式（`.pb.gz`）。

#### Scenario: Successful merge output
- **WHEN** 合并成功完成
- **THEN** 输出文件为有效的 gzip 压缩 pprof profile，可被现有 analyze/diff/flame 等命令正常读取

### Requirement: Output file flag required
系统 SHALL 要求用户通过 `-o`/`--output` flag 指定输出文件路径。

#### Scenario: No output flag provided
- **WHEN** 用户未指定 `-o` flag
- **THEN** 系统报错并提示必须指定输出文件路径

### Requirement: Merge summary output
系统 SHALL 在合并完成后输出简要的合并统计信息（输入文件数、总采样数、输出文件路径）。

#### Scenario: Successful merge
- **WHEN** 合并成功完成
- **THEN** 系统在 stdout 输出合并摘要，包含输入文件数量、合并后的采样总数和输出文件路径
