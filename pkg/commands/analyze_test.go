package commands

import (
	"bytes"
	"strings"
	"testing"
)

func TestAnalyzeCommand_InvalidFormat(t *testing.T) {
	var buf bytes.Buffer
	AnalyzeCmd.SetOut(&buf)
	AnalyzeCmd.SetErr(&buf)
	AnalyzeCmd.SetArgs([]string{"../../testdata/cpu.pb.gz", "--format", "yaml"})

	err := AnalyzeCmd.Execute()
	if err == nil {
		t.Fatal("expected error for invalid format")
	}
	if !strings.Contains(err.Error(), "invalid format") {
		t.Errorf("error should mention 'invalid format', got: %v", err)
	}
}

func TestAnalyzeCommand_InvalidFocusPattern(t *testing.T) {
	var buf bytes.Buffer
	AnalyzeCmd.SetOut(&buf)
	AnalyzeCmd.SetErr(&buf)
	AnalyzeCmd.SetArgs([]string{"../../testdata/cpu.pb.gz", "--format", "text", "--focus", "[invalid"})

	err := AnalyzeCmd.Execute()
	if err == nil {
		t.Fatal("expected error for invalid focus pattern")
	}
	if !strings.Contains(err.Error(), "invalid focus pattern") {
		t.Errorf("error should mention 'invalid focus pattern', got: %v", err)
	}
}

func TestAnalyzeCommand_FileNotFound(t *testing.T) {
	var buf bytes.Buffer
	AnalyzeCmd.SetOut(&buf)
	AnalyzeCmd.SetErr(&buf)
	AnalyzeCmd.SetArgs([]string{"/nonexistent/profile.pb.gz"})

	err := AnalyzeCmd.Execute()
	if err == nil {
		t.Fatal("expected error for nonexistent file")
	}
}
