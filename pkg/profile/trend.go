package profile

import (
	"fmt"
	"math"
	"regexp"
	"sort"

)

// TimePoint represents a single data point in the time series.
type TimePoint struct {
	Label string
	Time  int64 // Unix timestamp (mtime)
}

// OverallTrend represents the overall trend across all time points.
type OverallTrend struct {
	TotalSeries []int64  // Total samples at each time point
	Slope      float64  // Linear regression slope of totals
}

// FunctionTrend represents the trend of a single function across time points.
type FunctionTrend struct {
	Function   *string
	File       *string
	LocationID *uint64
	Address    *string
	Module     *string

	FlatSeries []*int64 // nil entries for missing time points
	CumSeries  []*int64

	Slope      float64 // Linear regression slope of flat values
	Trend      string  // "regressing", "improving", "stable"
	AvgFlat    float64 // Average of non-nil flat values
	Volatility float64 // Coefficient of variation

	PeakValue *int64
	PeakIndex int

	StartValue *int64 // First non-nil flat value
	EndValue   *int64 // Last non-nil flat value
}

// TrendResult represents the result of trend analysis across multiple profiles.
type TrendResult struct {
	TimePoints        []TimePoint
	ValueType         string
	Overall           OverallTrend
	Functions         []FunctionTrend
	TopRegressions    []FunctionTrend
	TopImprovements   []FunctionTrend
	NewHotspots       []FunctionTrend // populated when IncludeNew is true
	VolatileFunctions []FunctionTrend // populated when IncludeVolatile is true

	// Statistics
	RegressingCount  int
	ImprovingCount   int
	StableCount      int
}

// TrendConfig contains configuration for trend analysis.
type TrendConfig struct {
	MinImpact      float64 // Minimum flat percentage at any time point to include (default 1, 0 = no filter)
	Threshold      float64 // Trend threshold percentage (default 5)
	TopN           int     // Limit per category (default 10, 0 = unlimited)
	FocusPattern   string
	IgnorePattern  string
	ValueType      *ValueTypeConfig
	IncludeNew     bool
	IncludeVolatile bool
}

// Trend analyzes multiple profiles to detect performance trends.
func Trend(profiles []*Profile, timePoints []TimePoint, config TrendConfig) (*TrendResult, error) {
	if len(profiles) < 3 {
		return nil, fmt.Errorf("need at least 3 profiles for trend analysis, got %d (use 'diff' for 2 profiles)", len(profiles))
	}

	if len(profiles) != len(timePoints) {
		return nil, fmt.Errorf("profiles (%d) and time points (%d) count mismatch", len(profiles), len(timePoints))
	}

	if err := ValidateTypeConsistency(profiles); err != nil {
		return nil, err
	}

	if config.ValueType == nil {
		metadata := extractMetadata(profiles[0])
		config.ValueType = selectDefaultValueType(profiles[0], metadata.Type)
	}

	result := &TrendResult{
		ValueType:  config.ValueType.Name + "/" + config.ValueType.Unit,
		TimePoints: timePoints,
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

	// Build per-profile value maps (keyed by function name for cross-profile matching)
	valueIndex := config.ValueType.Index
	type funcValues struct {
		flat int64
		cum  int64
	}
	perProfileFuncs := make([]map[string]funcValues, len(profiles)) // funcName -> values
	perProfileTotal := make([]int64, len(profiles))
	allFuncInfo := make(map[string]*FunctionTrend) // funcName -> shared info

	for i, p := range profiles {
		fMap := make(map[string]funcValues)
		total := int64(0)

		for _, sample := range p.Sample {
			if valueIndex >= len(sample.Value) {
				continue
			}
			val := sample.Value[valueIndex]
			total += val
			if len(sample.Location) == 0 {
				continue
			}

			leaf := sample.Location[0]
			name := getFunctionNameFromProfile(p, leaf.ID)

			fv := fMap[name]
			fv.flat += val
			fMap[name] = fv

			for _, loc := range sample.Location {
				locName := getFunctionNameFromProfile(p, loc.ID)
				cv := fMap[locName]
				cv.cum += val
				fMap[locName] = cv
			}
		}

		perProfileFuncs[i] = fMap
		perProfileTotal[i] = total

		// Collect function info from first occurrence
		for name := range fMap {
			if _, exists := allFuncInfo[name]; !exists {
				locID := findFirstLocationID(p, name)
				info := &FunctionTrend{Function: &name}
				if locID > 0 {
					sym := extractLocationSymbol(p, locID)
					if sym != nil {
						info.File = sym.File
						info.Address = sym.Address
						info.Module = sym.Module
						info.LocationID = &locID
					}
				}
				allFuncInfo[name] = info
			}
		}
	}

	// Collect all unique function names
	allFuncNames := make(map[string]bool)
	for _, m := range perProfileFuncs {
		for name := range m {
			allFuncNames[name] = true
		}
	}

	// Build function trends
	n := len(profiles)
	for name := range allFuncNames {
		ft := FunctionTrend{
			FlatSeries: make([]*int64, n),
			CumSeries:  make([]*int64, n),
		}

		// Copy shared info
		if info, ok := allFuncInfo[name]; ok {
			ft.Function = info.Function
			ft.File = info.File
			ft.LocationID = info.LocationID
			ft.Address = info.Address
			ft.Module = info.Module
		}

		// L0: focus/ignore filter
		if !matchFunction(ft, focusRegex, ignoreRegex) {
			continue
		}

		var nonNilFlat []float64
		var nonNilIndices []int

		for i := 0; i < n; i++ {
			if fv, ok := perProfileFuncs[i][name]; ok {
				ft.FlatSeries[i] = &fv.flat
				ft.CumSeries[i] = &fv.cum
				nonNilFlat = append(nonNilFlat, float64(fv.flat))
				nonNilIndices = append(nonNilIndices, i)
			}
		}

		// L1: min-impact filter
		if config.MinImpact > 0 {
			maxImpact := 0.0
			for i := 0; i < n; i++ {
				if ft.FlatSeries[i] != nil && perProfileTotal[i] > 0 {
					impact := float64(*ft.FlatSeries[i]) / float64(perProfileTotal[i]) * 100
					if impact > maxImpact {
						maxImpact = impact
					}
				}
			}
			if maxImpact < config.MinImpact {
				continue
			}
		}

		// Compute average
		if len(nonNilFlat) > 0 {
			sum := 0.0
			for _, v := range nonNilFlat {
				sum += v
			}
			ft.AvgFlat = sum / float64(len(nonNilFlat))
		}

		// Linear regression on flat series
		ft.Slope = linearRegression(nonNilIndices, nonNilFlat)

		// L2: trend classification
		if ft.AvgFlat == 0 {
			ft.Trend = "stable"
		} else {
			pct := math.Abs(ft.Slope/ft.AvgFlat) * 100
			if ft.Slope/ft.AvgFlat*100 > config.Threshold {
				ft.Trend = "regressing"
			} else if ft.Slope/ft.AvgFlat*100 < -config.Threshold {
				ft.Trend = "improving"
			} else {
				ft.Trend = "stable"
			}
			_ = pct // used implicitly via threshold comparison
		}

		// Coefficient of variation
		if ft.AvgFlat > 0 {
			variance := 0.0
			for _, v := range nonNilFlat {
				diff := v - ft.AvgFlat
				variance += diff * diff
			}
			stddev := math.Sqrt(variance / float64(len(nonNilFlat)))
			ft.Volatility = stddev / ft.AvgFlat
		}

		// Peak and start/end values
		for i, v := range ft.FlatSeries {
			if v != nil {
				if ft.PeakValue == nil || *v > *ft.PeakValue {
					peak := *v
					ft.PeakValue = &peak
					ft.PeakIndex = i
				}
				if ft.StartValue == nil {
					start := *v
					ft.StartValue = &start
				}
				end := *v
				ft.EndValue = &end
			}
		}

		result.Functions = append(result.Functions, ft)
	}

	// Categorize
	for _, ft := range result.Functions {
		switch ft.Trend {
		case "regressing":
			result.RegressingCount++
		case "improving":
			result.ImprovingCount++
		default:
			result.StableCount++
		}
	}

	// Sort and apply top N
	sortRegressions(result.Functions, &result.TopRegressions, config.TopN)
	sortImprovements(result.Functions, &result.TopImprovements, config.TopN)

	// New hotspots detection
	if config.IncludeNew {
		cutoffIndex := int(float64(n) * 0.3)
		if cutoffIndex < 1 {
			cutoffIndex = 1
		}
		for _, ft := range result.Functions {
			firstIdx := -1
			for i, v := range ft.FlatSeries {
				if v != nil {
					firstIdx = i
					break
				}
			}
			if firstIdx > cutoffIndex && ft.EndValue != nil && perProfileTotal[n-1] > 0 {
				finalImpact := float64(*ft.EndValue) / float64(perProfileTotal[n-1]) * 100
				if finalImpact > config.MinImpact {
					result.NewHotspots = append(result.NewHotspots, ft)
				}
			}
		}
		sort.Slice(result.NewHotspots, func(i, j int) bool {
			return pctOfTotal(result.NewHotspots[i].EndValue, perProfileTotal[n-1]) >
				pctOfTotal(result.NewHotspots[j].EndValue, perProfileTotal[n-1])
		})
	}

	// Volatile detection
	if config.IncludeVolatile {
		for _, ft := range result.Functions {
			if ft.Trend == "stable" && ft.Volatility > 0.3 {
				result.VolatileFunctions = append(result.VolatileFunctions, ft)
			}
		}
		sort.Slice(result.VolatileFunctions, func(i, j int) bool {
			return result.VolatileFunctions[i].Volatility > result.VolatileFunctions[j].Volatility
		})
	}

	// Overall trend
	result.Overall = OverallTrend{
		TotalSeries: perProfileTotal,
	}
	var totalIndices []int
	var totalValues []float64
	for i, v := range perProfileTotal {
		totalIndices = append(totalIndices, i)
		totalValues = append(totalValues, float64(v))
	}
	result.Overall.Slope = linearRegression(totalIndices, totalValues)

	return result, nil
}

// locationSymbol holds extracted symbol information for a location.
type locationSymbol struct {
	Function   *string
	File       *string
	Address    *string
	Module     *string
}

func extractLocationSymbol(p *Profile, locID uint64) *locationSymbol {
	loc := findLocationByID(p, locID)
	if loc == nil {
		return nil
	}
	info := &locationSymbol{}
	if len(loc.Line) > 0 && loc.Line[0].Function != nil {
		info.Function = &loc.Line[0].Function.Name
		if loc.Line[0].Function.Filename != "" {
			file := fmt.Sprintf("%s:%d", loc.Line[0].Function.Filename, loc.Line[0].Line)
			info.File = &file
		}
	}
	addr := fmt.Sprintf("0x%x", loc.Address)
	info.Address = &addr
	if loc.Mapping != nil && loc.Mapping.File != "" {
		mod := normalizeMappingFile(loc.Mapping.File)
		info.Module = &mod
	}
	return info
}

func matchFunction(ft FunctionTrend, focusRegex, ignoreRegex *regexp.Regexp) bool {
	name := ""
	if ft.Function != nil {
		name = *ft.Function
	} else if ft.Address != nil {
		name = *ft.Address
	}

	if ignoreRegex != nil && ignoreRegex.MatchString(name) {
		return false
	}
	if focusRegex != nil && !focusRegex.MatchString(name) {
		return false
	}
	return true
}

func getFunctionNameFromProfile(p *Profile, locID uint64) string {
	loc := findLocationByID(p, locID)
	if loc == nil {
		return fmt.Sprintf("0x%x", locID)
	}
	return getFunctionNameFromLocation(loc)
}

func findFirstLocationID(p *Profile, funcName string) uint64 {
	for _, loc := range p.Location {
		name := getFunctionNameFromLocation(loc)
		if name == funcName {
			return loc.ID
		}
	}
	return 0
}

// linearRegression computes the slope using least squares.
// x values are indices, y values are the corresponding data points.
func linearRegression(xIndices []int, yValues []float64) float64 {
	n := len(xIndices)
	if n < 2 {
		return 0
	}

	sumX := 0.0
	sumY := 0.0
	sumXY := 0.0
	sumX2 := 0.0

	for i := 0; i < n; i++ {
		x := float64(xIndices[i])
		y := yValues[i]
		sumX += x
		sumY += y
		sumXY += x * y
		sumX2 += x * x
	}

	denom := float64(n)*sumX2 - sumX*sumX
	if denom == 0 {
		return 0
	}

	return (float64(n)*sumXY - sumX*sumY) / denom
}

func sortRegressions(all []FunctionTrend, result *[]FunctionTrend, topN int) {
	var regressing []FunctionTrend
	for _, ft := range all {
		if ft.Trend == "regressing" {
			regressing = append(regressing, ft)
		}
	}
	sort.Slice(regressing, func(i, j int) bool {
		return regressing[i].Slope > regressing[j].Slope
	})
	if topN > 0 && len(regressing) > topN {
		regressing = regressing[:topN]
	}
	*result = regressing
}

func sortImprovements(all []FunctionTrend, result *[]FunctionTrend, topN int) {
	var improving []FunctionTrend
	for _, ft := range all {
		if ft.Trend == "improving" {
			improving = append(improving, ft)
		}
	}
	sort.Slice(improving, func(i, j int) bool {
		return improving[i].Slope < improving[j].Slope
	})
	if topN > 0 && len(improving) > topN {
		improving = improving[:topN]
	}
	*result = improving
}

func pctOfTotal(value *int64, total int64) float64 {
	if value == nil || total == 0 {
		return 0
	}
	return float64(*value) / float64(total) * 100
}
