package commands

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/chenquan/agent-insight/pkg/skill"
)

func TestInitCommand_Generate(t *testing.T) {
	tmpDir := t.TempDir()

	var buf bytes.Buffer
	InitCmd.SetOut(&buf)
	InitCmd.SetArgs([]string{})

	// Run from tmpDir
	oldDir, _ := os.Getwd()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chdir(oldDir) }()

	err := InitCmd.Execute()
	if err != nil {
		t.Fatalf("init command error = %v", err)
	}

	expected := filepath.Join(tmpDir, skill.SkillDir, skill.SkillFile)
	if _, err := os.Stat(expected); os.IsNotExist(err) {
		t.Error("skill file was not generated")
	}

	output := buf.String()
	if !contains(output, "Generated") {
		t.Errorf("output should contain 'Generated', got: %s", output)
	}
}

func TestInitCommand_FileExistsWithoutForce(t *testing.T) {
	tmpDir := t.TempDir()

	oldDir, _ := os.Getwd()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chdir(oldDir) }()

	// First generation
	_, err := skill.Generate(".")
	if err != nil {
		t.Fatalf("first generate error = %v", err)
	}

	// Second without force should fail
	InitCmd.SetArgs([]string{})
	err = InitCmd.Execute()
	if err == nil {
		t.Fatal("expected error when file exists without --force")
	}

	if !contains(err.Error(), "already exists") {
		t.Errorf("error should mention 'already exists', got: %v", err)
	}
}

func TestInitCommand_FileExistsWithForce(t *testing.T) {
	tmpDir := t.TempDir()

	oldDir, _ := os.Getwd()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chdir(oldDir) }()

	// First generation
	_, err := skill.Generate(".")
	if err != nil {
		t.Fatalf("first generate error = %v", err)
	}

	// Second with force should succeed
	var buf bytes.Buffer
	InitCmd.SetOut(&buf)
	InitCmd.SetArgs([]string{"--force"})

	err = InitCmd.Execute()
	if err != nil {
		t.Fatalf("init --force error = %v", err)
	}

	output := buf.String()
	if !contains(output, "Overwritten") {
		t.Errorf("output should contain 'Overwritten', got: %s", output)
	}
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && searchString(s, sub)
}

func searchString(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
