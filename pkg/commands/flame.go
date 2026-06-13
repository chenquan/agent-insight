package commands

import (
	"fmt"
	"os"
	"regexp"

	"github.com/chenquan/agent-insight/pkg/output"
	"github.com/chenquan/agent-insight/pkg/profile"

	"github.com/spf13/cobra"
)

// FlameCmd represents the flame command
var FlameCmd = &cobra.Command{
	Use:   "flame <profile.pb.gz> [flags]",
	Short: "Generate folded stack format for flame graphs",
	Long: `Convert pprof profile data into collapsed stack format compatible with flame graph visualization tools.

Output format: "func1;func2;func3 count" where count is the sample value.
This can be piped directly to flame graph generation tools like flamegraph.pl.

Example usage:
  agent-insight flame profile.pb.gz > stacks.folded
  agent-insight flame profile.pb.gz --focus "encoding.*" | flamegraph.pl > graph.svg
  agent-insight flame heap.prof --ignore "runtime.*" --depth 10`,
	Args: cobra.ExactArgs(1),
	RunE: runFlame,
}

// Flame flags
var (
	flameFocus     string
	flameIgnore    string
	flameDepth     int
	flameTop       int
	flameValueType string
	flameStats     bool
	flameFormat    string
)

func init() {
	// Add flags
	FlameCmd.Flags().StringVar(&flameFocus, "focus", "", "Regex pattern to focus on specific functions")
	FlameCmd.Flags().StringVar(&flameIgnore, "ignore", "", "Regex pattern to ignore specific functions")
	FlameCmd.Flags().IntVar(&flameDepth, "depth", 0, "Maximum depth of stack traces (0 = unlimited)")
	FlameCmd.Flags().IntVar(&flameTop, "top", 0, "Limit to top N stacks (0 = unlimited)")
	FlameCmd.Flags().StringVar(&flameValueType, "value-type", "", "Specify which value type to use")
	FlameCmd.Flags().BoolVar(&flameStats, "stats", false, "Include statistics in output")
	FlameCmd.Flags().StringVar(&flameFormat, "format", "text", "Output format: text, json, markdown")
}

func runFlame(cmd *cobra.Command, args []string) error {
	profilePath := args[0]

	// Validate format
	if flameFormat != "text" && flameFormat != "json" && flameFormat != "markdown" {
		return fmt.Errorf("invalid format: %s (must be text, json, or markdown)", flameFormat)
	}

	// Validate patterns if provided
	if flameFocus != "" {
		if _, err := regexp.Compile(flameFocus); err != nil {
			return fmt.Errorf("invalid focus pattern: %w", err)
		}
	}

	if flameIgnore != "" {
		if _, err := regexp.Compile(flameIgnore); err != nil {
			return fmt.Errorf("invalid ignore pattern: %w", err)
		}
	}

	// Load profile
	loader := profile.NewLoader()
	p, err := loader.LoadFromFile(profilePath)
	if err != nil {
		return fmt.Errorf("failed to load profile: %w", err)
	}

	// Configure flame generation
	config := profile.FlameConfig{
		FocusPattern:  flameFocus,
		IgnorePattern: flameIgnore,
		Depth:         flameDepth,
		TopN:          flameTop,
	}

	// Perform flame generation
	result, err := profile.Flame(p, config)
	if err != nil {
		return fmt.Errorf("failed to generate folded stacks: %w", err)
	}

	// Output statistics if requested
	if flameStats && flameFormat == "text" {
		outputFlameStats(result)
	}

	// Output based on format
	formatter := output.NewFlameFormatter(os.Stdout)
	return formatter.Format(result, flameFormat)
}

// outputFlameStats outputs statistics about the flame generation
func outputFlameStats(result *profile.FlameResult) {
	fmt.Fprintf(os.Stderr, "Flame Graph Statistics\n")
	fmt.Fprintf(os.Stderr, "=====================\n")
	fmt.Fprintf(os.Stderr, "Total stacks: %d\n", result.TotalStacks)
	if result.FilteredStacks > 0 {
		fmt.Fprintf(os.Stderr, "Filtered stacks: %d\n", result.FilteredStacks)
	}
	fmt.Fprintf(os.Stderr, "Unique stacks: %d\n", result.UniqueStacks)
	fmt.Fprintf(os.Stderr, "Output stacks: %d\n\n", len(result.Stacks))

	if result.FlameConfig.FocusPattern != "" {
		fmt.Fprintf(os.Stderr, "Focus pattern: %s\n", result.FlameConfig.FocusPattern)
	}
	if result.FlameConfig.IgnorePattern != "" {
		fmt.Fprintf(os.Stderr, "Ignore pattern: %s\n", result.FlameConfig.IgnorePattern)
	}
	if result.FlameConfig.Depth > 0 {
		fmt.Fprintf(os.Stderr, "Depth limit: %d\n", result.FlameConfig.Depth)
	}
	if result.FlameConfig.TopN > 0 {
		fmt.Fprintf(os.Stderr, "Top N limit: %d\n", result.FlameConfig.TopN)
	}
	fmt.Fprintf(os.Stderr, "Value type: %s\n\n", result.FlameConfig.ValueType.Name+"/"+result.FlameConfig.ValueType.Unit)
}
