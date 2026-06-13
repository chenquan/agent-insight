package profile

import (
	"fmt"

	"github.com/google/pprof/profile"
)

// MergeResult contains statistics about a merge operation.
type MergeResult struct {
	InputCount   int
	TotalSamples int
	ValueType    string
}

// ValidateAndMerge validates profile type consistency and merges multiple profiles.
func ValidateAndMerge(profiles []*profile.Profile) (*profile.Profile, *MergeResult, error) {
	if len(profiles) < 2 {
		return nil, nil, fmt.Errorf("need at least 2 profiles to merge, got %d", len(profiles))
	}

	if err := ValidateTypeConsistency(profiles); err != nil {
		return nil, nil, err
	}

	merged, err := profile.Merge(profiles)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to merge profiles: %w", err)
	}

	valueType := "unknown"
	if merged.PeriodType != nil && merged.PeriodType.Type != "" {
		valueType = merged.PeriodType.Type
	}

	result := &MergeResult{
		InputCount:   len(profiles),
		TotalSamples: len(merged.Sample),
		ValueType:    valueType,
	}

	return merged, result, nil
}

// ValidateTypeConsistency checks that all profiles have the same PeriodType.
func ValidateTypeConsistency(profiles []*profile.Profile) error {
	var knownType string

	for _, p := range profiles {
		if p.PeriodType == nil || p.PeriodType.Type == "" {
			continue
		}
		if knownType == "" {
			knownType = p.PeriodType.Type
			continue
		}
		if p.PeriodType.Type != knownType {
			return fmt.Errorf("cannot merge: mixed profile types (%s vs %s)", knownType, p.PeriodType.Type)
		}
	}

	return nil
}
