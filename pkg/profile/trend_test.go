package profile

import (
	"testing"

	"github.com/google/pprof/profile"
)

func TestTrendRequiresMinimumProfiles(t *testing.T) {
	p := createTestCPUProfile(100)
	tp := TimePoint{Label: "p1", Time: 1}

	_, err := Trend([]*profile.Profile{p}, []TimePoint{tp}, TrendConfig{})
	if err == nil {
		t.Fatal("expected error for < 3 profiles")
	}

	_, err = Trend([]*profile.Profile{p, p}, []TimePoint{{Label: "a", Time: 1}, {Label: "b", Time: 2}}, TrendConfig{})
	if err == nil {
		t.Fatal("expected error for 2 profiles")
	}
}

func TestTrendProfileTimePointMismatch(t *testing.T) {
	p := createTestCPUProfile(100)
	tp := []TimePoint{{Label: "a", Time: 1}, {Label: "b", Time: 2}, {Label: "c", Time: 3}}

	_, err := Trend([]*profile.Profile{p, p}, tp, TrendConfig{})
	if err == nil {
		t.Fatal("expected error for profile/timepoint count mismatch")
	}
}

func TestTrendBasic(t *testing.T) {
	// Create 3 profiles with increasing mallocgc flat values
	p1 := createSingleFuncProfile("runtime.mallocgc", 100, "runtime/malloc.go")
	p2 := createSingleFuncProfile("runtime.mallocgc", 200, "runtime/malloc.go")
	p3 := createSingleFuncProfile("runtime.mallocgc", 350, "runtime/malloc.go")

	tp := []TimePoint{
		{Label: "p1", Time: 1},
		{Label: "p2", Time: 2},
		{Label: "p3", Time: 3},
	}

	result, err := Trend([]*profile.Profile{p1, p2, p3}, tp, TrendConfig{Threshold: 5, TopN: 10})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.RegressingCount != 1 {
		t.Errorf("expected 1 regressing, got %d", result.RegressingCount)
	}
	if len(result.TopRegressions) != 1 {
		t.Fatalf("expected 1 regression, got %d", len(result.TopRegressions))
	}
	if result.TopRegressions[0].Trend != "regressing" {
		t.Errorf("expected regressing, got %s", result.TopRegressions[0].Trend)
	}
	if result.TopRegressions[0].Slope <= 0 {
		t.Errorf("expected positive slope for increasing values, got %f", result.TopRegressions[0].Slope)
	}
}

func TestTrendImproving(t *testing.T) {
	p1 := createSingleFuncProfile("runtime.mallocgc", 300, "runtime/malloc.go")
	p2 := createSingleFuncProfile("runtime.mallocgc", 200, "runtime/malloc.go")
	p3 := createSingleFuncProfile("runtime.mallocgc", 100, "runtime/malloc.go")

	tp := []TimePoint{{Label: "p1", Time: 1}, {Label: "p2", Time: 2}, {Label: "p3", Time: 3}}

	result, err := Trend([]*profile.Profile{p1, p2, p3}, tp, TrendConfig{Threshold: 5})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.ImprovingCount != 1 {
		t.Errorf("expected 1 improving, got %d", result.ImprovingCount)
	}
}

func TestTrendStable(t *testing.T) {
	p1 := createSingleFuncProfile("runtime.mallocgc", 100, "runtime/malloc.go")
	p2 := createSingleFuncProfile("runtime.mallocgc", 102, "runtime/malloc.go")
	p3 := createSingleFuncProfile("runtime.mallocgc", 98, "runtime/malloc.go")

	tp := []TimePoint{{Label: "p1", Time: 1}, {Label: "p2", Time: 2}, {Label: "p3", Time: 3}}

	result, err := Trend([]*profile.Profile{p1, p2, p3}, tp, TrendConfig{Threshold: 5})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.StableCount != 1 {
		t.Errorf("expected 1 stable, got %d", result.StableCount)
	}
}

func TestTrendMissingFunction(t *testing.T) {
	// p1 and p3 have funcA, p2 has only funcB (different location)
	p1 := createSingleFuncProfile("funcA", 100, "main.go")
	p2 := createSingleFuncProfile("funcB", 50, "main.go")
	p3 := createSingleFuncProfile("funcA", 200, "main.go")

	tp := []TimePoint{{Label: "p1", Time: 1}, {Label: "p2", Time: 2}, {Label: "p3", Time: 3}}

	result, err := Trend([]*profile.Profile{p1, p2, p3}, tp, TrendConfig{Threshold: 5, MinImpact: 0})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// funcA should be in results with nil at p2
	var funcA *FunctionTrend
	for i := range result.Functions {
		if result.Functions[i].Function != nil && *result.Functions[i].Function == "funcA" {
			funcA = &result.Functions[i]
			break
		}
	}
	if funcA == nil {
		t.Fatal("funcA not found in results")
	}
	if funcA.FlatSeries[1] != nil {
		t.Error("expected nil for missing time point p2")
	}
	if funcA.FlatSeries[0] == nil || *funcA.FlatSeries[0] != 100 {
		t.Error("expected 100 at p1")
	}
	if funcA.FlatSeries[2] == nil || *funcA.FlatSeries[2] != 200 {
		t.Error("expected 200 at p3")
	}
}

func TestTrendFocusFilter(t *testing.T) {
	p1 := createMultiFuncProfile(map[string]int64{"runtime.mallocgc": 100, "main.handle": 50})
	p2 := createMultiFuncProfile(map[string]int64{"runtime.mallocgc": 200, "main.handle": 100})
	p3 := createMultiFuncProfile(map[string]int64{"runtime.mallocgc": 350, "main.handle": 150})

	tp := []TimePoint{{Label: "p1", Time: 1}, {Label: "p2", Time: 2}, {Label: "p3", Time: 3}}

	result, err := Trend([]*profile.Profile{p1, p2, p3}, tp, TrendConfig{
		Threshold: 5,
		FocusPattern: "^main\\.",
		MinImpact:   0,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should only have main.handle
	for _, ft := range result.Functions {
		if ft.Function != nil && *ft.Function != "main.handle" {
			t.Errorf("unexpected function in results: %s", *ft.Function)
		}
	}
}

func TestTrendMinImpactFilter(t *testing.T) {
	p1 := createMultiFuncProfile(map[string]int64{"runtime.mallocgc": 1000, "tiny.func": 1})
	p2 := createMultiFuncProfile(map[string]int64{"runtime.mallocgc": 1100, "tiny.func": 1})
	p3 := createMultiFuncProfile(map[string]int64{"runtime.mallocgc": 1200, "tiny.func": 1})

	tp := []TimePoint{{Label: "p1", Time: 1}, {Label: "p2", Time: 2}, {Label: "p3", Time: 3}}

	result, err := Trend([]*profile.Profile{p1, p2, p3}, tp, TrendConfig{
		Threshold: 5,
		MinImpact: 1, // tiny.func is < 1% at every point
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for _, ft := range result.Functions {
		if ft.Function != nil && *ft.Function == "tiny.func" {
			t.Error("tiny.func should be filtered by min-impact")
		}
	}
}

func TestTrendNewHotspots(t *testing.T) {
	// p1-p2: only funcA; p3: funcA + newFunc (appears late, significant)
	p1 := createSingleFuncProfile("funcA", 100, "main.go")
	p2 := createSingleFuncProfile("funcA", 100, "main.go")
	p3 := createMultiFuncProfile(map[string]int64{"funcA": 100, "newFunc": 200})

	tp := []TimePoint{{Label: "p1", Time: 1}, {Label: "p2", Time: 2}, {Label: "p3", Time: 3}}

	result, err := Trend([]*profile.Profile{p1, p2, p3}, tp, TrendConfig{
		Threshold:   5,
		MinImpact:   0,
		IncludeNew: true,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	found := false
	for _, ft := range result.NewHotspots {
		if ft.Function != nil && *ft.Function == "newFunc" {
			found = true
		}
	}
	if !found {
		t.Error("expected newFunc in new hotspots")
	}
}

func TestTrendVolatile(t *testing.T) {
	// Values oscillate wildly but average stays same → volatile
	p1 := createSingleFuncProfile("volatile.func", 100, "main.go")
	p2 := createSingleFuncProfile("volatile.func", 200, "main.go")
	p3 := createSingleFuncProfile("volatile.func", 50, "main.go")

	tp := []TimePoint{{Label: "p1", Time: 1}, {Label: "p2", Time: 2}, {Label: "p3", Time: 3}}

	result, err := Trend([]*profile.Profile{p1, p2, p3}, tp, TrendConfig{
		Threshold:       50, // high threshold so it stays "stable"
		MinImpact:       0,
		IncludeVolatile: true,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.VolatileFunctions) == 0 {
		t.Error("expected volatile function to be detected")
	}
}

func TestTrendOverallSlope(t *testing.T) {
	p1 := createSingleFuncProfile("funcA", 100, "main.go")
	p2 := createSingleFuncProfile("funcA", 200, "main.go")
	p3 := createSingleFuncProfile("funcA", 300, "main.go")

	tp := []TimePoint{{Label: "p1", Time: 1}, {Label: "p2", Time: 2}, {Label: "p3", Time: 3}}

	result, err := Trend([]*profile.Profile{p1, p2, p3}, tp, TrendConfig{MinImpact: 0})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Overall.Slope <= 0 {
		t.Errorf("expected positive overall slope, got %f", result.Overall.Slope)
	}
}

func TestLinearRegression(t *testing.T) {
	tests := []struct {
		name     string
		x        []int
		y        []float64
		wantSlope float64
	}{
		{"increasing", []int{0, 1, 2}, []float64{0, 10, 20}, 10},
		{"flat", []int{0, 1, 2}, []float64{5, 5, 5}, 0},
		{"single point", []int{0}, []float64{5}, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := linearRegression(tt.x, tt.y)
			if got != tt.wantSlope {
				t.Errorf("linearRegression() = %f, want %f", got, tt.wantSlope)
			}
		})
	}
}

// Helper functions

func createTestCPUProfile(totalSamples int) *profile.Profile {
	mallocgc := &profile.Function{ID: 1, Name: "runtime.mallocgc", SystemName: "runtime.mallocgc", Filename: "runtime/malloc.go"}
	mapping := &profile.Mapping{ID: 1, Start: 0x1000, Limit: 0x2000, File: "/usr/local/bin/myapp"}
	loc := &profile.Location{ID: 1, Mapping: mapping, Address: 0x1100, Line: []profile.Line{{Function: mallocgc, Line: 1020}}}

	return &profile.Profile{
		PeriodType:    &profile.ValueType{Type: "cpu", Unit: "nanoseconds"},
		SampleType:    []*profile.ValueType{{Type: "samples", Unit: "count"}, {Type: "cpu", Unit: "nanoseconds"}},
		Function:      []*profile.Function{mallocgc},
		Mapping:       []*profile.Mapping{mapping},
		Location:      []*profile.Location{loc},
		Sample:        []*profile.Sample{{Location: []*profile.Location{loc}, Value: []int64{int64(totalSamples), int64(totalSamples) * 10000000}}},
	}
}

func createSingleFuncProfile(funcName string, flatValue int, filename string) *profile.Profile {
	fn := &profile.Function{ID: 1, Name: funcName, SystemName: funcName, Filename: filename}
	mapping := &profile.Mapping{ID: 1, Start: 0x1000, Limit: 0x2000, File: "/usr/local/bin/myapp"}
	loc := &profile.Location{ID: 1, Mapping: mapping, Address: 0x1100, Line: []profile.Line{{Function: fn, Line: 10}}}

	return &profile.Profile{
		PeriodType:    &profile.ValueType{Type: "cpu", Unit: "nanoseconds"},
		SampleType:    []*profile.ValueType{{Type: "samples", Unit: "count"}, {Type: "cpu", Unit: "nanoseconds"}},
		Function:      []*profile.Function{fn},
		Mapping:       []*profile.Mapping{mapping},
		Location:      []*profile.Location{loc},
		Sample:        []*profile.Sample{{Location: []*profile.Location{loc}, Value: []int64{int64(flatValue), int64(flatValue) * 10000000}}},
	}
}

func createMultiFuncProfile(funcs map[string]int64) *profile.Profile {
	var fns []*profile.Function
	var locs []*profile.Location
	var samples []*profile.Sample

	mapping := &profile.Mapping{ID: 1, Start: 0x1000, Limit: 0x2000, File: "/usr/local/bin/myapp"}

	i := 1
	for name, val := range funcs {
		fn := &profile.Function{ID: uint64(i), Name: name, SystemName: name, Filename: "main.go"}
		loc := &profile.Location{ID: uint64(i), Mapping: mapping, Address: 0x1000 + uint64(i)*0x100, Line: []profile.Line{{Function: fn, Line: 10}}}
		fns = append(fns, fn)
		locs = append(locs, loc)
		samples = append(samples, &profile.Sample{
			Location: []*profile.Location{loc},
			Value:    []int64{val, val * 10000000},
		})
		i++
	}

	return &profile.Profile{
		PeriodType: &profile.ValueType{Type: "cpu", Unit: "nanoseconds"},
		SampleType: []*profile.ValueType{{Type: "samples", Unit: "count"}, {Type: "cpu", Unit: "nanoseconds"}},
		Function:   fns,
		Mapping:    []*profile.Mapping{mapping},
		Location:   locs,
		Sample:     samples,
	}
}
