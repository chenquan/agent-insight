## 1. Implement fallback inference

- [x] 1.1 在 `pkg/profile/analysis.go` 新增 `inferProfileType(p *profile.Profile) string` 纯函数，根据 SampleType 关键字推断类型
- [x] 1.2 修改 `extractMetadata`：当 `p.PeriodType == nil` 或 `p.PeriodType.Type == ""` 时调用 `inferProfileType`

## 2. Add unit tests

- [x] 2.1 在 `pkg/profile/analysis_test.go` 添加测试：构造 PeriodType 为 nil 的 heap profile，验证 metadata.Type == "heap"
- [x] 2.2 添加测试：构造含 `cpu` SampleType 的 profile，验证推断为 `cpu`
- [x] 2.3 添加测试：构造 SampleType 无可识别关键字的 profile，验证返回 `unknown`

## 3. Manual verification

- [x] 3.1 用 google/pprof testdata 的 `gobench.heap` 跑 analyze --format json，验证 type 字段非空
- [x] 3.2 用 `java.heap` 跑同样命令验证
- [x] 3.3 用 `go.crc32.cpu` 跑同样命令验证 PeriodType 存在时仍返回 `cpu`