package commands

import (
	"fmt"
	"os"

	"github.com/chenquan/agent-insight/pkg/output"
	"github.com/chenquan/agent-insight/pkg/profile"

	"github.com/spf13/cobra"
)

var (
	tracesFocus     string
	tracesIgnore    string
	tracesTop       int
	tracesValueType string
	tracesFormat    string
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
	TracesCmd.Flags().StringVar(&tracesValueType, "value-type", "", "Specify which value type to use")
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

	config := profile.TracesConfig{
		FocusPattern:  tracesFocus,
		IgnorePattern: tracesIgnore,
		TopN:          tracesTop,
	}

	result, err := profile.Traces(p, config)
	if err != nil {
		return fmt.Errorf("failed to query traces: %w", err)
	}

	switch tracesFormat {
	case "json":
		formatter := output.NewTracesJSONFormatter(os.Stdout)
		return formatter.FormatTracesResult(result)
	case "markdown":
		formatter := output.NewTracesMarkdownFormatter(os.Stdout)
		return formatter.FormatTracesResult(result)
	default:
		formatter := output.NewTracesTextFormatter(os.Stdout)
		return formatter.FormatTracesResult(result)
	}
}
