package skill

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
)

//go:embed template.md
var skillTemplate []byte

const (
	SkillDir  = ".claude/skills/agent-insight"
	SkillFile = "SKILL.md"
)

// Generate writes the SKILL.md file to the target directory.
// Returns the full path of the generated file.
func Generate(targetDir string) (string, error) {
	dir := filepath.Join(targetDir, SkillDir)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	path := filepath.Join(dir, SkillFile)
	if err := os.WriteFile(path, skillTemplate, 0o644); err != nil {
		return "", fmt.Errorf("failed to write %s: %w", path, err)
	}

	return path, nil
}

// Exists checks if the SKILL.md file already exists.
func Exists(targetDir string) bool {
	path := filepath.Join(targetDir, SkillDir, SkillFile)
	_, err := os.Stat(path)
	return err == nil
}
