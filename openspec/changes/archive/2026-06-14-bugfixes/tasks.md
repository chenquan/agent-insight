## 1. 共享校验函数

- [x] 1.1 在 `pkg/commands/validate.go` 中添加 `ValidatePositiveInt(value int, name string) error` 函数及其单元测试

## 2. 命令层修复

- [x] 2.1 `pkg/commands/info.go`: 用 `ValidateFormat(infoFormat)` 替换手写格式校验
- [x] 2.2 `pkg/commands/diagnose.go`: 在 RunE 中添加 `--top` 参数的 `ValidatePositiveInt` 校验
- [x] 2.3 `pkg/commands/flame.go`: 当 `--stats` 与非 text 格式同时使用时向 stderr 输出警告

## 3. 移除无效 --value-type flag

- [x] 3.1 从 `pkg/commands/{diff,tree,traces,list,flame,trend}.go` 移除 `--value-type` flag 注册和相关变量

## 4. formatter 修复

- [ ] 4.1 修复 `funcName()`: Function、Address、LocationID 全为 nil 时返回 "unknown" 而非 panic
- [x] 4.2 修复 diff 除零: BaseSamples=0 时 text 和 markdown 格式输出 "N/A"
- [x] 4.3 flame JSON: `flameJSONStack` 的 JSON tag 从 `value` 改为 `count`
- [x] 4.4 将 `TrendMarkdownFormatter.FormatTrendMarkdownResult` 重命名为 `FormatTrendResult`

## 5. 验证

- [x] 5.1 运行 `make test` 确保所有测试通过
- [x] 5.2 运行 `make lint` 确保无 lint 错误
