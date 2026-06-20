package commands

import (
	"fmt"
	"io"

	"github.com/chenquan/agent-insight/pkg/output"
	"github.com/chenquan/agent-insight/pkg/profile"

	"github.com/spf13/cobra"
)

// DiffCmd represents the diff command
var DiffCmd = &cobra.Command{
	Use:   "diff <base.prof> <target.prof> [flags]",
	Short: "Compare two profiles to identify performance changes",
	Long: `Compare two pprof profile files to identify performance regressions and improvements.

Supports:
- Identifying top regressions and improvements
- Detecting new and deleted functions
- Filtering by minimum change threshold
- Multiple output formats

Example usage:
  agent-insight diff before.prof after.prof
  agent-insight diff before.prof after.prof --min-delta 10
  agent-insight diff base.prof target.prof --focus "runtime.*" --format json`,
	Args: cobra.ExactArgs(2),
	RunE: runDiff,
}

// Diff flags
var (
	diffMinDelta    float64
	diffFocus       string
	diffIgnore      string
	diffFormat      string
	diffTop         int
	diffHideNew     bool
	diffHideDeleted bool
	diffTag         []string
	diffIgnoreTag   []string
)

func init() {
	// Add flags
	DiffCmd.Flags().Float64Var(&diffMinDelta, "min-delta", 0, "Minimum percentage change to include (0 = all)")
	DiffCmd.Flags().StringVar(&diffFocus, "focus", "", "Regex pattern to focus on specific functions")
	DiffCmd.Flags().StringVar(&diffIgnore, "ignore", "", "Regex pattern to ignore specific functions")
	DiffCmd.Flags().StringVar(&diffFormat, "format", "text", "Output format: text, json, markdown")
	DiffCmd.Flags().IntVar(&diffTop, "top", 15, "Limit to top N in each category")
	DiffCmd.Flags().BoolVar(&diffHideNew, "hide-new", false, "Hide new functions")
	DiffCmd.Flags().BoolVar(&diffHideDeleted, "hide-deleted", false, "Hide deleted functions")
	DiffCmd.Flags().StringSliceVar(&diffTag, "tag", nil, "Filter samples by pprof label key=value (repeatable; same key OR, across keys AND)")
	DiffCmd.Flags().StringSliceVar(&diffIgnoreTag, "tag-ignore", nil, "Exclude samples by pprof label key=value (same semantics as --tag)")
}

func runDiff(cmd *cobra.Command, args []string) error {
	basePath := args[0]
	targetPath := args[1]

	// Validate format
	if err := ValidateFormat(diffFormat); err != nil {
		return err
	}

	// Validate regex patterns
	if err := ValidateRegex(diffFocus, "focus"); err != nil {
		return err
	}

	if err := ValidateRegex(diffIgnore, "ignore"); err != nil {
		return err
	}

	// Load profiles
	loader := profile.NewLoader()

	base, err := loader.LoadFromFile(basePath)
	if err != nil {
		return fmt.Errorf("failed to load base profile: %w", err)
	}

	target, err := loader.LoadFromFile(targetPath)
	if err != nil {
		return fmt.Errorf("failed to load target profile: %w", err)
	}

	// Apply pprof label filter (--tag / --tag-ignore) to both profiles before
	// diffing. base is filtered first; if it matches 0 samples we error out
	// before touching target.
	labelFilter, err := profile.NewLabelFilter(diffTag, diffIgnoreTag)
	if err != nil {
		return err
	}
	base, err = labelFilter.Apply(base)
	if err != nil {
		return err
	}
	target, err = labelFilter.Apply(target)
	if err != nil {
		return err
	}

	// Configure diff
	config := profile.DiffConfig{
		MinDelta:      diffMinDelta,
		TopN:          diffTop,
		FocusPattern:  diffFocus,
		IgnorePattern: diffIgnore,
	}

	// Perform diff
	result, err := profile.Diff(base, target, config)
	if err != nil {
		return fmt.Errorf("failed to compare profiles: %w", err)
	}

	// Apply hide flags
	if diffHideNew {
		result.NewFunctions = nil
	}
	if diffHideDeleted {
		result.DeletedFunctions = nil
	}

	// Generate output based on format
	out := cmd.OutOrStdout()
	switch diffFormat {
	case "json":
		return outputDiffJSON(result, basePath, targetPath, out)
	case "markdown":
		return outputDiffMarkdown(result, basePath, targetPath, out)
	default:
		return outputDiffText(result, basePath, targetPath, out)
	}
}

// outputDiffText outputs diff result in text format
func outputDiffText(result *profile.DiffResult, base, target string, w io.Writer) error {
	formatter := output.NewDiffTextFormatter(w)
	return formatter.FormatDiffResult(result, base, target)
}

// outputDiffJSON outputs diff result in JSON format
func outputDiffJSON(result *profile.DiffResult, base, target string, w io.Writer) error {
	formatter := output.NewDiffJSONFormatter(w)
	return formatter.FormatDiffResult(result, base, target)
}

// outputDiffMarkdown outputs diff result in Markdown format
func outputDiffMarkdown(result *profile.DiffResult, base, target string, w io.Writer) error {
	formatter := output.NewDiffMarkdownFormatter(w)
	return formatter.FormatDiffResult(result, base, target)
}
