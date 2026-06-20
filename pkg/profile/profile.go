package profile

import (
	"github.com/google/pprof/profile"
)

// Profile wraps *profile.Profile with cached label summaries and inferred type.
// It is the primary type returned by Loader, allowing all commands to share
// a single pass of label extraction and type inference.
type Profile struct {
	*profile.Profile
	LabelSummaries []LabelSummary
	InferredType   string
}

// NewProfile constructs a Profile from a raw *profile.Profile, computing
// label summaries and inferring the profile type once.
func NewProfile(p *profile.Profile) *Profile {
	if p == nil {
		return nil
	}
	wp := &Profile{Profile: p}
	wp.LabelSummaries = ExtractLabelSummaries(wp)
	wp.InferredType = inferProfileType(wp)
	return wp
}
