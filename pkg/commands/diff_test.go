package commands

import (
	"bytes"
	"strings"
	"testing"
)

func TestDiffCommand_InvalidFormat(t *testing.T) {
	var buf bytes.Buffer
	DiffCmd.SetOut(&buf)
	DiffCmd.SetErr(&buf)
	DiffCmd.SetArgs([]string{"../../testdata/cpu.pb.gz", "../../testdata/cpu.pb.gz", "--format", "yaml"})

	err := DiffCmd.Execute()
	if err == nil {
		t.Fatal("expected error for invalid format")
	}
	if !strings.Contains(err.Error(), "invalid format") {
		t.Errorf("error should mention 'invalid format', got: %v", err)
	}
}

func TestDiffCommand_MissingFile(t *testing.T) {
	var buf bytes.Buffer
	DiffCmd.SetOut(&buf)
	DiffCmd.SetErr(&buf)
	DiffCmd.SetArgs([]string{"/nonexistent/a.pb.gz", "/nonexistent/b.pb.gz"})

	err := DiffCmd.Execute()
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestDiffCommand_NormalRun(t *testing.T) {
	var buf bytes.Buffer
	DiffCmd.SetOut(&buf)
	DiffCmd.SetErr(&buf)
	DiffCmd.SetArgs([]string{"../../testdata/cpu.pb.gz", "../../testdata/cpu.pb.gz", "--format", "text"})

	err := DiffCmd.Execute()
	if err != nil {
		t.Fatalf("diff with same file should not error, got: %v", err)
	}
}
