package output

import (
	"encoding/json"
	"fmt"
	"io"
	"math"
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
	Unit        string  `json:"unit"`
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
	unit := ""
	if analysis.Config.ValueType != nil {
		unit = analysis.Config.ValueType.Unit
	}
	for i, hotspot := range analysis.Hotspots {
		output.Top[i] = JSONHotspot{
			Function:    hotspot.Function,
			File:        hotspot.File,
			LocationID:  hotspot.LocationID,
			Address:     hotspot.Address,
			Module:      hotspot.Module,
			Unit:        unit,
			Flat:        hotspot.FlatValue,
			FlatPercent: roundPercent(hotspot.FlatPercent),
			Cum:         hotspot.CumValue,
			CumPercent:  roundPercent(hotspot.CumPercent),
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

func roundPercent(v float64) float64 {
	return math.Round(v*100) / 100
}

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

	// Determine wording based on profile type
	profileType := analysis.Metadata.Type
	hotspotLabel := "Top hotspot"
	concentratedLabel := "Highly concentrated performance bottleneck"
	switch profileType {
	case "heap", "space":
		hotspotLabel = "Top memory hotspot"
		concentratedLabel = "Highly concentrated memory allocation"
	case "goroutine":
		hotspotLabel = "Top blocking point"
		concentratedLabel = "Highly concentrated goroutine blocking"
	default:
		if profileType != "cpu" {
			concentratedLabel = "Highly concentrated workload"
		}
	}

	// Add top function info with more context
	if len(analysis.Hotspots) > 0 {
		top := analysis.Hotspots[0]
		topRef := "unknown location"
		if top.Function != nil {
			topRef = *top.Function
		} else if top.Address != nil {
			topRef = *top.Address
		}

		parts = append(parts, fmt.Sprintf("%s: %s (%.2f%%)", hotspotLabel, topRef, top.FlatPercent))

		// Add concentration info if top function dominates
		if top.FlatPercent > 30 {
			parts = append(parts, concentratedLabel)
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

// InfoTextFormatter formats info results in text format
type InfoTextFormatter struct {
	writer io.Writer
}

func NewInfoTextFormatter(w io.Writer) *InfoTextFormatter {
	return &InfoTextFormatter{writer: w}
}

func (f *InfoTextFormatter) FormatInfoResult(result *profile.InfoResult) error {
	fmt.Fprintf(f.writer, "Profile Info\n")
	fmt.Fprintf(f.writer, "============\n\n")

	fmt.Fprintf(f.writer, "Type:     %s\n", result.Type)
	if result.Duration > 0 {
		fmt.Fprintf(f.writer, "Duration: %s\n", result.Duration)
	}
	if result.Period > 0 && result.PeriodType != "" {
		fmt.Fprintf(f.writer, "Period:   %d (%s)\n", result.Period, result.PeriodType)
	}
	fmt.Fprintf(f.writer, "Samples:  %d\n", result.SampleCount)
	if result.TotalValue > 0 {
		fmt.Fprintf(f.writer, "Total Goroutines: %d\n", result.TotalValue)
	}
	fmt.Fprintf(f.writer, "Functions: %d\n", result.Functions)
	fmt.Fprintf(f.writer, "Locations: %d\n", result.Locations)

	if len(result.ValueTypes) > 0 {
		fmt.Fprintf(f.writer, "\nValue Types:\n")
		for _, vt := range result.ValueTypes {
			fmt.Fprintf(f.writer, "  %s/%s\n", vt.Type, vt.Unit)
		}
	}

	fmt.Fprintf(f.writer, "\nSymbols: ")
	if result.HasSymbols {
		fmt.Fprintf(f.writer, "available")
	} else {
		fmt.Fprintf(f.writer, "unavailable")
	}
	if result.HasFileLines {
		fmt.Fprintf(f.writer, " (file/lines: available)")
	}
	fmt.Fprintf(f.writer, "\n")

	if result.TimeRange.HasTime {
		fmt.Fprintf(f.writer, "\nTime Range:\n")
		fmt.Fprintf(f.writer, "  Start: %s\n", result.TimeRange.Start.Format("2006-01-02 15:04:05"))
		fmt.Fprintf(f.writer, "  End:   %s\n", result.TimeRange.End.Format("2006-01-02 15:04:05"))
	}

	if len(result.Mappings) > 0 {
		fmt.Fprintf(f.writer, "\nMappings (%d):\n", len(result.Mappings))
		for _, m := range result.Mappings {
			fmt.Fprintf(f.writer, "  %s", m.File)
			if m.BuildID != "" {
				fmt.Fprintf(f.writer, " (buildID: %s)", m.BuildID)
			}
			features := []string{}
			if m.HasFunctions {
				features = append(features, "functions")
			}
			if m.HasFilenames {
				features = append(features, "filenames")
			}
			if m.HasLineNumbers {
				features = append(features, "lines")
			}
			if m.HasInlineFrames {
				features = append(features, "inline")
			}
			if len(features) > 0 {
				fmt.Fprintf(f.writer, " [%s]", strings.Join(features, ", "))
			}
			fmt.Fprintf(f.writer, "\n")
		}
	}

	if len(result.Comments) > 0 {
		fmt.Fprintf(f.writer, "\nComments:\n")
		for _, c := range result.Comments {
			fmt.Fprintf(f.writer, "  %s\n", c)
		}
	}

	return nil
}

// InfoJSONFormatter formats info results in JSON format
type InfoJSONFormatter struct {
	writer io.Writer
}

func NewInfoJSONFormatter(w io.Writer) *InfoJSONFormatter {
	return &InfoJSONFormatter{writer: w}
}

func (f *InfoJSONFormatter) FormatInfoResult(result *profile.InfoResult) error {
	output := map[string]interface{}{
		"type":           result.Type,
		"duration":       result.Duration.String(),
		"period":         result.Period,
		"period_type":    result.PeriodType,
		"samples":        result.SampleCount,
		"functions":      result.Functions,
		"locations":      result.Locations,
		"has_symbols":    result.HasSymbols,
		"has_file_lines": result.HasFileLines,
		"value_types":    result.ValueTypes,
	}

	if result.TimeRange.HasTime {
		output["time_range"] = map[string]string{
			"start": result.TimeRange.Start.Format(time.RFC3339),
			"end":   result.TimeRange.End.Format(time.RFC3339),
		}
	}

	if len(result.Mappings) > 0 {
		output["mappings"] = result.Mappings
	}

	if result.TotalValue > 0 {
		output["total_value"] = result.TotalValue
	}

	if len(result.Comments) > 0 {
		output["comments"] = result.Comments
	}

	encoder := json.NewEncoder(f.writer)
	encoder.SetIndent("", "  ")
	return encoder.Encode(output)
}

// InfoMarkdownFormatter formats info results in Markdown format
type InfoMarkdownFormatter struct {
	writer io.Writer
}

func NewInfoMarkdownFormatter(w io.Writer) *InfoMarkdownFormatter {
	return &InfoMarkdownFormatter{writer: w}
}

func (f *InfoMarkdownFormatter) FormatInfoResult(result *profile.InfoResult) error {
	fmt.Fprintf(f.writer, "# Profile Info\n\n")

	fmt.Fprintf(f.writer, "| Property | Value |\n")
	fmt.Fprintf(f.writer, "|----------|-------|\n")
	fmt.Fprintf(f.writer, "| Type | %s |\n", result.Type)
	if result.Duration > 0 {
		fmt.Fprintf(f.writer, "| Duration | %s |\n", result.Duration)
	}
	if result.Period > 0 && result.PeriodType != "" {
		fmt.Fprintf(f.writer, "| Period | %d (%s) |\n", result.Period, result.PeriodType)
	}
	fmt.Fprintf(f.writer, "| Samples | %d |\n", result.SampleCount)
	if result.TotalValue > 0 {
		fmt.Fprintf(f.writer, "| Total Goroutines | %d |\n", result.TotalValue)
	}
	fmt.Fprintf(f.writer, "| Functions | %d |\n", result.Functions)
	fmt.Fprintf(f.writer, "| Locations | %d |\n", result.Locations)
	fmt.Fprintf(f.writer, "| Symbols | %v |\n", result.HasSymbols)
	fmt.Fprintf(f.writer, "| File/Lines | %v |\n", result.HasFileLines)

	if len(result.ValueTypes) > 0 {
		fmt.Fprintf(f.writer, "\n## Value Types\n\n")
		fmt.Fprintf(f.writer, "| Type | Unit |\n")
		fmt.Fprintf(f.writer, "|------|------|\n")
		for _, vt := range result.ValueTypes {
			fmt.Fprintf(f.writer, "| %s | %s |\n", vt.Type, vt.Unit)
		}
	}

	if result.TimeRange.HasTime {
		fmt.Fprintf(f.writer, "\n## Time Range\n\n")
		fmt.Fprintf(f.writer, "| Property | Value |\n")
		fmt.Fprintf(f.writer, "|----------|-------|\n")
		fmt.Fprintf(f.writer, "| Start | %s |\n", result.TimeRange.Start.Format("2006-01-02 15:04:05"))
		fmt.Fprintf(f.writer, "| End | %s |\n", result.TimeRange.End.Format("2006-01-02 15:04:05"))
	}

	if len(result.Mappings) > 0 {
		fmt.Fprintf(f.writer, "\n## Mappings\n\n")
		fmt.Fprintf(f.writer, "| File | Build ID | Features |\n")
		fmt.Fprintf(f.writer, "|------|----------|----------|\n")
		for _, m := range result.Mappings {
			features := []string{}
			if m.HasFunctions {
				features = append(features, "functions")
			}
			if m.HasFilenames {
				features = append(features, "filenames")
			}
			if m.HasLineNumbers {
				features = append(features, "lines")
			}
			if m.HasInlineFrames {
				features = append(features, "inline")
			}
			fmt.Fprintf(f.writer, "| `%s` | %s | %s |\n", m.File, m.BuildID, strings.Join(features, ", "))
		}
	}

	if len(result.Comments) > 0 {
		fmt.Fprintf(f.writer, "\n## Comments\n\n")
		for _, c := range result.Comments {
			fmt.Fprintf(f.writer, "- %s\n", c)
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

// Format dispatches to the appropriate format method based on the format string.
func (f *FlameFormatter) Format(result *profile.FlameResult, format string) error {
	switch format {
	case "json":
		return f.formatJSON(result)
	case "markdown":
		return f.formatMarkdown(result)
	default:
		return f.FormatFlameResult(result)
	}
}

// flameJSONOutput is the JSON output structure for flame results.
type flameJSONOutput struct {
	TotalStacks    int              `json:"total_stacks"`
	FilteredStacks int              `json:"filtered_stacks,omitempty"`
	UniqueStacks   int              `json:"unique_stacks"`
	Stacks         []flameJSONStack `json:"stacks"`
}

type flameJSONStack struct {
	Stack []string `json:"stack"`
	Value int64    `json:"value"`
}

func (f *FlameFormatter) formatJSON(result *profile.FlameResult) error {
	output := flameJSONOutput{
		TotalStacks:    result.TotalStacks,
		FilteredStacks: result.FilteredStacks,
		UniqueStacks:   result.UniqueStacks,
		Stacks:         make([]flameJSONStack, len(result.Stacks)),
	}
	for i, s := range result.Stacks {
		output.Stacks[i] = flameJSONStack{
			Stack: strings.Split(s.Stack, ";"),
			Value: s.Count,
		}
	}
	encoder := json.NewEncoder(f.writer)
	encoder.SetIndent("", "  ")
	return encoder.Encode(output)
}

func (f *FlameFormatter) formatMarkdown(result *profile.FlameResult) error {
	fmt.Fprintf(f.writer, "# Flame Graph Data\n\n")
	fmt.Fprintf(f.writer, "| Metric | Value |\n|---|---|\n")
	fmt.Fprintf(f.writer, "| Total Stacks | %d |\n", result.TotalStacks)
	if result.FilteredStacks > 0 {
		fmt.Fprintf(f.writer, "| Filtered Stacks | %d |\n", result.FilteredStacks)
	}
	fmt.Fprintf(f.writer, "| Unique Stacks | %d |\n", result.UniqueStacks)
	fmt.Fprintf(f.writer, "| Output Stacks | %d |\n\n", len(result.Stacks))

	// Top 20 stacks table
	limit := 20
	if len(result.Stacks) < limit {
		limit = len(result.Stacks)
	}
	fmt.Fprintf(f.writer, "## Top %d Stacks\n\n", limit)
	fmt.Fprintf(f.writer, "| # | Stack | Value |\n|---|---|---|\n")
	for i := 0; i < limit; i++ {
		s := result.Stacks[i]
		fmt.Fprintf(f.writer, "| %d | `%s` | %d |\n", i+1, s.Stack, s.Count)
	}

	// Full folded stacks in code block
	fmt.Fprintf(f.writer, "\n## Folded Stacks\n\n```\n")
	for _, s := range result.Stacks {
		fmt.Fprintf(f.writer, "%s %d\n", s.Stack, s.Count)
	}
	fmt.Fprintf(f.writer, "```\n")
	return nil
}

// DiffFormatter interface for diff command output
type DiffFormatter interface {
	FormatDiffResult(result *profile.DiffResult, base, target string) error
}

// TreeTextFormatter formats tree results in text format
type TreeTextFormatter struct {
	writer io.Writer
}

func NewTreeTextFormatter(w io.Writer) *TreeTextFormatter {
	return &TreeTextFormatter{writer: w}
}

func (f *TreeTextFormatter) FormatTreeResult(result *profile.TreeResult) error {
	fmt.Fprintf(f.writer, "Call Tree (Value Type: %s)\n", result.ValueType)
	fmt.Fprintf(f.writer, "%s\n\n", strings.Repeat("=", 50))

	for _, child := range result.VisibleChildren() {
		f.formatNode(child, 0, 5)
	}

	return nil
}

func (f *TreeTextFormatter) formatNode(node *profile.CallTreeNode, depth, maxDepth int) {
	if depth >= maxDepth {
		return
	}

	indent := strings.Repeat("  ", depth)
	fmt.Fprintf(f.writer, "%s%s  flat: %d (%.2f%%)  cum: %d (%.2f%%)\n",
		indent, node.Name, node.Flat, node.FlatPercent, node.Cum, node.CumPercent)

	for _, child := range node.Children {
		f.formatNode(child, depth+1, maxDepth)
	}
}

// TreeJSONFormatter formats tree results in JSON format
type TreeJSONFormatter struct {
	writer io.Writer
}

func NewTreeJSONFormatter(w io.Writer) *TreeJSONFormatter {
	return &TreeJSONFormatter{writer: w}
}

func (f *TreeJSONFormatter) FormatTreeResult(result *profile.TreeResult) error {
	type jsonNode struct {
		Name        string      `json:"name"`
		Flat        int64       `json:"flat"`
		FlatPercent float64     `json:"flat_percent"`
		Cum         int64       `json:"cum"`
		CumPercent  float64     `json:"cum_percent"`
		Children    []*jsonNode `json:"children,omitempty"`
	}

	var convertNode func(n *profile.CallTreeNode) *jsonNode
	convertNode = func(n *profile.CallTreeNode) *jsonNode {
		jn := &jsonNode{
			Name:        n.Name,
			Flat:        n.Flat,
			FlatPercent: n.FlatPercent,
			Cum:         n.Cum,
			CumPercent:  n.CumPercent,
		}
		for _, child := range n.Children {
			jn.Children = append(jn.Children, convertNode(child))
		}
		return jn
	}

	children := make([]*jsonNode, len(result.VisibleChildren()))
	for i, child := range result.VisibleChildren() {
		children[i] = convertNode(child)
	}

	output := map[string]interface{}{
		"value_type": result.ValueType,
		"tree":       children,
	}

	encoder := json.NewEncoder(f.writer)
	encoder.SetIndent("", "  ")
	return encoder.Encode(output)
}

// TreeMarkdownFormatter formats tree results in Markdown format
type TreeMarkdownFormatter struct {
	writer io.Writer
}

func NewTreeMarkdownFormatter(w io.Writer) *TreeMarkdownFormatter {
	return &TreeMarkdownFormatter{writer: w}
}

func (f *TreeMarkdownFormatter) FormatTreeResult(result *profile.TreeResult) error {
	fmt.Fprintf(f.writer, "# Call Tree\n\n")
	fmt.Fprintf(f.writer, "**Value Type:** %s\n\n", result.ValueType)

	for _, child := range result.VisibleChildren() {
		f.formatNode(child, 0, 5)
	}

	return nil
}

func (f *TreeMarkdownFormatter) formatNode(node *profile.CallTreeNode, depth, maxDepth int) {
	if depth >= maxDepth {
		return
	}

	indent := strings.Repeat("  ", depth)
	fmt.Fprintf(f.writer, "%s- `%s` — flat: %d (%.2f%%), cum: %d (%.2f%%)\n",
		indent, node.Name, node.Flat, node.FlatPercent, node.Cum, node.CumPercent)

	for _, child := range node.Children {
		f.formatNode(child, depth+1, maxDepth)
	}
}

// TracesTextFormatter formats traces results in text format
type TracesTextFormatter struct {
	writer io.Writer
}

func NewTracesTextFormatter(w io.Writer) *TracesTextFormatter {
	return &TracesTextFormatter{writer: w}
}

func (f *TracesTextFormatter) FormatTracesResult(result *profile.TracesResult) error {
	fmt.Fprintf(f.writer, "Traces (%d shown / %d total)\n", result.ShownTraces, result.TotalTraces)
	fmt.Fprintf(f.writer, "Value Type: %s\n\n", result.ValueType)

	for i, trace := range result.Traces {
		fmt.Fprintf(f.writer, "Trace #%d  Value: %d (%.2f%%)\n", i+1, trace.Value, trace.Percent)
		for j, fn := range trace.Stack {
			fmt.Fprintf(f.writer, "  %s%s\n", strings.Repeat("  ", j), fn)
		}
		fmt.Fprintf(f.writer, "\n")
	}

	return nil
}

// TracesJSONFormatter formats traces results in JSON format
type TracesJSONFormatter struct {
	writer io.Writer
}

func NewTracesJSONFormatter(w io.Writer) *TracesJSONFormatter {
	return &TracesJSONFormatter{writer: w}
}

func (f *TracesJSONFormatter) FormatTracesResult(result *profile.TracesResult) error {
	type jsonTrace struct {
		Stack   []string `json:"stack"`
		Value   int64    `json:"value"`
		Percent float64  `json:"percent"`
	}

	traces := make([]jsonTrace, len(result.Traces))
	for i, t := range result.Traces {
		traces[i] = jsonTrace{
			Stack:   t.Stack,
			Value:   t.Value,
			Percent: t.Percent,
		}
	}

	output := map[string]interface{}{
		"value_type":   result.ValueType,
		"total_traces": result.TotalTraces,
		"shown_traces": result.ShownTraces,
		"traces":       traces,
	}

	encoder := json.NewEncoder(f.writer)
	encoder.SetIndent("", "  ")
	return encoder.Encode(output)
}

// TracesMarkdownFormatter formats traces results in Markdown format
type TracesMarkdownFormatter struct {
	writer io.Writer
}

func NewTracesMarkdownFormatter(w io.Writer) *TracesMarkdownFormatter {
	return &TracesMarkdownFormatter{writer: w}
}

func (f *TracesMarkdownFormatter) FormatTracesResult(result *profile.TracesResult) error {
	fmt.Fprintf(f.writer, "# Traces (%d shown / %d total)\n\n", result.ShownTraces, result.TotalTraces)
	fmt.Fprintf(f.writer, "**Value Type:** %s\n\n", result.ValueType)

	for i, trace := range result.Traces {
		fmt.Fprintf(f.writer, "## Trace #%d — Value: %d (%.2f%%)\n\n", i+1, trace.Value, trace.Percent)
		for j, fn := range trace.Stack {
			fmt.Fprintf(f.writer, "%s- `%s`\n", strings.Repeat("  ", j), fn)
		}
		fmt.Fprintf(f.writer, "\n")
	}

	return nil
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
	if delta.CumDelta != 0 || delta.CumDeltaPercent != 0 {
		fmt.Fprintf(f.writer, "   Cum:     %s%d (%s%.2f%% → %d)\n",
			prefix, delta.CumDelta, prefix, delta.CumDeltaPercent, delta.TargetCum)
	}
}

// DiffJSONFormatter formats diff results in JSON format
type DiffJSONFormatter struct {
	writer io.Writer
}

// NewDiffJSONFormatter creates a new diff JSON formatter
func NewDiffJSONFormatter(w io.Writer) *DiffJSONFormatter {
	return &DiffJSONFormatter{writer: w}
}

// convertDeltaToMap converts a FunctionDelta to a snake_case map for JSON output.
func convertDeltaToMap(delta profile.FunctionDelta) map[string]interface{} {
	return map[string]interface{}{
		"function":          delta.Function,
		"file":              delta.File,
		"location_id":       delta.LocationID,
		"address":           delta.Address,
		"module":            delta.Module,
		"base_flat":         delta.BaseFlat,
		"target_flat":       delta.TargetFlat,
		"base_cum":          delta.BaseCum,
		"target_cum":        delta.TargetCum,
		"flat_delta":        delta.FlatDelta,
		"flat_delta_percent": roundPercent(delta.FlatDeltaPercent),
		"cum_delta":         delta.CumDelta,
		"cum_delta_percent":  roundPercent(delta.CumDeltaPercent),
		"is_new":            delta.IsNew,
		"is_deleted":        delta.IsDeleted,
	}
}

// FormatDiffResult formats and outputs the diff result as JSON
func (f *DiffJSONFormatter) FormatDiffResult(result *profile.DiffResult, base, target string) error {
	// Convert function delta slices to snake_case maps
	regressions := make([]map[string]interface{}, len(result.Regressions))
	for i, d := range result.Regressions {
		regressions[i] = convertDeltaToMap(d)
	}
	improvements := make([]map[string]interface{}, len(result.Improvements))
	for i, d := range result.Improvements {
		improvements[i] = convertDeltaToMap(d)
	}
	newFns := make([]map[string]interface{}, len(result.NewFunctions))
	for i, d := range result.NewFunctions {
		newFns[i] = convertDeltaToMap(d)
	}
	deletedFns := make([]map[string]interface{}, len(result.DeletedFunctions))
	for i, d := range result.DeletedFunctions {
		deletedFns[i] = convertDeltaToMap(d)
	}

	output := map[string]interface{}{
		"base":         base,
		"target":       target,
		"value_type":   result.ValueType,
		"regressions":  regressions,
		"improvements": improvements,
		"new":          newFns,
		"deleted":      deletedFns,
		"overall": map[string]interface{}{
			"base_total":     result.OverallDiff.BaseTotal,
			"target_total":   result.OverallDiff.TargetTotal,
			"total_delta":    result.OverallDiff.TotalDelta,
			"total_percent":  roundPercent(result.OverallDiff.TotalPercent),
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

// TrendFormatter interface for trend command output
type TrendFormatter interface {
	FormatTrendResult(result *profile.TrendResult) error
}

// TrendTextFormatter outputs trend result in human-readable text format
type TrendTextFormatter struct {
	writer io.Writer
}

// NewTrendTextFormatter creates a new trend text formatter
func NewTrendTextFormatter(w io.Writer) *TrendTextFormatter {
	return &TrendTextFormatter{writer: w}
}

// FormatTrendResult formats and outputs the trend result as text
func (f *TrendTextFormatter) FormatTrendResult(result *profile.TrendResult) error {
	w := f.writer

	fmt.Fprintf(w, "Performance Trend Analysis\n")
	fmt.Fprintf(w, "==========================\n\n")

	fmt.Fprintf(w, "Profiles: %d | Value Type: %s\n", len(result.TimePoints), result.ValueType)
	if len(result.TimePoints) > 0 {
		fmt.Fprintf(w, "Time Range: %s → %s\n", result.TimePoints[0].Label, result.TimePoints[len(result.TimePoints)-1].Label)
	}
	fmt.Fprintf(w, "Overall Slope: %.2f\n", result.Overall.Slope)
	fmt.Fprintf(w, "Functions: %d regressing, %d improving, %d stable\n\n",
		result.RegressingCount, result.ImprovingCount, result.StableCount)

	if len(result.TopRegressions) > 0 {
		fmt.Fprintf(w, "Top Regressions (slope descending)\n")
		fmt.Fprintf(w, "%s\n", strings.Repeat("-", 50))
		for i, ft := range result.TopRegressions {
			fmt.Fprintf(w, "\n%d. %s [slope: %.2f, %s]\n", i+1, funcName(ft), ft.Slope, ft.Trend)
			fmt.Fprintf(w, "   flat series: %s\n", formatSeries(ft.FlatSeries))
			fmt.Fprintf(w, "   start: %s → end: %s (%s)\n",
				formatIntPtr(ft.StartValue), formatIntPtr(ft.EndValue), changePercent(ft.StartValue, ft.EndValue))
		}
		fmt.Fprintf(w, "\n")
	}

	if len(result.TopImprovements) > 0 {
		fmt.Fprintf(w, "Top Improvements (slope ascending)\n")
		fmt.Fprintf(w, "%s\n", strings.Repeat("-", 50))
		for i, ft := range result.TopImprovements {
			fmt.Fprintf(w, "\n%d. %s [slope: %.2f, %s]\n", i+1, funcName(ft), ft.Slope, ft.Trend)
			fmt.Fprintf(w, "   flat series: %s\n", formatSeries(ft.FlatSeries))
			fmt.Fprintf(w, "   start: %s → end: %s (%s)\n",
				formatIntPtr(ft.StartValue), formatIntPtr(ft.EndValue), changePercent(ft.StartValue, ft.EndValue))
		}
		fmt.Fprintf(w, "\n")
	}

	if len(result.NewHotspots) > 0 {
		fmt.Fprintf(w, "New Hotspots\n")
		fmt.Fprintf(w, "%s\n", strings.Repeat("-", 50))
		for i, ft := range result.NewHotspots {
			fmt.Fprintf(w, "\n%d. %s\n", i+1, funcName(ft))
			fmt.Fprintf(w, "   flat series: %s\n", formatSeries(ft.FlatSeries))
		}
		fmt.Fprintf(w, "\n")
	}

	if len(result.VolatileFunctions) > 0 {
		fmt.Fprintf(w, "Volatile Functions (CV > 0.3)\n")
		fmt.Fprintf(w, "%s\n", strings.Repeat("-", 50))
		for i, ft := range result.VolatileFunctions {
			fmt.Fprintf(w, "\n%d. %s [CV: %.3f]\n", i+1, funcName(ft), ft.Volatility)
			fmt.Fprintf(w, "   flat series: %s\n", formatSeries(ft.FlatSeries))
		}
		fmt.Fprintf(w, "\n")
	}

	return nil
}

// TrendJSONFormatter outputs trend result as JSON
type TrendJSONFormatter struct {
	writer io.Writer
}

// NewTrendJSONFormatter creates a new trend JSON formatter
func NewTrendJSONFormatter(w io.Writer) *TrendJSONFormatter {
	return &TrendJSONFormatter{writer: w}
}

// FormatTrendResult formats and outputs the trend result as JSON
func (f *TrendJSONFormatter) FormatTrendResult(result *profile.TrendResult) error {
	output := map[string]interface{}{
		"value_type":  result.ValueType,
		"time_points": convertTimePoints(result.TimePoints),
		"overall": map[string]interface{}{
			"total_series": result.Overall.TotalSeries,
			"slope":        result.Overall.Slope,
		},
		"summary": map[string]interface{}{
			"regressing_count": result.RegressingCount,
			"improving_count":  result.ImprovingCount,
			"stable_count":     result.StableCount,
		},
		"top_regressions":  convertFunctionTrends(result.TopRegressions),
		"top_improvements": convertFunctionTrends(result.TopImprovements),
	}

	if len(result.NewHotspots) > 0 {
		output["new_hotspots"] = convertFunctionTrends(result.NewHotspots)
	}
	if len(result.VolatileFunctions) > 0 {
		output["volatile"] = convertFunctionTrends(result.VolatileFunctions)
	}

	enc := json.NewEncoder(f.writer)
	enc.SetIndent("", "  ")
	return enc.Encode(output)
}

// TrendMarkdownFormatter outputs trend result as Markdown
type TrendMarkdownFormatter struct {
	writer io.Writer
}

// NewTrendMarkdownFormatter creates a new trend markdown formatter
func NewTrendMarkdownFormatter(w io.Writer) *TrendMarkdownFormatter {
	return &TrendMarkdownFormatter{writer: w}
}

// FormatTrendMarkdownResult formats and outputs the trend result as Markdown
func (f *TrendMarkdownFormatter) FormatTrendMarkdownResult(result *profile.TrendResult) error {
	w := f.writer

	fmt.Fprintf(w, "# Performance Trend Analysis\n\n")
	fmt.Fprintf(w, "- **Profiles:** %d\n", len(result.TimePoints))
	fmt.Fprintf(w, "- **Value Type:** %s\n", result.ValueType)
	if len(result.TimePoints) > 0 {
		fmt.Fprintf(w, "- **Time Range:** `%s` → `%s`\n", result.TimePoints[0].Label, result.TimePoints[len(result.TimePoints)-1].Label)
	}
	fmt.Fprintf(w, "- **Overall Slope:** %.2f\n", result.Overall.Slope)
	fmt.Fprintf(w, "- **Functions:** %d regressing, %d improving, %d stable\n\n",
		result.RegressingCount, result.ImprovingCount, result.StableCount)

	if len(result.TopRegressions) > 0 {
		fmt.Fprintf(w, "## Top Regressions\n\n")
		fmt.Fprintf(w, "| Rank | Function | Slope | Trend | Flat Series | Change |\n")
		fmt.Fprintf(w, "|------|----------|-------|-------|-------------|--------|\n")
		for i, ft := range result.TopRegressions {
			fmt.Fprintf(w, "| %d | `%s` | %.2f | %s | %s | %s |\n",
				i+1, funcName(ft), ft.Slope, ft.Trend,
				formatSeries(ft.FlatSeries), changePercent(ft.StartValue, ft.EndValue))
		}
		fmt.Fprintf(w, "\n")
	}

	if len(result.TopImprovements) > 0 {
		fmt.Fprintf(w, "## Top Improvements\n\n")
		fmt.Fprintf(w, "| Rank | Function | Slope | Trend | Flat Series | Change |\n")
		fmt.Fprintf(w, "|------|----------|-------|-------|-------------|--------|\n")
		for i, ft := range result.TopImprovements {
			fmt.Fprintf(w, "| %d | `%s` | %.2f | %s | %s | %s |\n",
				i+1, funcName(ft), ft.Slope, ft.Trend,
				formatSeries(ft.FlatSeries), changePercent(ft.StartValue, ft.EndValue))
		}
		fmt.Fprintf(w, "\n")
	}

	if len(result.NewHotspots) > 0 {
		fmt.Fprintf(w, "## New Hotspots\n\n")
		fmt.Fprintf(w, "| Rank | Function | Flat Series |\n")
		fmt.Fprintf(w, "|------|----------|-------------|\n")
		for i, ft := range result.NewHotspots {
			fmt.Fprintf(w, "| %d | `%s` | %s |\n", i+1, funcName(ft), formatSeries(ft.FlatSeries))
		}
		fmt.Fprintf(w, "\n")
	}

	if len(result.VolatileFunctions) > 0 {
		fmt.Fprintf(w, "## Volatile Functions\n\n")
		fmt.Fprintf(w, "| Rank | Function | CV | Flat Series |\n")
		fmt.Fprintf(w, "|------|----------|----|-------------|\n")
		for i, ft := range result.VolatileFunctions {
			fmt.Fprintf(w, "| %d | `%s` | %.3f | %s |\n", i+1, funcName(ft), ft.Volatility, formatSeries(ft.FlatSeries))
		}
		fmt.Fprintf(w, "\n")
	}

	return nil
}

func funcName(ft profile.FunctionTrend) string {
	if ft.Function != nil {
		return *ft.Function
	}
	if ft.Address != nil {
		return *ft.Address
	}
	return fmt.Sprintf("Location %s", formatLocationID(*ft.LocationID))
}

func formatSeries(series []*int64) string {
	parts := make([]string, len(series))
	for i, v := range series {
		if v == nil {
			parts[i] = "-"
		} else {
			parts[i] = fmt.Sprintf("%d", *v)
		}
	}
	return "[" + strings.Join(parts, ", ") + "]"
}

func formatIntPtr(v *int64) string {
	if v == nil {
		return "-"
	}
	return fmt.Sprintf("%d", *v)
}

func changePercent(start, end *int64) string {
	if start == nil || end == nil || *start == 0 {
		return "-"
	}
	pct := float64(*end-*start) / float64(*start) * 100
	if pct >= 0 {
		return fmt.Sprintf("+%.1f%%", pct)
	}
	return fmt.Sprintf("%.1f%%", pct)
}

func convertTimePoints(tps []profile.TimePoint) []map[string]interface{} {
	result := make([]map[string]interface{}, len(tps))
	for i, tp := range tps {
		result[i] = map[string]interface{}{
			"label": tp.Label,
			"time":  tp.Time,
		}
	}
	return result
}

func convertFunctionTrends(fts []profile.FunctionTrend) []map[string]interface{} {
	result := make([]map[string]interface{}, len(fts))
	for i, ft := range fts {
		entry := map[string]interface{}{
			"slope":      ft.Slope,
			"trend":      ft.Trend,
			"avg_flat":   ft.AvgFlat,
			"volatility": ft.Volatility,
		}
		if ft.Function != nil {
			entry["function"] = *ft.Function
		}
		if ft.File != nil {
			entry["file"] = *ft.File
		}
		if ft.Address != nil {
			entry["address"] = *ft.Address
		}
		if ft.Module != nil {
			entry["module"] = *ft.Module
		}

		flatSeries := make([]interface{}, len(ft.FlatSeries))
		for j, v := range ft.FlatSeries {
			flatSeries[j] = v
		}
		cumSeries := make([]interface{}, len(ft.CumSeries))
		for j, v := range ft.CumSeries {
			cumSeries[j] = v
		}
		entry["flat_series"] = flatSeries
		entry["cum_series"] = cumSeries

		if ft.StartValue != nil {
			entry["start_value"] = *ft.StartValue
		}
		if ft.EndValue != nil {
			entry["end_value"] = *ft.EndValue
		}

		result[i] = entry
	}
	return result
}
