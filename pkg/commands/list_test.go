package commands

import (
	"bytes"
	"strings"
	"testing"
)

// resetListFlags restores the list command's package-level flag variables to
// their defaults. Cobra's SetArgs+Execute does not reset flags set by a prior
// test, so without this each test inherits the previous test's flag values.
func resetListFlags() {
	listDepth = 5
	listCallersOnly = false
	listCalleesOnly = false
	listIgnoreFunction = ""
	listFormat = "text"
	listTag = nil
	listIgnoreTag = nil
}

func TestListCommand_NormalRun(t *testing.T) {
	resetListFlags()
	var buf bytes.Buffer
	ListCmd.SetOut(&buf)
	ListCmd.SetErr(&buf)
	ListCmd.SetArgs([]string{"../../testdata/cpu.pb.gz", "runtime.*", "--format", "text"})

	err := ListCmd.Execute()
	if err != nil {
		t.Fatalf("list should not error, got: %v", err)
	}
}

func TestListCommand_NoMatch(t *testing.T) {
	resetListFlags()
	var buf bytes.Buffer
	ListCmd.SetOut(&buf)
	ListCmd.SetErr(&buf)
	ListCmd.SetArgs([]string{"../../testdata/cpu.pb.gz", "nonexistent_xyz_123.*"})

	err := ListCmd.Execute()
	if err != nil {
		t.Fatalf("no match should not error, got: %v", err)
	}
}

// 7.4: --ignore-function excludes matching functions (equivalent to old --exclude).
func TestListCommand_IgnoreFunction(t *testing.T) {
	resetListFlags()
	var buf bytes.Buffer
	ListCmd.SetOut(&buf)
	ListCmd.SetErr(&buf)
	// "encoding.*" matches encoding/json.Marshal, but --ignore-function drops it
	// entirely, leaving no matched functions.
	ListCmd.SetArgs([]string{"../../testdata/cpu.pb.gz", "encoding.*", "--ignore-function", "encoding.*"})

	if err := ListCmd.Execute(); err != nil {
		t.Fatalf("list should not error, got: %v", err)
	}
	if !strings.Contains(buf.String(), "No functions matched") {
		t.Errorf("expected ignore-function to drop all matches, got: %s", buf.String())
	}
}

// 7.5: --tag filters samples before querying.
func TestListCommand_TagFilter(t *testing.T) {
	resetListFlags()
	var buf bytes.Buffer
	ListCmd.SetOut(&buf)
	ListCmd.SetErr(&buf)
	ListCmd.SetArgs([]string{"../../testdata/goroutine.pb.gz", "Query", "--tag", "state=blocked", "--format", "json"})

	if err := ListCmd.Execute(); err != nil {
		t.Fatalf("list should not error, got: %v", err)
	}
	if !strings.Contains(buf.String(), "QueryContext") {
		t.Errorf("expected QueryContext in filtered output, got: %s", buf.String())
	}
}

// 7.6: the old --exclude flag is no longer recognized.
func TestListCommand_OldExcludeRejected(t *testing.T) {
	resetListFlags()
	var buf bytes.Buffer
	ListCmd.SetOut(&buf)
	ListCmd.SetErr(&buf)
	ListCmd.SetArgs([]string{"../../testdata/cpu.pb.gz", "main.*", "--exclude", "database.*"})

	err := ListCmd.Execute()
	if err == nil {
		t.Fatal("expected error for unknown --exclude flag")
	}
	if !strings.Contains(err.Error(), "unknown flag") {
		t.Errorf("expected 'unknown flag' error, got: %v", err)
	}
}

// 7.7: --tag-ignore requires key=value format (no bare regex).
func TestListCommand_TagIgnoreInvalidFormat(t *testing.T) {
	resetListFlags()
	var buf bytes.Buffer
	ListCmd.SetOut(&buf)
	ListCmd.SetErr(&buf)
	ListCmd.SetArgs([]string{"../../testdata/cpu.pb.gz", "main.*", "--tag-ignore", "database.*"})

	err := ListCmd.Execute()
	if err == nil {
		t.Fatal("expected error for --tag-ignore without '='")
	}
	if !strings.Contains(err.Error(), "key=value") {
		t.Errorf("expected key=value format error, got: %v", err)
	}
}

// 7.8: --ignore-function and --tag-ignore coexist.
func TestListCommand_IgnoreFunctionAndTagCoexist(t *testing.T) {
	resetListFlags()
	var buf bytes.Buffer
	ListCmd.SetOut(&buf)
	ListCmd.SetErr(&buf)
	ListCmd.SetArgs([]string{
		"../../testdata/goroutine.pb.gz", "Query",
		"--ignore-function", "runtime.*", "--tag-ignore", "state=running",
		"--format", "json",
	})

	if err := ListCmd.Execute(); err != nil {
		t.Fatalf("list should not error, got: %v", err)
	}
}
