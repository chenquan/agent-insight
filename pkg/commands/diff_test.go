package commands

import (
	"bytes"
	"strings"
	"testing"
)

// resetDiffFlags restores the diff command's package-level flag variables to
// their defaults. Cobra's SetArgs+Execute does not reset flags set by a prior
// test, so without this each test inherits the previous test's values.
func resetDiffFlags() {
	diffMinDelta = 0
	diffFocus = ""
	diffIgnore = ""
	diffFormat = "text"
	diffTop = 15
	diffHideNew = false
	diffHideDeleted = false
	diffTag = nil
	diffIgnoreTag = nil
}

func TestDiffCommand_InvalidFormat(t *testing.T) {
	resetDiffFlags()
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
	resetDiffFlags()
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
	resetDiffFlags()
	var buf bytes.Buffer
	DiffCmd.SetOut(&buf)
	DiffCmd.SetErr(&buf)
	DiffCmd.SetArgs([]string{"../../testdata/cpu.pb.gz", "../../testdata/cpu.pb.gz", "--format", "text"})

	err := DiffCmd.Execute()
	if err != nil {
		t.Fatalf("diff with same file should not error, got: %v", err)
	}
}

// 9.3: --tag applies the same filter to both profiles before diffing.
func TestDiffCommand_TagFilter(t *testing.T) {
	resetDiffFlags()
	var buf bytes.Buffer
	DiffCmd.SetOut(&buf)
	DiffCmd.SetErr(&buf)
	// Same profile on both sides; --tag filters both to state=blocked.
	DiffCmd.SetArgs([]string{
		"../../testdata/goroutine.pb.gz", "../../testdata/goroutine.pb.gz",
		"--tag", "state=blocked", "--format", "json",
	})

	if err := DiffCmd.Execute(); err != nil {
		t.Fatalf("diff should not error, got: %v", err)
	}
	if !strings.Contains(buf.String(), `"regressions"`) {
		t.Errorf("expected regressions field in diff output, got: %s", buf.String())
	}
}

// base 0 samples errors out before touching target.
func TestDiffCommand_TagFilterBaseZeroSamples(t *testing.T) {
	resetDiffFlags()
	var buf bytes.Buffer
	DiffCmd.SetOut(&buf)
	DiffCmd.SetErr(&buf)
	DiffCmd.SetArgs([]string{
		"../../testdata/cpu.pb.gz", "../../testdata/cpu.pb.gz",
		"--tag", "state=blocked",
	})

	err := DiffCmd.Execute()
	if err == nil {
		t.Fatal("expected error when base tag filter matches 0 samples")
	}
	if !strings.Contains(err.Error(), "matched 0 of 5 samples") {
		t.Errorf("expected 'matched 0 of 5 samples' in error, got: %v", err)
	}
}
