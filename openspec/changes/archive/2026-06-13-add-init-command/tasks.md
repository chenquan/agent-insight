## 1. Skill 模板

- [x] 1.1 创建 `pkg/skill/` 包，编写 SKILL.md 模板内容（包含 frontmatter、触发条件、命令速查、工作流、输出解读）
- [x] 1.2 使用 `//go:embed` 嵌入模板文件，编写 `generator.go` 提供生成函数

## 2. Init 命令

- [x] 2.1 创建 `pkg/commands/init.go`，实现 cobra 命令注册（支持 `--force` flag）
- [x] 2.2 实现 init 逻辑：检测目标路径、创建目录、写入文件、处理已存在文件的情况

## 3. 注册与测试

- [x] 3.1 在 `cmd/root.go` 中注册 InitCmd
- [x] 3.2 为 init 命令编写测试（首次生成、已存在提示、--force 覆盖）
- [x] 3.3 运行全量测试和 lint，确保无回归
