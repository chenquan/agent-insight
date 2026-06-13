## Why

生产环境中经常需要合并多个短时间采样的 pprof profile 才能获得有代表性的性能数据。当前 agent-insight 缺少这一基础能力，用户只能在合并前依赖外部工具。pprof 库原生提供 `profile.Merge()` 支持，实现成本低、收益高。

## What Changes

- 新增 `merge` 子命令，支持合并多个同类 pprof profile 文件
- 支持显式指定多个文件路径（至少 2 个）或输入目录（自动发现目录下所有同类 profile）
- 校验所有输入 profile 类型一致（如不能将 cpu 和 heap 混合）
- 合并后输出为 `.pb.gz` 格式，可被现有 analyze/diff/flame 等命令直接使用
- 同步更新 `pkg/skill/template.md` 模板以告知 Claude Code 新命令的存在

## Capabilities

### New Capabilities
- `profile-merge`: 合并多个同类 pprof profile 为一个，校验类型一致性，支持文件列表和目录输入

### Modified Capabilities

## Impact

- **新增文件**: `pkg/commands/merge.go`、`pkg/profile/merge.go`
- **修改文件**: `cmd/root.go`（注册子命令）、`pkg/skill/template.md`（更新模板）
- **依赖**: 无新增外部依赖，使用已有的 `github.com/google/pprof/profile`
- **测试数据**: 需新增测试用例验证合并逻辑
