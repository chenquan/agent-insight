package commands

import (
	"fmt"

	"github.com/chenquan/agent-insight/pkg/skill"

	"github.com/spf13/cobra"
)

var initForce bool

var InitCmd = &cobra.Command{
	Use:   "init [flags]",
	Short: "Generate Claude Code skill file for agent-insight",
	Long: `Generate a SKILL.md file in .claude/skills/agent-insight/ that teaches
Claude Code when and how to use agent-insight for pprof analysis.

After running this command, Claude Code will automatically recognize
performance analysis scenarios and use agent-insight appropriately.`,
	Args: cobra.NoArgs,
	RunE: runInit,
}

func init() {
	InitCmd.Flags().BoolVarP(&initForce, "force", "f", false, "Overwrite existing SKILL.md without prompt")
}

func runInit(cmd *cobra.Command, _ []string) error {
	targetDir := "."
	existed := skill.Exists(targetDir)

	if existed && !initForce {
		return fmt.Errorf("skill file already exists at %s/%s. Use --force to overwrite", skill.SkillDir, skill.SkillFile)
	}

	path, err := skill.Generate(targetDir)
	if err != nil {
		return err
	}

	if existed {
		fmt.Fprintln(cmd.OutOrStdout(), "Overwritten skill file: "+path)
	} else {
		fmt.Fprintln(cmd.OutOrStdout(), "Generated skill file: "+path)
	}
	return nil
}
