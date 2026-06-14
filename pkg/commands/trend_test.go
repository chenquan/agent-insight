package commands

import (
	"bytes"
	"strings"
	"testing"
)

func TestTrendCommand_TooFewProfiles(t *testing.T) {
	var buf bytes.Buffer
	TrendCmd.SetOut(&buf)
	TrendCmd.SetErr(&buf)
	TrendCmd.SetArgs([]string{"../../testdata/cpu.pb.gz", "../../testdata/heap.pb.gz"})

	err := TrendCmd.Execute()
	if err == nil {
		t.Fatal("expected error for fewer than 3 profiles")
	}
	if !strings.Contains(err.Error(), "at least 3") {
		t.Errorf("error should mention 'at least 3', got: %v", err)
	}
}

func TestTrendCommand_InvalidSortBy(t *testing.T) {
	var buf bytes.Buffer
	TrendCmd.SetOut(&buf)
	TrendCmd.SetErr(&buf)
	TrendCmd.SetArgs([]string{"../../testdata/cpu.pb.gz", "../../testdata/heap.pb.gz", "--sort-by", "invalid"})

	err := TrendCmd.Execute()
	if err == nil {
		t.Fatal("expected error for invalid sort-by")
	}
}
