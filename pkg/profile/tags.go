package profile

import "fmt"

// TagsResult is the label overview produced by the tags command.
type TagsResult struct {
	ProfilePath  string
	Type         string
	TotalSamples int
	Labels       []LabelSummary
}

// Tags builds a label overview for the profile. topN limits how many values are
// shown for numeric labels (string labels are always shown in full). It reads
// from the cached LabelSummaries computed at load time.
func Tags(p *Profile, profilePath string, topN int) (*TagsResult, error) {
	if p == nil {
		return nil, fmt.Errorf("profile is nil")
	}

	result := &TagsResult{
		ProfilePath:  profilePath,
		Type:         displayType(p),
		TotalSamples: len(p.Sample),
	}

	for _, ls := range p.LabelSummaries {
		// Numeric labels may carry many values; honor --top. String labels are
		// shown in full (they rarely exceed a handful of values).
		if ls.Type == "numeric" && topN > 0 && len(ls.Values) > topN {
			copy := ls
			copy.Values = ls.Values[:topN]
			result.Labels = append(result.Labels, copy)
			continue
		}
		result.Labels = append(result.Labels, ls)
	}
	return result, nil
}

// displayType returns the profile's display type, preferring the explicit
// PeriodType and falling back to the inferred type. Matches the info command.
func displayType(p *Profile) string {
	if p.PeriodType != nil && p.PeriodType.Type != "" {
		return p.PeriodType.Type
	}
	return p.InferredType
}
