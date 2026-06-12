package profile

import (
	"fmt"
	"regexp"
	"sort"

	"github.com/google/pprof/profile"
)

// ListResult represents the result of a function query
type ListResult struct {
	// Query information
	QueryPattern string
	MatchedFunctions []FunctionInfo

	// Options used
	ListConfig ListConfig
}

// FunctionInfo represents information about a single function
type FunctionInfo struct {
	// Symbol information
	Function    *string
	File        *string
	LocationID  *uint64
	Address     *string
	Module      *string

	// Performance data
	FlatValue   int64
	FlatPercent float64
	CumValue    int64
	CumPercent  float64

	// Call relationships
	Callers []CallRelationship
	Callees []CallRelationship

	// Special properties
	IsLeaf     bool // No callees
	IsRecursive bool // Calls itself
}

// CallRelationship represents a call relationship between functions
type CallRelationship struct {
	// The other function in this relationship
	Function    *string
	File        *string
	LocationID  *uint64
	Address     *string
	Module      *string

	// Performance contribution
	FlatValue   int64
	FlatPercent float64
	// Percentage of the parent/child function
	RelationPercent float64
}

// ListConfig contains configuration for list command
type ListConfig struct {
	Pattern      string // Regex pattern to match functions
	Depth        int    // Maximum depth to traverse (0 = unlimited)
	CallersOnly  bool   // Show only callers
	CalleesOnly  bool   // Show only callees
	ExcludePattern string // Pattern to exclude
	ValueType    *ValueTypeConfig
}

// List performs a function query analysis on the profile
func List(p *profile.Profile, config ListConfig) (*ListResult, error) {
	if p == nil {
		return nil, fmt.Errorf("profile is nil")
	}

	// Set defaults
	if config.Depth == 0 {
		config.Depth = 5
	}

	// Set default value type if not specified
	if config.ValueType == nil {
		metadata := extractMetadata(p)
		config.ValueType = selectDefaultValueType(p, metadata.Type)
	}

	result := &ListResult{
		QueryPattern: config.Pattern,
		ListConfig:   config,
	}

	// Compile patterns
	pattern, err := regexp.Compile(config.Pattern)
	if err != nil {
		return nil, fmt.Errorf("invalid pattern: %w", err)
	}

	var excludeRegex *regexp.Regexp
	if config.ExcludePattern != "" {
		excludeRegex, err = regexp.Compile(config.ExcludePattern)
		if err != nil {
			return nil, fmt.Errorf("invalid exclude pattern: %w", err)
		}
	}

	// Find matching functions and build their call relationships
	matched := findMatchingFunctions(p, pattern, excludeRegex, config)

	// Calculate total value for percentages
	totalValue := calculateTotalValue(p, config.ValueType.Index)

	// Fill in function info and calculate percentages
	for i := range matched {
		calculatePercentages(&matched[i], totalValue)
	}

	sort.Slice(matched, func(i, j int) bool {
		return matched[i].FlatValue > matched[j].FlatValue
	})

	result.MatchedFunctions = matched
	return result, nil
}

// findMatchingFunctions finds functions matching the pattern and builds their relationships
func findMatchingFunctions(p *profile.Profile, pattern, exclude *regexp.Regexp, config ListConfig) []FunctionInfo {
	valueIndex := config.ValueType.Index

	// Build location -> function name mapping
	locationNames := make(map[uint64]string)
	for _, loc := range p.Location {
		if len(loc.Line) > 0 && loc.Line[0].Function != nil {
			locationNames[loc.ID] = loc.Line[0].Function.Name
		}
	}

	// Find all locations that match the pattern
	matchedLocs := make(map[uint64]bool)
	for _, loc := range p.Location {
		name := locationNames[loc.ID]
		if name == "" {
			name = fmt.Sprintf("0x%x", loc.Address)
		}

		if pattern.MatchString(name) {
			if exclude != nil && exclude.MatchString(name) {
				continue
			}
			matchedLocs[loc.ID] = true
		}
	}

	// Build function info for matched locations
	funcInfos := make(map[uint64]FunctionInfo)
	callersMap := make(map[uint64][]CallRelationship)
	calleesMap := make(map[uint64][]CallRelationship)

	// Process each sample to build relationships
	for _, sample := range p.Sample {
		if valueIndex >= len(sample.Value) {
			continue
		}
		value := sample.Value[valueIndex]

		if len(sample.Location) == 0 {
			continue
		}

		// Build call stack from leaf (index 0) to root
		for i, loc := range sample.Location {
			locID := loc.ID

			// Only process if this location matches our query
			if !matchedLocs[locID] {
				continue
			}

			// Initialize function info if not exists
			info := funcInfos[locID]
			if info.Function == nil && info.LocationID == nil {
				info = FunctionInfo{}
				extractSymbolInfoForLocation(loc, &info)
				info.LocationID = &locID
				if loc.Mapping != nil && loc.Mapping.File != "" {
					info.Module = &loc.Mapping.File
				}
				funcInfos[locID] = info
			}

			// Get current info for updates
			info = funcInfos[locID]

			// Add flat value for leaf locations
			if i == 0 {
				info.FlatValue += value
			}

			// Add cumulative value
			info.CumValue += value

			// Add caller relationship (not leaf)
			if i < len(sample.Location)-1 {
				callerLoc := sample.Location[i+1]
				callerRel := buildCallRelationship(callerLoc, value, p)
				callersMap[locID] = append(callersMap[locID], callerRel)

				// Check for recursion
				if callerLoc.ID == locID {
					info.IsRecursive = true
				}
			}

			// Store updated info
			funcInfos[locID] = info

			// Add callee relationship (if not leaf)
			if i > 0 && !config.CallersOnly {
				calleeLoc := sample.Location[i-1]
				calleeRel := buildCallRelationship(calleeLoc, value, p)
				calleesMap[locID] = append(calleesMap[locID], calleeRel)
			}
		}
	}

	// Convert to slice and add relationships
	result := make([]FunctionInfo, 0, len(funcInfos))
	for locID, info := range funcInfos {
		// Add callers
		if !config.CalleesOnly {
			info.Callers = callersMap[locID]
		}

		// Add callees
		if !config.CallersOnly {
			info.Callees = calleesMap[locID]
		}

		// Mark as leaf if no callees
		info.IsLeaf = len(info.Callees) == 0

		result = append(result, info)
	}

	return result
}

// buildCallRelationship builds a call relationship from a location
func buildCallRelationship(loc *profile.Location, value int64, p *profile.Profile) CallRelationship {
	rel := CallRelationship{}

	if len(loc.Line) > 0 && loc.Line[0].Function != nil {
		rel.Function = &loc.Line[0].Function.Name
		if loc.Line[0].Function.Filename != "" {
			file := fmt.Sprintf("%s:%d", loc.Line[0].Function.Filename, loc.Line[0].Line)
			rel.File = &file
		}
	} else {
		locID := loc.ID
		rel.LocationID = &locID
		addr := fmt.Sprintf("0x%x", loc.Address)
		rel.Address = &addr
		if loc.Mapping != nil && loc.Mapping.File != "" {
			rel.Module = &loc.Mapping.File
		}
	}

	rel.FlatValue = value
	return rel
}

// extractSymbolInfoForLocation extracts symbol information for a location
func extractSymbolInfoForLocation(loc *profile.Location, info *FunctionInfo) {
	if len(loc.Line) == 0 {
		return
	}

	line := loc.Line[0]
	if line.Function == nil {
		return
	}

	info.Function = &line.Function.Name

	if line.Function.Filename != "" {
		file := fmt.Sprintf("%s:%d", line.Function.Filename, line.Line)
		info.File = &file
	}

	// Set address
	addr := fmt.Sprintf("0x%x", loc.Address)
	info.Address = &addr
}

// calculateTotalValue calculates total value across all samples
func calculateTotalValue(p *profile.Profile, valueIndex int) int64 {
	total := int64(0)
	for _, sample := range p.Sample {
		if valueIndex < len(sample.Value) {
			total += sample.Value[valueIndex]
		}
	}
	return total
}

// calculatePercentages calculates percentages for function info and relationships
func calculatePercentages(info *FunctionInfo, totalValue int64) {
	if totalValue > 0 {
		info.FlatPercent = float64(info.FlatValue) / float64(totalValue) * 100
		info.CumPercent = float64(info.CumValue) / float64(totalValue) * 100
	}

	// Calculate percentages for callers
	for i := range info.Callers {
		if info.FlatValue > 0 {
			info.Callers[i].RelationPercent = float64(info.Callers[i].FlatValue) / float64(info.FlatValue) * 100
		}
		if totalValue > 0 {
			info.Callers[i].FlatPercent = float64(info.Callers[i].FlatValue) / float64(totalValue) * 100
		}
	}

	// Calculate percentages for callees
	for i := range info.Callees {
		if info.FlatValue > 0 {
			info.Callees[i].RelationPercent = float64(info.Callees[i].FlatValue) / float64(info.FlatValue) * 100
		}
		if totalValue > 0 {
			info.Callees[i].FlatPercent = float64(info.Callees[i].FlatValue) / float64(totalValue) * 100
		}
	}
}
