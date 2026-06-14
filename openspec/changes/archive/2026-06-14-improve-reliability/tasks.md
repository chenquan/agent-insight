## 1. 公共验证函数

- [x] 1.1 在 `pkg/commands/validate.go` 中实现 `ValidateFormat(format string) error`
- [x] 1.2 在 `pkg/commands/validate.go` 中实现 `ValidateRegex(pattern, name string) error`
- [x] 1.3 为 ValidateFormat 和 ValidateRegex 编写单元测试

## 2. analysis.go 除零防护

- [x] 2.1 在 `calculateHotspots` 中对 totalValue == 0 返回空切片
- [x] 2.2 编写测试：全部样本被过滤时返回空结果且无 NaN

## 3. Commands 层迁移到公共验证函数

- [x] 3.1 analyze.go 替换重复的 format/regex 验证为公共函数调用
- [x] 3.2 diff.go 替换重复的 format/regex 验证为公共函数调用
- [x] 3.3 trend.go 替换重复的 format/regex 验证为公共函数调用
- [x] 3.4 flame.go 替换重复的 format/regex 验证为公共函数调用
- [x] 3.5 traces.go 替换重复的 format/regex 验证为公共函数调用
- [x] 3.6 tree.go 替换重复的 format/regex 验证为公共函数调用
- [x] 3.7 list.go 替换重复的 format/regex 验证为公共函数调用
- [x] 3.8 diagnose.go 替换重复的 format/regex 验证为公共函数调用

## 4. Commands 层测试补全

- [x] 4.1 为 analyze.go RunE 编写测试（正常路径 + 空 profile）
- [x] 4.2 为 diff.go RunE 编写测试（正常路径 + 缺失文件）
- [x] 4.3 为 trend.go RunE 编写测试（正常路径 + 不足 3 个 profile）
- [x] 4.4 为 flame.go RunE 编写测试（正常路径）
- [x] 4.5 为 traces.go RunE 编写测试（正常路径）
- [x] 4.6 为 tree.go RunE 编写测试（正常路径）
- [x] 4.7 为 list.go RunE 编写测试（正常路径 + 无匹配）
- [x] 4.8 为 diagnose.go RunE 编写测试（正常路径）

## 5. 验证

- [x] 5.1 运行 `make test` 确认所有测试通过
- [x] 5.2 运行 `make lint` 确认无 lint 警告（4 个已有 staticcheck 警告非本次引入）
- [x] 5.3 运行 `make build` 确认构建成功
