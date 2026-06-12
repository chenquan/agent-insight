package output

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/chenquan/agent-insight/pkg/profile"
)

// Formatter interface for different output formats
type Formatter interface {
	FormatAnalysis(analysis *profile.Analysis) error
}

// ListFormatter interface for list command output
type ListFormatter interface {
	FormatListResult(result *profile.ListResult) error
}

// TextFormatter outputs analysis in human-readable text format
type TextFormatter struct {
	writer io.Writer
}

// NewTextFormatter creates a new text formatter
func NewTextFormatter(w io.Writer) *TextFormatter {
	return &TextFormatter{writer: w}
}

// FormatAnalysis formats and outputs the analysis
func (f *TextFormatter) FormatAnalysis(analysis *profile.Analysis) error {
	// Output header
	fmt.Fprintf(f.writer, "Profile Analysis\n")
	fmt.Fprintf(f.writer, "================\n\n")

	// Output metadata
	fmt.Fprintf(f.writer, "Type: %s\n", analysis.Metadata.Type)
	if analysis.Metadata.Duration > 0 {
		fmt.Fprintf(f.writer, "Duration: %s\n", analysis.Metadata.Duration)
	}
	fmt.Fprintf(f.writer, "Samples: %d\n", analysis.SampleCount)

	if len(analysis.Metadata.SampleTypes) > 0 {
		fmt.Fprintf(f.writer, "Sample Types: %s\n", strings.Join(analysis.Metadata.SampleTypes, ", "))
	}

	fmt.Fprintf(f.writer, "\n")

	// Output hotspots
	fmt.Fprintf(f.writer, "Top %d Hotspots\n", len(analysis.Hotspots))
	fmt.Fprintf(f.writer, "%s\n", strings.Repeat("=", 50))

	for i, hotspot := range analysis.Hotspots {
		fmt.Fprintf(f.writer, "\n%d. ", i+1)

		// Output function name or location ID
		if hotspot.Function != nil {
			fmt.Fprintf(f.writer, "%s", *hotspot.Function)
		} else if hotspot.LocationID != nil {
			if hotspot.Address != nil {
				fmt.Fprintf(f.writer, "%s (%s)", *hotspot.Address, formatLocationID(*hotspot.LocationID))
			} else {
				fmt.Fprintf(f.writer, "Location %s", formatLocationID(*hotspot.LocationID))
			}
		}

		// Output file info
		if hotspot.File != nil {
			fmt.Fprintf(f.writer, " [%s]", *hotspot.File)
		} else if hotspot.Module != nil {
			fmt.Fprintf(f.writer, " [%s]", *hotspot.Module)
		}

		fmt.Fprintf(f.writer, "\n")

		// Output metrics
		fmt.Fprintf(f.writer, "   Flat:      %d (%.2f%%)\n", hotspot.FlatValue, hotspot.FlatPercent)
		fmt.Fprintf(f.writer, "   Cumulative: %d (%.2f%%)\n", hotspot.CumValue, hotspot.CumPercent)
	}

	// Output call paths if available
	if len(analysis.CallPaths) > 0 {
		fmt.Fprintf(f.writer, "\nTop Call Stack Paths\n")
		fmt.Fprintf(f.writer, "%s\n", strings.Repeat("=", 50))
		for i, path := range analysis.CallPaths {
			fmt.Fprintf(f.writer, "\n%d. %s\n", i+1, path.Path)
			fmt.Fprintf(f.writer, "   Count: %d (%.2f%%)\n", path.Count, path.Percent)
		}
	}
	return nil
}

// JSONFormatter outputs analysis in JSON format
type JSONFormatter struct {
	writer io.Writer
}

// NewJSONFormatter creates a new JSON formatter
func NewJSONFormatter(w io.Writer) *JSONFormatter {
	return &JSONFormatter{writer: w}
}

// FormatAnalysis formats and outputs the analysis as JSON
func (f *JSONFormatter) FormatAnalysis(analysis *profile.Analysis) error {
	output := f.convertToJSONFormat(analysis)
	encoder := json.NewEncoder(f.writer)
	encoder.SetIndent("", "  ")
	return encoder.Encode(output)
}

// JSONOutput represents the JSON output structure
type JSONOutput struct {
	Type         string        `json:"type"`
	Duration     string        `json:"duration,omitempty"`
	Samples      int           `json:"samples"`
	SampleTypes  []string      `json:"sample_types,omitempty"`
	Top          []JSONHotspot `json:"top"`
	Summary      string        `json:"summary"`
	AnalyzedType string        `json:"analyzed_type,omitempty"`
}

// JSONHotspot represents a hotspot in JSON format
type JSONHotspot struct {
	Function    *string `json:"function,omitempty"`
	File        *string `json:"file,omitempty"`
	LocationID  *uint64 `json:"location_id,omitempty"`
	Address     *string `json:"address,omitempty"`
	Module      *string `json:"module,omitempty"`
	Flat        int64   `json:"flat"`
	FlatPercent float64 `json:"flat_percent"`
	Cum         int64   `json:"cum"`
	CumPercent  float64 `json:"cum_percent"`
}

func (f *JSONFormatter) convertToJSONFormat(analysis *profile.Analysis) *JSONOutput {
	output := &JSONOutput{
		Type:    analysis.Metadata.Type,
		Samples: analysis.SampleCount,
	}

	// Add duration if available
	if analysis.Metadata.Duration > 0 {
		output.Duration = formatDuration(analysis.Metadata.Duration)
	}

	// Add sample types
	if len(analysis.Metadata.SampleTypes) > 0 {
		output.SampleTypes = analysis.Metadata.SampleTypes
	}

	// Add analyzed value type
	if analysis.Config.ValueType != nil {
		output.AnalyzedType = analysis.Config.ValueType.Name + "/" + analysis.Config.ValueType.Unit
	}

	// Convert hotspots
	output.Top = make([]JSONHotspot, len(analysis.Hotspots))
	for i, hotspot := range analysis.Hotspots {
		output.Top[i] = JSONHotspot{
			Function:    hotspot.Function,
			File:        hotspot.File,
			LocationID:  hotspot.LocationID,
			Address:     hotspot.Address,
			Module:      hotspot.Module,
			Flat:        hotspot.FlatValue,
			FlatPercent: hotspot.FlatPercent,
			Cum:         hotspot.CumValue,
			CumPercent:  hotspot.CumPercent,
		}
	}

	// Generate summary
	output.Summary = generateSummary(analysis)

	return output
}

// MarkdownFormatter outputs analysis in Markdown format
type MarkdownFormatter struct {
	writer io.Writer
}

// NewMarkdownFormatter creates a new Markdown formatter
func NewMarkdownFormatter(w io.Writer) *MarkdownFormatter {
	return &MarkdownFormatter{writer: w}
}

// FormatAnalysis formats and outputs the analysis as Markdown
func (f *MarkdownFormatter) FormatAnalysis(analysis *profile.Analysis) error {
	// Output header
	fmt.Fprintf(f.writer, "# Profile Analysis\n\n")

	// Output metadata in a table
	fmt.Fprintf(f.writer, "## Profile Metadata\n\n")
	fmt.Fprintf(f.writer, "| Property | Value |\n")
	fmt.Fprintf(f.writer, "|----------|-------|\n")
	fmt.Fprintf(f.writer, "| Type | %s |\n", analysis.Metadata.Type)

	if analysis.Metadata.Duration > 0 {
		fmt.Fprintf(f.writer, "| Duration | %s |\n", analysis.Metadata.Duration)
	}

	fmt.Fprintf(f.writer, "| Samples | %d |\n", analysis.SampleCount)

	if len(analysis.Metadata.SampleTypes) > 0 {
		fmt.Fprintf(f.writer, "| Sample Types | %s |\n", strings.Join(analysis.Metadata.SampleTypes, ", "))
	}

	fmt.Fprintf(f.writer, "\n")

	// Output hotspots table
	fmt.Fprintf(f.writer, "## Top %d Hotspots\n\n", len(analysis.Hotspots))
	fmt.Fprintf(f.writer, "| Rank | Function | File | Flat | Flat%% | Cum | Cum%% |\n")
	fmt.Fprintf(f.writer, "|------|----------|------|------|-------|-----|-------|\n")

	for i, hotspot := range analysis.Hotspots {
		function := formatFunctionRef(&hotspot)
		file := formatFileRef(&hotspot)

		fmt.Fprintf(f.writer, "| %d | %s | %s | %d | %.2f%% | %d | %.2f%% |\n",
			i+1, function, file,
			hotspot.FlatValue, hotspot.FlatPercent,
			hotspot.CumValue, hotspot.CumPercent)
	}

	// Output summary
	fmt.Fprintf(f.writer, "\n## Summary\n\n")
	fmt.Fprintf(f.writer, "%s\n", generateSummary(analysis))

	return nil
}

// Helper functions

func formatLocationID(id uint64) string {
	return fmt.Sprintf("#%d", id)
}

func formatDuration(d time.Duration) string {
	return d.String()
}

func formatFunctionRef(hotspot *profile.Hotspot) string {
	if hotspot.Function != nil {
		return fmt.Sprintf("`%s`", *hotspot.Function)
	} else if hotspot.LocationID != nil {
		if hotspot.Address != nil {
			return fmt.Sprintf("%s (%s)", *hotspot.Address, formatLocationID(*hotspot.LocationID))
		}
		return fmt.Sprintf("Location %s", formatLocationID(*hotspot.LocationID))
	}
	return "unknown"
}

func formatFileRef(hotspot *profile.Hotspot) string {
	if hotspot.File != nil {
		return fmt.Sprintf("`%s`", *hotspot.File)
	} else if hotspot.Module != nil {
		return fmt.Sprintf("`%s`", *hotspot.Module)
	}
	return "-"
}

func generateSummary(analysis *profile.Analysis) string {
	var parts []string

	// Add profile type
	parts = append(parts, fmt.Sprintf("Profile type: %s", analysis.Metadata.Type))

	// Add sample count
	parts = append(parts, fmt.Sprintf("Total samples: %d", analysis.SampleCount))

	// Add top function info with more context
	if len(analysis.Hotspots) > 0 {
		top := analysis.Hotspots[0]
		topRef := "unknown location"
		if top.Function != nil {
			topRef = *top.Function
		} else if top.Address != nil {
			topRef = *top.Address
		}

		parts = append(parts, fmt.Sprintf("Top hotspot: %s (%.2f%%)", topRef, top.FlatPercent))

		// Add concentration info if top function dominates
		if top.FlatPercent > 30 {
			parts = append(parts, fmt.Sprintf("Highly concentrated performance bottleneck"))
		}
	}

	// Add symbol availability info
	symbolCount := 0
	for _, h := range analysis.Hotspots {
		if h.Function != nil {
			symbolCount++
		}
	}
	symbolPercent := float64(symbolCount) / float64(len(analysis.Hotspots)) * 100

	if symbolPercent < 50 {
		parts = append(parts, fmt.Sprintf("Limited symbol information available (%.0f%% of top functions)", symbolPercent))
	}

	return strings.Join(parts, ". ")
}

// ListTextFormatter formats list results in text format
type ListTextFormatter struct {
	writer io.Writer
}

// NewListTextFormatter creates a new list text formatter
func NewListTextFormatter(w io.Writer) *ListTextFormatter {
	return &ListTextFormatter{writer: w}
}

// FormatListResult formats and outputs the list result
func (f *ListTextFormatter) FormatListResult(result *profile.ListResult) error {
	fmt.Fprintf(f.writer, "Function Query Results\n")
	fmt.Fprintf(f.writer, "Pattern: %s\n", result.QueryPattern)
	fmt.Fprintf(f.writer, "Matches: %d\n\n", len(result.MatchedFunctions))

	for i, fn := range result.MatchedFunctions {
		fmt.Fprintf(f.writer, "%d. ", i+1)
		if fn.Function != nil {
			fmt.Fprintf(f.writer, "%s", *fn.Function)
		} else if fn.LocationID != nil {
			if fn.Address != nil {
				fmt.Fprintf(f.writer, "%s (Location #%d)", *fn.Address, *fn.LocationID)
			} else {
				fmt.Fprintf(f.writer, "Location #%d", *fn.LocationID)
			}
		}
		if fn.File != nil {
			fmt.Fprintf(f.writer, " [%s]", *fn.File)
		}
		fmt.Fprintf(f.writer, "\n")

		fmt.Fprintf(f.writer, "   Flat: %d (%.2f%%), Cumulative: %d (%.2f%%)\n",
			fn.FlatValue, fn.FlatPercent, fn.CumValue, fn.CumPercent)

		indicators := []string{}
		if fn.IsLeaf {
			indicators = append(indicators, "leaf")
		}
		if fn.IsRecursive {
			indicators = append(indicators, "recursive")
		}
		if len(indicators) > 0 {
			fmt.Fprintf(f.writer, "   [%s]\n", strings.Join(indicators, ", "))
		}

		if len(fn.Callers) > 0 && !result.ListConfig.CalleesOnly {
			fmt.Fprintf(f.writer, "   Callers:\n")
			for _, caller := range fn.Callers {
				f.formatCallRelationship(caller, "   ")
			}
		}

		if len(fn.Callees) > 0 && !result.ListConfig.CallersOnly {
			fmt.Fprintf(f.writer, "   Callees:\n")
			for _, callee := range fn.Callees {
				f.formatCallRelationship(callee, "   ")
			}
		}

		if fn.IsLeaf && !result.ListConfig.CallersOnly {
			fmt.Fprintf(f.writer, "   No callees (leaf function)\n")
		}

		fmt.Fprintf(f.writer, "\n")
	}

	return nil
}

func (f *ListTextFormatter) formatCallRelationship(rel profile.CallRelationship, indent string) {
	fmt.Fprintf(f.writer, "%s  ", indent)
	if rel.Function != nil {
		fmt.Fprintf(f.writer, "%s", *rel.Function)
	} else if rel.Address != nil {
		fmt.Fprintf(f.writer, "%s", *rel.Address)
	}
	fmt.Fprintf(f.writer, ": %.2f%% (of parent), %.2f%% (total)\n",
		rel.RelationPercent, rel.FlatPercent)
}

// ListJSONFormatter formats list results in JSON format
type ListJSONFormatter struct {
	writer io.Writer
}

// NewListJSONFormatter creates a new list JSON formatter
func NewListJSONFormatter(w io.Writer) *ListJSONFormatter {
	return &ListJSONFormatter{writer: w}
}

// FormatListResult formats and outputs the list result as JSON
func (f *ListJSONFormatter) FormatListResult(result *profile.ListResult) error {
	functions := make([]map[string]interface{}, len(result.MatchedFunctions))
	for i, fn := range result.MatchedFunctions {
		functions[i] = map[string]interface{}{
			"function":     fn.Function,
			"file":         fn.File,
			"location_id":  fn.LocationID,
			"address":      fn.Address,
			"module":       fn.Module,
			"flat":         fn.FlatValue,
			"flat_percent": fn.FlatPercent,
			"cum":          fn.CumValue,
			"cum_percent":  fn.CumPercent,
			"is_leaf":      fn.IsLeaf,
			"is_recursive": fn.IsRecursive,
			"callers":      fn.Callers,
			"callees":      fn.Callees,
		}
	}

	output := map[string]interface{}{
		"pattern":   result.QueryPattern,
		"functions": functions,
	}

	encoder := json.NewEncoder(f.writer)
	encoder.SetIndent("", "  ")
	return encoder.Encode(output)
}

// ListMarkdownFormatter formats list results in Markdown format
type ListMarkdownFormatter struct {
	writer io.Writer
}

// NewListMarkdownFormatter creates a new list markdown formatter
func NewListMarkdownFormatter(w io.Writer) *ListMarkdownFormatter {
	return &ListMarkdownFormatter{writer: w}
}

// FormatListResult formats and outputs the list result as Markdown
func (f *ListMarkdownFormatter) FormatListResult(result *profile.ListResult) error {
	fmt.Fprintf(f.writer, "# Function Query Results\n\n")
	fmt.Fprintf(f.writer, "**Pattern:** `%s`\n\n", result.QueryPattern)
	fmt.Fprintf(f.writer, "**Matches:** %d\n\n", len(result.MatchedFunctions))

	for i, fn := range result.MatchedFunctions {
		fmt.Fprintf(f.writer, "## %d. ", i+1)
		if fn.Function != nil {
			fmt.Fprintf(f.writer, "`%s`", *fn.Function)
		} else if fn.LocationID != nil {
			if fn.Address != nil {
				fmt.Fprintf(f.writer, "%s (Location #%d)", *fn.Address, *fn.LocationID)
			} else {
				fmt.Fprintf(f.writer, "Location #%d", *fn.LocationID)
			}
		}
		fmt.Fprintf(f.writer, "\n\n")

		fmt.Fprintf(f.writer, "| Metric | Value | Percentage |\n")
		fmt.Fprintf(f.writer, "|--------|-------|------------|\n")
		fmt.Fprintf(f.writer, "| Flat | %d | %.2f%% |\n", fn.FlatValue, fn.FlatPercent)
		fmt.Fprintf(f.writer, "| Cumulative | %d | %.2f%% |\n\n", fn.CumValue, fn.CumPercent)

		tags := []string{}
		if fn.IsLeaf {
			tags = append(tags, "leaf")
		}
		if fn.IsRecursive {
			tags = append(tags, "recursive")
		}
		if len(tags) > 0 {
			fmt.Fprintf(f.writer, "**Tags:** %s\n\n", strings.Join(tags, ", "))
		}

		if len(fn.Callers) > 0 && !result.ListConfig.CalleesOnly {
			fmt.Fprintf(f.writer, "### Callers\n\n")
			fmt.Fprintf(f.writer, "| Function | %% of Parent | %% Total |\n")
			fmt.Fprintf(f.writer, "|----------|-------------|----------|\n")
			for _, caller := range fn.Callers {
				name := "unknown"
				if caller.Function != nil {
					name = fmt.Sprintf("`%s`", *caller.Function)
				} else if caller.Address != nil {
					name = *caller.Address
				}
				fmt.Fprintf(f.writer, "| %s | %.2f%% | %.2f%% |\n",
					name, caller.RelationPercent, caller.FlatPercent)
			}
			fmt.Fprintf(f.writer, "\n")
		}

		if len(fn.Callees) > 0 && !result.ListConfig.CallersOnly {
			fmt.Fprintf(f.writer, "### Callees\n\n")
			fmt.Fprintf(f.writer, "| Function | %% of Parent | %% Total |\n")
			fmt.Fprintf(f.writer, "|----------|-------------|----------|\n")
			for _, callee := range fn.Callees {
				name := "unknown"
				if callee.Function != nil {
					name = fmt.Sprintf("`%s`", *callee.Function)
				} else if callee.Address != nil {
					name = *callee.Address
				}
				fmt.Fprintf(f.writer, "| %s | %.2f%% | %.2f%% |\n",
					name, callee.RelationPercent, callee.FlatPercent)
			}
			fmt.Fprintf(f.writer, "\n")
		}

		if fn.IsLeaf && !result.ListConfig.CallersOnly {
			fmt.Fprintf(f.writer, "**No callees** (leaf function)\n\n")
		}
	}

	return nil
}

// FlameFormatter formats flame results in folded stack format
type FlameFormatter struct {
	writer io.Writer
}

// NewFlameFormatter creates a new flame formatter
func NewFlameFormatter(w io.Writer) *FlameFormatter {
	return &FlameFormatter{writer: w}
}

// FormatFlameResult formats and outputs the flame result
func (f *FlameFormatter) FormatFlameResult(result *profile.FlameResult) error {
	for _, stack := range result.Stacks {
		fmt.Fprintf(f.writer, "%s %d\n", stack.Stack, stack.Count)
	}
	return nil
}

// DiffFormatter interface for diff command output
type DiffFormatter interface {
	FormatDiffResult(result *profile.DiffResult, base, target string) error
}

// DiffTextFormatter formats diff results in text format
type DiffTextFormatter struct {
	writer io.Writer
}

// NewDiffTextFormatter creates a new diff text formatter
func NewDiffTextFormatter(w io.Writer) *DiffTextFormatter {
	return &DiffTextFormatter{writer: w}
}

// FormatDiffResult formats and outputs the diff result
func (f *DiffTextFormatter) FormatDiffResult(result *profile.DiffResult, base, target string) error {
	fmt.Fprintf(f.writer, "Profile Comparison\n")
	fmt.Fprintf(f.writer, "==================\n\n")
	fmt.Fprintf(f.writer, "Base:   %s\n", base)
	fmt.Fprintf(f.writer, "Target: %s\n\n", target)

	// Overall statistics
	fmt.Fprintf(f.writer, "Overall Changes:\n")
	fmt.Fprintf(f.writer, "  Base:   %d samples, total value: %d\n", result.OverallDiff.BaseSamples, result.OverallDiff.BaseTotal)
	fmt.Fprintf(f.writer, "  Target: %d samples, total value: %d\n", result.OverallDiff.TargetSamples, result.OverallDiff.TargetTotal)
	fmt.Fprintf(f.writer, "  Delta:  %d (%.2f%%)\n\n", result.OverallDiff.TotalDelta, result.OverallDiff.TotalPercent)

	// Regressions
	if len(result.Regressions) > 0 {
		fmt.Fprintf(f.writer, "Top Regressions (Performance Degradation)\n")
		fmt.Fprintf(f.writer, "%s\n", strings.Repeat("=", 50))
		for i, delta := range result.Regressions {
			f.formatFunctionDelta(i+1, delta, true)
		}
		fmt.Fprintf(f.writer, "\n")
	}

	// Improvements
	if len(result.Improvements) > 0 {
		fmt.Fprintf(f.writer, "Top Improvements\n")
		fmt.Fprintf(f.writer, "%s\n", strings.Repeat("=", 50))
		for i, delta := range result.Improvements {
			f.formatFunctionDelta(i+1, delta, false)
		}
		fmt.Fprintf(f.writer, "\n")
	}

	// New functions
	if len(result.NewFunctions) > 0 {
		fmt.Fprintf(f.writer, "New Functions (%d)\n", len(result.NewFunctions))
		fmt.Fprintf(f.writer, "%s\n", strings.Repeat("-", 50))
		for _, delta := range result.NewFunctions {
			if delta.Function != nil {
				fmt.Fprintf(f.writer, "  %s: %d\n", *delta.Function, delta.TargetFlat)
			}
		}
		fmt.Fprintf(f.writer, "\n")
	}

	// Deleted functions
	if len(result.DeletedFunctions) > 0 {
		fmt.Fprintf(f.writer, "Deleted Functions (%d)\n", len(result.DeletedFunctions))
		fmt.Fprintf(f.writer, "%s\n", strings.Repeat("-", 50))
		for _, delta := range result.DeletedFunctions {
			if delta.Function != nil {
				fmt.Fprintf(f.writer, "  %s: %d\n", *delta.Function, delta.BaseFlat)
			}
		}
		fmt.Fprintf(f.writer, "\n")
	}

	return nil
}

func (f *DiffTextFormatter) formatFunctionDelta(rank int, delta profile.FunctionDelta, isRegression bool) {
	fmt.Fprintf(f.writer, "\n%d. ", rank)
	if delta.Function != nil {
		fmt.Fprintf(f.writer, "%s", *delta.Function)
	} else if delta.Address != nil {
		fmt.Fprintf(f.writer, "%s", *delta.Address)
	}
	if delta.File != nil {
		fmt.Fprintf(f.writer, " [%s]", *delta.File)
	}
	fmt.Fprintf(f.writer, "\n")

	prefix := "+"
	if !isRegression {
		prefix = ""
	}

	fmt.Fprintf(f.writer, "   Flat:    %s%d (%s%.2f%% → %d)\n",
		prefix, delta.FlatDelta, prefix, delta.FlatDeltaPercent, delta.TargetFlat)
	fmt.Fprintf(f.writer, "   Cum:     %s%d (%s%.2f%% → %d)\n",
		prefix, delta.CumDelta, prefix, delta.CumDeltaPercent, delta.TargetCum)
}

// DiffJSONFormatter formats diff results in JSON format
type DiffJSONFormatter struct {
	writer io.Writer
}

// NewDiffJSONFormatter creates a new diff JSON formatter
func NewDiffJSONFormatter(w io.Writer) *DiffJSONFormatter {
	return &DiffJSONFormatter{writer: w}
}

// FormatDiffResult formats and outputs the diff result as JSON
func (f *DiffJSONFormatter) FormatDiffResult(result *profile.DiffResult, base, target string) error {
	output := map[string]interface{}{
		"base":         base,
		"target":       target,
		"value_type":   result.ValueType,
		"regressions":  result.Regressions,
		"improvements": result.Improvements,
		"new":          result.NewFunctions,
		"deleted":      result.DeletedFunctions,
		"overall": map[string]interface{}{
			"base_total":     result.OverallDiff.BaseTotal,
			"target_total":   result.OverallDiff.TargetTotal,
			"total_delta":    result.OverallDiff.TotalDelta,
			"total_percent":  result.OverallDiff.TotalPercent,
			"base_samples":   result.OverallDiff.BaseSamples,
			"target_samples": result.OverallDiff.TargetSamples,
		},
	}

	encoder := json.NewEncoder(f.writer)
	encoder.SetIndent("", "  ")
	return encoder.Encode(output)
}

// DiffMarkdownFormatter formats diff results in Markdown format
type DiffMarkdownFormatter struct {
	writer io.Writer
}

// NewDiffMarkdownFormatter creates a new diff markdown formatter
func NewDiffMarkdownFormatter(w io.Writer) *DiffMarkdownFormatter {
	return &DiffMarkdownFormatter{writer: w}
}

// FormatDiffResult formats and outputs the diff result as Markdown
func (f *DiffMarkdownFormatter) FormatDiffResult(result *profile.DiffResult, base, target string) error {
	fmt.Fprintf(f.writer, "# Profile Comparison\n\n")
	fmt.Fprintf(f.writer, "**Base:** `%s`\n\n", base)
	fmt.Fprintf(f.writer, "**Target:** `%s`\n\n", target)

	// Overall statistics
	fmt.Fprintf(f.writer, "## Overall Changes\n\n")
	fmt.Fprintf(f.writer, "| Metric | Base | Target | Delta | Percentage |\n")
	fmt.Fprintf(f.writer, "|--------|------|--------|-------|------------|\n")
	fmt.Fprintf(f.writer, "| Samples | %d | %d | %d | %.2f%% |\n",
		result.OverallDiff.BaseSamples,
		result.OverallDiff.TargetSamples,
		result.OverallDiff.TargetSamples-result.OverallDiff.BaseSamples,
		float64(result.OverallDiff.TargetSamples-result.OverallDiff.BaseSamples)/float64(result.OverallDiff.BaseSamples)*100)
	fmt.Fprintf(f.writer, "| Total Value | %d | %d | %d | %.2f%% |\n\n",
		result.OverallDiff.BaseTotal,
		result.OverallDiff.TargetTotal,
		result.OverallDiff.TotalDelta,
		result.OverallDiff.TotalPercent)

	// Regressions
	if len(result.Regressions) > 0 {
		fmt.Fprintf(f.writer, "## Top Regressions\n\n")
		fmt.Fprintf(f.writer, "| Rank | Function | Flat Delta | Flat %% | Cum Delta | Cum %% |\n")
		fmt.Fprintf(f.writer, "|------|----------|------------|--------|-----------|-------|\n")
		for i, delta := range result.Regressions {
			name := "unknown"
			if delta.Function != nil {
				name = fmt.Sprintf("`%s`", *delta.Function)
			} else if delta.Address != nil {
				name = *delta.Address
			}
			fmt.Fprintf(f.writer, "| %d | %s | +%d | +%.2f%% | +%d | +%.2f%% |\n",
				i+1, name, delta.FlatDelta, delta.FlatDeltaPercent,
				delta.CumDelta, delta.CumDeltaPercent)
		}
		fmt.Fprintf(f.writer, "\n")
	}

	// Improvements
	if len(result.Improvements) > 0 {
		fmt.Fprintf(f.writer, "## Top Improvements\n\n")
		fmt.Fprintf(f.writer, "| Rank | Function | Flat Delta | Flat %% | Cum Delta | Cum %% |\n")
		fmt.Fprintf(f.writer, "|------|----------|------------|--------|-----------|-------|\n")
		for i, delta := range result.Improvements {
			name := "unknown"
			if delta.Function != nil {
				name = fmt.Sprintf("`%s`", *delta.Function)
			} else if delta.Address != nil {
				name = *delta.Address
			}
			fmt.Fprintf(f.writer, "| %d | %s | %d | %.2f%% | %d | %.2f%% |\n",
				i+1, name, delta.FlatDelta, delta.FlatDeltaPercent,
				delta.CumDelta, delta.CumDeltaPercent)
		}
		fmt.Fprintf(f.writer, "\n")
	}

	return nil
}
