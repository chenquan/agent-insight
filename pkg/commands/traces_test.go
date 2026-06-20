package commands

import (
	"bytes"
	"strings"
	"testing"
)

// resetTracesFlags restores the traces command's package-level flag variables
// to their defaults. Cobra's SetArgs+Execute does not reset flags set by a
// prior test, so without this each test inherits the previous test's values.
func resetTracesFlags() {
	tracesFocus = ""
	tracesIgnore = ""
	tracesTop = 20
	tracesFormat = "text"
	tracesTag = nil
	tracesIgnoreTag = nil
}

func TestTracesCommand_NormalRun(t *testing.T) {
	resetTracesFlags()
	var buf bytes.Buffer
	TracesCmd.SetOut(&buf)
	TracesCmd.SetErr(&buf)
	TracesCmd.SetArgs([]string{"../../testdata/cpu.pb.gz", "--format", "text"})

	err := TracesCmd.Execute()
	if err != nil {
		t.Fatalf("traces should not error, got: %v", err)
	}
}

// 8.3: --tag filters samples before querying traces.
func TestTracesCommand_TagFilter(t *testing.T) {
	resetTracesFlags()
	var buf bytes.Buffer
	TracesCmd.SetOut(&buf)
	TracesCmd.SetErr(&buf)
	TracesCmd.SetArgs([]string{"../../testdata/goroutine.pb.gz", "--tag", "state=blocked", "--format", "json"})

	if err := TracesCmd.Execute(); err != nil {
		t.Fatalf("traces should not error, got: %v", err)
	}
	// 8 samples total, 4 are state=blocked -> only blocked traces shown.
	if !strings.Contains(buf.String(), `"total_traces": 4`) {
		t.Errorf("expected total_traces filtered to 4, got: %s", buf.String())
	}
}

// Zero-sample tag filter errors out (consistent with analyze/list/diff).
func TestTracesCommand_TagFilterZeroSamples(t *testing.T) {
	resetTracesFlags()
	var buf bytes.Buffer
	TracesCmd.SetOut(&buf)
	TracesCmd.SetErr(&buf)
	TracesCmd.SetArgs([]string{"../../testdata/cpu.pb.gz", "--tag", "state=blocked"})

	err := TracesCmd.Execute()
	if err == nil {
		t.Fatal("expected error when tag filter matches 0 samples")
	}
	if !strings.Contains(err.Error(), "matched 0 of 5 samples") {
		t.Errorf("expected 'matched 0 of 5 samples' in error, got: %v", err)
	}
}
