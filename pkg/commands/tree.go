package commands

import (
	"fmt"
	"os"

	"github.com/chenquan/agent-insight/pkg/output"
	"github.com/chenquan/agent-insight/pkg/profile"

	"github.com/spf13/cobra"
)

var (
	treeFocus     string
	treeIgnore    string
	treeDepth     int
	treeTop       int
	treeCum       bool
	treeValueType string
	treeFormat    string
)

var TreeCmd = &cobra.Command{
	Use:   "tree <profile.pb.gz> [flags]",
	Short: "Show hierarchical call tree",
	Long: `Build and display a hierarchical call tree from profile samples.

Shows the global call structure from root to leaf, with flat and cumulative
values at each level. Children are sorted by cumulative value by default.

Example usage:
  agent-insight tree profile.pb.gz
  agent-insight tree cpu.pb.gz --depth 3 --top 5
  agent-insight tree heap.pb.gz --focus "main.*" --format json`,
	Args: cobra.ExactArgs(1),
	RunE: runTree,
}

func init() {
	TreeCmd.Flags().StringVar(&treeFocus, "focus", "", "Regex pattern to focus on specific functions")
	TreeCmd.Flags().StringVar(&treeIgnore, "ignore", "", "Regex pattern to ignore specific functions")
	TreeCmd.Flags().IntVar(&treeDepth, "depth", 5, "Maximum tree depth to display")
	TreeCmd.Flags().IntVar(&treeTop, "top", 10, "Max children per node")
	TreeCmd.Flags().BoolVar(&treeCum, "cum", true, "Sort by cumulative value (default true)")
	TreeCmd.Flags().StringVar(&treeValueType, "value-type", "", "Specify which value type to use")
	TreeCmd.Flags().StringVar(&treeFormat, "format", "text", "Output format: text, json, markdown")
}

func runTree(cmd *cobra.Command, args []string) error {
	profilePath := args[0]

	if err := ValidateFormat(treeFormat); err != nil {
		return err
	}

	if err := ValidateRegex(treeFocus, "focus"); err != nil {
		return err
	}

	if err := ValidateRegex(treeIgnore, "ignore"); err != nil {
		return err
	}

	loader := profile.NewLoader()
	p, err := loader.LoadFromFile(profilePath)
	if err != nil {
		return fmt.Errorf("failed to load profile: %w", err)
	}

	config := profile.TreeConfig{
		FocusPattern:  treeFocus,
		IgnorePattern: treeIgnore,
		Depth:         treeDepth,
		TopN:          treeTop,
		SortByCum:     treeCum,
	}

	result, err := profile.Tree(p, config)
	if err != nil {
		return fmt.Errorf("failed to build call tree: %w", err)
	}

	switch treeFormat {
	case "json":
		formatter := output.NewTreeJSONFormatter(os.Stdout)
		return formatter.FormatTreeResult(result)
	case "markdown":
		formatter := output.NewTreeMarkdownFormatter(os.Stdout)
		return formatter.FormatTreeResult(result)
	default:
		formatter := output.NewTreeTextFormatter(os.Stdout)
		return formatter.FormatTreeResult(result)
	}
}
