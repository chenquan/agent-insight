package profile

import (
	"fmt"
	"regexp"
	"sort"

)

// DiffResult represents the result of comparing two profiles
type DiffResult struct {
	// Profile information
	BaseProfile   string
	TargetProfile string

	// Value type used for comparison
	ValueType string

	// Comparison results
	Regressions []FunctionDelta
	Improvements []FunctionDelta
	Unchanged   []FunctionDelta
	NewFunctions []FunctionDelta
	DeletedFunctions []FunctionDelta

	// Overall statistics
	OverallDiff OverallDiff

	// Configuration used
	DiffConfig DiffConfig
}

// FunctionDelta represents the change in a function between two profiles
type FunctionDelta struct {
	// Symbol information
	Function    *string
	File        *string
	LocationID  *uint64
	Address     *string
	Module      *string

	// Values
	BaseFlat    int64
	TargetFlat  int64
	BaseCum     int64
	TargetCum   int64

	// Deltas
	FlatDelta    int64
	FlatDeltaPercent float64
	CumDelta     int64
	CumDeltaPercent  float64

	// Status
	IsNew     bool
	IsDeleted bool
}

// OverallDiff represents overall changes between profiles
type OverallDiff struct {
	BaseTotal     int64
	TargetTotal   int64
	TotalDelta    int64
	TotalPercent  float64
	BaseSamples   int
	TargetSamples int
}

// DiffConfig contains configuration for diff command
type DiffConfig struct {
	MinDelta      float64 // Minimum percentage change to include (0 = all)
	TopN          int     // Limit to top N in each category (0 = unlimited)
	FocusPattern  string  // Regex pattern to focus on
	IgnorePattern string  // Regex pattern to ignore
	ValueType     *ValueTypeConfig
}

// Diff compares two profiles and identifies performance changes
func Diff(base, target *Profile, config DiffConfig) (*DiffResult, error) {
	if base == nil || target == nil {
		return nil, fmt.Errorf("both profiles must be non-nil")
	}

	if err := ValidateTypeConsistency([]*Profile{base, target}); err != nil {
		return nil, err
	}

	// Set default value type if not specified
	if config.ValueType == nil {
		metadata := extractMetadata(base)
		config.ValueType = selectDefaultValueType(base, metadata.Type)
	}

	result := &DiffResult{
		DiffConfig: config,
		ValueType:  config.ValueType.Name + "/" + config.ValueType.Unit,
	}

	// Compile patterns if provided
	var focusRegex, ignoreRegex *regexp.Regexp
	var err error

	if config.FocusPattern != "" {
		focusRegex, err = regexp.Compile(config.FocusPattern)
		if err != nil {
			return nil, fmt.Errorf("invalid focus pattern: %w", err)
		}
	}

	if config.IgnorePattern != "" {
		ignoreRegex, err = regexp.Compile(config.IgnorePattern)
		if err != nil {
			return nil, fmt.Errorf("invalid ignore pattern: %w", err)
		}
	}

	// Calculate function deltas
	deltas := calculateDeltas(base, target, config.ValueType.Index)

	// Apply filters
	filteredDeltas := applyDiffFilters(deltas, focusRegex, ignoreRegex)

	// Categorize deltas
	for _, delta := range filteredDeltas {
		if delta.IsNew {
			result.NewFunctions = append(result.NewFunctions, delta)
		} else if delta.IsDeleted {
			result.DeletedFunctions = append(result.DeletedFunctions, delta)
		} else if delta.FlatDeltaPercent < 0 {
			result.Improvements = append(result.Improvements, delta)
		} else if delta.FlatDeltaPercent > 0 {
			result.Regressions = append(result.Regressions, delta)
		} else {
			result.Unchanged = append(result.Unchanged, delta)
		}
	}

	// Apply minimum delta filter
	if config.MinDelta > 0 {
		result.Regressions = filterByMinDelta(result.Regressions, config.MinDelta)
		result.Improvements = filterByMinDelta(result.Improvements, config.MinDelta)
	}

	// Sort by impact (percentage change)
	sort.Slice(result.Regressions, func(i, j int) bool {
		return result.Regressions[i].FlatDeltaPercent > result.Regressions[j].FlatDeltaPercent
	})

	sort.Slice(result.Improvements, func(i, j int) bool {
		// For improvements, larger negative % = better improvement
		return result.Improvements[i].FlatDeltaPercent < result.Improvements[j].FlatDeltaPercent
	})

	// Apply top N limits
	if config.TopN > 0 {
		if len(result.Regressions) > config.TopN {
			result.Regressions = result.Regressions[:config.TopN]
		}
		if len(result.Improvements) > config.TopN {
			result.Improvements = result.Improvements[:config.TopN]
		}
	}

	// Calculate overall statistics
	result.OverallDiff = calculateOverallDiff(base, target, config.ValueType.Index)

	return result, nil
}

// calculateDeltas calculates deltas for all functions between two profiles
func calculateDeltas(base, target *Profile, valueIndex int) []FunctionDelta {
	// Build maps of location ID -> function info and values
	baseValues := buildLocationValueMap(base, valueIndex)
	targetValues := buildLocationValueMap(target, valueIndex)

	// Track all locations
	allLocs := make(map[uint64]bool)
	for locID := range baseValues {
		allLocs[locID] = true
	}
	for locID := range targetValues {
		allLocs[locID] = true
	}

	// Calculate deltas
	var deltas []FunctionDelta
	for locID := range allLocs {
		baseData, baseExists := baseValues[locID]
		targetData, targetExists := targetValues[locID]

		delta := FunctionDelta{
			LocationID: &locID,
		}

		// Handle new/deleted functions
		if !baseExists && targetExists {
			delta.IsNew = true
			delta.TargetFlat = targetData.flat
			delta.TargetCum = targetData.cum
			delta.BaseFlat = 0
			delta.BaseCum = 0
		} else if baseExists && !targetExists {
			delta.IsDeleted = true
			delta.BaseFlat = baseData.flat
			delta.BaseCum = baseData.cum
			delta.TargetFlat = 0
			delta.TargetCum = 0
		} else {
			// Both exist
			delta.BaseFlat = baseData.flat
			delta.BaseCum = baseData.cum
			delta.TargetFlat = targetData.flat
			delta.TargetCum = targetData.cum
		}

		// Calculate deltas
		delta.FlatDelta = delta.TargetFlat - delta.BaseFlat
		delta.CumDelta = delta.TargetCum - delta.BaseCum

		if delta.BaseFlat != 0 {
			delta.FlatDeltaPercent = float64(delta.FlatDelta) / float64(delta.BaseFlat) * 100
		}
		if delta.BaseCum != 0 {
			delta.CumDeltaPercent = float64(delta.CumDelta) / float64(delta.BaseCum) * 100
		}

		// Extract symbol info from base or target
		if baseExists {
			extractDeltaSymbolInfo(base, locID, &delta)
		} else if targetExists {
			extractDeltaSymbolInfo(target, locID, &delta)
		}

		deltas = append(deltas, delta)
	}

	return deltas
}

// locationValue holds value data for a location
type locationValue struct {
	flat int64
	cum  int64
}

// buildLocationValueMap builds a map of location ID to value data
func buildLocationValueMap(p *Profile, valueIndex int) map[uint64]locationValue {
	valueMap := make(map[uint64]locationValue)

	for _, sample := range p.Sample {
		if valueIndex >= len(sample.Value) {
			continue
		}
		value := sample.Value[valueIndex]

		if len(sample.Location) == 0 {
			continue
		}

		// Leaf location gets flat value
		leaf := sample.Location[0]
		data := valueMap[leaf.ID]
		data.flat += value

		// All locations get cumulative value
		for _, loc := range sample.Location {
			locData := valueMap[loc.ID]
			locData.cum += value
			valueMap[loc.ID] = locData
		}

		valueMap[leaf.ID] = data
	}

	return valueMap
}

// extractDeltaSymbolInfo extracts symbol information for a delta
func extractDeltaSymbolInfo(p *Profile, locID uint64, delta *FunctionDelta) {
	loc := findLocationByID(p, locID)
	if loc == nil {
		return
	}

	if len(loc.Line) > 0 && loc.Line[0].Function != nil {
		delta.Function = &loc.Line[0].Function.Name
		if loc.Line[0].Function.Filename != "" {
			file := fmt.Sprintf("%s:%d", loc.Line[0].Function.Filename, loc.Line[0].Line)
			delta.File = &file
		}
	}

	addr := fmt.Sprintf("0x%x", loc.Address)
	delta.Address = &addr

	if loc.Mapping != nil && loc.Mapping.File != "" {
		mod := normalizeMappingFile(loc.Mapping.File)
		delta.Module = &mod
	}
}

// applyDiffFilters applies focus/ignore patterns to deltas
func applyDiffFilters(deltas []FunctionDelta, focusRegex, ignoreRegex *regexp.Regexp) []FunctionDelta {
	var filtered []FunctionDelta

	for _, delta := range deltas {
		name := ""
		if delta.Function != nil {
			name = *delta.Function
		} else if delta.Address != nil {
			name = *delta.Address
		}

		// Apply ignore filter
		if ignoreRegex != nil && ignoreRegex.MatchString(name) {
			continue
		}

		// Apply focus filter
		if focusRegex != nil && !focusRegex.MatchString(name) {
			continue
		}

		filtered = append(filtered, delta)
	}

	return filtered
}

// filterByMinDelta filters deltas by minimum percentage change
func filterByMinDelta(deltas []FunctionDelta, minDelta float64) []FunctionDelta {
	var filtered []FunctionDelta

	for _, delta := range deltas {
		if abs(delta.FlatDeltaPercent) >= minDelta {
			filtered = append(filtered, delta)
		}
	}

	return filtered
}

// calculateOverallDiff calculates overall statistics
func calculateOverallDiff(base, target *Profile, valueIndex int) OverallDiff {
	baseTotal := int64(0)
	targetTotal := int64(0)

	for _, sample := range base.Sample {
		if valueIndex < len(sample.Value) {
			baseTotal += sample.Value[valueIndex]
		}
	}

	for _, sample := range target.Sample {
		if valueIndex < len(sample.Value) {
			targetTotal += sample.Value[valueIndex]
		}
	}

	totalDelta := targetTotal - baseTotal
	var totalPercent float64
	if baseTotal != 0 {
		totalPercent = float64(totalDelta) / float64(baseTotal) * 100
	}

	return OverallDiff{
		BaseTotal:     baseTotal,
		TargetTotal:   targetTotal,
		TotalDelta:    totalDelta,
		TotalPercent:  totalPercent,
		BaseSamples:   len(base.Sample),
		TargetSamples: len(target.Sample),
	}
}

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}
