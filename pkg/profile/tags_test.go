package profile

import (
	"testing"

	pprofprofile "github.com/google/pprof/profile"
)

func TestTags_NoLabels(t *testing.T) {
	p := NewProfile(&pprofprofile.Profile{
		PeriodType: &pprofprofile.ValueType{Type: "cpu", Unit: "nanoseconds"},
		SampleType: []*pprofprofile.ValueType{{Type: "samples", Unit: "count"}},
		Sample: []*pprofprofile.Sample{
			{Location: []*pprofprofile.Location{}, Value: []int64{1}},
			{Location: []*pprofprofile.Location{}, Value: []int64{1}},
		},
	})
	result, err := Tags(p, "cpu.pb.gz", 50)
	if err != nil {
		t.Fatalf("Tags failed: %v", err)
	}
	if len(result.Labels) != 0 {
		t.Errorf("expected 0 labels, got %d", len(result.Labels))
	}
	if result.TotalSamples != 2 {
		t.Errorf("expected 2 samples, got %d", result.TotalSamples)
	}
	if result.Type != "cpu" {
		t.Errorf("expected type=cpu, got %s", result.Type)
	}
}

func TestTags_WithLabels(t *testing.T) {
	p := makeStringLabelProfile()
	result, err := Tags(p, "goroutine.pb.gz", 50)
	if err != nil {
		t.Fatalf("Tags failed: %v", err)
	}
	if len(result.Labels) != 2 {
		t.Fatalf("expected 2 labels (state, wait_reason), got %d", len(result.Labels))
	}
	keys := map[string]bool{}
	for _, ls := range result.Labels {
		keys[ls.Key] = true
	}
	if !keys["state"] || !keys["wait_reason"] {
		t.Errorf("missing expected label keys: %v", keys)
	}
}

func TestTags_TopNTruncatesNumeric(t *testing.T) {
	// Build a numeric label with 60 distinct values, all count 1.
	samples := make([]*pprofprofile.Sample, 0, 60)
	for i := 0; i < 60; i++ {
		samples = append(samples, &pprofprofile.Sample{
			Location: []*pprofprofile.Location{},
			NumLabel: map[string][]int64{"cpu": {int64(1000 + i)}},
			NumUnit:  map[string][]string{"cpu": {"nanoseconds"}},
			Value:    []int64{1},
		})
	}
	p := NewProfile(&pprofprofile.Profile{
		PeriodType: &pprofprofile.ValueType{Type: "cpu", Unit: "nanoseconds"},
		SampleType: []*pprofprofile.ValueType{{Type: "cpu", Unit: "nanoseconds"}},
		Sample:     samples,
	})

	// top=20 should cap the numeric label to 20 values.
	result, err := Tags(p, "cpu.pb.gz", 20)
	if err != nil {
		t.Fatalf("Tags failed: %v", err)
	}
	if len(result.Labels) != 1 {
		t.Fatalf("expected 1 label, got %d", len(result.Labels))
	}
	ls := result.Labels[0]
	if len(ls.Values) != 20 {
		t.Errorf("expected top=20 to limit values to 20, got %d", len(ls.Values))
	}
	if ls.Distinct != 60 {
		t.Errorf("expected distinct=60, got %d", ls.Distinct)
	}
}
