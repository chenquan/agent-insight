//nolint:staticcheck // QF1012: b.WriteString(fmt.Sprintf(...)) is intentional for readability in the template builder
package profile

import (
	"fmt"
	"strings"

	"github.com/google/pprof/profile"
)

// DiagnosePrompt holds all components of a diagnostic prompt.
type DiagnosePrompt struct {
	Language    string
	ProfileType string
	Metadata    Metadata
	Analysis    *Analysis
	Tree        *TreeResult
	Traces      *TracesResult
	Guidance    string
	UserContext string
}

// baseGuidance maps profile types to diagnostic guidance text.
var baseGuidance = map[string]string{
	"cpu": `请重点分析以下维度：
1. 计算热点 — 哪些函数消耗最多资源？是否合理？
2. 调用路径 — 热点是怎么被触发的？关键调用链是什么？
3. 算法效率 — 是否存在可优化的算法复杂度（如 O(n²) 优化为 O(n)）？
4. 并发效率 — 是否有并发度不足或竞争的迹象？`,

	"heap": `请重点分析以下维度：
1. 分配热点 — 哪些函数分配内存最多？分配量是否合理？
2. 内存泄漏 — 是否有持续增长不释放的模式？（结合 inuse vs alloc 判断）
3. 大对象 — 是否有不必要的大对象分配？
4. 分配优化 — 能否减少分配频率（对象复用、预分配等）？
5. GC 影响 — 分配模式是否会给 GC 造成压力？`,

	"goroutine": `请重点分析以下维度：
1. goroutine 数量 — 总量是否异常？
2. 阻塞模式 — 大量 goroutine 停在什么位置？（channel 阻塞、锁等待、IO 等待？）
3. 泄漏迹象 — 是否有不断创建但不退出的模式？
4. 并发结构 — goroutine 的使用模式是否合理？（worker pool vs 无限创建）`,

	"contentions": `请重点分析以下维度：
1. 竞争热点 — 哪些锁竞争最激烈？
2. 竞争路径 — 锁是在什么调用链中被持有的？
3. 优化方向 — 能否用无锁结构或更细粒度的锁替代？
4. 持有时间 — 锁持有时间是否过长？是否在锁内做了不必要的重操作？`,

	"thread": `请重点分析以下维度：
1. 线程数量 — 总线程数是否异常？是否有不断创建的趋势？
2. 创建来源 — 线程是在哪里被创建的？
3. 线程生命周期 — 是否有过早创建或不及时回收的模式？`,

	"unknown": `请根据以下 profile 数据进行分析。结合数据中的函数名、调用路径和采样分布，判断可能存在的性能问题并给出优化建议。`,
}

// languageAddition maps languages to additional diagnostic guidance.
var languageAddition = map[string]string{
	"Go": `

该程序使用 Go 语言。请额外关注：
- runtime 包相关函数（如 runtime.mallocgc）的占比，这通常是 GC 压力的信号
- sync.Pool、切片预分配等 Go 特有的减少分配手段
- string 和 []byte 之间的转换是否造成不必要的额外分配
- goroutine 并发效率，是否存在过多 goroutine 竞争`,

	"C++": `

该程序使用 C++ 语言。请额外关注：
- 虚函数调用开销，是否可以用 final 或 CRTP 替代
- 缓存局部性（cache miss），数据访问模式是否对缓存友好
- SIMD/OpenMP 等并行化机会
- custom allocator 或内存池的使用
- RAII 对象的生命周期管理`,

	"Rust": `

该程序使用 Rust 语言。请额外关注：
- clone() 的开销，能否通过引用或 Copy 语义避免
- unsafe 代码段的热点
- 迭代器链 vs 手写循环的性能差异
- async/await 运行时开销
- Arc/Rc 的分配开销，是否可以用借用替代`,

	"Java": `

该程序使用 Java 语言。请额外关注：
- JIT 编译是否生效（解释执行的方法会异常消耗资源）
- GC 策略对性能的影响（G1/ZGC/Shenandoah）
- 对象分配频率和生命周期
- 线程阻塞和锁竞争
- 静态集合持有对象引用导致的内存泄漏`,

	"C": `

该程序使用 C 语言。请额外关注：
- malloc/free 模式，是否存在频繁分配释放
- 缓冲区管理策略，是否能复用缓冲区
- struct 内存布局和缓存行对齐
- 系统调用开销`,

	"Unknown": "",
}

// BuildDiagnosePrompt constructs a diagnostic prompt from profile data.
func BuildDiagnosePrompt(p *profile.Profile, topN int, userContext string) (*DiagnosePrompt, error) {
	metadata := extractMetadata(p)
	lang := DetectLanguage(p)

	analysisConfig := AnalysisConfig{
		TopN:      topN,
		CallDepth: 5,
	}
	analysis, err := NewAnalysis(p, analysisConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze profile: %w", err)
	}

	treeResult, err := Tree(p, TreeConfig{
		TopN:  topN,
		Depth: 5,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to build tree: %w", err)
	}

	tracesResult, err := Traces(p, TracesConfig{
		TopN: 5,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get traces: %w", err)
	}

	guidance := buildGuidance(metadata.Type, lang)

	return &DiagnosePrompt{
		Language:    lang,
		ProfileType: metadata.Type,
		Metadata:    metadata,
		Analysis:    analysis,
		Tree:        treeResult,
		Traces:      tracesResult,
		Guidance:    guidance,
		UserContext: userContext,
	}, nil
}

func buildGuidance(profileType, lang string) string {
	// Normalize heap aliases: pprof uses "space" as PeriodType for heap profiles
	normalized := profileType
	if profileType == "space" {
		normalized = "heap"
	}

	base, ok := baseGuidance[normalized]
	if !ok {
		base = baseGuidance["unknown"]
	}

	addition := languageAddition[lang]

	return base + addition
}

// Text renders the diagnostic prompt as plain text.
func (dp *DiagnosePrompt) Text() string {
	var b strings.Builder

	b.WriteString("你是一个")
	if dp.Language != string(langUnknown) {
		b.WriteString(dp.Language)
	}
	b.WriteString("性能诊断专家。以下是")
	b.WriteString("一个 pprof profile 的分析数据，请根据数据给出性能诊断。\n\n")

	// Profile metadata
	b.WriteString("## Profile 概况\n\n")
	displayType := dp.ProfileType
	if displayType == "space" {
		displayType = "heap"
	}
	b.WriteString(fmt.Sprintf("- 类型: %s\n", displayType))
	b.WriteString(fmt.Sprintf("- 检测到语言: %s\n", dp.Language))
	if dp.Metadata.Duration > 0 {
		b.WriteString(fmt.Sprintf("- 采样时长: %s\n", dp.Metadata.Duration))
	}
	b.WriteString(fmt.Sprintf("- 采样数: %d\n", dp.Analysis.SampleCount))
	if len(dp.Metadata.SampleTypes) > 0 {
		b.WriteString(fmt.Sprintf("- 值类型: %s\n", strings.Join(dp.Metadata.SampleTypes, ", ")))
	}
	b.WriteString(fmt.Sprintf("- 函数数量: %d\n", dp.Metadata.Functions))

	hasSymbols := false
	for _, h := range dp.Analysis.Hotspots {
		if h.Function != nil {
			hasSymbols = true
			break
		}
	}
	symStatus := "可用"
	if !hasSymbols {
		symStatus = "不可用（部分函数显示为地址）"
	}
	b.WriteString(fmt.Sprintf("- 符号状态: %s\n", symStatus))
	b.WriteString("\n")

	// Hotspot data
	b.WriteString("## 热点函数\n\n")
	for i, h := range dp.Analysis.Hotspots {
		name := "unknown"
		if h.Function != nil {
			name = *h.Function
		} else if h.Address != nil {
			name = *h.Address
		}
		file := ""
		if h.File != nil {
			file = *h.File
		}
		b.WriteString(fmt.Sprintf("%d. %s", i+1, name))
		if file != "" {
			b.WriteString(fmt.Sprintf(" [%s]", file))
		}
		b.WriteString(fmt.Sprintf("\n   Flat: %d (%.2f%%), Cum: %d (%.2f%%)\n",
			h.FlatValue, h.FlatPercent, h.CumValue, h.CumPercent))
	}
	b.WriteString("\n")

	// Call tree summary
	b.WriteString("## 调用树\n\n")
	if dp.Tree != nil {
		dp.renderTreeNode(&b, dp.Tree.VisibleChildren(), 0, 5)
	}
	b.WriteString("\n")

	// Key traces
	b.WriteString("## 关键调用路径\n\n")
	if dp.Traces != nil {
		for i, trace := range dp.Traces.Traces {
			b.WriteString(fmt.Sprintf("路径 %d (占比 %.2f%%):\n", i+1, trace.Percent))
			for j, fn := range trace.Stack {
				b.WriteString(fmt.Sprintf("%s%s\n", strings.Repeat("  ", j), fn))
			}
			b.WriteString("\n")
		}
	}

	// Diagnostic guidance
	b.WriteString("## 诊断引导\n\n")
	b.WriteString(dp.Guidance)
	b.WriteString("\n\n")

	// Output format requirements
	b.WriteString("## 输出要求\n\n")
	b.WriteString("请按以下格式输出诊断结果：\n")
	b.WriteString("1. **根因分析** — 一句话概括核心问题\n")
	b.WriteString("2. **优化建议** — 按优先级排序（影响大、改动小优先），每条包含改动点和预期收益\n")
	b.WriteString("3. **异常模式** — 如果检测到异常模式（内存泄漏、锁竞争、无限循环等），说明判断依据\n")
	b.WriteString("\n")

	// User context
	if dp.UserContext != "" {
		b.WriteString("## 用户上下文\n\n")
		b.WriteString(dp.UserContext)
		b.WriteString("\n")
	}

	return b.String()
}

func (dp *DiagnosePrompt) renderTreeNode(b *strings.Builder, children []*CallTreeNode, depth, maxDepth int) {
	if depth >= maxDepth {
		return
	}
	for _, node := range children {
		if node.FlatPercent < 1.0 && node.CumPercent < 1.0 {
			continue
		}
		indent := strings.Repeat("  ", depth)
		fmt.Fprintf(b, "%s%s  flat: %d (%.2f%%)  cum: %d (%.2f%%)\n",
			indent, node.Name, node.Flat, node.FlatPercent, node.Cum, node.CumPercent)
		dp.renderTreeNode(b, node.Children, depth+1, maxDepth)
	}
}
