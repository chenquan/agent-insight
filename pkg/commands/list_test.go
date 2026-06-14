package commands

import (
	"bytes"
	"testing"
)

func TestListCommand_NormalRun(t *testing.T) {
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
	var buf bytes.Buffer
	ListCmd.SetOut(&buf)
	ListCmd.SetErr(&buf)
	ListCmd.SetArgs([]string{"../../testdata/cpu.pb.gz", "nonexistent_xyz_123.*"})

	err := ListCmd.Execute()
	if err != nil {
		t.Fatalf("no match should not error, got: %v", err)
	}
}
