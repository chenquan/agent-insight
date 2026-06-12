## ADDED Requirements

### Requirement: traces 命令展示原始采样调用链

系统 SHALL 提供 `traces` 命令，接受一个 profile 文件路径和一个可选的正则 pattern，输出匹配的原始采样调用链。

每条 trace SHALL 展示：
- 完整调用路径（从根到叶的函数名序列）
- 该采样的采样值

#### Scenario: 查看所有调用链
- **WHEN** 用户执行 `agent-insight traces profile.pb.gz`
- **THEN** 输出所有采样的完整调用链及其值，按值降序排列

#### Scenario: 按 pattern 过滤调用链
- **WHEN** 用户执行 `agent-insight traces profile.pb.gz "runtime.mallocgc"`
- **THEN** 只输出调用栈中包含匹配 pattern 函数的采样链

### Requirement: traces 命令支持过滤和限制

系统 SHALL 支持以下 flags：
- `--focus pattern`：正则，只展示调用栈中包含匹配函数的 trace
- `--ignore pattern`：正则，排除调用栈中包含匹配函数的 trace
- `--top N`：限制输出条数，默认 20
- `--value-type`：指定值类型（多值 profile）
- `--format`：输出格式，默认 text，可选 json/markdown

#### Scenario: 使用 focus 过滤
- **WHEN** 用户执行 `agent-insight traces profile.pb.gz --focus "main.*"`
- **THEN** 只输出包含 main 包函数的调用链

#### Scenario: 使用 top 限制输出
- **WHEN** 用户执行 `agent-insight traces profile.pb.gz --top 5`
- **THEN** 最多输出 5 条调用链（值最大的 5 条）

#### Scenario: 使用 ignore 排除
- **WHEN** 用户执行 `agent-insight traces profile.pb.gz --ignore "runtime.*"`
- **THEN** 排除调用栈中包含 runtime 函数的 trace

### Requirement: traces 命令参数验证

系统 SHALL 验证：恰好一个 profile 文件路径，pattern 为可选参数，format 必须为 text/json/markdown。

#### Scenario: 无效正则 pattern
- **WHEN** 用户执行 `agent-insight traces profile.pb.gz "[invalid"`
- **THEN** 返回正则编译错误
