## 1. info 命令

- [x] 1.1 实现 `pkg/profile/info.go`：Info 函数，读取 Profile 元信息返回 InfoResult 结构体
- [x] 1.2 实现 `pkg/commands/info.go`：InfoCmd cobra 命令，解析参数并调用 profile.Info
- [x] 1.3 在 `pkg/output/formatter.go` 中添加 InfoResult 的 text/json/markdown 格式化器
- [x] 1.4 在 `cmd/root.go` 中注册 InfoCmd
- [x] 1.5 验证 info 命令在 testdata 上的输出

## 2. traces 命令

- [x] 2.1 实现 `pkg/profile/traces.go`：Traces 函数，遍历 Sample 并按 pattern 过滤，返回 TracesResult
- [x] 2.2 实现 `pkg/commands/traces.go`：TracesCmd cobra 命令，支持 --focus/--ignore/--top/--value-type/--format flags
- [x] 2.3 在 `pkg/output/formatter.go` 中添加 TracesResult 的 text/json/markdown 格式化器
- [x] 2.4 在 `cmd/root.go` 中注册 TracesCmd
- [x] 2.5 验证 traces 命令在 testdata 上的输出

## 3. tree 命令

- [x] 3.1 实现 `pkg/profile/tree.go`：Tree 函数，构建 CallTreeNode 递归树结构并返回 TreeResult
- [x] 3.2 实现 `pkg/commands/tree.go`：TreeCmd cobra 命令，支持 --focus/--ignore/--depth/--top/--cum/--value-type/--format flags
- [x] 3.3 在 `pkg/output/formatter.go` 中添加 TreeResult 的 text/json/markdown 格式化器
- [x] 3.4 在 `cmd/root.go` 中注册 TreeCmd
- [x] 3.5 验证 tree 命令在 testdata 上的输出

## 4. Skill 模板更新

- [x] 4.1 更新 frontmatter 触发词：新增 概况、元信息、调用链、执行路径、调用树、层级结构
- [x] 4.2 更新"何时使用"表：新增 3 行触发场景（概况→info，调用链→traces，调用树→tree）
- [x] 4.3 新增 info/traces/tree 三个命令速查段（flags 表 + 示例）
- [x] 4.4 新增"快速概览"和"调用路径追踪"两个典型工作流
- [x] 4.5 新增 info/traces/tree 三个输出解读字段表
- [x] 4.6 新增注意事项："分析前可先用 info 了解概况"

## 5. 其他文档更新

- [x] 5.1 更新 `README.md`：添加三个新命令的使用说明
- [x] 5.2 更新 `cmd/root.go` 中的 Long 描述：列出所有命令
