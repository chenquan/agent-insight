package commands

import (
	"bytes"
	"strings"
	"testing"

	"github.com/chenquan/agent-insight/pkg/output"
	"github.com/chenquan/agent-insight/pkg/profile"
)

func TestFlameCommand_InvalidFormat(t *testing.T) {
	var buf bytes.Buffer
	FlameCmd.SetOut(&buf)
	FlameCmd.SetErr(&buf)
	FlameCmd.SetArgs([]string{"../../testdata/cpu.pb.gz", "--format", "yaml"})

	err := FlameCmd.Execute()
	if err == nil {
		t.Fatal("expected error for invalid format")
	}
	if !strings.Contains(err.Error(), "invalid format") {
		t.Errorf("error should mention 'invalid format', got: %v", err)
	}
}

func TestFlameCommand_DefaultTextNoError(t *testing.T) {
	var buf bytes.Buffer
	FlameCmd.SetOut(&buf)
	FlameCmd.SetErr(&buf)
	FlameCmd.SetArgs([]string{"../../testdata/cpu.pb.gz", "--format", "text"})

	err := FlameCmd.Execute()
	if err != nil {
		t.Fatalf("default text format should not error, got: %v", err)
	}
}

func TestFlameFormatter_JSON(t *testing.T) {
	result := &profile.FlameResult{
		TotalStacks:  10,
		UniqueStacks: 3,
		Stacks: []profile.FoldedStack{
			{Stack: "main;handle;doWork", Count: 500},
			{Stack: "main;handle;alloc", Count: 200},
		},
	}

	var buf bytes.Buffer
	formatter := output.NewFlameFormatter(&buf)
	err := formatter.Format(result, "json")
	if err != nil {
		t.Fatalf("Format json failed: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "total_stacks") {
		t.Error("JSON should contain 'total_stacks'")
	}
	if !strings.Contains(out, `"stack"`) {
		t.Error("JSON should contain 'stack' field")
	}
	if !strings.Contains(out, `"handle"`) {
		t.Error("JSON stack should be split into array")
	}
}

func TestFlameFormatter_Markdown(t *testing.T) {
	result := &profile.FlameResult{
		TotalStacks:  10,
		UniqueStacks: 3,
		Stacks: []profile.FoldedStack{
			{Stack: "main;handle;doWork", Count: 500},
			{Stack: "main;handle;alloc", Count: 200},
		},
	}

	var buf bytes.Buffer
	formatter := output.NewFlameFormatter(&buf)
	err := formatter.Format(result, "markdown")
	if err != nil {
		t.Fatalf("Format markdown failed: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "|---|") {
		t.Error("Markdown should contain table separator")
	}
	if !strings.Contains(out, "```") {
		t.Error("Markdown should contain fenced code block")
	}
	if !strings.Contains(out, "doWork") {
		t.Error("Markdown should contain function name")
	}
}

func TestFlameFormatter_Text(t *testing.T) {
	result := &profile.FlameResult{
		Stacks: []profile.FoldedStack{
			{Stack: "main;doWork", Count: 100},
		},
	}

	var buf bytes.Buffer
	formatter := output.NewFlameFormatter(&buf)
	err := formatter.Format(result, "text")
	if err != nil {
		t.Fatalf("Format text failed: %v", err)
	}

	if !strings.Contains(buf.String(), "main;doWork 100") {
		t.Errorf("text output should contain 'main;doWork 100', got: %s", buf.String())
	}
}
