package commands

import (
	"fmt"
	"io"

	"github.com/chenquan/agent-insight/pkg/output"
	"github.com/chenquan/agent-insight/pkg/profile"

	"github.com/spf13/cobra"
)

// ListCmd represents the list command
var ListCmd = &cobra.Command{
	Use:   "list <profile.pb.gz> <function-pattern> [flags]",
	Short: "Query specific function call relationships",
	Long: `Query specific functions and display their caller/callee relationships.

Shows which functions call the target function (callers) and which functions
are called by the target function (callees), along with their performance contributions.

Example usage:
  agent-insight list profile.pb.gz "main.*"
  agent-insight list profile.pb.gz "runtime.mallocgc" --callers-only
  agent-insight list heap.prof "encoding.*" --depth 3 --format json`,
	Args: cobra.ExactArgs(2),
	RunE: runList,
}

// List flags
var (
	listDepth          int
	listCallersOnly    bool
	listCalleesOnly    bool
	listIgnoreFunction string
	listFormat         string
	listTag            []string
	listIgnoreTag      []string
)

func init() {
	// Add flags
	ListCmd.Flags().IntVar(&listDepth, "depth", 5, "Maximum depth of caller/callee relationships to show")
	ListCmd.Flags().BoolVar(&listCallersOnly, "callers-only", false, "Show only callers, exclude callees")
	ListCmd.Flags().BoolVar(&listCalleesOnly, "callees-only", false, "Show only callees, exclude callers")
	ListCmd.Flags().StringVar(&listIgnoreFunction, "ignore-function", "", "Regex pattern to exclude matching functions")
	ListCmd.Flags().StringSliceVar(&listTag, "tag", nil, "Filter samples by pprof label key=value (repeatable; same key OR, across keys AND)")
	ListCmd.Flags().StringSliceVar(&listIgnoreTag, "tag-ignore", nil, "Exclude samples by pprof label key=value (same semantics as --tag)")
	ListCmd.Flags().StringVar(&listFormat, "format", "text", "Output format: text, json, markdown")
}

func runList(cmd *cobra.Command, args []string) error {
	profilePath := args[0]
	pattern := args[1]

	// Validate format
	if err := ValidateFormat(listFormat); err != nil {
		return err
	}

	// Validate regex
	if err := ValidateRegex(pattern, "pattern"); err != nil {
		return err
	}

	// Validate ignore-function pattern (function-name regex exclusion)
	if err := ValidateRegex(listIgnoreFunction, "ignore-function"); err != nil {
		return err
	}

	// Load profile
	loader := profile.NewLoader()
	p, err := loader.LoadFromFile(profilePath)
	if err != nil {
		return fmt.Errorf("failed to load profile: %w", err)
	}

	// Apply pprof label filter (--tag / --tag-ignore) before querying.
	labelFilter, err := profile.NewLabelFilter(listTag, listIgnoreTag)
	if err != nil {
		return err
	}
	p, err = labelFilter.Apply(p)
	if err != nil {
		return err
	}

	// Configure list analysis
	config := profile.ListConfig{
		Pattern:     pattern,
		Depth:       listDepth,
		CallersOnly: listCallersOnly,
		CalleesOnly: listCalleesOnly,
	}

	if listIgnoreFunction != "" {
		config.ExcludePattern = listIgnoreFunction
	}

	// Perform list analysis
	result, err := profile.List(p, config)
	if err != nil {
		return fmt.Errorf("failed to query functions: %w", err)
	}

	out := cmd.OutOrStdout()

	// Check if any functions matched
	if len(result.MatchedFunctions) == 0 {
		fmt.Fprintf(out, "No functions matched pattern: %s\n", pattern)
		return nil
	}

	// Generate output based on format
	switch listFormat {
	case "json":
		return outputListJSON(result, out)
	case "markdown":
		return outputListMarkdown(result, out)
	default:
		return outputListText(result, out)
	}
}

// outputListText outputs list result in text format
func outputListText(result *profile.ListResult, w io.Writer) error {
	formatter := output.NewListTextFormatter(w)
	return formatter.FormatListResult(result)
}

// outputListJSON outputs list result in JSON format
func outputListJSON(result *profile.ListResult, w io.Writer) error {
	formatter := output.NewListJSONFormatter(w)
	return formatter.FormatListResult(result)
}

// outputListMarkdown outputs list result in Markdown format
func outputListMarkdown(result *profile.ListResult, w io.Writer) error {
	formatter := output.NewListMarkdownFormatter(w)
	return formatter.FormatListResult(result)
}
