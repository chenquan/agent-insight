package commands

import (
	"fmt"
	"os"
	"regexp"

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
	analyzeTop       int
	analyzeCum       bool
	analyzeFocus     string
	analyzeIgnore    string
	analyzeFormat    string
	analyzeCallDepth int
	analyzeCollapse  bool
	analyzeValueType string
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
}

func runAnalyze(cmd *cobra.Command, args []string) error {
	profilePath := args[0]

	// Validate format
	if analyzeFormat != "text" && analyzeFormat != "json" && analyzeFormat != "markdown" {
		return fmt.Errorf("invalid format: %s (must be text, json, or markdown)", analyzeFormat)
	}

	// Load profile
	loader := profile.NewLoader()
	p, err := loader.LoadFromFile(profilePath)
	if err != nil {
		return fmt.Errorf("failed to load profile: %w", err)
	}

	// Configure analysis
	config := profile.AnalysisConfig{
		TopN:      analyzeTop,
		SortByCum: analyzeCum,
		CallDepth: analyzeCallDepth,
	}

	// Apply focus/ignore filters if provided
	if analyzeFocus != "" {
		config.FocusPattern = analyzeFocus
		// Validate regex
		if _, err := regexp.Compile(analyzeFocus); err != nil {
			return fmt.Errorf("invalid focus regex: %w", err)
		}
	}

	if analyzeIgnore != "" {
		config.IgnorePattern = analyzeIgnore
		// Validate regex
		if _, err := regexp.Compile(analyzeIgnore); err != nil {
			return fmt.Errorf("invalid ignore regex: %w", err)
		}
	}

	// Apply value type if specified
	if analyzeValueType != "" {
		// This will be resolved during analysis based on available value types
		// Store the requested type name for later validation
		_ = analyzeValueType // Will be used in analysis
	}

	// Perform analysis
	analysis, err := profile.NewAnalysis(p, config)
	if err != nil {
		return fmt.Errorf("failed to analyze profile: %w", err)
	}

	// Generate output based on format
	switch analyzeFormat {
	case "json":
		if err := outputJSON(analysis); err != nil {
			return err
		}
	case "markdown":
		if err := outputMarkdown(analysis); err != nil {
			return err
		}
	default:
		if err := outputText(analysis); err != nil {
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
		fmt.Fprintln(os.Stdout, "\n--- Collapsed Stacks ---")
		formatter := output.NewFlameFormatter(os.Stdout)
		return formatter.FormatFlameResult(flameResult)
	}

	return nil
}

// outputText outputs analysis in text format
func outputText(analysis *profile.Analysis) error {
	formatter := output.NewTextFormatter(os.Stdout)
	return formatter.FormatAnalysis(analysis)
}

// outputJSON outputs analysis in JSON format
func outputJSON(analysis *profile.Analysis) error {
	formatter := output.NewJSONFormatter(os.Stdout)
	return formatter.FormatAnalysis(analysis)
}

// outputMarkdown outputs analysis in Markdown format
func outputMarkdown(analysis *profile.Analysis) error {
	formatter := output.NewMarkdownFormatter(os.Stdout)
	return formatter.FormatAnalysis(analysis)
}
