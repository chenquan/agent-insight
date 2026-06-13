## Why

C++ profile mappings 的 `File` 字段在不同文件间格式不一致：
- `cppbench.heap` mappings 出现 `/home/cppbench_server_main`
- `cppbench.cpu` mappings 出现 `cppbench_server_main`（无前缀）
- `cppbench.growth` mappings 出现 `/libnss_cache-2.15.so`（相对路径）

影响：用户对比多个 profile 时，无法快速识别"是否是同一个 binary"；输出可读性差。

## What Changes

- 在 `pkg/profile/analysis.go` 新增 `normalizeMappingFile(file string) string` 工具函数：用 `filepath.Base()` 提取文件名（如 `/home/cppbench_server_main` → `cppbench_server_main`）；保留 `[vdso]`/`[vsyscall]` 等方括号包裹的特殊标识；空字符串保持空。
- 在 `pkg/profile/analysis.go:319-321`（Module 字段）和 `pkg/output/formatter.go:570`（mappings 列表输出）调用该函数。
- 在 profile 层一次性规范化，避免在多处重复处理。

## Capabilities

### New Capabilities

无。

### Modified Capabilities

- `profile-analyze`：补充 REQUIREMENT —— mappings 的 `File` 字段 SHALL 规范化（统一为 `filepath.Base()` 形式），便于跨 profile 对比。

## Impact

- **代码**：`pkg/profile/analysis.go`（新增 helper + 两处调用）、`pkg/output/formatter.go`（一处调用）
- **输出**：所有命令输出中的 mapping file 字段统一为 basename
- **行为**：JSON 输出字段值变化（无前缀路径 → basename），下游若做字符串匹配可能受影响