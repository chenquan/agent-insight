package profile

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/google/pprof/profile"
)

// LabelSummary describes a single label key and the distribution of its values.
type LabelSummary struct {
	Key      string
	Type     string // "string" or "numeric"
	Unit     *string
	Values   []LabelValueSummary
	Distinct int
}

// LabelValueSummary describes a single value of a label.
type LabelValueSummary struct {
	Value   string
	Count   int64
	Percent float64
}

// LabelFilter filters samples by pprof labels.
type LabelFilter struct {
	Focus  map[string][]string
	Ignore map[string][]string
}

// BreakdownConfig controls label breakdown computation.
type BreakdownConfig struct {
	Keys       []string
	Top        int
	ValueIndex int // which column of sample.Value to accumulate (must match the value type used to compute hotspots)
}

// FunctionLabelBreakdown describes label breakdown for a single function.
type FunctionLabelBreakdown struct {
	Function Hotspot
	Labels   []LabelBreakdown
}

// LabelBreakdown is the per-label contribution for a function.
type LabelBreakdown struct {
	Key    string
	Values []LabelValueContribution
}

// LabelValueContribution is a single (key, value) bucket with flat and cum.
type LabelValueContribution struct {
	Value   string
	Flat    int64
	FlatPct float64
	Cum     int64
	CumPct  float64
}

// ExtractLabelSummaries returns a label summary for each unique key in p.Sample.
// String labels from Sample.Label; numeric labels from Sample.NumLabel.
// Sorted by distinct value count descending.
func ExtractLabelSummaries(p *Profile) []LabelSummary {
	if p == nil {
		return nil
	}

	type valueCount struct {
		value string
		count int64
	}
	acc := make(map[string]map[string]int64)
	types := make(map[string]string) // "string" or "numeric"
	units := make(map[string]*string)
	numUnits, _ := p.NumLabelUnits()

	for _, s := range p.Sample {
		// String labels
		for k, vals := range s.Label {
			if acc[k] == nil {
				acc[k] = make(map[string]int64)
				types[k] = "string"
			}
			for _, v := range vals {
				acc[k][v]++
			}
		}
		// Numeric labels
		for k, vals := range s.NumLabel {
			if acc[k] == nil {
				acc[k] = make(map[string]int64)
				types[k] = "numeric"
				if u, ok := numUnits[k]; ok {
					uCopy := u
					units[k] = &uCopy
				}
			}
			for _, v := range vals {
				acc[k][strconv.FormatInt(v, 10)]++
			}
		}
	}

	if len(acc) == 0 {
		return nil
	}

	out := make([]LabelSummary, 0, len(acc))
	total := int64(len(p.Sample))
	for k, vals := range acc {
		distinct := len(vals)
		pairs := make([]valueCount, 0, distinct)
		for v, c := range vals {
			pairs = append(pairs, valueCount{value: v, count: c})
		}
		sort.Slice(pairs, func(i, j int) bool {
			if pairs[i].count != pairs[j].count {
				return pairs[i].count > pairs[j].count
			}
			return pairs[i].value < pairs[j].value
		})
		const topN = 50
		n := len(pairs)
		truncated := n > topN
		if truncated {
			pairs = pairs[:topN]
		}
		values := make([]LabelValueSummary, 0, len(pairs))
		for _, pc := range pairs {
			var pct float64
			if total > 0 {
				pct = float64(pc.count) / float64(total) * 100
			}
			values = append(values, LabelValueSummary{Value: pc.value, Count: pc.count, Percent: pct})
		}
		ls := LabelSummary{
			Key:      k,
			Type:     types[k],
			Unit:     units[k],
			Values:   values,
			Distinct: distinct,
		}
		out = append(out, ls)
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].Distinct > out[j].Distinct
	})
	return out
}

// NewLabelFilter parses a list of "key=value" strings into a LabelFilter.
// Same key repeated -> OR; cross key -> AND.
func NewLabelFilter(focusFlags, ignoreFlags []string) (*LabelFilter, error) {
	focus, err := parseKeyValueFlags(focusFlags)
	if err != nil {
		return nil, fmt.Errorf("--tag: %w", err)
	}
	ignore, err := parseKeyValueFlags(ignoreFlags)
	if err != nil {
		return nil, fmt.Errorf("--tag-ignore: %w", err)
	}
	return &LabelFilter{Focus: focus, Ignore: ignore}, nil
}

func parseKeyValueFlags(flags []string) (map[string][]string, error) {
	if len(flags) == 0 {
		return nil, nil
	}
	out := make(map[string][]string)
	for _, f := range flags {
		idx := strings.Index(f, "=")
		if idx < 0 {
			return nil, fmt.Errorf("invalid value %q: expected key=value format", f)
		}
		key := f[:idx]
		val := f[idx+1:]
		if key == "" {
			return nil, fmt.Errorf("invalid value %q: empty key", f)
		}
		if val == "" {
			return nil, fmt.Errorf("invalid value %q: empty value", f)
		}
		out[key] = append(out[key], val)
	}
	return out, nil
}

// Apply returns a shallow copy of p with samples filtered by the filter.
// Returns an error if 0 samples match. Empty filter is a no-op.
func (f *LabelFilter) Apply(p *Profile) (*Profile, error) {
	if f == nil || (len(f.Focus) == 0 && len(f.Ignore) == 0) {
		return p, nil
	}
	if p == nil {
		return nil, fmt.Errorf("profile is nil")
	}

	filtered := make([]*profile.Sample, 0, len(p.Sample))
	unlabeled := 0
	for _, s := range p.Sample {
		if f.matches(s) {
			filtered = append(filtered, s)
		} else {
			unlabeled++
		}
	}
	if len(filtered) == 0 {
		return nil, fmt.Errorf(
			"tag filter matched 0 of %d samples (%d samples have no matching labels); "+
				"check --tag key=value spelling, or use a profile that has labels",
			len(p.Sample), unlabeled)
	}

	newP := p.Copy()
	newP.Sample = filtered
	return &Profile{
		Profile:        newP,
		LabelSummaries: p.LabelSummaries,
		InferredType:   p.InferredType,
	}, nil
}

func (f *LabelFilter) matches(sample *profile.Sample) bool {
	if sample == nil {
		return false
	}
	// Build sample's label map: key -> []string (combining string and numeric)
	m := make(map[string][]string)
	for k, vs := range sample.Label {
		m[k] = append(m[k], vs...)
	}
	for k, vs := range sample.NumLabel {
		for _, v := range vs {
			m[k] = append(m[k], strconv.FormatInt(v, 10))
		}
	}

	// Focus: all keys must be present, and at least one value must match per key (OR within key).
	for k, wantVals := range f.Focus {
		sampleVals, ok := m[k]
		if !ok {
			return false
		}
		if !anyValueMatch(wantVals, sampleVals) {
			return false
		}
	}

	// Ignore: if sample has the key, none of the ignored values may match.
	// If sample lacks the key, vacuously true.
	for k, badVals := range f.Ignore {
		sampleVals, ok := m[k]
		if !ok {
			continue
		}
		if anyValueMatch(badVals, sampleVals) {
			return false
		}
	}
	return true
}

func anyValueMatch(want, have []string) bool {
	for _, w := range want {
		for _, h := range have {
			if w == h {
				return true
			}
		}
	}
	return false
}

// ComputeFunctionBreakdowns computes per-function label-value flat distribution
// for the top-N functions. v1 only computes flat; cum/cum_pct are emitted as 0.
func ComputeFunctionBreakdowns(p *Profile, hotspots []Hotspot, cfg BreakdownConfig) []FunctionLabelBreakdown {
	if p == nil || len(cfg.Keys) == 0 || len(hotspots) == 0 {
		return nil
	}
	top := cfg.Top
	if top <= 0 {
		top = 20
	}
	if top > len(hotspots) {
		top = len(hotspots)
	}
	chosen := hotspots[:top]

	// Map function display name to chosen index. We match by display name
	// (mirroring getLocationDisplayName) rather than file, because the hotspot's
	// File field is formatted as "file:line" while a sample location exposes the
	// bare Filename — they never compare equal. The function name is the stable
	// identity a breakdown is reported against.
	keyIdx := make(map[string]int, len(chosen))
	for i, h := range chosen {
		keyIdx[hotspotDisplayName(h)] = i
	}

	// accum[chosen idx][label key][label value] -> flat
	accum := make([]map[string]map[string]int64, len(chosen))
	valueIndex := cfg.ValueIndex
	for _, sample := range p.Sample {
		if len(sample.Location) == 0 || len(sample.Value) == 0 {
			continue
		}
		vi := valueIndex
		if vi < 0 || vi >= len(sample.Value) {
			vi = 0
		}
		flat := sample.Value[vi]
		leaf := sample.Location[0]
		idx, ok := keyIdx[getLocationDisplayName(leaf)]
		if !ok {
			continue
		}
		if accum[idx] == nil {
			accum[idx] = make(map[string]map[string]int64)
		}
		for k, vs := range sample.Label {
			if accum[idx][k] == nil {
				accum[idx][k] = make(map[string]int64)
			}
			for _, v := range vs {
				accum[idx][k][v] += flat
			}
		}
		for k, vs := range sample.NumLabel {
			if accum[idx][k] == nil {
				accum[idx][k] = make(map[string]int64)
			}
			for _, v := range vs {
				accum[idx][k][strconv.FormatInt(v, 10)] += flat
			}
		}
	}

	// Emit exactly one breakdown entry per chosen function, in hotspots order,
	// so callers can align breakdowns[i] with hotspots[i] positionally. A
	// function whose samples carry none of the requested keys still gets an
	// entry (with empty Labels) rather than being dropped.
	out := make([]FunctionLabelBreakdown, 0, len(chosen))
	for i, h := range chosen {
		bd := FunctionLabelBreakdown{Function: h}
		for _, k := range cfg.Keys {
			vals := accum[i][k]
			if len(vals) == 0 {
				bd.Labels = append(bd.Labels, LabelBreakdown{Key: k})
				continue
			}
			total := int64(0)
			for _, c := range vals {
				total += c
			}
			pairs := make([]LabelValueContribution, 0, len(vals))
			for v, c := range vals {
				var pct float64
				if total > 0 {
					pct = float64(c) / float64(total) * 100
				}
				pairs = append(pairs, LabelValueContribution{
					Value:   v,
					Flat:    c,
					FlatPct: pct,
					Cum:     0,
					CumPct:  0,
				})
			}
			sort.Slice(pairs, func(a, b int) bool {
				if pairs[a].Flat != pairs[b].Flat {
					return pairs[a].Flat > pairs[b].Flat
				}
				return pairs[a].Value < pairs[b].Value
			})
			bd.Labels = append(bd.Labels, LabelBreakdown{Key: k, Values: pairs})
		}
		out = append(out, bd)
	}
	return out
}

// hotspotDisplayName returns the canonical display name for a hotspot. It
// mirrors getLocationDisplayName so that a sample's leaf location can be
// matched back to the hotspot it contributed to.
func hotspotDisplayName(h Hotspot) string {
	if h.Function != nil && *h.Function != "" {
		return *h.Function
	}
	if h.LocationID != nil {
		return fmt.Sprintf("loc_%d", *h.LocationID)
	}
	return ""
}
