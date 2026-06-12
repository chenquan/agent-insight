package profile

import (
	"testing"

	pprofprofile "github.com/google/pprof/profile"
)

func TestNoSymbolInfo(t *testing.T) {
	p := &pprofprofile.Profile{
		PeriodType: &pprofprofile.ValueType{Type: "cpu", Unit: "nanoseconds"},
		Period:     10000000,
		SampleType: []*pprofprofile.ValueType{
			{Type: "samples", Unit: "count"},
		},
	}

	mapping := &pprofprofile.Mapping{ID: 1, Start: 0x1000, Limit: 0x2000, File: "/usr/local/bin/myapp"}
	p.Mapping = []*pprofprofile.Mapping{mapping}

	// Location without any symbol info
	noSymLoc := &pprofprofile.Location{ID: 1, Mapping: mapping, Address: 0x1234}
	// Location with symbol info
	symFunc := &pprofprofile.Function{ID: 1, Name: "main.main", Filename: "main.go"}
	symLoc := &pprofprofile.Location{ID: 2, Mapping: mapping, Address: 0x2000, Line: []pprofprofile.Line{{Function: symFunc, Line: 10}}}

	p.Function = []*pprofprofile.Function{symFunc}
	p.Location = []*pprofprofile.Location{noSymLoc, symLoc}

	p.Sample = []*pprofprofile.Sample{
		{Location: []*pprofprofile.Location{noSymLoc}, Value: []int64{100}},
		{Location: []*pprofprofile.Location{symLoc}, Value: []int64{200}},
	}

	analysis, err := NewAnalysis(p, AnalysisConfig{TopN: 10, CallDepth: 0})
	if err != nil {
		t.Fatalf("NewAnalysis failed: %v", err)
	}

	if len(analysis.Hotspots) != 2 {
		t.Fatalf("expected 2 hotspots, got %d", len(analysis.Hotspots))
	}

	// Check that symbol info is correctly handled
	for _, h := range analysis.Hotspots {
		if h.Function != nil {
			// This hotspot should have symbol info
			if *h.Function != "main.main" {
				t.Errorf("expected function 'main.main', got '%s'", *h.Function)
			}
		} else {
			// This hotspot should have fallback info
			if h.Address == nil || *h.Address != "0x1234" {
				t.Errorf("expected address '0x1234', got %v", h.Address)
			}
			if h.Module == nil || *h.Module != "/usr/local/bin/myapp" {
				t.Errorf("expected module '/usr/local/bin/myapp', got %v", h.Module)
			}
		}
	}
}
