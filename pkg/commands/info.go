package commands

import (
	"fmt"
	"os"

	"github.com/chenquan/agent-insight/pkg/output"
	"github.com/chenquan/agent-insight/pkg/profile"

	"github.com/spf13/cobra"
)

var infoFormat string

var InfoCmd = &cobra.Command{
	Use:   "info <profile.pb.gz> [flags]",
	Short: "Show profile metadata overview",
	Long: `Display profile metadata without performing sample-level analysis.

Shows: profile type, duration, sample count, value types, symbol status,
mapping information, time range, and comments.

Example usage:
  agent-insight info profile.pb.gz
  agent-insight info cpu.pb.gz --format json
  agent-insight info heap.prof --format markdown`,
	Args: cobra.ExactArgs(1),
	RunE: runInfo,
}

func init() {
	InfoCmd.Flags().StringVar(&infoFormat, "format", "text", "Output format: text, json, markdown")
}

func runInfo(cmd *cobra.Command, args []string) error {
	profilePath := args[0]

	if infoFormat != "text" && infoFormat != "json" && infoFormat != "markdown" {
		return fmt.Errorf("invalid format: %s (must be text, json, or markdown)", infoFormat)
	}

	loader := profile.NewLoader()
	p, err := loader.LoadFromFile(profilePath)
	if err != nil {
		return fmt.Errorf("failed to load profile: %w", err)
	}

	result, err := profile.Info(p)
	if err != nil {
		return fmt.Errorf("failed to read profile info: %w", err)
	}

	switch infoFormat {
	case "json":
		formatter := output.NewInfoJSONFormatter(os.Stdout)
		return formatter.FormatInfoResult(result)
	case "markdown":
		formatter := output.NewInfoMarkdownFormatter(os.Stdout)
		return formatter.FormatInfoResult(result)
	default:
		formatter := output.NewInfoTextFormatter(os.Stdout)
		return formatter.FormatInfoResult(result)
	}
}
