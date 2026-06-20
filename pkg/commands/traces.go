package commands

import (
	"fmt"

	"github.com/chenquan/agent-insight/pkg/output"
	"github.com/chenquan/agent-insight/pkg/profile"

	"github.com/spf13/cobra"
)

var (
	tracesFocus     string
	tracesIgnore    string
	tracesTop       int
	tracesFormat    string
	tracesTag       []string
	tracesIgnoreTag []string
)

var TracesCmd = &cobra.Command{
	Use:   "traces <profile.pb.gz> [flags]",
	Short: "Show individual sample call traces",
	Long: `Display raw sample call chains from a profile.

Shows each sample's full call stack (root to leaf) with its value.
Unlike flame (which aggregates), traces preserves individual sample detail.

Example usage:
  agent-insight traces profile.pb.gz
  agent-insight traces profile.pb.gz --focus "runtime.*"
  agent-insight traces cpu.pb.gz --ignore "runtime.*" --top 10`,
	Args: cobra.ExactArgs(1),
	RunE: runTraces,
}

func init() {
	TracesCmd.Flags().StringVar(&tracesFocus, "focus", "", "Regex pattern to focus on specific functions")
	TracesCmd.Flags().StringVar(&tracesIgnore, "ignore", "", "Regex pattern to ignore specific functions")
	TracesCmd.Flags().IntVar(&tracesTop, "top", 20, "Limit to top N traces")
	TracesCmd.Flags().StringSliceVar(&tracesTag, "tag", nil, "Filter samples by pprof label key=value (repeatable; same key OR, across keys AND)")
	TracesCmd.Flags().StringSliceVar(&tracesIgnoreTag, "tag-ignore", nil, "Exclude samples by pprof label key=value (same semantics as --tag)")
	TracesCmd.Flags().StringVar(&tracesFormat, "format", "text", "Output format: text, json, markdown")
}

func runTraces(cmd *cobra.Command, args []string) error {
	profilePath := args[0]

	if err := ValidateFormat(tracesFormat); err != nil {
		return err
	}

	if err := ValidateRegex(tracesFocus, "focus"); err != nil {
		return err
	}

	if err := ValidateRegex(tracesIgnore, "ignore"); err != nil {
		return err
	}

	loader := profile.NewLoader()
	p, err := loader.LoadFromFile(profilePath)
	if err != nil {
		return fmt.Errorf("failed to load profile: %w", err)
	}

	// Apply pprof label filter (--tag / --tag-ignore) before querying.
	labelFilter, err := profile.NewLabelFilter(tracesTag, tracesIgnoreTag)
	if err != nil {
		return err
	}
	p, err = labelFilter.Apply(p)
	if err != nil {
		return err
	}

	config := profile.TracesConfig{
		FocusPattern:  tracesFocus,
		IgnorePattern: tracesIgnore,
		TopN:          tracesTop,
	}

	result, err := profile.Traces(p, config)
	if err != nil {
		return fmt.Errorf("failed to query traces: %w", err)
	}

	out := cmd.OutOrStdout()
	switch tracesFormat {
	case "json":
		return output.NewTracesJSONFormatter(out).FormatTracesResult(result)
	case "markdown":
		return output.NewTracesMarkdownFormatter(out).FormatTracesResult(result)
	default:
		return output.NewTracesTextFormatter(out).FormatTracesResult(result)
	}
}
