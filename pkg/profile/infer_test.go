package profile

import (
	"testing"

	pprofprofile "github.com/google/pprof/profile"
)

func TestInferProfileType_HeapFromNilPeriodType(t *testing.T) {
	p := &pprofprofile.Profile{
		PeriodType: nil,
		SampleType: []*pprofprofile.ValueType{
			{Type: "alloc_objects", Unit: "count"},
			{Type: "alloc_space", Unit: "bytes"},
			{Type: "inuse_objects", Unit: "count"},
			{Type: "inuse_space", Unit: "bytes"},
		},
	}

	analysis, err := NewAnalysis(p, AnalysisConfig{TopN: 5, CallDepth: 0})
	if err != nil {
		t.Fatalf("NewAnalysis failed: %v", err)
	}
	if analysis.Metadata.Type != "heap" {
		t.Errorf("expected type=heap, got %q", analysis.Metadata.Type)
	}
}

func TestInferProfileType_CPU(t *testing.T) {
	p := &pprofprofile.Profile{
		PeriodType: nil,
		SampleType: []*pprofprofile.ValueType{
			{Type: "samples", Unit: "count"},
			{Type: "cpu", Unit: "nanoseconds"},
		},
	}

	got := inferProfileType(p)
	if got != "cpu" {
		t.Errorf("expected cpu, got %q", got)
	}
}

func TestInferProfileType_Unknown(t *testing.T) {
	p := &pprofprofile.Profile{
		PeriodType: nil,
		SampleType: []*pprofprofile.ValueType{
			{Type: "custom_metric", Unit: "widgets"},
		},
	}

	got := inferProfileType(p)
	if got != "unknown" {
		t.Errorf("expected unknown, got %q", got)
	}
}

func TestInferProfileType_PeriodTypeWinsOverInference(t *testing.T) {
	p := &pprofprofile.Profile{
		PeriodType: &pprofprofile.ValueType{Type: "cpu", Unit: "nanoseconds"},
		SampleType: []*pprofprofile.ValueType{
			{Type: "inuse_space", Unit: "bytes"},
		},
	}

	analysis, err := NewAnalysis(p, AnalysisConfig{TopN: 5, CallDepth: 0})
	if err != nil {
		t.Fatalf("NewAnalysis failed: %v", err)
	}
	if analysis.Metadata.Type != "cpu" {
		t.Errorf("PeriodType should win, expected cpu, got %q", analysis.Metadata.Type)
	}
}
