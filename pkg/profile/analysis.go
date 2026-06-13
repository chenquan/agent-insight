package profile

import (
	"fmt"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/google/pprof/profile"
)

// Analysis represents the complete analysis results of a profile
type Analysis struct {
	// Profile metadata
	Metadata Metadata

	// Analysis results
	Hotspots  []Hotspot
	SampleCount int
	TotalValue  int64

	// Call stack paths (aggregated by path)
	CallPaths []CallPath

	// Configuration used for analysis
	Config AnalysisConfig
}

// Metadata contains profile metadata information
type Metadata struct {
	Type         string    // cpu, heap, goroutine, etc.
	Duration     time.Duration
	SampleTypes  []string  // Available value types
	Functions    int       // Number of functions in profile
	Locations    int       // Number of locations in profile
}

// Hotspot represents a function with performance data
type Hotspot struct {
	// Symbol information (may be null if unavailable)
	Function    *string  // Function name, e.g., "runtime.mallocgc"
	File        *string  // File name, e.g., "runtime/malloc.go:1020"

	// Fallback information when symbols are missing
	LocationID  *uint64  // Location ID from profile
	Address     *string  // Memory address, e.g., "0x430bac"
	Module      *string  // Module/binary name from Mapping

	// Performance metrics
	FlatValue   int64     // Direct value (samples, bytes, etc.)
	FlatPercent float64   // Percentage of total
	CumValue    int64     // Cumulative value
	CumPercent  float64   // Percentage of total
}

// CallPath represents an aggregated call stack path
type CallPath struct {
	// Stack trace as semicolon-separated function names (root to leaf)
	Path string

	// Sample count for this path
	Count int64

	// Percentage of total samples
	Percent float64
}

// ValueTypeConfig describes how to handle a specific value type
type ValueTypeConfig struct {
	Name     string
	Unit     string
	Index    int  // Index in Sample.Value array
	IsDefault bool // Whether this is the default for this profile type
}

// AnalysisConfig contains configuration for profile analysis
type AnalysisConfig struct {
	TopN          int           // Number of top hotspots to return
	SortByCum     bool          // Sort by cumulative instead of flat
	FocusPattern  string        // Regex pattern to focus on
	IgnorePattern string        // Regex pattern to ignore
	ValueType     *ValueTypeConfig // Which value type to analyze
	CallDepth     int           // Maximum depth of call stack paths to extract
}

// NewAnalysis creates a new analysis from a pprof profile
func NewAnalysis(p *profile.Profile, config AnalysisConfig) (*Analysis, error) {
	if p == nil {
		return nil, fmt.Errorf("profile is nil")
	}

	// Set defaults
	if config.TopN <= 0 {
		config.TopN = 15
	}

	// Detect profile type and determine default value type
	metadata := extractMetadata(p)
	if config.ValueType == nil {
		config.ValueType = selectDefaultValueType(p, metadata.Type)
	}

	analysis := &Analysis{
		Metadata: metadata,
		Config:   config,
	}

	// Calculate hotspots
	hotspots, err := analysis.calculateHotspots(p)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate hotspots: %w", err)
	}

	analysis.Hotspots = hotspots
	analysis.SampleCount = len(p.Sample)

	// Calculate total value
	for _, sample := range p.Sample {
		if config.ValueType.Index < len(sample.Value) {
			analysis.TotalValue += sample.Value[config.ValueType.Index]
		}
	}

	// Calculate call paths
	if config.CallDepth > 0 {
		analysis.CallPaths = analysis.calculateCallPaths(p)
	}

	return analysis, nil
}

// extractMetadata extracts metadata from the profile
func extractMetadata(p *profile.Profile) Metadata {
	metadata := Metadata{
		Functions: len(p.Function),
		Locations: len(p.Location),
	}

	// Extract duration
	if p.DurationNanos > 0 {
		metadata.Duration = time.Duration(p.DurationNanos) * time.Nanosecond
	}

	// Extract sample types
	if len(p.SampleType) > 0 {
		metadata.SampleTypes = make([]string, len(p.SampleType))
		for i, st := range p.SampleType {
			metadata.SampleTypes[i] = st.Type + "/" + st.Unit
		}
	}

	// Detect profile type
	if p.PeriodType != nil && p.PeriodType.Type != "" {
		metadata.Type = p.PeriodType.Type
	} else {
		metadata.Type = inferProfileType(p)
	}

	return metadata
}

// inferProfileType infers the profile type from SampleType when PeriodType is missing or empty.
func inferProfileType(p *profile.Profile) string {
	for _, st := range p.SampleType {
		switch st.Type {
		case "inuse_space", "alloc_space", "inuse_objects", "alloc_objects":
			return "heap"
		case "cpu":
			return "cpu"
		case "goroutine":
			return "goroutine"
		case "contentions":
			return "contentions"
		case "thread", "threadcreate":
			return "thread"
		}
	}
	return "unknown"
}

// normalizeMappingFile reduces a mapping file path to its basename.
func normalizeMappingFile(file string) string {
	if file == "" {
		return ""
	}
	return filepath.Base(file)
}

// selectDefaultValueType intelligently selects the default value type based on profile type
func selectDefaultValueType(p *profile.Profile, profileType string) *ValueTypeConfig {
	if len(p.SampleType) == 0 {
		return nil
	}

	// Default to first value type if no special handling
	defaultIndex := 0

	// Smart defaults based on profile type
	switch profileType {
	case "cpu":
		// For CPU, prefer cpu/nanoseconds
		for i, st := range p.SampleType {
			if st.Type == "cpu" || st.Type == "samples" {
				defaultIndex = i
				break
			}
		}
	case "space", "heap":
		// For heap, prefer inuse_space/inuse_bytes if available
		for i, st := range p.SampleType {
			if st.Type == "inuse_space" || (st.Type == "inuse" && st.Unit == "bytes") {
				defaultIndex = i
				break
			}
		}
		if defaultIndex == 0 {
			for i, st := range p.SampleType {
				if st.Type == "inuse_objects" {
					defaultIndex = i
					break
				}
			}
		}
	case "threadcreate", "goroutine":
		// For goroutine/thread profiles, prefer count
		for i, st := range p.SampleType {
			if st.Type == "count" || st.Type == "threads" || st.Type == "goroutine" {
				defaultIndex = i
				break
			}
		}
	default:
		// For unknown types, use the first value type
		defaultIndex = 0
	}

	return &ValueTypeConfig{
		Name:      p.SampleType[defaultIndex].Type,
		Unit:      p.SampleType[defaultIndex].Unit,
		Index:     defaultIndex,
		IsDefault: true,
	}
}

// calculateHotspots calculates flat and cumulative values for each function
func (a *Analysis) calculateHotspots(p *profile.Profile) ([]Hotspot, error) {
	if a.Config.ValueType == nil {
		return nil, fmt.Errorf("value type not configured")
	}

	valueIndex := a.Config.ValueType.Index

	// Compile patterns if provided
	var focusRegex, ignoreRegex *regexp.Regexp
	var err error

	if a.Config.FocusPattern != "" {
		focusRegex, err = regexp.Compile(a.Config.FocusPattern)
		if err != nil {
			return nil, fmt.Errorf("invalid focus pattern: %w", err)
		}
	}

	if a.Config.IgnorePattern != "" {
		ignoreRegex, err = regexp.Compile(a.Config.IgnorePattern)
		if err != nil {
			return nil, fmt.Errorf("invalid ignore pattern: %w", err)
		}
	}

	// Maps to accumulate flat and cumulative values
	flatMap := make(map[uint64]int64)  // locationID -> flat value
	cumMap := make(map[uint64]int64)   // locationID -> cumulative value
	nameMap := make(map[uint64]string) // locationID -> display name

	// Process each sample
	for _, sample := range p.Sample {
		if valueIndex >= len(sample.Value) {
			continue
		}
		value := sample.Value[valueIndex]

		if len(sample.Location) == 0 {
			continue
		}

		// Check if this sample should be included based on focus/ignore patterns
		if !shouldIncludeSample(sample, focusRegex, ignoreRegex, p) {
			continue
		}

		// Leaf location gets flat value
		leaf := sample.Location[0]
		flatMap[leaf.ID] += value

		// All locations in the stack get cumulative value
		for _, loc := range sample.Location {
			cumMap[loc.ID] += value

			// Store display name (function name or location ID)
			if _, exists := nameMap[loc.ID]; !exists {
				nameMap[loc.ID] = getLocationDisplayName(loc)
			}
		}
	}

	// Convert to hotspots
	var hotspots []Hotspot
	totalValue := int64(0)
	for _, v := range flatMap {
		totalValue += v
	}

	for locID, flat := range flatMap {
		name := nameMap[locID]

		// Apply ignore filter at hotspot level
		if ignoreRegex != nil && ignoreRegex.MatchString(name) {
			continue
		}

		// Apply focus filter at hotspot level
		if focusRegex != nil && !focusRegex.MatchString(name) {
			continue
		}

		hotspot := Hotspot{
			FlatValue:   flat,
			FlatPercent: float64(flat) / float64(totalValue) * 100,
			CumValue:    cumMap[locID],
			CumPercent:  float64(cumMap[locID]) / float64(totalValue) * 100,
		}

		// Extract symbol information from location
		if loc := findLocationByID(p, locID); loc != nil {
			extractSymbolInfo(loc, &hotspot)
		}

		// Set fallback if symbol info is missing
		if hotspot.Function == nil {
			hotspot.LocationID = &locID
			if loc := findLocationByID(p, locID); loc != nil {
				addr := fmt.Sprintf("0x%x", loc.Address)
				hotspot.Address = &addr

				// Extract module info from mapping
				if loc.Mapping != nil && loc.Mapping.File != "" {
					mod := normalizeMappingFile(loc.Mapping.File)
					hotspot.Module = &mod
				}
			}
		}

		hotspots = append(hotspots, hotspot)
	}

	// Sort hotspots
	sort.Slice(hotspots, func(i, j int) bool {
		if a.Config.SortByCum {
			return hotspots[i].CumValue > hotspots[j].CumValue
		}
		return hotspots[i].FlatValue > hotspots[j].FlatValue
	})

	// Limit to Top N
	if a.Config.TopN > 0 && len(hotspots) > a.Config.TopN {
		hotspots = hotspots[:a.Config.TopN]
	}

	return hotspots, nil
}

// getLocationDisplayName generates a display name for a location
func getLocationDisplayName(loc *profile.Location) string {
	if len(loc.Line) > 0 && loc.Line[0].Function != nil {
		return loc.Line[0].Function.Name
	}
	return fmt.Sprintf("loc_%d", loc.ID)
}

// findLocationByID finds a location by its ID in the profile
func findLocationByID(p *profile.Profile, id uint64) *profile.Location {
	for _, loc := range p.Location {
		if loc.ID == id {
			return loc
		}
	}
	return nil
}

// extractSymbolInfo extracts symbol information from a location
func extractSymbolInfo(loc *profile.Location, hotspot *Hotspot) {
	if len(loc.Line) == 0 {
		return
	}

	// Get the first line (most specific function)
	line := loc.Line[0]
	if line.Function == nil {
		return
	}

	// Set function name
	hotspot.Function = &line.Function.Name

	// Set file information if available
	if line.Function.Filename != "" {
		fileStr := fmt.Sprintf("%s:%d", line.Function.Filename, line.Line)
		hotspot.File = &fileStr
	}
}

// shouldIncludeSample checks if a sample should be included based on focus/ignore patterns
func shouldIncludeSample(sample *profile.Sample, focusRegex, ignoreRegex *regexp.Regexp, p *profile.Profile) bool {
	// Build list of function names in this sample's stack
	var functions []string
	for _, loc := range sample.Location {
		name := getLocationDisplayName(loc)
		functions = append(functions, name)
	}

	// If no functions in stack, skip
	if len(functions) == 0 {
		return false
	}

	// Check ignore pattern - if any function matches, exclude the sample
	if ignoreRegex != nil {
		for _, name := range functions {
			if ignoreRegex.MatchString(name) {
				return false
			}
		}
	}

	// Check focus pattern - if set, only include if at least one function matches
	if focusRegex != nil {
		hasMatch := false
		for _, name := range functions {
			if focusRegex.MatchString(name) {
				hasMatch = true
				break
			}
		}
		return hasMatch
	}

	// No filters, include the sample
	return true
}

// calculateCallPaths aggregates call stack paths from samples
func (a *Analysis) calculateCallPaths(p *profile.Profile) []CallPath {
	if a.Config.ValueType == nil {
		return nil
	}

	valueIndex := a.Config.ValueType.Index
	depth := a.Config.CallDepth

	// Aggregate paths by their string representation
	pathMap := make(map[string]int64)

	for _, sample := range p.Sample {
		if valueIndex >= len(sample.Value) || len(sample.Location) == 0 {
			continue
		}

		value := sample.Value[valueIndex]

		// Build path string (root to leaf, limited by depth)
		path := a.buildPathString(sample.Location, depth)
		if path == "" {
			continue
		}

		pathMap[path] += value
	}

	// Convert to slice and calculate percentages
	paths := make([]CallPath, 0, len(pathMap))
	for path, count := range pathMap {
		percent := float64(count) / float64(a.TotalValue) * 100
		paths = append(paths, CallPath{
			Path:     path,
			Count:    count,
			Percent:  percent,
		})
	}

	// Sort by count (descending)
	sort.Slice(paths, func(i, j int) bool {
		return paths[i].Count > paths[j].Count
	})

	return paths
}

// buildPathString builds a semicolon-separated path string from locations
func (a *Analysis) buildPathString(locations []*profile.Location, maxDepth int) string {
	if len(locations) == 0 {
		return ""
	}

	var parts []string

	// Traverse from leaf (index 0) to root
	// Limit depth as specified
	for i := 0; i < len(locations) && i < maxDepth; i++ {
		loc := locations[i]
		name := getFunctionNameFromLocation(loc)
		if name == "" {
			name = fmt.Sprintf("0x%x", loc.Address)
		}
		parts = append(parts, name)
	}

	if len(parts) == 0 {
		return ""
	}

	// Reverse to get root-first order (standard for call paths)
	for i, j := 0, len(parts)-1; i < j; i, j = i+1, j-1 {
		parts[i], parts[j] = parts[j], parts[i]
	}

	return strings.Join(parts, ";")
}

// getFunctionNameFromLocation extracts function name from a location
func getFunctionNameFromLocation(loc *profile.Location) string {
	if len(loc.Line) > 0 && loc.Line[0].Function != nil {
		return loc.Line[0].Function.Name
	}
	return ""
}
