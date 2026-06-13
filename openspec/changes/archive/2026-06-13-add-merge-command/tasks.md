## 1. 核心合并逻辑

- [x] 1.1 在 `pkg/profile/merge.go` 中实现 `Merge(profiles []*profile.Profile) (*profile.Profile, error)` 函数，调用 `profile.Merge()` 完成合并
- [x] 1.2 实现类型一致性校验逻辑，比对所有 profile 的 `PeriodType`，不一致时返回清晰错误信息
- [x] 1.3 实现 `MergeResult` 结构体，包含输入文件数、合并后采样数等统计信息

## 2. 命令层

- [x] 2.1 在 `pkg/commands/merge.go` 中创建 `MergeCmd` cobra 命令，支持多文件路径和目录两种输入方式
- [x] 2.2 添加 `-o`/`--output` flag（必需），校验输出路径合法性
- [x] 2.3 实现目录递归自动发现逻辑：递归扫描 `.pb` 和 `.pb.gz` 文件，按完整路径排序
- [x] 2.4 实现输入校验：至少 2 个 profile、类型一致性、文件存在性
- [x] 2.5 实现合并结果写入：将合并后的 profile 序列化为 gzip protobuf 写入 `-o` 指定的文件
- [x] 2.6 实现合并摘要输出（输入文件数、总采样数、输出路径）

## 3. 注册和集成

- [x] 3.1 在 `cmd/root.go` 中注册 `MergeCmd` 子命令
- [x] 3.2 更新 rootCmd.Long 描述中的命令列表，添加 merge

## 4. 模板更新

- [x] 4.1 在 `pkg/skill/template.md` 中添加 merge 命令的说明、用法示例和输出字段描述

## 5. 测试

- [x] 5.1 在 `pkg/profile/merge_test.go` 中编写单元测试：成功合并、类型不一致报错、单 profile 报错
- [x] 5.2 验证合并产出的 `.pb.gz` 文件可被现有 analyze/diff/flame 命令正常读取
