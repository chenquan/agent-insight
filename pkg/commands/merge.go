package commands

import (
	"compress/gzip"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	pprofprofile "github.com/google/pprof/profile"

	"github.com/chenquan/agent-insight/pkg/profile"

	"github.com/spf13/cobra"
)

// MergeCmd represents the merge command
var MergeCmd = &cobra.Command{
	Use:   "merge <profile...> -o <output>",
	Short: "Merge multiple pprof profiles of the same type",
	Long: `Merge multiple pprof profile files of the same type into one.

Supports:
- Merging multiple profile files
- Directory mode: recursively discover .pb and .pb.gz files
- Profile type consistency validation

Example usage:
  agent-insight merge cpu1.pb.gz cpu2.pb.gz cpu3.pb.gz -o merged.pb.gz
  agent-insight merge ./profiles/ -o merged.pb.gz`,
	Args: cobra.MinimumNArgs(1),
	RunE: runMerge,
}

var mergeOutput string

func init() {
	MergeCmd.Flags().StringVarP(&mergeOutput, "output", "o", "", "Output file path (required)")
	if err := MergeCmd.MarkFlagRequired("output"); err != nil {
		panic(fmt.Sprintf("failed to mark output flag as required: %v", err))
	}
}

func runMerge(cmd *cobra.Command, args []string) error {
	outputPath := mergeOutput
	if outputPath == "" {
		return fmt.Errorf("output path is required (use -o)")
	}

	// Resolve input paths: separate files from directories
	var profilePaths []string
	for _, arg := range args {
		info, err := os.Stat(arg)
		if err != nil {
			return fmt.Errorf("failed to access %s: %w", arg, err)
		}

		if info.IsDir() {
			paths, err := discoverProfiles(arg)
			if err != nil {
				return err
			}
			profilePaths = append(profilePaths, paths...)
		} else {
			profilePaths = append(profilePaths, arg)
		}
	}

	if len(profilePaths) < 2 {
		return fmt.Errorf("need at least 2 profile files to merge, found %d", len(profilePaths))
	}

	// Load all profiles
	loader := profile.NewLoader()
	var profiles []*pprofprofile.Profile

	for _, p := range profilePaths {
		pf, err := loader.LoadFromFile(p)
		if err != nil {
			return fmt.Errorf("failed to load %s: %w", p, err)
		}
		profiles = append(profiles, pf)
	}

	// Merge
	merged, result, err := profile.ValidateAndMerge(profiles)
	if err != nil {
		return err
	}

	// Write output
	if err := writeMergeOutput(merged, outputPath); err != nil {
		return fmt.Errorf("failed to write output: %w", err)
	}

	// Print summary
	fmt.Fprintf(cmd.OutOrStdout(), "Merged %d profiles (%d samples, %s) → %s\n",
		result.InputCount, result.TotalSamples, result.ValueType, outputPath)

	return nil
}

// discoverProfiles recursively finds .pb and .pb.gz files in a directory.
func discoverProfiles(dir string) ([]string, error) {
	var paths []string

	err := filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}

		base := strings.ToLower(d.Name())
		if strings.HasSuffix(base, ".pb.gz") || strings.HasSuffix(base, ".pb") {
			paths = append(paths, path)
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to scan directory %s: %w", dir, err)
	}

	if len(paths) == 0 {
		return nil, fmt.Errorf("no profile files (.pb or .pb.gz) found in %s", dir)
	}

	sort.Strings(paths)
	return paths, nil
}

// writeMergeOutput writes the merged profile as gzip-compressed protobuf.
func writeMergeOutput(p *pprofprofile.Profile, outputPath string) error {
	f, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()

	gzWriter := gzip.NewWriter(f)
	defer func() { _ = gzWriter.Close() }()

	if err := p.Write(gzWriter); err != nil {
		return fmt.Errorf("failed to serialize profile: %w", err)
	}

	return gzWriter.Close()
}
