package commands

import (
	"fmt"
	"os"
	"sort"

	"github.com/chenquan/agent-insight/pkg/output"
	"github.com/chenquan/agent-insight/pkg/profile"

	"github.com/spf13/cobra"
)

// TrendCmd represents the trend command
var TrendCmd = &cobra.Command{
	Use:   "trend <path...>",
	Short: "Analyze performance trends across multiple profiles",
	Long: `Analyze multiple pprof profile files to detect performance trends over time.

Supports:
- Directory mode: recursively discover .pb and .pb.gz files
- Explicit file list
- Linear regression trend detection (regressing/improving/stable)
- Four-layer filtering (focus/ignore, min-impact, threshold, top N)
- Optional new hotspot and volatile function detection

Example usage:
  agent-insight trend ./profiles/cpu/
  agent-insight trend p1.pb.gz p2.pb.gz p3.pb.gz --format json
  agent-insight trend ./cpu/ --focus "pkg/server.*" --include-new
  agent-insight trend ./cpu/ --min-impact 0.5 --threshold 3 --top 5`,
	Args: cobra.MinimumNArgs(1),
	RunE: runTrend,
}

var (
	trendFormat        string
	trendFocus         string
	trendIgnore        string
	trendMinImpact     float64
	trendThreshold     float64
	trendTop           int
	trendSortBy        string
	trendIncludeNew    bool
	trendIncludeVolatile bool
)

func init() {
	TrendCmd.Flags().StringVar(&trendFormat, "format", "text", "Output format: text, json, markdown")
	TrendCmd.Flags().StringVar(&trendFocus, "focus", "", "Regex pattern to focus on specific functions")
	TrendCmd.Flags().StringVar(&trendIgnore, "ignore", "", "Regex pattern to ignore specific functions")
	TrendCmd.Flags().Float64Var(&trendMinImpact, "min-impact", 1, "Minimum flat percentage at any time point to include (0 = all)")
	TrendCmd.Flags().Float64Var(&trendThreshold, "threshold", 5, "Trend threshold percentage for classification")
	TrendCmd.Flags().IntVar(&trendTop, "top", 10, "Limit to top N in each category")
	TrendCmd.Flags().StringVar(&trendSortBy, "sort-by", "mtime", "Sort profiles by: mtime, name")
	TrendCmd.Flags().BoolVar(&trendIncludeNew, "include-new", false, "Include new hotspots in output")
	TrendCmd.Flags().BoolVar(&trendIncludeVolatile, "include-volatile", false, "Include volatile functions in output")
}

type profileEntry struct {
	path string
	mtime int64
}

func runTrend(cmd *cobra.Command, args []string) error {
	if err := ValidateFormat(trendFormat); err != nil {
		return err
	}

	if trendSortBy != "mtime" && trendSortBy != "name" {
		return fmt.Errorf("invalid sort-by: %s (must be mtime or name)", trendSortBy)
	}

	if err := ValidateRegex(trendFocus, "focus"); err != nil {
		return err
	}

	if err := ValidateRegex(trendIgnore, "ignore"); err != nil {
		return err
	}

	loader := profile.NewLoader()

	// Resolve input paths
	var entries []profileEntry
	for _, arg := range args {
		info, err := os.Stat(arg)
		if err != nil {
			return fmt.Errorf("failed to access %s: %w", arg, err)
		}

		if info.IsDir() {
			paths, err := loader.DiscoverProfiles(arg)
			if err != nil {
				return err
			}
			for _, p := range paths {
				mtime, err := getFileMtime(p)
				if err != nil {
					return fmt.Errorf("failed to get mtime for %s: %w", p, err)
				}
				entries = append(entries, profileEntry{path: p, mtime: mtime})
			}
		} else {
			mtime, err := getFileMtime(arg)
			if err != nil {
				return fmt.Errorf("failed to get mtime for %s: %w", arg, err)
			}
			entries = append(entries, profileEntry{path: arg, mtime: mtime})
		}
	}

	if len(entries) < 3 {
		return fmt.Errorf("need at least 3 profiles for trend analysis, found %d (use 'diff' for 2 profiles)", len(entries))
	}

	// Sort entries
	switch trendSortBy {
	case "mtime":
		sort.Slice(entries, func(i, j int) bool {
			return entries[i].mtime < entries[j].mtime
		})
	case "name":
		sort.Slice(entries, func(i, j int) bool {
			return entries[i].path < entries[j].path
		})
	}

	// Load profiles and build time points
	var profiles []*profile.Profile
	var timePoints []profile.TimePoint
	for _, e := range entries {
		p, err := loader.LoadFromFile(e.path)
		if err != nil {
			return fmt.Errorf("failed to load %s: %w", e.path, err)
		}
		profiles = append(profiles, p)
		timePoints = append(timePoints, profile.TimePoint{
			Label: e.path,
			Time:  e.mtime,
		})
	}

	// Configure trend
	config := profile.TrendConfig{
		MinImpact:       trendMinImpact,
		Threshold:       trendThreshold,
		TopN:            trendTop,
		FocusPattern:    trendFocus,
		IgnorePattern:   trendIgnore,
		IncludeNew:      trendIncludeNew,
		IncludeVolatile: trendIncludeVolatile,
	}

	result, err := profile.Trend(profiles, timePoints, config)
	if err != nil {
		return err
	}

	// Output
	switch trendFormat {
	case "json":
		return outputTrendJSON(result)
	case "markdown":
		return outputTrendMarkdown(result)
	default:
		return outputTrendText(result)
	}
}

func getFileMtime(path string) (int64, error) {
	info, err := os.Stat(path)
	if err != nil {
		return 0, err
	}
	return info.ModTime().Unix(), nil
}

func outputTrendText(result *profile.TrendResult) error {
	formatter := output.NewTrendTextFormatter(os.Stdout)
	return formatter.FormatTrendResult(result)
}

func outputTrendJSON(result *profile.TrendResult) error {
	formatter := output.NewTrendJSONFormatter(os.Stdout)
	return formatter.FormatTrendResult(result)
}

func outputTrendMarkdown(result *profile.TrendResult) error {
	formatter := output.NewTrendMarkdownFormatter(os.Stdout)
	return formatter.FormatTrendResult(result)
}
