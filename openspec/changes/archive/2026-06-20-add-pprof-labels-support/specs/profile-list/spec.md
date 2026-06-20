## MODIFIED Requirements

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
