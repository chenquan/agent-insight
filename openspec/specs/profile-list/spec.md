## Requirements

### Requirement: Query function by regex pattern
The system SHALL allow users to query specific functions using regular expression patterns.

#### Scenario: List exact function match
- **WHEN** user provides "mainhandleRequest" as the function pattern
- **THEN** system displays all functions matching the pattern with their performance data

#### Scenario: List wildcard function match
- **WHEN** user provides "encoding.*" as the function pattern
- **THEN** system displays all functions in the encoding package with their performance data

#### Scenario: No matches found
- **WHEN** user provides a pattern that matches no functions
- **THEN** system outputs a clear message indicating no functions matched the pattern

### Requirement: Display caller information
The system SHALL display which functions call the target function (callers) along with their contribution.

#### Scenario: Show direct callers
- **WHEN** user queries a function that has multiple callers
- **THEN** system lists each caller function with its flat and cumulative contribution to the target

#### Scenario: Show caller percentages
- **WHEN** displaying caller information
- **THEN** each caller entry includes the percentage of total samples attributed through that call path

#### Scenario: Format caller output
- **WHEN** outputting caller information in text format
- **THEN** output uses indentation to show the call hierarchy and aligns numerical columns

### Requirement: Display callee information
The system SHALL display which functions are called by the target function (callees) along with their contribution.

#### Scenario: Show direct callees
- **WHEN** user queries a function that calls multiple other functions
- **THEN** system lists each callee function with its flat and cumulative contribution from the target

#### Scenario: Show callee percentages
- **WHEN** displaying callee information
- **THEN** each callee entry includes the percentage of the target's samples attributed to that callee

#### Scenario: Recursive call detection
- **WHEN** the target function recursively calls itself
- **THEN** system clearly indicates the recursive nature in the output

### Requirement: Support output format options
The system SHALL support text, JSON, and markdown output formats for the list command.

#### Scenario: Text format output
- **WHEN** user runs list without --format flag or with --format text
- **THEN** system outputs formatted text showing caller/callee relationships with aligned columns

#### Scenario: JSON format output
- **WHEN** user specifies --format json
- **THEN** system outputs JSON with structure: {function, callers[], callees[]} containing full performance data

#### Scenario: Markdown format output
- **WHEN** user specifies --format markdown
- **THEN** system outputs formatted markdown with tables for callers and callees

### Requirement: Filter and sort list results
The system SHALL support filtering and sorting options for the list output.

#### Scenario: Limit output depth
- **WHEN** user specifies --depth 3
- **THEN** system shows only up to 3 levels of caller/callee relationships

#### Scenario: Show only callers
- **WHEN** user specifies --callers-only flag
- **THEN** system displays only caller information and excludes callees

#### Scenario: Show only callees
- **WHEN** user specifies --callees-only flag
- **THEN** system displays only callee information and excludes callers

### Requirement: Handle inline and leaf functions
The system SHALL correctly handle functions that are either leaf nodes (no callees) or deeply inline.

#### Scenario: Leaf function listing
- **WHEN** user queries a leaf function with no callees
- **THEN** system indicates "No callees" or similar message in the callees section

#### Scenario: Inline function expansion
- **WHEN** the target function contains inlined calls
- **THEN** system shows the inlined functions as part of the call tree with appropriate attribution

### Requirement: Support negative matching
The system SHALL support excluding functions from the list output using negative patterns.

#### Scenario: Exclude runtime functions
- **WHEN** user specifies --exclude "runtime.*" 
- **THEN** system excludes all runtime package functions from the caller/callee listings

#### Scenario: Combine include and exclude
- **WHEN** user provides both function pattern and --exclude flag
- **THEN** system first matches by function pattern, then applies exclusions to results

### Requirement: Query by Location ID when symbols unavailable
The system SHALL allow users to query functions by Location ID when symbol information is not available.

#### Scenario: Query by Location ID
- **WHEN** profile lacks function symbols and user provides location ID pattern
- **THEN** system matches locations by their numeric IDs
- **AND** displays results using Location IDs and memory addresses

#### Scenario: Query by memory address
- **WHEN** user provides memory address pattern (e.g., "0x1234")
- **THEN** system matches locations by their memory addresses
- **AND** includes module information when available for context

#### Scenario: Mixed pattern matching
- **WHEN** profile has both symbols and Location IDs
- **THEN** user can query using either function name patterns or Location ID patterns
- **AND** system correctly handles both types of queries in the same execution

### Requirement: Filter by pprof labels

`list` 命令 SHALL 支持 `--tag key=value` 和 `--tag-ignore key=value` flag，对 profile 的 sample 做 label 维度过滤，再在过滤后的 sample 上做调用关系分析。

`list` 中的 `--tag-ignore` SHALL **仅** 用于 label 过滤（与 `analyze` / `traces` / `diff` 中 `--tag-ignore` 语义一致），不接受正则字符串。函数名正则排除改用 `--ignore-function`（见下一条 Requirement）。

- `--tag` 可重复多次
- 同 key 多次 → OR
- 跨 key → AND
- 数字 label value 必须是十进制整数字符串
- 0 样本匹配时退出并报错

#### Scenario: 单 tag 过滤
- **WHEN** 用户跑 `agent-insight list goroutine.pb.gz "Query" --tag state=blocked`
- **THEN** 命令分析 state=blocked 样本中的 "Query" 函数调用关系

#### Scenario: 同 key OR
- **WHEN** 用户跑 `agent-insight list goroutine.pb.gz "Query" --tag state=blocked --tag state=running`
- **THEN** 命令分析 state 是 blocked 或 running 的样本

#### Scenario: 跨 key AND
- **WHEN** 用户跑 `agent-insight list goroutine.pb.gz "Query" --tag state=blocked --tag wait_reason=IO`
- **THEN** 命令分析同时满足两个条件的样本

#### Scenario: --tag-ignore 排除 label
- **WHEN** 用户跑 `agent-insight list goroutine.pb.gz "Query" --tag-ignore state=running`
- **THEN** 命令排除所有 state=running 的 sample

#### Scenario: --tag-ignore 不接受正则字符串
- **WHEN** 用户跑 `agent-insight list goroutine.pb.gz "Query" --tag-ignore "database.*"`
- **THEN** 报 "invalid --tag-ignore value 'database.*': missing '=key' format"（因为没有 `=`）

#### Scenario: 0 样本退出
- **WHEN** 用户跑 `agent-insight list cpu.pb.gz "main.*" --tag state=blocked` 且 cpu.pb.gz 无 state label
- **THEN** 命令以非零状态退出，错误信息含 "tag filter matched 0 of N samples"

### Requirement: Rename --exclude to --ignore-function

`list` 命令 SHALL 把 `--exclude pattern` flag 改名为 `--ignore-function pattern`。语义不变（正则排除匹配函数），仅 flag 名称改变。

`--ignore-function` 是 `list` 命令独有的（其他命令的样本过滤用 `--tag-ignore` 走 label 维度），两个 flag SHALL 完全独立、互不重叠。

迁移路径：
- v0.X 之前：`--exclude "database.*"`
- v0.X 之后：`--ignore-function "database.*"`

旧 `--exclude` flag SHALL 不再被识别。

#### Scenario: 旧 --exclude 不再工作
- **WHEN** 用户跑 `agent-insight list cpu.pb.gz "main.*" --exclude "database.*"`
- **THEN** 报 "unknown flag: --exclude"

#### Scenario: 新 --ignore-function 等价于旧 --exclude
- **WHEN** 用户跑 `agent-insight list cpu.pb.gz "main.*" --ignore-function "database.*"`
- **THEN** 命令行为与 v0.X 之前的 `--exclude "database.*"` 完全一致（正则排除匹配函数）

#### Scenario: --ignore-function 接受正则字符串
- **WHEN** 用户跑 `agent-insight list cpu.pb.gz "main.*" --ignore-function "database\\..*"`
- **THEN** 命令正确按正则排除函数

#### Scenario: --ignore-function 与 --tag-ignore 共存
- **WHEN** 用户跑 `agent-insight list goroutine.pb.gz "Query" --ignore-function "runtime.*" --tag-ignore state=running`
- **THEN** 先按 `--tag-ignore` 过滤 sample（排除 state=running），再在结果上按 `--ignore-function` 过滤函数（排除 runtime.*）
