package commands

import (
	"fmt"
	"os"

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
	listDepth       int
	listCallersOnly bool
	listCalleesOnly bool
	listExclude     string
	listFormat      string
)

func init() {
	// Add flags
	ListCmd.Flags().IntVar(&listDepth, "depth", 5, "Maximum depth of caller/callee relationships to show")
	ListCmd.Flags().BoolVar(&listCallersOnly, "callers-only", false, "Show only callers, exclude callees")
	ListCmd.Flags().BoolVar(&listCalleesOnly, "callees-only", false, "Show only callees, exclude callers")
	ListCmd.Flags().StringVar(&listExclude, "exclude", "", "Regex pattern to exclude from results")
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

	// Validate exclude pattern
	if err := ValidateRegex(listExclude, "exclude"); err != nil {
		return err
	}

	// Load profile
	loader := profile.NewLoader()
	p, err := loader.LoadFromFile(profilePath)
	if err != nil {
		return fmt.Errorf("failed to load profile: %w", err)
	}

	// Configure list analysis
	config := profile.ListConfig{
		Pattern:     pattern,
		Depth:       listDepth,
		CallersOnly: listCallersOnly,
		CalleesOnly: listCalleesOnly,
	}

	if listExclude != "" {
		config.ExcludePattern = listExclude
	}

	// Perform list analysis
	result, err := profile.List(p, config)
	if err != nil {
		return fmt.Errorf("failed to query functions: %w", err)
	}

	// Check if any functions matched
	if len(result.MatchedFunctions) == 0 {
		fmt.Printf("No functions matched pattern: %s\n", pattern)
		return nil
	}

	// Generate output based on format
	switch listFormat {
	case "json":
		return outputListJSON(result)
	case "markdown":
		return outputListMarkdown(result)
	default:
		return outputListText(result)
	}
}

// outputListText outputs list result in text format
func outputListText(result *profile.ListResult) error {
	formatter := output.NewListTextFormatter(os.Stdout)
	return formatter.FormatListResult(result)
}

// outputListJSON outputs list result in JSON format
func outputListJSON(result *profile.ListResult) error {
	formatter := output.NewListJSONFormatter(os.Stdout)
	return formatter.FormatListResult(result)
}

// outputListMarkdown outputs list result in Markdown format
func outputListMarkdown(result *profile.ListResult) error {
	formatter := output.NewListMarkdownFormatter(os.Stdout)
	return formatter.FormatListResult(result)
}
