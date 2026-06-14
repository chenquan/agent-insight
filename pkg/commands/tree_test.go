package commands

import (
	"bytes"
	"testing"
)

func TestTreeCommand_NormalRun(t *testing.T) {
	var buf bytes.Buffer
	TreeCmd.SetOut(&buf)
	TreeCmd.SetErr(&buf)
	TreeCmd.SetArgs([]string{"../../testdata/cpu.pb.gz", "--format", "text"})

	err := TreeCmd.Execute()
	if err != nil {
		t.Fatalf("tree should not error, got: %v", err)
	}
}
