## 1. Fix flag passthrough in analyze command

- [x] 1.1 修改 `pkg/commands/analyze.go` 第97-102行：把 `analyzeValueType` 字符串解析为 `ValueTypeConfig` 并赋值给 `config.ValueType`
- [x] 1.2 修改 `pkg/commands/analyze.go`：当 `--value-type` 值不在 profile 的 `SampleType` 列表中时返回错误，错误信息包含可用 value types

## 2. Verify profile layer accepts user-specified value type

- [x] 2.1 在 `pkg/profile/analysis.go` 确认 `NewAnalysis` 在 `config.ValueType != nil` 时不再调用 `selectDefaultValueType`
- [x] 2.2 若 `ValueTypeConfig` 在 profile 层没有按名字匹配 SampleType 的逻辑，补齐该解析（参照 `selectDefaultValueType` 实现）

## 3. Add unit tests

- [x] 3.1 在 `pkg/commands/analyze_test.go`（或新建）添加测试：`--value-type alloc_objects` 与 `--value-type inuse_space` 在 `gobench.heap` 上输出不同
- [x] 3.2 添加测试：传入不存在的 value-type 返回非零退出码与明确错误信息

## 4. Manual verification

- [x] 4.1 `make build && ./agent-insight analyze testdata/cpu.pb.gz --value-type samples` 验证未引入回归
- [x] 4.2 用 google/pprof testdata 的 `gobench.heap` 跑 `--value-type alloc_objects` 与 `--value-type inuse_space` 确认输出不同