## Context

测试发现同一个 binary `cppbench_server_main` 在不同 profile 中的 mapping File 字段格式不一致：
- `cppbench.cpu`：`cppbench_server_main`（无路径）
- `cppbench.heap`：`/home/cppbench_server_main`（绝对路径）
- `cppbench.thread.*`：`/home/rsilvera/cppbench/cppbench_server_main[.unstripped]`（绝对路径）

这种不一致来自原始 profile 数据的差异（不同时间/不同机器采集），但对分析用户来说，"是否是同一个 binary" 应该通过规范化后的 basename 判断。

## Goals / Non-Goals

**Goals:**
- 所有 mapping file 字段输出统一为 basename
- 保留特殊标识 `[vdso]` / `[vsyscall]` / 空字符串
- 跨 profile 对比时同一个 binary 呈现一致名称

**Non-Goals:**
- 不修改原始 pprof 数据
- 不修改 BuildID 字段（已经是规范化的十六进制串）
- 不影响其他字段（如 function names、addresses）

## Decisions

### Decision 1: 在 profile 层一次性规范化

把 `normalizeMappingFile` 函数放在 `pkg/profile/analysis.go`（被多命令共享），理由：
- `Module` 字段（analyze）和 `Mappings` 列表（info）都需要规范化
- 单一数据源，避免 formatter 层重复处理
- 测试简单（helper 函数 + 两处调用点的 unit test）

### Decision 2: 保留方括号标识

`[vdso]` / `[vsyscall]` 是 Linux 内核虚拟动态共享对象标识，对 profile 分析有意义。不去除方括号，仅 `filepath.Base` 处理后仍是 `[vdso]`（因为 `filepath.Base("/[vdso]") == "[vdso]"`）。理由：
- 保持语义
- 测试已验证现有 `vdso` 字符串在输出中正常

### Decision 3: 空字符串保持空

如果 `loc.Mapping.File == ""`，规范化后仍是 `""`，不强制替换为占位符。理由：
- 区分"无文件信息"与"已知是空"的语义
- formatter 层已有 `if m.BuildID != ""` 判断，行为可预测

## Risks / Trade-offs

- [风险] 下游脚本若依赖完整路径字符串做匹配会失效 → 缓解：低概率（C++ profile 的 mapping file 多数场景下用户不直接解析，且 basename 才是更友好的标识）
- [风险] `filepath.Base` 在 Windows 上行为不同（处理 `\`）→ 缓解：agent-insight 主要服务 Linux/macOS 用户，跨平台不是核心场景

## Migration Plan

无需数据迁移。输出格式调整。