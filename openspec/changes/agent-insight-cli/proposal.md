## Why

当前 pprof 性能分析依赖 `go tool pprof` 的交互式界面或复杂的 Python 工具链，不适合 Claude Code 等 AI 编码助手直接解析和操作。需要一个零配置、单二进制的 CLI 工具，输出结构化数据，便于 AI 理解性能瓶颈并提供优化建议。

## What Changes

- **新增 CLI 工具**：创建 `agent-insight` 命令行工具，支持 pprof 文件解析和分析
- **核心命令集**：
  - `analyze`: 分析 pprof 文件，输出性能热点摘要
  - `list`: 查询指定函数的调用关系和性能贡献
  - `flame`: 生成折叠栈格式，供火焰图工具使用
  - `diff`: 对比两个 profile 文件的性能差异
- **多格式输出**：支持 text、json、markdown 三种输出格式
- **AI 友好设计**：JSON 输出结构清晰，字段命名一致，易于大模型解析

## Capabilities

### New Capabilities
- `profile-analyze`: 解析并分析 pprof 文件，输出性能热点、调用栈和摘要信息
- `profile-list`: 查询指定函数的调用方/被调方及其性能贡献
- `profile-flame`: 将 profile 转换为折叠栈格式（flame graph folded format）
- `profile-diff`: 对比两个 profile，识别性能变化的热点

### Modified Capabilities
- 无（这是全新项目）

## Impact

- **依赖项**：核心依赖 `github.com/google/pprof/profile`，使用 Go 标准库构建 CLI
- **项目结构**：新建 Go module，采用 cmd/ + pkg/ 标准布局
- **兼容性**：完全兼容 `go tool pprof` 生成的 protobuf 格式文件
- **性能**：流式解析大文件，内存占用可控，输出精简避免上下文溢出
- **符号信息处理**：优雅处理生产环境 profile 常见的符号信息缺失情况，支持 Location ID 和内存地址作为备选方案
- **多值类型支持**：正确处理不同 profile 类型的多值结构（如 Go heap profile 的四值系统），提供智能默认值和用户可选值
