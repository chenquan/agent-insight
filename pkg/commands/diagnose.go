package commands

import (
	"fmt"
	"os"

	"github.com/chenquan/agent-insight/pkg/output"
	"github.com/chenquan/agent-insight/pkg/profile"

	"github.com/spf13/cobra"
)

var DiagnoseCmd = &cobra.Command{
	Use:   "diagnose <profile.pb.gz> [flags]",
	Short: "Generate AI diagnostic prompt from pprof profile",
	Long: `Analyze a pprof profile and generate a structured diagnostic prompt
for AI coding assistants (e.g., Claude Code).

The diagnose command extracts profile data (hotspots, call tree, traces),
detects the programming language, and assembles a diagnostic prompt with
language-specific and profile-type-specific guidance.

Example usage:
  agent-insight diagnose cpu.pb.gz
  agent-insight diagnose cpu.pb.gz --context "HTTP microservice"
  agent-insight diagnose heap.prof --top 5 --format json`,
	Args: cobra.ExactArgs(1),
	RunE: runDiagnose,
}

var (
	diagnoseTop    int
	diagnoseContext string
	diagnoseFormat string
)

func init() {
	DiagnoseCmd.Flags().IntVar(&diagnoseTop, "top", 10, "Number of top hotspots to include")
	DiagnoseCmd.Flags().StringVar(&diagnoseContext, "context", "", "User-provided application context to embed in the prompt")
	DiagnoseCmd.Flags().StringVar(&diagnoseFormat, "format", "text", "Output format: text, json, markdown")
}

func runDiagnose(cmd *cobra.Command, args []string) error {
	profilePath := args[0]

	if diagnoseFormat != "text" && diagnoseFormat != "json" && diagnoseFormat != "markdown" {
		return fmt.Errorf("invalid format: %s (must be text, json, or markdown)", diagnoseFormat)
	}

	loader := profile.NewLoader()
	p, err := loader.LoadFromFile(profilePath)
	if err != nil {
		return fmt.Errorf("failed to load profile: %w", err)
	}

	dp, err := profile.BuildDiagnosePrompt(p, diagnoseTop, diagnoseContext)
	if err != nil {
		return fmt.Errorf("failed to build diagnose prompt: %w", err)
	}

	switch diagnoseFormat {
	case "json":
		f := output.NewDiagnoseJSONFormatter(os.Stdout)
		return f.FormatDiagnose(dp)
	case "markdown":
		f := output.NewDiagnoseMarkdownFormatter(os.Stdout)
		return f.FormatDiagnose(dp)
	default:
		f := output.NewDiagnoseTextFormatter(os.Stdout)
		return f.FormatDiagnose(dp)
	}
}
