package profile

import (
	"os"
	"path/filepath"
	"testing"

	pprofprofile "github.com/google/pprof/profile"
)

// helper: create a minimal profile and write to a temp file
func createTestProfile(t *testing.T) *Profile {
	t.Helper()
	raw := &pprofprofile.Profile{
		PeriodType: &pprofprofile.ValueType{Type: "cpu", Unit: "nanoseconds"},
		Period:     10000000,
		SampleType: []*pprofprofile.ValueType{
			{Type: "samples", Unit: "count"},
		},
	}
	fn := &pprofprofile.Function{ID: 1, Name: "main.foo", Filename: "main.go"}
	m := &pprofprofile.Mapping{ID: 1, Start: 0x1000, Limit: 0x2000, File: "/bin/app"}
	loc := &pprofprofile.Location{ID: 1, Mapping: m, Address: 0x1100, Line: []pprofprofile.Line{{Function: fn, Line: 10}}}
	raw.Function = []*pprofprofile.Function{fn}
	raw.Mapping = []*pprofprofile.Mapping{m}
	raw.Location = []*pprofprofile.Location{loc}
	raw.Sample = []*pprofprofile.Sample{
		{Location: []*pprofprofile.Location{loc}, Value: []int64{100}},
	}
	return NewProfile(raw)
}

func writeProfileToTemp(t *testing.T, p *Profile) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "test.pb.gz")
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if err := p.Write(f); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestLoaderLoadFromFile(t *testing.T) {
	p := createTestProfile(t)
	path := writeProfileToTemp(t, p)

	loader := NewLoader()
	loaded, err := loader.LoadFromFile(path)
	if err != nil {
		t.Fatalf("LoadFromFile failed: %v", err)
	}
	if len(loaded.Sample) != 1 {
		t.Errorf("expected 1 sample, got %d", len(loaded.Sample))
	}
	if loaded.Sample[0].Value[0] != 100 {
		t.Errorf("expected value 100, got %d", loaded.Sample[0].Value[0])
	}
}

func TestLoaderFileNotFound(t *testing.T) {
	loader := NewLoader()
	_, err := loader.LoadFromFile("/nonexistent/path.pb.gz")
	if err == nil {
		t.Fatal("expected error for nonexistent file")
	}
}

func TestAnalysisHotspots(t *testing.T) {
	p := createTestProfile(t)
	analysis, err := NewAnalysis(p, AnalysisConfig{TopN: 10, CallDepth: 0})
	if err != nil {
		t.Fatalf("NewAnalysis failed: %v", err)
	}
	if len(analysis.Hotspots) != 1 {
		t.Fatalf("expected 1 hotspot, got %d", len(analysis.Hotspots))
	}
	if analysis.Hotspots[0].FlatValue != 100 {
		t.Errorf("expected flat 100, got %d", analysis.Hotspots[0].FlatValue)
	}
	if analysis.Hotspots[0].Function == nil || *analysis.Hotspots[0].Function != "main.foo" {
		t.Errorf("expected function 'main.foo'")
	}
}

func TestAnalysisSortByCum(t *testing.T) {
	p := createTestProfile(t)
	analysis, err := NewAnalysis(p, AnalysisConfig{TopN: 10, SortByCum: true, CallDepth: 0})
	if err != nil {
		t.Fatalf("NewAnalysis failed: %v", err)
	}
	if analysis.Config.SortByCum != true {
		t.Error("expected SortByCum to be true")
	}
}

func TestListCommand(t *testing.T) {
	p := createTestProfile(t)
	result, err := List(p, ListConfig{Pattern: "main.*"})
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(result.MatchedFunctions) != 1 {
		t.Fatalf("expected 1 match, got %d", len(result.MatchedFunctions))
	}
	if result.MatchedFunctions[0].Function == nil || *result.MatchedFunctions[0].Function != "main.foo" {
		t.Errorf("expected function 'main.foo'")
	}
}

func TestListNoMatch(t *testing.T) {
	p := createTestProfile(t)
	result, err := List(p, ListConfig{Pattern: "nonexistent.*"})
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(result.MatchedFunctions) != 0 {
		t.Errorf("expected 0 matches, got %d", len(result.MatchedFunctions))
	}
}

func TestFlameCommand(t *testing.T) {
	p := createTestProfile(t)
	result, err := Flame(p, FlameConfig{})
	if err != nil {
		t.Fatalf("Flame failed: %v", err)
	}
	if len(result.Stacks) != 1 {
		t.Fatalf("expected 1 stack, got %d", len(result.Stacks))
	}
	if result.Stacks[0].Count != 100 {
		t.Errorf("expected count 100, got %d", result.Stacks[0].Count)
	}
}

func TestFlameWithDepth(t *testing.T) {
	p := createTestProfile(t)
	result, err := Flame(p, FlameConfig{Depth: 1})
	if err != nil {
		t.Fatalf("Flame failed: %v", err)
	}
	if len(result.Stacks) != 1 {
		t.Fatalf("expected 1 stack, got %d", len(result.Stacks))
	}
}

func TestDiffCommand(t *testing.T) {
	base := createTestProfile(t)

	// Create target with different values
	rawTarget := &pprofprofile.Profile{
		PeriodType: &pprofprofile.ValueType{Type: "cpu", Unit: "nanoseconds"},
		Period:     10000000,
		SampleType: []*pprofprofile.ValueType{
			{Type: "samples", Unit: "count"},
		},
	}
	fn := &pprofprofile.Function{ID: 1, Name: "main.foo", Filename: "main.go"}
	m := &pprofprofile.Mapping{ID: 1, Start: 0x1000, Limit: 0x2000, File: "/bin/app"}
	loc := &pprofprofile.Location{ID: 1, Mapping: m, Address: 0x1100, Line: []pprofprofile.Line{{Function: fn, Line: 10}}}
	rawTarget.Function = []*pprofprofile.Function{fn}
	rawTarget.Mapping = []*pprofprofile.Mapping{m}
	rawTarget.Location = []*pprofprofile.Location{loc}
	rawTarget.Sample = []*pprofprofile.Sample{
		{Location: []*pprofprofile.Location{loc}, Value: []int64{200}}, // Doubled
	}
	target := NewProfile(rawTarget)

	result, err := Diff(base, target, DiffConfig{})
	if err != nil {
		t.Fatalf("Diff failed: %v", err)
	}

	// Should show improvement (doubled = less work per sample means... actually it's more samples, so regression)
	if len(result.Regressions)+len(result.Improvements) == 0 {
		t.Error("expected some deltas")
	}
}

func TestDiffNilProfile(t *testing.T) {
	_, err := Diff(nil, nil, DiffConfig{})
	if err == nil {
		t.Fatal("expected error for nil profiles")
	}
}
