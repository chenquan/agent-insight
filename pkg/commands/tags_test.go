package commands

import (
	"bytes"
	"strings"
	"testing"
)

func TestTagsCommand_Text(t *testing.T) {
	var buf bytes.Buffer
	TagsCmd.SetOut(&buf)
	TagsCmd.SetErr(&buf)
	TagsCmd.SetArgs([]string{"../../testdata/goroutine.pb.gz", "--format", "text"})

	if err := TagsCmd.Execute(); err != nil {
		t.Fatalf("TagsCmd failed: %v", err)
	}
	out := buf.String()
	for _, want := range []string{"state", "wait_reason", "blocked", "Labels"} {
		if !strings.Contains(out, want) {
			t.Errorf("expected output to contain %q", want)
		}
	}
}

func TestTagsCommand_JSON(t *testing.T) {
	var buf bytes.Buffer
	TagsCmd.SetOut(&buf)
	TagsCmd.SetErr(&buf)
	TagsCmd.SetArgs([]string{"../../testdata/goroutine.pb.gz", "--format", "json"})

	if err := TagsCmd.Execute(); err != nil {
		t.Fatalf("TagsCmd failed: %v", err)
	}
	out := buf.String()
	for _, want := range []string{`"key":`, `"values":`, `"state"`, `"total_samples": 8`} {
		if !strings.Contains(out, want) {
			t.Errorf("expected JSON to contain %q", want)
		}
	}
}

func TestTagsCommand_NoLabelsProfile(t *testing.T) {
	var buf bytes.Buffer
	TagsCmd.SetOut(&buf)
	TagsCmd.SetErr(&buf)
	TagsCmd.SetArgs([]string{"../../testdata/cpu.pb.gz", "--format", "text"})

	if err := TagsCmd.Execute(); err != nil {
		t.Fatalf("TagsCmd failed: %v", err)
	}
	if !strings.Contains(buf.String(), "no labels found") {
		t.Errorf("expected 'no labels found' for label-less profile, got: %s", buf.String())
	}
}

func TestTagsCommand_InvalidFormat(t *testing.T) {
	var buf bytes.Buffer
	TagsCmd.SetOut(&buf)
	TagsCmd.SetErr(&buf)
	TagsCmd.SetArgs([]string{"../../testdata/cpu.pb.gz", "--format", "xml"})

	if err := TagsCmd.Execute(); err == nil {
		t.Fatal("expected error for invalid format")
	}
}

func TestTagsCommand_MissingArg(t *testing.T) {
	var buf bytes.Buffer
	TagsCmd.SetOut(&buf)
	TagsCmd.SetErr(&buf)
	TagsCmd.SetArgs([]string{})

	if err := TagsCmd.Execute(); err == nil {
		t.Fatal("expected error for missing profile arg")
	}
}
