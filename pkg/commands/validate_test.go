package commands

import "testing"

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
