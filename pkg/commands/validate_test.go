package commands

import (
	"strings"
	"testing"
)

func TestValidateFormat(t *testing.T) {
	tests := []struct {
		format string
		want   bool
	}{
		{"text", true},
		{"json", true},
		{"markdown", true},
		{"yaml", false},
		{"", false},
		{"TEXT", false},
	}
	for _, tt := range tests {
		err := ValidateFormat(tt.format)
		if (err == nil) != tt.want {
			t.Errorf("ValidateFormat(%q) error = %v, want error = %v", tt.format, err, !tt.want)
		}
	}
}

func TestValidatePositiveInt(t *testing.T) {
	tests := []struct {
		value int
		name  string
		want  bool
	}{
		{10, "top", true},
		{1, "top", true},
		{0, "top", false},
		{-1, "top", false},
	}
	for _, tt := range tests {
		err := ValidatePositiveInt(tt.value, tt.name)
		if (err == nil) != tt.want {
			t.Errorf("ValidatePositiveInt(%d, %q) error = %v, want error = %v", tt.value, tt.name, err, !tt.want)
		}
		if err != nil && tt.name != "" {
			if !strings.Contains(err.Error(), tt.name) {
				t.Errorf("ValidatePositiveInt(%d, %q) error = %v, want error to contain %q", tt.value, tt.name, err, tt.name)
			}
		}
	}
}

func TestValidateRegex(t *testing.T) {
	tests := []struct {
		pattern string
		name    string
		want    bool
	}{
		{"runtime.*", "focus", true},
		{"", "focus", true},
		{"[invalid", "focus", false},
		{"encoding\\.json", "ignore", true},
	}
	for _, tt := range tests {
		err := ValidateRegex(tt.pattern, tt.name)
		if (err == nil) != tt.want {
			t.Errorf("ValidateRegex(%q, %q) error = %v, want error = %v", tt.pattern, tt.name, err, !tt.want)
		}
	}
}
