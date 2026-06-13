package profile

import (
	"testing"

	pprofprofile "github.com/google/pprof/profile"
)

func TestMultiValueType(t *testing.T) {
	// Go heap profile has 4 value types
	p := &pprofprofile.Profile{
		PeriodType: &pprofprofile.ValueType{Type: "space", Unit: "bytes"},
		Period:     524288,
		SampleType: []*pprofprofile.ValueType{
			{Type: "alloc_objects", Unit: "count"},
			{Type: "alloc_space", Unit: "bytes"},
			{Type: "inuse_objects", Unit: "count"},
			{Type: "inuse_space", Unit: "bytes"},
		},
	}

	fn := &pprofprofile.Function{ID: 1, Name: "main.makeSlice", Filename: "main.go"}
	mapping := &pprofprofile.Mapping{ID: 1, Start: 0x1000, Limit: 0x2000, File: "/usr/local/bin/myapp"}
	loc := &pprofprofile.Location{ID: 1, Mapping: mapping, Address: 0x1300, Line: []pprofprofile.Line{{Function: fn, Line: 30}}}

	p.Function = []*pprofprofile.Function{fn}
	p.Mapping = []*pprofprofile.Mapping{mapping}
	p.Location = []*pprofprofile.Location{loc}
	p.Sample = []*pprofprofile.Sample{
		{Location: []*pprofprofile.Location{loc}, Value: []int64{1000, 838860800, 500, 419430400}},
	}

	// Test: default value type should be selected automatically
	analysis, err := NewAnalysis(p, AnalysisConfig{TopN: 10, CallDepth: 0})
	if err != nil {
		t.Fatalf("NewAnalysis failed: %v", err)
	}

	// For heap profiles, default should prefer inuse_bytes (index 3)
	if analysis.Config.ValueType == nil {
		t.Fatal("expected ValueType to be set")
	}

	// Verify hotspot has correct flat value
	if len(analysis.Hotspots) != 1 {
		t.Fatalf("expected 1 hotspot, got %d", len(analysis.Hotspots))
	}

	// Verify that all sample types are reported
	if len(analysis.Metadata.SampleTypes) != 4 {
		t.Errorf("expected 4 sample types, got %d", len(analysis.Metadata.SampleTypes))
	}

	t.Logf("Selected value type: %s/%s (index %d)", analysis.Config.ValueType.Name, analysis.Config.ValueType.Unit, analysis.Config.ValueType.Index)
	t.Logf("Hotspot flat: %d, cum: %d", analysis.Hotspots[0].FlatValue, analysis.Hotspots[0].CumValue)
}

func TestUserSpecifiedValueType(t *testing.T) {
	p := &pprofprofile.Profile{
		PeriodType: &pprofprofile.ValueType{Type: "space", Unit: "bytes"},
		Period:     524288,
		SampleType: []*pprofprofile.ValueType{
			{Type: "alloc_objects", Unit: "count"},
			{Type: "alloc_space", Unit: "bytes"},
			{Type: "inuse_objects", Unit: "count"},
			{Type: "inuse_space", Unit: "bytes"},
		},
	}

	fn := &pprofprofile.Function{ID: 1, Name: "main.makeSlice", Filename: "main.go"}
	loc := &pprofprofile.Location{ID: 1, Address: 0x1300, Line: []pprofprofile.Line{{Function: fn, Line: 30}}}

	p.Function = []*pprofprofile.Function{fn}
	p.Location = []*pprofprofile.Location{loc}
	p.Sample = []*pprofprofile.Sample{
		{Location: []*pprofprofile.Location{loc}, Value: []int64{1000, 838860800, 500, 419430400}},
	}

	// Analyze with alloc_objects (index 0) — value should be 1000
	allocAnalysis, err := NewAnalysis(p, AnalysisConfig{
		TopN: 10,
		ValueType: &ValueTypeConfig{Name: "alloc_objects", Unit: "count", Index: 0},
	})
	if err != nil {
		t.Fatalf("alloc_objects analysis failed: %v", err)
	}
	if len(allocAnalysis.Hotspots) == 0 {
		t.Fatal("expected at least 1 hotspot for alloc_objects")
	}
	if allocAnalysis.Hotspots[0].FlatValue != 1000 {
		t.Errorf("alloc_objects: expected flat=1000, got %d", allocAnalysis.Hotspots[0].FlatValue)
	}

	// Analyze with inuse_space (index 3) — value should be 419430400
	inuseAnalysis, err := NewAnalysis(p, AnalysisConfig{
		TopN: 10,
		ValueType: &ValueTypeConfig{Name: "inuse_space", Unit: "bytes", Index: 3},
	})
	if err != nil {
		t.Fatalf("inuse_space analysis failed: %v", err)
	}
	if len(inuseAnalysis.Hotspots) == 0 {
		t.Fatal("expected at least 1 hotspot for inuse_space")
	}
	if inuseAnalysis.Hotspots[0].FlatValue != 419430400 {
		t.Errorf("inuse_space: expected flat=419430400, got %d", inuseAnalysis.Hotspots[0].FlatValue)
	}

	// Verify the two results are actually different
	if allocAnalysis.Hotspots[0].FlatValue == inuseAnalysis.Hotspots[0].FlatValue {
		t.Errorf("alloc_objects and inuse_space should produce different flat values, both got %d", allocAnalysis.Hotspots[0].FlatValue)
	}

	// Verify analyzed_type reflects the user choice
	if allocAnalysis.Config.ValueType.Name != "alloc_objects" {
		t.Errorf("alloc_analysis: expected ValueType.Name=alloc_objects, got %s", allocAnalysis.Config.ValueType.Name)
	}
	if inuseAnalysis.Config.ValueType.Name != "inuse_space" {
		t.Errorf("inuse_analysis: expected ValueType.Name=inuse_space, got %s", inuseAnalysis.Config.ValueType.Name)
	}
}
