package commands

import (
	"bytes"
	"testing"
)

func TestTracesCommand_NormalRun(t *testing.T) {
	var buf bytes.Buffer
	TracesCmd.SetOut(&buf)
	TracesCmd.SetErr(&buf)
	TracesCmd.SetArgs([]string{"../../testdata/cpu.pb.gz", "--format", "text"})

	err := TracesCmd.Execute()
	if err != nil {
		t.Fatalf("traces should not error, got: %v", err)
	}
}
