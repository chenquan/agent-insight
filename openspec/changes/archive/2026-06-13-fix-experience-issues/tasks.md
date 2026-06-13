## 1. P1: diff 类型校验

- [x] 1.1 将 `pkg/profile/merge.go` 中的 `validateTypeConsistency` 提取为导出函数
- [x] 1.2 在 `pkg/profile/diff.go` 的 `Diff` 函数中调用类型校验，不一致时返回错误
- [x] 1.3 验证: `diff cpu.pb.gz heap.pb.gz` 报错

## 2. P2: summary 措辞自适应

- [x] 2.1 在 `pkg/output/formatter.go` 中根据 profile type 选择 summary 措辞（cpu→热点、heap→内存热点、goroutine→阻塞点）
- [x] 2.2 验证: heap profile 的 summary 不再出现 "performance bottleneck"

## 3. P3: JSON 输出附带单位

- [x] 3.1 在 `pkg/output/formatter.go` 的 JSON formatter 中为 flat/cum 值添加 unit 字段
- [x] 3.2 验证: heap JSON 输出包含 "unit": "bytes"

## 4. P4: info 输出 goroutine 总数

- [x] 4.1 在 `pkg/profile/info.go` 的 `InfoResult` 中添加 `TotalValue int64`，对 goroutine profile 聚合所有 sample 的 value
- [x] 4.2 在 `pkg/output/formatter.go` 中输出 goroutine 总数
- [x] 4.3 验证: goroutine profile 的 info 输出包含 total goroutine 数

## 5. P5: help text 完善

- [x] 5.1 审查所有 9 个子命令的 help text 和 examples，补全缺失内容

## 6. P6: 百分比精度统一

- [x] 6.1 在 `pkg/output/formatter.go` 中统一百分比输出为 2 位小数
- [x] 6.2 验证: JSON 和 text 输出的百分比均为 2 位小数

## 7. P7: diff text 输出优化

- [x] 7.1 在 diff text formatter 中，当 cum 值为 0 时不显示 cum 列

## 8. P8: JSON 字段命名统一

- [x] 8.1 扫描所有 JSON formatter 输出，将 camelCase 字段名改为 snake_case
- [x] 8.2 验证: 所有 JSON 输出字段名均为 snake_case

## 9. 测试

- [x] 9.1 运行全量测试确认无回归
- [x] 9.2 逐个 P 验证修复效果
