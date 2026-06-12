package profile

import (
	"fmt"
	"testing"

	pprofprofile "github.com/google/pprof/profile"
)

// buildLargeProfile creates a profile with many samples for benchmarking
func buildLargeProfile(b *testing.B) *pprofprofile.Profile {
	const numFunctions = 100
	const numSamples = 10000

	funcs := make([]*pprofprofile.Function, numFunctions)
	locs := make([]*pprofprofile.Location, numFunctions)

	m := &pprofprofile.Mapping{ID: 1, Start: 0x1000, Limit: 0x100000, File: "/bin/app"}

	for i := 0; i < numFunctions; i++ {
		funcs[i] = &pprofprofile.Function{
			ID:       uint64(i + 1),
			Name:     fmt.Sprintf("pkg%d.Func%d", i/10, i),
			Filename: fmt.Sprintf("pkg%d/main.go", i/10),
		}
		locs[i] = &pprofprofile.Location{
			ID:      uint64(i + 1),
			Mapping: m,
			Address: uint64(0x1000 + i*0x100),
			Line:    []pprofprofile.Line{{Function: funcs[i], Line: int64(i + 1)}},
		}
	}

	samples := make([]*pprofprofile.Sample, numSamples)
	for i := 0; i < numSamples; i++ {
		// Create a stack of depth 5-10
		depth := 5 + (i % 6)
		stack := make([]*pprofprofile.Location, depth)
		for j := 0; j < depth; j++ {
			stack[j] = locs[(i+j)%numFunctions]
		}
		samples[i] = &pprofprofile.Sample{
			Location: stack,
			Value:    []int64{int64(i%100 + 1)},
		}
	}

	return &pprofprofile.Profile{
		PeriodType:    &pprofprofile.ValueType{Type: "cpu", Unit: "nanoseconds"},
		Period:        10000000,
		DurationNanos: 30e9,
		SampleType:    []*pprofprofile.ValueType{{Type: "samples", Unit: "count"}},
		Function:      funcs,
		Mapping:       []*pprofprofile.Mapping{m},
		Location:      locs,
		Sample:        samples,
	}
}

func BenchmarkAnalysis(b *testing.B) {
	p := buildLargeProfile(b)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := NewAnalysis(p, AnalysisConfig{TopN: 15, CallDepth: 5})
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkFlame(b *testing.B) {
	p := buildLargeProfile(b)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := Flame(p, FlameConfig{})
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkList(b *testing.B) {
	p := buildLargeProfile(b)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := List(p, ListConfig{Pattern: "pkg1.*"})
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkDiff(b *testing.B) {
	base := buildLargeProfile(b)
	target := buildLargeProfile(b)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := Diff(base, target, DiffConfig{})
		if err != nil {
			b.Fatal(err)
		}
	}
}
