package profile

import (
	"os"
	"path/filepath"
	"testing"
)

func TestErrorLoadInvalidFile(t *testing.T) {
	dir := t.TempDir()
	invalidPath := filepath.Join(dir, "invalid.pb.gz")
	if err := os.WriteFile(invalidPath, []byte("not a valid profile"), 0644); err != nil {
		t.Fatal(err)
	}

	loader := NewLoader()
	_, err := loader.LoadFromFile(invalidPath)
	if err == nil {
		t.Fatal("expected error for invalid profile file")
	}
}

func TestErrorLoadEmptyFile(t *testing.T) {
	dir := t.TempDir()
	emptyPath := filepath.Join(dir, "empty.pb.gz")
	if err := os.WriteFile(emptyPath, []byte{}, 0644); err != nil {
		t.Fatal(err)
	}

	loader := NewLoader()
	_, err := loader.LoadFromFile(emptyPath)
	if err == nil {
		t.Fatal("expected error for empty file")
	}
}

func TestErrorAnalysisNilProfile(t *testing.T) {
	_, err := NewAnalysis(nil, AnalysisConfig{TopN: 10, CallDepth: 0})
	if err == nil {
		t.Fatal("expected error for nil profile")
	}
}

func TestErrorDiffNilProfiles(t *testing.T) {
	_, err := Diff(nil, nil, DiffConfig{})
	if err == nil {
		t.Fatal("expected error for nil profiles")
	}
}

func TestErrorDiffNilBase(t *testing.T) {
	p := createTestProfile(t)
	_, err := Diff(nil, p, DiffConfig{})
	if err == nil {
		t.Fatal("expected error for nil base profile")
	}
}

func TestErrorDiffNilTarget(t *testing.T) {
	p := createTestProfile(t)
	_, err := Diff(p, nil, DiffConfig{})
	if err == nil {
		t.Fatal("expected error for nil target profile")
	}
}

func TestErrorListInvalidPattern(t *testing.T) {
	p := createTestProfile(t)
	_, err := List(p, ListConfig{Pattern: "[invalid"})
	if err == nil {
		t.Fatal("expected error for invalid regex pattern")
	}
}

func TestErrorFlameInvalidFocusPattern(t *testing.T) {
	p := createTestProfile(t)
	_, err := Flame(p, FlameConfig{FocusPattern: "[invalid"})
	if err == nil {
		t.Fatal("expected error for invalid focus pattern")
	}
}

func TestErrorFlameInvalidIgnorePattern(t *testing.T) {
	p := createTestProfile(t)
	_, err := Flame(p, FlameConfig{IgnorePattern: "[invalid"})
	if err == nil {
		t.Fatal("expected error for invalid ignore pattern")
	}
}

func TestAnalysisZeroTotalValueNoNaN(t *testing.T) {
	p := createTestProfile(t)
	// Use an ignore pattern that filters out all functions
	analysis, err := NewAnalysis(p, AnalysisConfig{TopN: 10, IgnorePattern: ".*", CallDepth: 0})
	if err != nil {
		t.Fatalf("NewAnalysis failed: %v", err)
	}
	if len(analysis.Hotspots) != 0 {
		t.Errorf("expected 0 hotspots when all filtered, got %d", len(analysis.Hotspots))
	}
}
