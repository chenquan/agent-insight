## 1. Implement normalization helper

- [x] 1.1 在 `pkg/profile/analysis.go` 新增 `normalizeMappingFile(file string) string` 函数：调用 `filepath.Base(file)`，处理空字符串特殊 case
- [x] 1.2 在 `pkg/profile/analysis.go:319-321`（hotspot.Module 赋值）调用 `normalizeMappingFile`
- [x] 1.3 在 `pkg/profile/analysis.go` 查找构造 `Mappings` 列表的位置，也调用 `normalizeMappingFile`（如已构造则覆盖 File 字段）

## 2. Update formatter to use normalized values

- [x] 2.1 在 `pkg/output/formatter.go:570` 确认 `m.File` 已是规范化值（因 profile 层已处理）；无需修改，但需在测试中验证

## 3. Add unit tests

- [x] 3.1 在 `pkg/profile/analysis_test.go` 添加测试：`normalizeMappingFile("/home/user/binary") == "binary"`
- [x] 3.2 添加测试：`normalizeMappingFile("[vdso]") == "[vdso]"`
- [x] 3.3 添加测试：`normalizeMappingFile("") == ""`
- [x] 3.4 添加测试：在 google/pprof testdata 的 cppbench.heap 上跑 info --format json，验证所有 mapping file 都是 basename 形式

## 4. Manual verification

- [x] 4.1 用 google/pprof testdata 的 `cppbench.cpu` 和 `cppbench.heap` 跑 info --format json，对比同一个 binary 的 file 字段是否一致
- [x] 4.2 用 `cppbench.contention` 跑 analyze，验证 Module 字段无绝对路径前缀