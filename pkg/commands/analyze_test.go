package commands

import (
	"bytes"
	"strings"
	"testing"
)

// resetAnalyzeFlags restores the analyze command's package-level flag
// variables to their defaults. Cobra's SetArgs+Execute does not reset flags
// that a prior test set, so without this each test inherits the previous
// test's flag values (e.g. --focus "[invalid" leaking into later tests).
func resetAnalyzeFlags() {
	analyzeTop = 15
	analyzeCum = false
	analyzeFocus = ""
	analyzeIgnore = ""
	analyzeFormat = "text"
	analyzeCallDepth = 5
	analyzeCollapse = false
	analyzeValueType = ""
	analyzeTag = nil
	analyzeTagIgnore = nil
	analyzeBreakdownOn = ""
	analyzeBreakdownTop = 20
}

func TestAnalyzeCommand_InvalidFormat(t *testing.T) {
	resetAnalyzeFlags()
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
	resetAnalyzeFlags()
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
	resetAnalyzeFlags()
	var buf bytes.Buffer
	AnalyzeCmd.SetOut(&buf)
	AnalyzeCmd.SetErr(&buf)
	AnalyzeCmd.SetArgs([]string{"/nonexistent/profile.pb.gz"})

	err := AnalyzeCmd.Execute()
	if err == nil {
		t.Fatal("expected error for nonexistent file")
	}
}

func TestAnalyzeCommand_TagFilter(t *testing.T) {
	resetAnalyzeFlags()
	var buf bytes.Buffer
	AnalyzeCmd.SetOut(&buf)
	AnalyzeCmd.SetErr(&buf)
	AnalyzeCmd.SetArgs([]string{"../../testdata/goroutine.pb.gz", "--format", "json", "--tag", "state=blocked"})

	if err := AnalyzeCmd.Execute(); err != nil {
		t.Fatalf("AnalyzeCmd failed: %v", err)
	}
	// 8 samples total, 4 are state=blocked -> filtered to 4.
	if !strings.Contains(buf.String(), `"samples": 4`) {
		t.Errorf("expected samples filtered to 4, got: %s", buf.String())
	}
}

func TestAnalyzeCommand_TagFilterZeroSamples(t *testing.T) {
	resetAnalyzeFlags()
	var buf bytes.Buffer
	AnalyzeCmd.SetOut(&buf)
	AnalyzeCmd.SetErr(&buf)
	AnalyzeCmd.SetArgs([]string{"../../testdata/cpu.pb.gz", "--format", "text", "--tag", "state=blocked"})

	err := AnalyzeCmd.Execute()
	if err == nil {
		t.Fatal("expected error when tag filter matches 0 samples")
	}
	if !strings.Contains(err.Error(), "matched 0 of 5 samples") {
		t.Errorf("expected 'matched 0 of 5 samples' in error, got: %v", err)
	}
}

func TestAnalyzeCommand_Breakdown(t *testing.T) {
	resetAnalyzeFlags()
	var buf bytes.Buffer
	AnalyzeCmd.SetOut(&buf)
	AnalyzeCmd.SetErr(&buf)
	AnalyzeCmd.SetArgs([]string{
		"../../testdata/goroutine.pb.gz", "--format", "json",
		"--tag-breakdown-on", "state", "--tag-breakdown-top", "1",
	})

	if err := AnalyzeCmd.Execute(); err != nil {
		t.Fatalf("AnalyzeCmd failed: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "label_breakdown") {
		t.Errorf("expected label_breakdown in output, got: %s", out)
	}
	if !strings.Contains(out, `"key": "state"`) {
		t.Errorf("expected state breakdown key, got: %s", out)
	}
}

func TestAnalyzeCommand_NoBreakdownByDefault(t *testing.T) {
	resetAnalyzeFlags()
	var buf bytes.Buffer
	AnalyzeCmd.SetOut(&buf)
	AnalyzeCmd.SetErr(&buf)
	AnalyzeCmd.SetArgs([]string{"../../testdata/goroutine.pb.gz", "--format", "json"})

	if err := AnalyzeCmd.Execute(); err != nil {
		t.Fatalf("AnalyzeCmd failed: %v", err)
	}
	if strings.Contains(buf.String(), "label_breakdown") {
		t.Errorf("expected no label_breakdown by default, got: %s", buf.String())
	}
}
