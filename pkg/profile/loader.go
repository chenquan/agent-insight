package profile

import (
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/google/pprof/profile"
)

// Loader handles loading pprof profiles from various formats
type Loader struct {
	// Future: add configuration options here
}

// NewLoader creates a new profile loader
func NewLoader() *Loader {
	return &Loader{}
}

// LoadFromFile loads a profile from a file path
// Supports both .pb.gz (gzipped protobuf) and .pb (raw protobuf) formats
func (l *Loader) LoadFromFile(path string) (*Profile, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open profile file: %w", err)
	}
	defer func() { _ = file.Close() }()

	return l.ParseFromReader(file, path)
}

// ParseFromReader parses a profile from an io.Reader
// The filename is used to determine if gzip decompression should be applied
func (l *Loader) ParseFromReader(r io.Reader, filename string) (*Profile, error) {
	var reader = r

	// Check if the file is gzipped by extension
	if strings.HasSuffix(filename, ".gz") {
		gzReader, err := gzip.NewReader(r)
		if err != nil {
			return nil, fmt.Errorf("failed to create gzip reader: %w", err)
		}
		defer func() { _ = gzReader.Close() }()
		reader = gzReader
	}

	// Parse the profile using the pprof library
	p, err := profile.Parse(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to parse profile: %w", err)
	}

	return NewProfile(p), nil
}

// DiscoverProfiles recursively finds .pb and .pb.gz files in a directory.
func (l *Loader) DiscoverProfiles(dir string) ([]string, error) {
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
