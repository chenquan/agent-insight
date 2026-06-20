package commands

import (
	"fmt"
	"io"
	"strings"

	"github.com/chenquan/agent-insight/pkg/output"
	"github.com/chenquan/agent-insight/pkg/profile"

	"github.com/spf13/cobra"
)

// AnalyzeCmd represents the analyze command
var AnalyzeCmd = &cobra.Command{
	Use:   "analyze <profile.pb.gz> [flags]",
	Short: "Analyze pprof files and output performance hotspots",
	Long: `Analyze pprof protobuf files and output performance hotspot analysis.

Supports various profile types (CPU, heap, goroutine, etc.) and provides:
- Top N hotspots by flat or cumulative value
- Function-level performance metrics
- Call stack paths
- Natural language summary

Example usage:
  agent-insight analyze profile.pb.gz
  agent-insight analyze profile.pb.gz --top 20 --cum
  agent-insight analyze heap.prof --format json --focus runtime.*`,
	Args: cobra.ExactArgs(1),
	RunE: runAnalyze,
}

// Analyze flags
var (
	analyzeTop            int
	analyzeCum            bool
	analyzeFocus          string
	analyzeIgnore         string
	analyzeFormat         string
	analyzeCallDepth      int
	analyzeCollapse       bool
	analyzeValueType      string
	analyzeTag            []string
	analyzeTagIgnore      []string
	analyzeBreakdownOn    string
	analyzeBreakdownTop   int
)

func init() {
	// Add flags
	AnalyzeCmd.Flags().IntVar(&analyzeTop, "top", 15, "Output top N hotspots")
	AnalyzeCmd.Flags().BoolVar(&analyzeCum, "cum", false, "Sort by cumulative value instead of flat")
	AnalyzeCmd.Flags().StringVar(&analyzeFocus, "focus", "", "Regex pattern to focus on specific functions")
	AnalyzeCmd.Flags().StringVar(&analyzeIgnore, "ignore", "", "Regex pattern to ignore specific functions")
	AnalyzeCmd.Flags().StringVar(&analyzeFormat, "format", "text", "Output format: text, json, markdown")
	AnalyzeCmd.Flags().IntVar(&analyzeCallDepth, "call-depth", 5, "Call stack depth for path output")
	AnalyzeCmd.Flags().BoolVar(&analyzeCollapse, "collapse", false, "Include collapsed stack format output")
	AnalyzeCmd.Flags().StringVar(&analyzeValueType, "value-type", "", "Specify which value type to analyze (for multi-value profiles)")
	AnalyzeCmd.Flags().StringSliceVar(&analyzeTag, "tag", nil, "Filter samples by pprof label key=value (repeatable; same key OR, across keys AND)")
	AnalyzeCmd.Flags().StringSliceVar(&analyzeTagIgnore, "tag-ignore", nil, "Exclude samples by pprof label key=value (same semantics as --tag)")
	AnalyzeCmd.Flags().StringVar(&analyzeBreakdownOn, "tag-breakdown-on", "", "Comma-separated label keys to expand per-function label breakdown")
	AnalyzeCmd.Flags().IntVar(&analyzeBreakdownTop, "tag-breakdown-top", 20, "Number of top functions to expand label breakdown for")
}

func runAnalyze(cmd *cobra.Command, args []string) error {
	profilePath := args[0]

	// Validate format
	if err := ValidateFormat(analyzeFormat); err != nil {
		return err
	}

	// Load profile
	loader := profile.NewLoader()
	p, err := loader.LoadFromFile(profilePath)
	if err != nil {
		return fmt.Errorf("failed to load profile: %w", err)
	}

	// Apply pprof label filter (--tag / --tag-ignore) before analysis.
	labelFilter, err := profile.NewLabelFilter(analyzeTag, analyzeTagIgnore)
	if err != nil {
		return err
	}
	p, err = labelFilter.Apply(p)
	if err != nil {
		return err
	}

	// Configure analysis
	config := profile.AnalysisConfig{
		TopN:      analyzeTop,
		SortByCum: analyzeCum,
		CallDepth: analyzeCallDepth,
	}

	// Apply focus/ignore filters if provided
	if err := ValidateRegex(analyzeFocus, "focus"); err != nil {
		return err
	}
	if analyzeFocus != "" {
		config.FocusPattern = analyzeFocus
	}

	if err := ValidateRegex(analyzeIgnore, "ignore"); err != nil {
		return err
	}
	if analyzeIgnore != "" {
		config.IgnorePattern = analyzeIgnore
	}

	// Configure label breakdown if requested.
	if analyzeBreakdownOn != "" {
		keys := splitCSV(analyzeBreakdownOn)
		config.Breakdown = &profile.BreakdownConfig{
			Keys: keys,
			Top:  analyzeBreakdownTop,
		}
	}

	// Apply value type if specified
	if analyzeValueType != "" {
		found := false
		for i, st := range p.SampleType {
			if st.Type == analyzeValueType {
				config.ValueType = &profile.ValueTypeConfig{
					Name:  st.Type,
					Unit:  st.Unit,
					Index: i,
				}
				found = true
				break
			}
		}
		if !found {
			var available []string
			for _, st := range p.SampleType {
				available = append(available, st.Type+"/"+st.Unit)
			}
			return fmt.Errorf("unknown value type %q, available: %s", analyzeValueType, strings.Join(available, ", "))
		}
	}

	// Perform analysis
	analysis, err := profile.NewAnalysis(p, config)
	if err != nil {
		return fmt.Errorf("failed to analyze profile: %w", err)
	}

	// Generate output based on format
	out := cmd.OutOrStdout()
	switch analyzeFormat {
	case "json":
		if err := outputJSON(analysis, out); err != nil {
			return err
		}
	case "markdown":
		if err := outputMarkdown(analysis, out); err != nil {
			return err
		}
	default:
		if err := outputText(analysis, out); err != nil {
			return err
		}
	}

	// Append collapsed stack output if requested
	if analyzeCollapse {
		flameConfig := profile.FlameConfig{
			FocusPattern:  config.FocusPattern,
			IgnorePattern: config.IgnorePattern,
			ValueType:     config.ValueType,
		}
		flameResult, err := profile.Flame(p, flameConfig)
		if err != nil {
			return fmt.Errorf("failed to generate collapsed stacks: %w", err)
		}
		fmt.Fprintln(out, "\n--- Collapsed Stacks ---")
		formatter := output.NewFlameFormatter(out)
		return formatter.FormatFlameResult(flameResult)
	}

	return nil
}

// outputText outputs analysis in text format
func outputText(analysis *profile.Analysis, w io.Writer) error {
	formatter := output.NewTextFormatter(w)
	return formatter.FormatAnalysis(analysis)
}

// outputJSON outputs analysis in JSON format
func outputJSON(analysis *profile.Analysis, w io.Writer) error {
	formatter := output.NewJSONFormatter(w)
	return formatter.FormatAnalysis(analysis)
}

// outputMarkdown outputs analysis in Markdown format
func outputMarkdown(analysis *profile.Analysis, w io.Writer) error {
	formatter := output.NewMarkdownFormatter(w)
	return formatter.FormatAnalysis(analysis)
}

// splitCSV splits a comma-separated string into trimmed, non-empty fields.
func splitCSV(s string) []string {
	var out []string
	for part := range strings.SplitSeq(s, ",") {
		part = strings.TrimSpace(part)
		if part != "" {
			out = append(out, part)
		}
	}
	return out
}
