package commands

import (
	"bytes"
	"testing"
)

func TestDiagnoseCommand_NormalRun(t *testing.T) {
	var buf bytes.Buffer
	DiagnoseCmd.SetOut(&buf)
	DiagnoseCmd.SetErr(&buf)
	DiagnoseCmd.SetArgs([]string{"../../testdata/cpu.pb.gz", "--format", "text"})

	err := DiagnoseCmd.Execute()
	if err != nil {
		t.Fatalf("diagnose should not error, got: %v", err)
	}
}
