package skill

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGenerate(t *testing.T) {
	tmpDir := t.TempDir()

	path, err := Generate(tmpDir)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	expected := filepath.Join(tmpDir, SkillDir, SkillFile)
	if path != expected {
		t.Errorf("Generate() path = %q, want %q", path, expected)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read generated file: %v", err)
	}

	if len(data) == 0 {
		t.Error("generated file is empty")
	}

	if string(data) != string(skillTemplate) {
		t.Error("generated file content does not match template")
	}
}

func TestGenerateCreatesDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	// Directory should not exist initially
	dir := filepath.Join(tmpDir, SkillDir)
	if _, err := os.Stat(dir); !os.IsNotExist(err) {
		t.Fatalf("directory %s should not exist initially", dir)
	}

	_, err := Generate(tmpDir)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	if _, err := os.Stat(dir); os.IsNotExist(err) {
		t.Error("directory was not created")
	}
}

func TestExists(t *testing.T) {
	tmpDir := t.TempDir()

	if Exists(tmpDir) {
		t.Error("Exists() = true before generation")
	}

	_, err := Generate(tmpDir)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	if !Exists(tmpDir) {
		t.Error("Exists() = false after generation")
	}
}

func TestGenerateOverwrite(t *testing.T) {
	tmpDir := t.TempDir()

	// First generation
	_, err := Generate(tmpDir)
	if err != nil {
		t.Fatalf("first Generate() error = %v", err)
	}

	// Overwrite should succeed
	path, err := Generate(tmpDir)
	if err != nil {
		t.Fatalf("second Generate() error = %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}

	if string(data) != string(skillTemplate) {
		t.Error("overwritten file content does not match template")
	}
}
