package profile

import (
	"strings"
	"testing"

	pprofprofile "github.com/google/pprof/profile"
)

func TestBuildGuidance_CPU_Go(t *testing.T) {
	guidance := buildGuidance("cpu", "Go")
	if !strings.Contains(guidance, "计算热点") {
		t.Error("CPU guidance should contain '计算热点'")
	}
	if !strings.Contains(guidance, "sync.Pool") {
		t.Error("Go addition should contain 'sync.Pool'")
	}
	if !strings.Contains(guidance, "runtime") {
		t.Error("Go addition should mention 'runtime'")
	}
}

func TestBuildGuidance_Heap_CPP(t *testing.T) {
	guidance := buildGuidance("heap", "C++")
	if !strings.Contains(guidance, "分配热点") {
		t.Error("Heap guidance should contain '分配热点'")
	}
	if !strings.Contains(guidance, "虚函数") {
		t.Error("C++ addition should contain '虚函数'")
	}
}

func TestBuildGuidance_Goroutine_Rust(t *testing.T) {
	guidance := buildGuidance("goroutine", "Rust")
	if !strings.Contains(guidance, "阻塞模式") {
		t.Error("Goroutine guidance should contain '阻塞模式'")
	}
	if !strings.Contains(guidance, "clone()") {
		t.Error("Rust addition should contain 'clone()'")
	}
}

func TestBuildGuidance_Unknown(t *testing.T) {
	guidance := buildGuidance("unknown", "Unknown")
	if !strings.Contains(guidance, "profile 数据进行分析") {
		t.Error("Unknown guidance should contain generic analysis text")
	}
	if strings.Contains(guidance, "sync.Pool") {
		t.Error("Unknown language should have no language-specific addition")
	}
}

func TestBuildDiagnosePrompt(t *testing.T) {
	p := makeTestCPUProfile()
	dp, err := BuildDiagnosePrompt(p, 5, "")
	if err != nil {
		t.Fatalf("BuildDiagnosePrompt() error: %v", err)
	}

	if dp.Language != string(langGo) {
		t.Errorf("Language = %q, want %q", dp.Language, langGo)
	}
	if dp.ProfileType != "cpu" {
		t.Errorf("ProfileType = %q, want %q", dp.ProfileType, "cpu")
	}
}

func TestDiagnosePrompt_Text(t *testing.T) {
	p := makeTestCPUProfile()
	dp, err := BuildDiagnosePrompt(p, 5, "HTTP API server")
	if err != nil {
		t.Fatalf("BuildDiagnosePrompt() error: %v", err)
	}

	text := dp.Text()

	// Check structure
	sections := []string{
		"Profile 概况",
		"热点函数",
		"调用树",
		"关键调用路径",
		"诊断引导",
		"输出要求",
		"用户上下文",
	}
	for _, s := range sections {
		if !strings.Contains(text, s) {
			t.Errorf("prompt text missing section: %s", s)
		}
	}

	// Check language-specific content
	if !strings.Contains(text, "Go") {
		t.Error("prompt should mention Go language")
	}
	if !strings.Contains(text, "HTTP API server") {
		t.Error("prompt should contain user context")
	}
}

func TestDiagnosePrompt_TextNoContext(t *testing.T) {
	p := makeTestCPUProfile()
	dp, err := BuildDiagnosePrompt(p, 5, "")
	if err != nil {
		t.Fatalf("BuildDiagnosePrompt() error: %v", err)
	}

	text := dp.Text()
	if strings.Contains(text, "用户上下文") {
		t.Error("prompt without context should not contain '用户上下文' section")
	}
}

func makeTestCPUProfile() *Profile {
	mallocgc := &pprofprofile.Function{ID: 1, Name: "runtime.mallocgc", SystemName: "runtime.mallocgc", Filename: "runtime/malloc.go"}
	mainFunc := &pprofprofile.Function{ID: 2, Name: "main.main", SystemName: "main.main", Filename: "main.go"}
	handleReq := &pprofprofile.Function{ID: 3, Name: "main.handleRequest", SystemName: "main.handleRequest", Filename: "main.go"}
	jsonMarshal := &pprofprofile.Function{ID: 4, Name: "encoding/json.Marshal", SystemName: "encoding/json.Marshal", Filename: "encoding/json/encode.go"}

	mapping := &pprofprofile.Mapping{ID: 1, Start: 0x1000, Limit: 0x2000, File: "/usr/local/bin/myapp"}

	locMallocgc := &pprofprofile.Location{ID: 1, Mapping: mapping, Address: 0x1100, Line: []pprofprofile.Line{{Function: mallocgc, Line: 1020}}}
	locMain := &pprofprofile.Location{ID: 2, Mapping: mapping, Address: 0x1200, Line: []pprofprofile.Line{{Function: mainFunc, Line: 15}}}
	locHandleReq := &pprofprofile.Location{ID: 3, Mapping: mapping, Address: 0x1300, Line: []pprofprofile.Line{{Function: handleReq, Line: 42}}}
	locJsonMarshal := &pprofprofile.Location{ID: 4, Mapping: mapping, Address: 0x1400, Line: []pprofprofile.Line{{Function: jsonMarshal, Line: 160}}}

	return NewProfile(&pprofprofile.Profile{
		PeriodType:   &pprofprofile.ValueType{Type: "cpu", Unit: "nanoseconds"},
		Period:       10000000,
		DurationNanos: 30e9,
		SampleType: []*pprofprofile.ValueType{
			{Type: "samples", Unit: "count"},
			{Type: "cpu", Unit: "nanoseconds"},
		},
		Function: []*pprofprofile.Function{mallocgc, mainFunc, handleReq, jsonMarshal},
		Mapping:  []*pprofprofile.Mapping{mapping},
		Location: []*pprofprofile.Location{locMallocgc, locMain, locHandleReq, locJsonMarshal},
		Sample: []*pprofprofile.Sample{
			{Location: []*pprofprofile.Location{locMallocgc, locHandleReq, locMain}, Value: []int64{500, 5000000000}},
			{Location: []*pprofprofile.Location{locJsonMarshal, locHandleReq, locMain}, Value: []int64{300, 3000000000}},
		},
	})
}
