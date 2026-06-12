package profile

import (
	"fmt"
	"time"

	"github.com/google/pprof/profile"
)

// InfoResult contains profile metadata without performing sample-level computation.
type InfoResult struct {
	Type         string
	Duration     time.Duration
	Period       int64
	PeriodType   string
	SampleCount  int
	ValueTypes   []ValueTypeDesc
	Functions    int
	Locations    int
	HasSymbols   bool
	HasFileLines bool
	Mappings     []MappingInfo
	TimeRange    TimeRangeInfo
	Comments     []string
}

// ValueTypeDesc describes a sample value type.
type ValueTypeDesc struct {
	Type string
	Unit string
}

// MappingInfo describes a binary mapping in the profile.
type MappingInfo struct {
	File            string
	BuildID         string
	HasFunctions    bool
	HasFilenames    bool
	HasLineNumbers  bool
	HasInlineFrames bool
}

// TimeRangeInfo describes the time range of profile collection.
type TimeRangeInfo struct {
	Start   time.Time
	End     time.Time
	HasTime bool
}

// Info reads profile metadata without performing sample-level computation.
func Info(p *profile.Profile) (*InfoResult, error) {
	if p == nil {
		return nil, fmt.Errorf("profile is nil")
	}

	result := &InfoResult{
		SampleCount:  len(p.Sample),
		Functions:    len(p.Function),
		Locations:    len(p.Location),
		HasSymbols:   p.HasFunctions(),
		HasFileLines: p.HasFileLines(),
	}

	// Period type
	if p.PeriodType != nil {
		result.PeriodType = p.PeriodType.Type
		if p.PeriodType.Unit != "" {
			result.PeriodType = p.PeriodType.Type + "/" + p.PeriodType.Unit
		}
	}

	// Duration
	if p.DurationNanos > 0 {
		result.Duration = time.Duration(p.DurationNanos) * time.Nanosecond
	}

	// Period
	result.Period = p.Period

	// Value types
	for _, st := range p.SampleType {
		result.ValueTypes = append(result.ValueTypes, ValueTypeDesc{
			Type: st.Type,
			Unit: st.Unit,
		})
	}

	// Time range
	if p.TimeNanos > 0 {
		result.TimeRange.Start = time.Unix(0, p.TimeNanos)
		result.TimeRange.End = result.TimeRange.Start.Add(result.Duration)
		result.TimeRange.HasTime = true
	}

	// Mappings
	for _, m := range p.Mapping {
		result.Mappings = append(result.Mappings, MappingInfo{
			File:            m.File,
			BuildID:         m.BuildID,
			HasFunctions:    m.HasFunctions,
			HasFilenames:    m.HasFilenames,
			HasLineNumbers:  m.HasLineNumbers,
			HasInlineFrames: m.HasInlineFrames,
		})
	}

	// Comments
	result.Comments = p.Comments

	// Detect profile type from PeriodType
	if p.PeriodType != nil {
		result.Type = p.PeriodType.Type
	}

	return result, nil
}
