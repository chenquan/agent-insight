package profile

import (
	"os"
	"path/filepath"
	"testing"

	pprofprofile "github.com/google/pprof/profile"
)

func TestMergeTwoProfiles(t *testing.T) {
	p1 := buildCPUProfile(t, 2, 100)
	p2 := buildCPUProfile(t, 3, 200)

	merged, result, err := ValidateAndMerge([]*Profile{p1, p2})
	if err != nil {
		t.Fatalf("ValidateAndMerge failed: %v", err)
	}

	if result.InputCount != 2 {
		t.Errorf("expected InputCount 2, got %d", result.InputCount)
	}
	if result.ValueType != "cpu" {
		t.Errorf("expected ValueType 'cpu', got '%s'", result.ValueType)
	}
	if merged == nil {
		t.Fatal("expected non-nil merged profile")
	}

	// profile.Merge aggregates samples with the same stack trace.
	// Both profiles have the same single-location stack, so they merge into 1 sample.
	if len(merged.Sample) != 1 {
		t.Fatalf("expected 1 merged sample (aggregated), got %d", len(merged.Sample))
	}

	// Verify the values were summed: 100*2 + 200*3 = 800
	expectedValue := int64(100*2 + 200*3)
	if merged.Sample[0].Value[0] != expectedValue {
		t.Errorf("expected merged value %d, got %d", expectedValue, merged.Sample[0].Value[0])
	}
}

func TestMergeProfilesWithDifferentStacks(t *testing.T) {
	p1 := buildCPUProfileWithFunc(t, 2, 100, "funcA")
	p2 := buildCPUProfileWithFunc(t, 3, 200, "funcB")

	merged, result, err := ValidateAndMerge([]*Profile{p1, p2})
	if err != nil {
		t.Fatalf("ValidateAndMerge failed: %v", err)
	}

	if result.InputCount != 2 {
		t.Errorf("expected InputCount 2, got %d", result.InputCount)
	}

	// Different stacks should remain as separate samples
	if len(merged.Sample) != 2 {
		t.Fatalf("expected 2 merged samples (different stacks), got %d", len(merged.Sample))
	}
}

func TestMergeSingleProfile(t *testing.T) {
	p := buildCPUProfile(t, 2, 100)

	_, _, err := ValidateAndMerge([]*Profile{p})
	if err == nil {
		t.Fatal("expected error for single profile")
	}
}

func TestMergeMixedTypes(t *testing.T) {
	cpu := buildCPUProfile(t, 2, 100)
	heap := buildHeapProfile(t, 2)

	_, _, err := ValidateAndMerge([]*Profile{cpu, heap})
	if err == nil {
		t.Fatal("expected error for mixed types")
	}
	t.Logf("expected error: %v", err)
}

func TestMergeOutputReadable(t *testing.T) {
	p1 := buildCPUProfile(t, 3, 100)
	p2 := buildCPUProfile(t, 2, 200)

	merged, _, err := ValidateAndMerge([]*Profile{p1, p2})
	if err != nil {
		t.Fatalf("ValidateAndMerge failed: %v", err)
	}

	dir := t.TempDir()
	outPath := filepath.Join(dir, "merged.pb.gz")

	f, err := os.Create(outPath)
	if err != nil {
		t.Fatalf("create file: %v", err)
	}
	if err := merged.Write(f); err != nil {
		t.Fatalf("Write failed: %v", err)
	}
	f.Close()

	info, err := os.Stat(outPath)
	if err != nil {
		t.Fatalf("stat output: %v", err)
	}
	if info.Size() == 0 {
		t.Error("output file is empty")
	}

	// Verify it can be loaded back
	loader := NewLoader()
	loaded, err := loader.LoadFromFile(outPath)
	if err != nil {
		t.Fatalf("failed to load merged profile back: %v", err)
	}
	if len(loaded.Sample) != 1 {
		t.Errorf("expected 1 sample after reload, got %d", len(loaded.Sample))
	}
}

func TestMergeThreeProfiles(t *testing.T) {
	p1 := buildCPUProfileWithFunc(t, 1, 100, "funcA")
	p2 := buildCPUProfileWithFunc(t, 1, 200, "funcB")
	p3 := buildCPUProfileWithFunc(t, 1, 300, "funcC")

	merged, result, err := ValidateAndMerge([]*Profile{p1, p2, p3})
	if err != nil {
		t.Fatalf("ValidateAndMerge failed: %v", err)
	}

	if result.InputCount != 3 {
		t.Errorf("expected InputCount 3, got %d", result.InputCount)
	}
	if len(merged.Sample) != 3 {
		t.Errorf("expected 3 merged samples, got %d", len(merged.Sample))
	}
}

func buildCPUProfile(t *testing.T, numSamples int, value int64) *Profile {
	t.Helper()
	return buildCPUProfileWithFunc(t, numSamples, value, "main.work")
}

func buildCPUProfileWithFunc(t *testing.T, numSamples int, value int64, funcName string) *Profile {
	t.Helper()

	fn := &pprofprofile.Function{ID: 1, Name: funcName, Filename: "main.go"}
	loc := &pprofprofile.Location{ID: 1, Address: 0x1000, Line: []pprofprofile.Line{{Function: fn, Line: 10}}}

	var samples []*pprofprofile.Sample
	for range numSamples {
		samples = append(samples, &pprofprofile.Sample{
			Location: []*pprofprofile.Location{loc},
			Value:    []int64{value},
		})
	}

	return NewProfile(&pprofprofile.Profile{
		PeriodType: &pprofprofile.ValueType{Type: "cpu", Unit: "nanoseconds"},
		Period:     10000000,
		SampleType: []*pprofprofile.ValueType{{Type: "samples", Unit: "count"}},
		Function:   []*pprofprofile.Function{fn},
		Location:   []*pprofprofile.Location{loc},
		Sample:     samples,
	})
}

func buildHeapProfile(t *testing.T, numSamples int) *Profile {
	t.Helper()

	fn := &pprofprofile.Function{ID: 1, Name: "main.alloc", Filename: "main.go"}
	loc := &pprofprofile.Location{ID: 1, Address: 0x1000, Line: []pprofprofile.Line{{Function: fn, Line: 20}}}

	var samples []*pprofprofile.Sample
	for range numSamples {
		samples = append(samples, &pprofprofile.Sample{
			Location: []*pprofprofile.Location{loc},
			Value:    []int64{1024, 1},
		})
	}

	return NewProfile(&pprofprofile.Profile{
		PeriodType: &pprofprofile.ValueType{Type: "space", Unit: "bytes"},
		Period:     512,
		SampleType: []*pprofprofile.ValueType{
			{Type: "inuse_space", Unit: "bytes"},
			{Type: "inuse_objects", Unit: "count"},
		},
		Function: []*pprofprofile.Function{fn},
		Location: []*pprofprofile.Location{loc},
		Sample:   samples,
	})
}
