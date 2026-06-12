# agent-insight

A lightweight pprof analysis CLI designed specifically for Claude Code and other AI coding assistants.

## Features

- **Zero Configuration**: Single binary, no external dependencies
- **AI-Friendly Output**: Structured JSON format optimized for LLM parsing
- **Symbol Fallback**: Gracefully handles production profiles without debug symbols
- **Multi-Value Support**: Smart defaults for complex profile types (Go heap, etc.)
- **Multiple Formats**: Text, JSON, and Markdown output
- **Four Core Commands**: analyze, list, flame, diff

## Installation

```bash
# Build from source
make build

# The binary will be at: ./build/agent-insight
```

## Usage

### analyze - Analyze performance hotspots

```bash
# Basic analysis
agent-insight analyze profile.pb.gz

# Top 20 hotspots, sorted by cumulative value
agent-insight analyze profile.pb.gz --top 20 --cum

# JSON output with filtering
agent-insight analyze profile.pb.gz --format json --focus "runtime.*"

# Include collapsed stacks for flame graph generation
agent-insight analyze profile.pb.gz --collapse

# Custom call stack depth
agent-insight analyze profile.pb.gz --call-depth 10
```

### list - Query function call relationships

```bash
# List functions matching a pattern
agent-insight list profile.pb.gz "main.*"

# Show only callers
agent-insight list profile.pb.gz "runtime.mallocgc" --callers-only

# JSON output with depth limit
agent-insight list profile.pb.gz "encoding.*" --depth 3 --format json

# Exclude runtime functions
agent-insight list profile.pb.gz "main.*" --exclude "runtime.*"
```

### flame - Generate folded stack format

```bash
# Generate collapsed stacks
agent-insight flame profile.pb.gz > stacks.folded

# With filtering and stats
agent-insight flame profile.pb.gz --focus "encoding.*" --stats

# Limit depth and top stacks
agent-insight flame profile.pb.gz --depth 10 --top 50

# Pipe to flamegraph tool
agent-insight flame profile.pb.gz | flamegraph.pl > graph.svg
```

### diff - Compare two profiles

```bash
# Compare profiles
agent-insight diff before.prof after.prof

# With minimum delta threshold
agent-insight diff base.prof target.prof --min-delta 10

# Focus on specific package
agent-insight diff base.prof target.prof --focus "runtime.*" --format json

# Hide new/deleted functions
agent-insight diff base.prof target.prof --hide-new --hide-deleted
```

### Output Formats

All commands support `--format text|json|markdown`:

```bash
# JSON output (AI-friendly)
agent-insight analyze profile.pb.gz --format json

# Markdown output
agent-insight analyze profile.pb.gz --format markdown

# Text output (default)
agent-insight analyze profile.pb.gz
```

### Filtering

```bash
# Focus on specific functions
agent-insight analyze profile.pb.gz --focus "runtime.*"

# Ignore specific functions
agent-insight analyze profile.pb.gz --ignore "runtime.*"

# Combine both
agent-insight analyze profile.pb.gz --focus "main.*" --ignore "runtime.*"
```

### Profile Types

Supports various profile types:
- CPU profiles
- Heap profiles (allocations, in-use memory)
- Goroutine profiles
- Contention profiles
- Custom profile types

## Output Examples

### JSON Output

```json
{
  "type": "cpu",
  "duration": "30s",
  "samples": 69,
  "sample_types": ["samples/count", "cpu/nanoseconds"],
  "top": [
    {
      "function": "runtime.mallocgc",
      "file": "runtime/malloc.go:1020",
      "flat": 85,
      "flat_percent": 40.28,
      "cum": 85,
      "cum_percent": 40.28
    }
  ],
  "summary": "Profile type: cpu. Total samples: 69. Top hotspot: runtime.mallocgc (40.28%)."
}
```

### Handling Missing Symbols

When profiles lack symbol information (common in production):

```json
{
  "top": [
    {
      "location_id": 9,
      "address": "0x430bac",
      "module": "/usr/bin/myapp",
      "flat": 85,
      "flat_percent": 40.28
    }
  ],
  "summary": "Limited symbol information available (0% of top functions)"
}
```

## Design Principles

1. **AI-First**: Structured output optimized for LLM consumption
2. **Production-Ready**: Handles profiles without debug symbols
3. **Lightweight**: Single binary, no Python/Node.js dependencies
4. **Compatible**: Works with `go tool pprof` generated files

## Development

```bash
# Build
make build

# Run tests
make test

# Clean build artifacts
make clean
```

## Compatibility

- Requires Go 1.26+
- Compatible with pprof protobuf format (.pb.gz and .pb)
- Tested with profiles from Go applications

## License

See LICENSE file for details.
