package profile

import (
	"strings"
	"testing"

	pprofprofile "github.com/google/pprof/profile"
)

// makeStringLabelProfile builds a profile with string labels on samples.
func makeStringLabelProfile() *Profile {
	return NewProfile(&pprofprofile.Profile{
		SampleType: []*pprofprofile.ValueType{{Type: "cpu", Unit: "nanoseconds"}},
		Sample: []*pprofprofile.Sample{
			{
				Location: []*pprofprofile.Location{},
				Label:    map[string][]string{"state": {"blocked"}, "wait_reason": {"IO"}},
				Value:    []int64{1},
			},
			{
				Location: []*pprofprofile.Location{},
				Label:    map[string][]string{"state": {"blocked"}, "wait_reason": {"IO"}},
				Value:    []int64{1},
			},
			{
				Location: []*pprofprofile.Location{},
				Label:    map[string][]string{"state": {"running"}},
				Value:    []int64{1},
			},
			{
				Location: []*pprofprofile.Location{},
				Label:    map[string][]string{"state": {"syscall"}},
				Value:    []int64{1},
			},
		},
	})
}

func TestExtractLabelSummaries_NoLabels(t *testing.T) {
	p := NewProfile(&pprofprofile.Profile{
		Sample: []*pprofprofile.Sample{
			{Location: []*pprofprofile.Location{}, Value: []int64{1}},
			{Location: []*pprofprofile.Location{}, Value: []int64{2}},
		},
	})
	if got := ExtractLabelSummaries(p); got != nil {
		t.Errorf("expected nil for profile with no labels, got %v", got)
	}
}

func TestExtractLabelSummaries_StringLabels(t *testing.T) {
	p := makeStringLabelProfile()
	summaries := ExtractLabelSummaries(p)
	if len(summaries) != 2 {
		t.Fatalf("expected 2 label summaries, got %d", len(summaries))
	}

	stateSum := summaries[0]
	if stateSum.Key != "state" {
		t.Errorf("expected first key=state, got %s", stateSum.Key)
	}
	if stateSum.Type != "string" {
		t.Errorf("expected type=string, got %s", stateSum.Type)
	}
	if stateSum.Distinct != 3 {
		t.Errorf("expected 3 distinct values, got %d", stateSum.Distinct)
	}
	if len(stateSum.Values) != 3 {
		t.Fatalf("expected 3 values, got %d", len(stateSum.Values))
	}
	if stateSum.Values[0].Value != "blocked" || stateSum.Values[0].Count != 2 {
		t.Errorf("expected top value=blocked count=2, got %s count=%d", stateSum.Values[0].Value, stateSum.Values[0].Count)
	}
	if stateSum.Values[1].Count != 1 {
		t.Errorf("expected second value count=1, got %d", stateSum.Values[1].Count)
	}
}

func TestExtractLabelSummaries_NumericLabels(t *testing.T) {
	p := NewProfile(&pprofprofile.Profile{
		SampleType: []*pprofprofile.ValueType{{Type: "cpu", Unit: "nanoseconds"}},
		Sample: []*pprofprofile.Sample{
			{
				Location:  []*pprofprofile.Location{},
				NumLabel:  map[string][]int64{"cpu": {1500000}},
				NumUnit:   map[string][]string{"cpu": {"nanoseconds"}},
				Value:     []int64{1},
			},
			{
				Location:  []*pprofprofile.Location{},
				NumLabel:  map[string][]int64{"cpu": {2000000}},
				NumUnit:   map[string][]string{"cpu": {"nanoseconds"}},
				Value:     []int64{1},
			},
		},
	})
	summaries := ExtractLabelSummaries(p)
	if len(summaries) != 1 {
		t.Fatalf("expected 1 summary, got %d", len(summaries))
	}
	ls := summaries[0]
	if ls.Type != "numeric" {
		t.Errorf("expected type=numeric, got %s", ls.Type)
	}
	if ls.Unit == nil || *ls.Unit != "nanoseconds" {
		t.Errorf("expected unit=nanoseconds, got %v", ls.Unit)
	}
	if ls.Distinct != 2 {
		t.Errorf("expected 2 distinct values, got %d", ls.Distinct)
	}
}

func TestExtractLabelSummaries_Mixed(t *testing.T) {
	raw := &pprofprofile.Profile{
		SampleType: []*pprofprofile.ValueType{{Type: "cpu", Unit: "nanoseconds"}},
		Sample: []*pprofprofile.Sample{
			{
				Location: []*pprofprofile.Location{},
				Label:    map[string][]string{"state": {"blocked"}},
				NumLabel: map[string][]int64{"cpu": {1500000}},
				NumUnit:  map[string][]string{"cpu": {"nanoseconds"}},
				Value:    []int64{1},
			},
			{
				Location: []*pprofprofile.Location{},
				Label:    map[string][]string{"state": {"running"}},
				NumLabel: map[string][]int64{"cpu": {2000000}},
				NumUnit:  map[string][]string{"cpu": {"nanoseconds"}},
				Value:    []int64{1},
			},
		},
	}
	p := NewProfile(raw)

	summaries := ExtractLabelSummaries(p)
	if len(summaries) != 2 {
		t.Fatalf("expected 2 summaries (state, cpu), got %d", len(summaries))
	}
	found := map[string]bool{}
	for _, ls := range summaries {
		found[ls.Key] = true
	}
	if !found["state"] || !found["cpu"] {
		t.Errorf("missing expected keys: %v", found)
	}
}

func TestNewLabelFilter_InvalidFormat(t *testing.T) {
	tests := []struct {
		name  string
		focus []string
	}{
		{"missing equals", []string{"stateblocked"}},
		{"empty key", []string{"=blocked"}},
		{"empty value", []string{"state="}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewLabelFilter(tt.focus, nil)
			if err == nil {
				t.Errorf("expected error for %q, got nil", tt.focus)
			}
		})
	}
}

func TestLabelFilter_Focus_WithinKeyOR(t *testing.T) {
	f, err := NewLabelFilter([]string{"state=blocked", "state=running"}, nil)
	if err != nil {
		t.Fatalf("NewLabelFilter failed: %v", err)
	}
	if len(f.Focus["state"]) != 2 {
		t.Errorf("expected 2 values for state, got %d", len(f.Focus["state"]))
	}

	p := makeStringLabelProfile()
	filtered, err := f.Apply(p)
	if err != nil {
		t.Fatalf("Apply failed: %v", err)
	}
	if len(filtered.Sample) != 3 {
		t.Errorf("expected 3 samples (2 blocked + 1 running), got %d", len(filtered.Sample))
	}
}

func TestLabelFilter_Focus_AcrossKeyAND(t *testing.T) {
	f, err := NewLabelFilter([]string{"state=blocked", "wait_reason=IO"}, nil)
	if err != nil {
		t.Fatalf("NewLabelFilter failed: %v", err)
	}
	p := makeStringLabelProfile()
	filtered, err := f.Apply(p)
	if err != nil {
		t.Fatalf("Apply failed: %v", err)
	}
	if len(filtered.Sample) != 2 {
		t.Errorf("expected 2 samples matching both keys, got %d", len(filtered.Sample))
	}
}

func TestLabelFilter_Ignore(t *testing.T) {
	f, err := NewLabelFilter(nil, []string{"state=syscall"})
	if err != nil {
		t.Fatalf("NewLabelFilter failed: %v", err)
	}
	p := makeStringLabelProfile()
	filtered, err := f.Apply(p)
	if err != nil {
		t.Fatalf("Apply failed: %v", err)
	}
	if len(filtered.Sample) != 3 {
		t.Errorf("expected 3 samples (syscall excluded), got %d", len(filtered.Sample))
	}
	for _, s := range filtered.Sample {
		for _, v := range s.Label["state"] {
			if v == "syscall" {
				t.Errorf("syscall sample should have been filtered out")
			}
		}
	}
}

func TestLabelFilter_Combined(t *testing.T) {
	f, err := NewLabelFilter(
		[]string{"state=blocked"},
		[]string{"wait_reason=GC"},
	)
	if err != nil {
		t.Fatalf("NewLabelFilter failed: %v", err)
	}
	p := makeStringLabelProfile()
	p.Sample = append(p.Sample, &pprofprofile.Sample{
		Location: []*pprofprofile.Location{},
		Label:    map[string][]string{"state": {"blocked"}, "wait_reason": {"GC"}},
		Value:    []int64{1},
	})
	filtered, err := f.Apply(p)
	if err != nil {
		t.Fatalf("Apply failed: %v", err)
	}
	if len(filtered.Sample) != 2 {
		t.Errorf("expected 2 samples (state=blocked but not wait_reason=GC), got %d", len(filtered.Sample))
	}
}

func TestLabelFilter_ZeroSamples(t *testing.T) {
	f, err := NewLabelFilter([]string{"state=blocked"}, nil)
	if err != nil {
		t.Fatalf("NewLabelFilter failed: %v", err)
	}
	p := NewProfile(&pprofprofile.Profile{
		Sample: []*pprofprofile.Sample{
			{Location: []*pprofprofile.Location{}, Label: map[string][]string{"state": {"running"}}, Value: []int64{1}},
		},
	})
	_, err = f.Apply(p)
	if err == nil {
		t.Fatal("expected error for 0 matching samples")
	}
	if !strings.Contains(err.Error(), "matched 0 of 1 samples") {
		t.Errorf("expected 'matched 0 of 1 samples' in error, got: %v", err)
	}
	if !strings.Contains(err.Error(), "check --tag key=value spelling") {
		t.Errorf("expected spelling hint in error, got: %v", err)
	}
}

func TestLabelFilter_EmptyFilter(t *testing.T) {
	f, err := NewLabelFilter(nil, nil)
	if err != nil {
		t.Fatalf("NewLabelFilter failed: %v", err)
	}
	p := makeStringLabelProfile()
	out, err := f.Apply(p)
	if err != nil {
		t.Fatalf("Apply failed: %v", err)
	}
	if out != p {
		t.Errorf("empty filter should return same profile pointer")
	}
}

func TestLabelFilter_NumericLabel(t *testing.T) {
	p := NewProfile(&pprofprofile.Profile{
		SampleType: []*pprofprofile.ValueType{{Type: "cpu", Unit: "nanoseconds"}},
		Sample: []*pprofprofile.Sample{
			{Location: []*pprofprofile.Location{}, NumLabel: map[string][]int64{"cpu": {1500000}}, NumUnit: map[string][]string{"cpu": {"nanoseconds"}}, Value: []int64{1}},
			{Location: []*pprofprofile.Location{}, NumLabel: map[string][]int64{"cpu": {2000000}}, NumUnit: map[string][]string{"cpu": {"nanoseconds"}}, Value: []int64{1}},
		},
	})
	f, err := NewLabelFilter([]string{"cpu=1500000"}, nil)
	if err != nil {
		t.Fatalf("NewLabelFilter failed: %v", err)
	}
	filtered, err := f.Apply(p)
	if err != nil {
		t.Fatalf("Apply failed: %v", err)
	}
	if len(filtered.Sample) != 1 {
		t.Errorf("expected 1 sample, got %d", len(filtered.Sample))
	}
}

func TestComputeFunctionBreakdowns_Empty(t *testing.T) {
	p := makeStringLabelProfile()
	tests := []struct {
		name string
		keys []string
		hs   []Hotspot
	}{
		{"no keys", nil, []Hotspot{{Function: ptrStr("blocked")}, {Function: ptrStr("running")}}},
		{"no hotspots", []string{"state"}, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ComputeFunctionBreakdowns(p, tt.hs, BreakdownConfig{Keys: tt.keys, Top: 5})
			if got != nil {
				t.Errorf("expected nil, got %v", got)
			}
		})
	}
}

func TestComputeFunctionBreakdowns_TopN(t *testing.T) {
	mkFn := func(name string) *pprofprofile.Function {
		return &pprofprofile.Function{ID: hashID(name), Name: name, Filename: "main.go"}
	}
	mkLoc := func(fn *pprofprofile.Function) *pprofprofile.Location {
		return &pprofprofile.Location{ID: hashID(fn.Name), Line: []pprofprofile.Line{{Function: fn, Line: 1}}}
	}
	fn1 := mkFn("fn1")
	fn2 := mkFn("fn2")
	fn3 := mkFn("fn3")
	p := NewProfile(&pprofprofile.Profile{
		SampleType: []*pprofprofile.ValueType{{Type: "cpu", Unit: "nanoseconds"}},
		Function:   []*pprofprofile.Function{fn1, fn2, fn3},
		Location:   []*pprofprofile.Location{mkLoc(fn1), mkLoc(fn2), mkLoc(fn3)},
		Sample: []*pprofprofile.Sample{
			{Location: []*pprofprofile.Location{mkLoc(fn1)}, Label: map[string][]string{"state": {"blocked"}}, Value: []int64{10}},
			{Location: []*pprofprofile.Location{mkLoc(fn2)}, Label: map[string][]string{"state": {"running"}}, Value: []int64{20}},
			{Location: []*pprofprofile.Location{mkLoc(fn3)}, Label: map[string][]string{"state": {"syscall"}}, Value: []int64{30}},
		},
	})
	hotspots := []Hotspot{
		{Function: ptrStr("fn1")},
		{Function: ptrStr("fn2")},
		{Function: ptrStr("fn3")},
	}
	got := ComputeFunctionBreakdowns(p, hotspots, BreakdownConfig{Keys: []string{"state"}, Top: 2})
	if len(got) != 2 {
		t.Errorf("expected 2 breakdown entries (Top=2), got %d", len(got))
	}
}

func TestComputeFunctionBreakdowns_FlatDistribution(t *testing.T) {
	mkFn := func(name string) *pprofprofile.Function {
		return &pprofprofile.Function{ID: hashID(name), Name: name, Filename: "main.go"}
	}
	mkLoc := func(fn *pprofprofile.Function) *pprofprofile.Location {
		return &pprofprofile.Location{ID: hashID(fn.Name), Line: []pprofprofile.Line{{Function: fn, Line: 1}}}
	}
	fn1 := mkFn("fn1")
	p := NewProfile(&pprofprofile.Profile{
		SampleType: []*pprofprofile.ValueType{{Type: "cpu", Unit: "nanoseconds"}},
		Function:   []*pprofprofile.Function{fn1},
		Location:   []*pprofprofile.Location{mkLoc(fn1)},
		Sample: []*pprofprofile.Sample{
			{Location: []*pprofprofile.Location{mkLoc(fn1)}, Label: map[string][]string{"state": {"blocked"}}, Value: []int64{60}},
			{Location: []*pprofprofile.Location{mkLoc(fn1)}, Label: map[string][]string{"state": {"blocked"}}, Value: []int64{60}},
			{Location: []*pprofprofile.Location{mkLoc(fn1)}, Label: map[string][]string{"state": {"blocked"}}, Value: []int64{60}},
			{Location: []*pprofprofile.Location{mkLoc(fn1)}, Label: map[string][]string{"state": {"blocked"}}, Value: []int64{60}},
			{Location: []*pprofprofile.Location{mkLoc(fn1)}, Label: map[string][]string{"state": {"blocked"}}, Value: []int64{60}},
			{Location: []*pprofprofile.Location{mkLoc(fn1)}, Label: map[string][]string{"state": {"blocked"}}, Value: []int64{60}},
			{Location: []*pprofprofile.Location{mkLoc(fn1)}, Label: map[string][]string{"state": {"running"}}, Value: []int64{40}},
			{Location: []*pprofprofile.Location{mkLoc(fn1)}, Label: map[string][]string{"state": {"running"}}, Value: []int64{40}},
			{Location: []*pprofprofile.Location{mkLoc(fn1)}, Label: map[string][]string{"state": {"running"}}, Value: []int64{40}},
			{Location: []*pprofprofile.Location{mkLoc(fn1)}, Label: map[string][]string{"state": {"running"}}, Value: []int64{40}},
		},
	})
	hotspots := []Hotspot{{Function: ptrStr("fn1")}}
	got := ComputeFunctionBreakdowns(p, hotspots, BreakdownConfig{Keys: []string{"state"}, Top: 1})
	if len(got) != 1 {
		t.Fatalf("expected 1 breakdown, got %d", len(got))
	}
	if len(got[0].Labels) != 1 {
		t.Fatalf("expected 1 label key, got %d", len(got[0].Labels))
	}
	values := got[0].Labels[0].Values
	if len(values) != 2 {
		t.Fatalf("expected 2 values (blocked + running), got %d", len(values))
	}
	if values[0].Value != "blocked" || values[0].Flat != 360 {
		t.Errorf("expected blocked flat=360, got %s flat=%d", values[0].Value, values[0].Flat)
	}
	if values[1].Value != "running" || values[1].Flat != 160 {
		t.Errorf("expected running flat=160, got %s flat=%d", values[1].Value, values[1].Flat)
	}
}

func ptrStr(s string) *string { return &s }

func hashID(s string) uint64 {
	var h uint64 = 1469598103934665603
	for i := 0; i < len(s); i++ {
		h ^= uint64(s[i])
		h *= 1099511628211
	}
	return h
}
