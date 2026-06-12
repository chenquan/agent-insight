## ADDED Requirements

### Requirement: tree 命令展示层级调用树

系统 SHALL 提供 `tree` 命令，接受一个 profile 文件路径，构建并输出层级调用树。

调用树 SHALL 从根（调用栈最底部）到叶（调用栈最顶部）展示，每个节点包含：
- 函数名
- flat 值和百分比（直接归到该函数的值）
- cum 值和百分比（经过该函数的累计值）
- 子节点列表（按 cum 值降序排列）

#### Scenario: 查看 CPU profile 调用树
- **WHEN** 用户执行 `agent-insight tree cpu.pb.gz`
- **THEN** 输出层级缩进的调用树，展示根函数到叶函数的调用关系，每层包含 flat/cum 值

#### Scenario: 查看调用树中子节点的占比
- **THEN** 每个子节点的 cum 占其父节点 cum 的百分比 SHALL 被展示

### Requirement: tree 命令支持过滤和限制

系统 SHALL 支持以下 flags：
- `--focus pattern`：正则，只展示包含匹配函数的分支
- `--ignore pattern`：正则，排除包含匹配函数的分支
- `--depth N`：最大展示深度，默认 5
- `--top N`：每层最多展示的子节点数，默认 10
- `--cum`：按 cum 排序（默认），否则按 flat 排序
- `--value-type`：指定值类型（多值 profile）
- `--format`：输出格式，默认 text，可选 json/markdown

#### Scenario: 限制深度
- **WHEN** 用户执行 `agent-insight tree profile.pb.gz --depth 3`
- **THEN** 调用树最多展示 3 层深度

#### Scenario: 限制每层子节点数
- **WHEN** 用户执行 `agent-insight tree profile.pb.gz --top 5`
- **THEN** 每个节点最多展示 cum 最高的 5 个子节点

#### Scenario: 使用 focus 过滤
- **WHEN** 用户执行 `agent-insight tree profile.pb.gz --focus "main.*"`
- **THEN** 只展示调用路径中包含 main 包函数的分支

### Requirement: tree 命令参数验证

系统 SHALL 验证：恰好一个 profile 文件路径，format 必须为 text/json/markdown。

#### Scenario: 无效正则 pattern
- **WHEN** 用户执行 `agent-insight tree profile.pb.gz --focus "[invalid"`
- **THEN** 返回正则编译错误
