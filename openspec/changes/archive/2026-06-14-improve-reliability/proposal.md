## Why

现有命令存在两个可靠性问题：(1) `analysis.go` 的 `calculateHotspots` 对零 totalValue 无防护，当所有 sample 被 focus/ignore 过滤后会产生 NaN 百分比；(2) commands 层 10 个命令各自重复实现 format 和 regex 验证逻辑（~20 段重复代码），且措辞不一致（`regex` vs `pattern` vs 无前缀）。此外 commands 层测试覆盖率接近零。

注：diff/trend 的 profile 类型一致性校验（`ValidateTypeConsistency`）已实现；list/traces/tree/diff 的百分比计算已有 `if totalValue > 0` 防护；loader.go 错误消息已包含路径和原因。这些无需重复处理。

## What Changes

- **analysis.go 除零防护**：在 `calculateHotspots` 中对 totalValue == 0 做防护，返回空结果而非 NaN
- **抽取公共验证函数**：从 commands 层提取 `ValidateFormat` 和 `ValidateRegex` 共享函数，消除 10 个命令中 ~20 段重复的验证代码并统一措辞
- **commands 层测试补全**：为关键命令的 RunE 函数添加单元测试

## Capabilities

### New Capabilities

- `shared-validation`: commands 层公共验证函数（ValidateFormat、ValidateRegex），供所有命令复用

### Modified Capabilities

- `profile-analyze`: `calculateHotspots` 增加 totalValue == 0 防护

## Impact

- **pkg/commands/**：新增 validate.go，所有命令文件迁移到公共验证函数
- **pkg/profile/analysis.go**：`calculateHotspots` 增加除零防护
- **pkg/output/**：无变更
- **pkg/skill/template.md**：无变更
- **README.md**：无变更
