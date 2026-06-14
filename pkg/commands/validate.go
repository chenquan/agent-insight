package commands

import (
	"fmt"
	"regexp"
)

// ValidateFormat checks that the output format is one of the supported values.
func ValidateFormat(format string) error {
	switch format {
	case "text", "json", "markdown":
		return nil
	}
	return fmt.Errorf("invalid format: %s (must be text, json, or markdown)", format)
}

// ValidateRegex checks that a regex pattern compiles successfully.
// Empty patterns are valid (no filtering).
func ValidateRegex(pattern, name string) error {
	if pattern == "" {
		return nil
	}
	if _, err := regexp.Compile(pattern); err != nil {
		return fmt.Errorf("invalid %s pattern: %w", name, err)
	}
	return nil
}
