package commands

import (
	"bytes"
	"strings"
	"testing"
)

func TestAnalyzeCommand_InvalidValueType(t *testing.T) {
	var buf bytes.Buffer
	AnalyzeCmd.SetOut(&buf)
	AnalyzeCmd.SetErr(&buf)
	AnalyzeCmd.SetArgs([]string{"../../testdata/heap.pb.gz", "--value-type", "nonexistent_metric"})

	err := AnalyzeCmd.Execute()
	if err == nil {
		t.Fatal("expected error for invalid value type")
	}

	if !strings.Contains(err.Error(), "unknown value type") {
		t.Errorf("error should mention 'unknown value type', got: %v", err)
	}

	if !strings.Contains(err.Error(), "available:") {
		t.Errorf("error should list available types, got: %v", err)
	}
}

func TestAnalyzeCommand_ValidValueType_NoError(t *testing.T) {
	var buf bytes.Buffer
	AnalyzeCmd.SetOut(&buf)
	AnalyzeCmd.SetErr(&buf)
	AnalyzeCmd.SetArgs([]string{"../../testdata/heap.pb.gz", "--value-type", "alloc_objects"})

	err := AnalyzeCmd.Execute()
	if err != nil {
		t.Fatalf("valid value-type should not error, got: %v", err)
	}
}
