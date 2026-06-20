package commands

import (
	"fmt"

	"github.com/chenquan/agent-insight/pkg/output"
	"github.com/chenquan/agent-insight/pkg/profile"

	"github.com/spf13/cobra"
)

var (
	tagsFormat string
	tagsTop    int
)

var TagsCmd = &cobra.Command{
	Use:   "tags <profile.pb.gz> [flags]",
	Short: "List pprof labels and their value distribution",
	Long: `List all pprof labels (Sample.Label) in a profile along with the value
distribution and sample counts.

This is the discovery layer for label-based analysis: run it before filtering
with --tag on analyze/list/traces/diff to see which labels exist and how
samples are distributed across their values.

Example usage:
  agent-insight tags goroutine.pb.gz
  agent-insight tags service.pb.gz --top 20 --format json
  agent-insight tags cpu.pb.gz --format markdown`,
	Args: cobra.ExactArgs(1),
	RunE: runTags,
}

func init() {
	TagsCmd.Flags().StringVar(&tagsFormat, "format", "text", "Output format: text, json, markdown")
	TagsCmd.Flags().IntVar(&tagsTop, "top", 50, "Max values shown per numeric label (string labels are shown in full)")
}

func runTags(cmd *cobra.Command, args []string) error {
	profilePath := args[0]

	if err := ValidateFormat(tagsFormat); err != nil {
		return err
	}
	if err := ValidatePositiveInt(tagsTop, "top"); err != nil {
		return err
	}

	loader := profile.NewLoader()
	p, err := loader.LoadFromFile(profilePath)
	if err != nil {
		return fmt.Errorf("failed to load profile: %w", err)
	}

	result, err := profile.Tags(p, profilePath, tagsTop)
	if err != nil {
		return fmt.Errorf("failed to read tags: %w", err)
	}

	out := cmd.OutOrStdout()
	switch tagsFormat {
	case "json":
		return output.NewTagsJSONFormatter(out).FormatTagsResult(result)
	case "markdown":
		return output.NewTagsMarkdownFormatter(out).FormatTagsResult(result)
	default:
		return output.NewTagsTextFormatter(out).FormatTagsResult(result)
	}
}
