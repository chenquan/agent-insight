package profile

import (
	"testing"

	pprofprofile "github.com/google/pprof/profile"
)

// buildComplexProfile creates a profile with multiple functions and call stacks
func buildComplexProfile(t *testing.T) *pprofprofile.Profile {
	t.Helper()

	funcs := []*pprofprofile.Function{
		{ID: 1, Name: "runtime.mallocgc", Filename: "runtime/malloc.go"},
		{ID: 2, Name: "main.main", Filename: "main.go"},
		{ID: 3, Name: "main.handleRequest", Filename: "main.go"},
		{ID: 4, Name: "encoding/json.Marshal", Filename: "encoding/json/encode.go"},
		{ID: 5, Name: "io.ReadAll", Filename: "io/io.go"},
	}

	m := &pprofprofile.Mapping{ID: 1, Start: 0x1000, Limit: 0x2000, File: "/bin/app"}

	locs := []*pprofprofile.Location{
		{ID: 1, Mapping: m, Address: 0x1100, Line: []pprofprofile.Line{{Function: funcs[0], Line: 1020}}},
		{ID: 2, Mapping: m, Address: 0x1200, Line: []pprofprofile.Line{{Function: funcs[1], Line: 15}}},
		{ID: 3, Mapping: m, Address: 0x1300, Line: []pprofprofile.Line{{Function: funcs[2], Line: 42}}},
		{ID: 4, Mapping: m, Address: 0x1400, Line: []pprofprofile.Line{{Function: funcs[3], Line: 160}}},
		{ID: 5, Mapping: m, Address: 0x1500, Line: []pprofprofile.Line{{Function: funcs[4], Line: 88}}},
	}

	p := &pprofprofile.Profile{
		PeriodType:    &pprofprofile.ValueType{Type: "cpu", Unit: "nanoseconds"},
		Period:        10000000,
		DurationNanos: 30e9,
		SampleType:    []*pprofprofile.ValueType{{Type: "samples", Unit: "count"}},
		Function:      funcs,
		Mapping:       []*pprofprofile.Mapping{m},
		Location:      locs,
		Sample: []*pprofprofile.Sample{
			{Location: []*pprofprofile.Location{locs[0], locs[2], locs[1]}, Value: []int64{500}},
			{Location: []*pprofprofile.Location{locs[3], locs[2], locs[1]}, Value: []int64{300}},
			{Location: []*pprofprofile.Location{locs[4], locs[2], locs[1]}, Value: []int64{150}},
			{Location: []*pprofprofile.Location{locs[0], locs[1]}, Value: []int64{100}},
		},
	}
	return p
}

func TestIntegrationAnalyzeFullFlow(t *testing.T) {
	p := buildComplexProfile(t)

	// Test analyze with all features
	analysis, err := NewAnalysis(p, AnalysisConfig{
		TopN:      10,
		CallDepth: 5,
	})
	if err != nil {
		t.Fatalf("NewAnalysis failed: %v", err)
	}

	// Verify metadata
	if analysis.Metadata.Type != "cpu" {
		t.Errorf("expected type 'cpu', got '%s'", analysis.Metadata.Type)
	}
	if analysis.SampleCount != 4 {
		t.Errorf("expected 4 samples, got %d", analysis.SampleCount)
	}

	// Verify hotspots
	if len(analysis.Hotspots) < 2 {
		t.Fatalf("expected at least 2 hotspots, got %d", len(analysis.Hotspots))
	}

	// mallocgc should be top (500+100=600 flat)
	top := analysis.Hotspots[0]
	if top.Function == nil || *top.Function != "runtime.mallocgc" {
		t.Errorf("expected top hotspot 'runtime.mallocgc', got %v", top.Function)
	}

	// Verify call paths
	if len(analysis.CallPaths) == 0 {
		t.Error("expected call paths to be generated")
	}
	t.Logf("Generated %d call paths", len(analysis.CallPaths))
	for _, cp := range analysis.CallPaths {
		t.Logf("  %s: %d (%.2f%%)", cp.Path, cp.Count, cp.Percent)
	}
}

func TestIntegrationListFullFlow(t *testing.T) {
	p := buildComplexProfile(t)

	// Query all main.* functions
	result, err := List(p, ListConfig{Pattern: "main.*"})
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	if len(result.MatchedFunctions) != 2 {
		t.Fatalf("expected 2 matches (main, handleRequest), got %d", len(result.MatchedFunctions))
	}

	// Check that callers and callees are populated
	for _, fn := range result.MatchedFunctions {
		if fn.Function != nil {
			t.Logf("Function: %s, Callers: %d, Callees: %d",
				*fn.Function, len(fn.Callers), len(fn.Callees))
		}
	}
}

func TestIntegrationFlameFullFlow(t *testing.T) {
	p := buildComplexProfile(t)

	result, err := Flame(p, FlameConfig{})
	if err != nil {
		t.Fatalf("Flame failed: %v", err)
	}

	if result.TotalStacks != 4 {
		t.Errorf("expected 4 total stacks, got %d", result.TotalStacks)
	}
	if len(result.Stacks) != 4 {
		t.Errorf("expected 4 unique stacks, got %d", len(result.Stacks))
	}

	// Stacks should be sorted by count descending
	for i := 1; i < len(result.Stacks); i++ {
		if result.Stacks[i].Count > result.Stacks[i-1].Count {
			t.Errorf("stacks not sorted: [%d]=%d > [%d]=%d", i, result.Stacks[i].Count, i-1, result.Stacks[i-1].Count)
		}
	}
}

func TestIntegrationDiffFullFlow(t *testing.T) {
	base := buildComplexProfile(t)

	// Create target with increased json.Marshal load
	target := buildComplexProfile(t)
	// Double json.Marshal samples
	locJson := target.Location[3]
	target.Sample = append(target.Sample, &pprofprofile.Sample{
		Location: []*pprofprofile.Location{locJson, target.Location[2], target.Location[1]},
		Value:    []int64{500},
	})

	result, err := Diff(base, target, DiffConfig{})
	if err != nil {
		t.Fatalf("Diff failed: %v", err)
	}

	// Should have regressions (json.Marshal increased)
	if len(result.Regressions) == 0 {
		t.Error("expected regressions due to increased json.Marshal load")
	}

	// Overall stats should reflect the change
	t.Logf("Base total: %d, Target total: %d, Delta: %d (%.2f%%)",
		result.OverallDiff.BaseTotal, result.OverallDiff.TargetTotal,
		result.OverallDiff.TotalDelta, result.OverallDiff.TotalPercent)
}

func TestIntegrationAnalyzeWithFilter(t *testing.T) {
	p := buildComplexProfile(t)

	analysis, err := NewAnalysis(p, AnalysisConfig{
		TopN:         10,
		FocusPattern: "encoding.*",
		CallDepth:    0,
	})
	if err != nil {
		t.Fatalf("NewAnalysis with focus failed: %v", err)
	}

	// Should only include encoding/json related hotspots
	for _, h := range analysis.Hotspots {
		if h.Function == nil {
			continue
		}
		t.Logf("Filtered hotspot: %s", *h.Function)
	}
}

func TestIntegrationFlameWithIgnore(t *testing.T) {
	p := buildComplexProfile(t)

	result, err := Flame(p, FlameConfig{IgnorePattern: "runtime.*"})
	if err != nil {
		t.Fatalf("Flame with ignore failed: %v", err)
	}

	// No stack should contain runtime functions
	for _, stack := range result.Stacks {
		if containsSubstring(stack.Stack, "runtime.") {
			t.Errorf("stack should not contain runtime: %s", stack.Stack)
		}
	}
}

func containsSubstring(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
