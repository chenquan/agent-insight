package profile

import (
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/google/pprof/profile"
)

// TracesResult contains the result of a traces query.
type TracesResult struct {
	Traces      []Trace
	TotalTraces int
	ShownTraces int
	ValueType   string
}

// Trace represents a single sample call chain.
type Trace struct {
	Stack   []string
	Value   int64
	Percent float64
}

// TracesConfig contains configuration for the traces command.
type TracesConfig struct {
	FocusPattern  string
	IgnorePattern string
	TopN          int
	ValueType     *ValueTypeConfig
}

// Traces iterates over samples and returns matching call chains.
func Traces(p *profile.Profile, config TracesConfig) (*TracesResult, error) {
	if p == nil {
		return nil, fmt.Errorf("profile is nil")
	}

	if config.ValueType == nil {
		metadata := extractMetadata(p)
		config.ValueType = selectDefaultValueType(p, metadata.Type)
	}

	var focusRegex, ignoreRegex *regexp.Regexp
	var err error

	if config.FocusPattern != "" {
		focusRegex, err = regexp.Compile(config.FocusPattern)
		if err != nil {
			return nil, fmt.Errorf("invalid focus pattern: %w", err)
		}
	}

	if config.IgnorePattern != "" {
		ignoreRegex, err = regexp.Compile(config.IgnorePattern)
		if err != nil {
			return nil, fmt.Errorf("invalid ignore pattern: %w", err)
		}
	}

	valueIndex := config.ValueType.Index

	// Calculate total value for percentages
	var totalValue int64
	for _, sample := range p.Sample {
		if valueIndex < len(sample.Value) {
			totalValue += sample.Value[valueIndex]
		}
	}

	var traces []Trace

	for _, sample := range p.Sample {
		if valueIndex >= len(sample.Value) || len(sample.Location) == 0 {
			continue
		}

		value := sample.Value[valueIndex]

		// Build stack from root to leaf
		stack := make([]string, 0, len(sample.Location))
		matched := false

		for i := len(sample.Location) - 1; i >= 0; i-- {
			loc := sample.Location[i]
			name := getLocationDisplayName(loc)
			stack = append(stack, name)

			if focusRegex != nil && focusRegex.MatchString(name) {
				matched = true
			}
		}

		// Check ignore: if any frame matches ignore, skip
		if ignoreRegex != nil {
			ignored := false
			for _, name := range stack {
				if ignoreRegex.MatchString(name) {
					ignored = true
					break
				}
			}
			if ignored {
				continue
			}
		}

		// Check focus: if focus is set, at least one frame must match
		if focusRegex != nil && !matched {
			continue
		}

		var percent float64
		if totalValue > 0 {
			percent = float64(value) / float64(totalValue) * 100
		}

		traces = append(traces, Trace{
			Stack:   stack,
			Value:   value,
			Percent: percent,
		})
	}

	// Sort by value descending
	sort.Slice(traces, func(i, j int) bool {
		return traces[i].Value > traces[j].Value
	})

	totalTraces := len(traces)

	// Apply top N limit
	if config.TopN > 0 && len(traces) > config.TopN {
		traces = traces[:config.TopN]
	}

	return &TracesResult{
		Traces:      traces,
		TotalTraces: totalTraces,
		ShownTraces: len(traces),
		ValueType:   config.ValueType.Name + "/" + config.ValueType.Unit,
	}, nil
}

// FormatStack returns the trace stack as a semicolon-separated string.
func (t *Trace) FormatStack() string {
	return strings.Join(t.Stack, ";")
}
