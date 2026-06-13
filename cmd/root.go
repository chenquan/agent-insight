/*
Copyright © 2026 chenquan
*/
package cmd

import (
	"os"

	"github.com/chenquan/agent-insight/pkg/commands"

	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "agent-insight",
	Short: "A lightweight pprof analysis CLI for AI coding assistants",
	Long: `A lightweight CLI tool designed specifically for Claude Code and other AI coding assistants.
Analyzes pprof performance profiles and outputs structured, AI-friendly results.

Core commands:
  analyze  - Analyze pprof files and output performance hotspots
  list     - Query specific function call relationships
  flame    - Generate folded stack format for flame graphs
  diff     - Compare two profiles to identify performance changes
  merge    - Merge multiple profiles of the same type
  info     - Show profile metadata overview
  traces   - Show individual sample call traces
  tree     - Show hierarchical call tree`,
	Version: "0.1.0",
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Add subcommands
	rootCmd.AddCommand(commands.AnalyzeCmd)
	rootCmd.AddCommand(commands.ListCmd)
	rootCmd.AddCommand(commands.FlameCmd)
	rootCmd.AddCommand(commands.DiffCmd)
	rootCmd.AddCommand(commands.InfoCmd)
	rootCmd.AddCommand(commands.TracesCmd)
	rootCmd.AddCommand(commands.TreeCmd)
	rootCmd.AddCommand(commands.MergeCmd)
	rootCmd.AddCommand(commands.InitCmd)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.agent-insight.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
