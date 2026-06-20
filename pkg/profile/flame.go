package profile

import (
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/google/pprof/profile"
)

// FlameResult represents the result of flame graph folded stack generation
type FlameResult struct {
	// Folded stacks
	Stacks []FoldedStack

	// Statistics
	TotalStacks int
	FilteredStacks int
	UniqueStacks int

	// Configuration used
	FlameConfig FlameConfig
}

// FoldedStack represents a single folded stack entry
type FoldedStack struct {
	// Stack trace as semicolon-separated function names
	Stack string

	// Sample count/value for this stack
	Count int64

	// Value type used for this count
	ValueType string
}

// FlameConfig contains configuration for flame command
type FlameConfig struct {
	FocusPattern   string // Regex pattern to focus on
	IgnorePattern   string // Regex pattern to ignore
	Depth          int    // Maximum depth (0 = unlimited)
	ValueType      *ValueTypeConfig
	TopN           int    // Limit to top N stacks (0 = unlimited)
}

// Flame generates folded stack format from a profile
func Flame(p *Profile, config FlameConfig) (*FlameResult, error) {
	if p == nil {
		return nil, fmt.Errorf("profile is nil")
	}

	// Set defaults
	if config.Depth == 0 {
		config.Depth = 0 // Unlimited by default
	}

	// Set default value type if not specified
	if config.ValueType == nil {
		metadata := extractMetadata(p)
		config.ValueType = selectDefaultValueType(p, metadata.Type)
	}

	result := &FlameResult{
		FlameConfig: config,
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

	// Aggregate stacks
	stackMap := make(map[string]int64)
	totalStacks := 0
	filteredStacks := 0

	valueIndex := config.ValueType.Index

	for _, sample := range p.Sample {
		if valueIndex >= len(sample.Value) {
			continue
		}
		value := sample.Value[valueIndex]
		totalStacks++

		// Build stack trace
		stack := buildStackString(sample.Location, focusRegex, ignoreRegex)

		// Skip if empty (filtered out)
		if stack == "" {
			filteredStacks++
			continue
		}

		// Apply depth limit if set
		if config.Depth > 0 {
			stack = applyDepthLimit(stack, config.Depth)
		}

		// Aggregate
		stackMap[stack] += value
	}

	result.TotalStacks = totalStacks
	result.FilteredStacks = filteredStacks
	result.UniqueStacks = len(stackMap)

	// Convert to sorted slice
	result.Stacks = make([]FoldedStack, 0, len(stackMap))
	for stack, count := range stackMap {
		result.Stacks = append(result.Stacks, FoldedStack{
			Stack:     stack,
			Count:     count,
			ValueType: config.ValueType.Name + "/" + config.ValueType.Unit,
		})
	}

	// Sort by count (descending)
	sort.Slice(result.Stacks, func(i, j int) bool {
		return result.Stacks[i].Count > result.Stacks[j].Count
	})

	// Apply top N limit if set
	if config.TopN > 0 && len(result.Stacks) > config.TopN {
		result.Stacks = result.Stacks[:config.TopN]
	}

	return result, nil
}

// buildStackString builds a semicolon-separated stack string from locations
func buildStackString(locations []*profile.Location, focusRegex, ignoreRegex *regexp.Regexp) string {
	if len(locations) == 0 {
		return ""
	}

	var parts []string

	// Process from leaf (index 0) to root
	for i := 0; i < len(locations); i++ {
		loc := locations[i]
		name := getFunctionName(loc)

		// Apply ignore filter
		if ignoreRegex != nil && ignoreRegex.MatchString(name) {
			return "" // Skip this entire stack
		}

		// Apply focus filter
		if focusRegex != nil && !focusRegex.MatchString(name) {
			// Only include stacks that match the focus pattern
			// Check if any frame in the stack matches
			hasFocus := false
			for _, l := range locations {
				if focusRegex.MatchString(getFunctionName(l)) {
					hasFocus = true
					break
				}
			}
			if !hasFocus {
				return "" // Skip stacks that don't have any focus match
			}
		}

		parts = append(parts, name)
	}

	if len(parts) == 0 {
		return ""
	}

	// Reverse to put root first (standard flame graph format)
	reverseParts(parts)
	return strings.Join(parts, ";")
}

// getFunctionName gets the function name from a location
func getFunctionName(loc *profile.Location) string {
	if len(loc.Line) > 0 && loc.Line[0].Function != nil {
		name := loc.Line[0].Function.Name
		if name == "" {
			return fmt.Sprintf("0x%x", loc.Address)
		}

		// Check if this is an inlined function
		if len(loc.Line) > 1 {
			return name + " [inlined]"
		}

		return name
	}

	// Fallback to address
	return fmt.Sprintf("0x%x", loc.Address)
}

// reverseParts reverses a slice of strings in place
func reverseParts(parts []string) {
	for i, j := 0, len(parts)-1; i < j; i, j = i+1, j-1 {
		parts[i], parts[j] = parts[j], parts[i]
	}
}

// applyDepthLimit truncates a stack string to the specified depth
func applyDepthLimit(stack string, depth int) string {
	if depth <= 0 {
		return stack
	}

	parts := strings.Split(stack, ";")
	if len(parts) <= depth {
		return stack
	}

	// Keep the bottom N frames (deepest functions)
	// In flame graph format, leaf functions are at the end
	start := len(parts) - depth
	if start < 0 {
		start = 0
	}

	return strings.Join(parts[start:], ";")
}
