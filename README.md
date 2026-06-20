# agent-insight

A lightweight pprof analysis CLI for AI coding assistants. Works with any pprof-compatible profiles — Go, C++, Rust, and more.

## Features

- **Zero Configuration**: Single binary, no external dependencies
- **AI-Friendly Output**: Structured JSON format optimized for LLM parsing
- **Symbol Fallback**: Gracefully handles production profiles without debug symbols
- **Multi-Value Support**: Smart defaults for complex profile types (Go, C++, etc.)
- **Core Commands**: info, analyze, tags, list, traces, tree, flame, diff, merge, trend, diagnose, init

## Installation

```bash
# From source (binary at ./agent-insight)
make build

# Or via go install (requires Go 1.24+)
go install github.com/chenquan/agent-insight@latest
```

## Usage

### info - Show profile metadata

```bash
agent-insight info profile.pb.gz
agent-insight info cpu.pb.gz --format json
```

Zero-computation overview: type, duration, sample count, value types, symbol status, mappings.

### tags - List pprof labels

```bash
agent-insight tags goroutine.pb.gz
agent-insight tags service.pb.gz --top 20 --format json
```

Discovery layer: lists all pprof labels (`Sample.Label`) and their value distribution. Run it before `--tag` filtering to see which labels exist.

### analyze - Analyze performance hotspots

```bash
agent-insight analyze profile.pb.gz
agent-insight analyze profile.pb.gz --top 20 --cum
agent-insight analyze profile.pb.gz --format json --focus "runtime.*"
agent-insight analyze profile.pb.gz --collapse              # include folded stacks
agent-insight analyze goroutine.pb.gz --tag state=blocked        # filter samples by label
agent-insight analyze goroutine.pb.gz --tag-breakdown-on state   # per-function label distribution
```

### list - Query function call relationships

```bash
agent-insight list profile.pb.gz "main.*"
agent-insight list profile.pb.gz "runtime.mallocgc" --callers-only
agent-insight list profile.pb.gz "encoding.*" --depth 3 --format json
agent-insight list goroutine.pb.gz "Query" --tag state=blocked           # filter samples by label
agent-insight list profile.pb.gz "main.*" --ignore-function "runtime.*"  # exclude functions (was --exclude)
```

### flame - Generate folded stack format

```bash
agent-insight flame profile.pb.gz > stacks.folded
agent-insight flame profile.pb.gz --focus "encoding.*" --stats
agent-insight flame profile.pb.gz | flamegraph.pl > graph.svg
```

### traces - Show individual sample call traces

```bash
agent-insight traces profile.pb.gz
agent-insight traces profile.pb.gz --focus "runtime.*" --top 10
agent-insight traces goroutine.pb.gz --tag state=blocked --format json
```

Complements flame's aggregated view with per-sample call chains.

### tree - Show hierarchical call tree

```bash
agent-insight tree profile.pb.gz
agent-insight tree profile.pb.gz --depth 3 --top 5 --focus "main.*"
```

### diff - Compare two profiles

```bash
agent-insight diff before.prof after.prof
agent-insight diff base.prof target.prof --min-delta 10 --format json
agent-insight diff base.prof target.prof --hide-new --hide-deleted
agent-insight diff v1.pb.gz v2.pb.gz --tag http.status=500   # diff only 5xx samples (same filter on both)
```

### trend - Analyze performance trends

```bash
agent-insight trend ./profiles/cpu/
agent-insight trend p1.pb.gz p2.pb.gz p3.pb.gz --format json
agent-insight trend ./cpu/ --focus "pkg/server.*" --include-new
agent-insight trend ./cpu/ --min-impact 0.5 --threshold 3 --top 5
```

Requires at least 3 profiles. Detects regressing/improving functions via linear regression, with optional new hotspot and volatile function detection.

### init - Generate Claude Code skill

```bash
agent-insight init          # generate .claude/skills/agent-insight/SKILL.md
agent-insight init --force  # overwrite existing
```

## Output Formats

All commands support `--format text|json|markdown`. JSON is recommended for LLM consumption.

## Filtering

Most commands support `--focus <regex>` and `--ignore <regex>`:

```bash
agent-insight analyze profile.pb.gz --focus "main.*" --ignore "runtime.*"
```

## pprof Labels (tags)

Profiles may carry pprof labels on each sample (e.g. goroutine `state`/`wait_reason`, service `http.method`/`http.status`). `analyze`, `list`, `traces`, and `diff` support label filtering:

```bash
# Same key repeated = OR; different keys = AND
agent-insight analyze goroutine.pb.gz --tag state=blocked --tag state=running
agent-insight analyze goroutine.pb.gz --tag state=blocked --tag-ignore wait_reason=IO
agent-insight analyze goroutine.pb.gz --tag-breakdown-on state --tag-breakdown-top 10
```

- `--tag key=value`: keep samples matching (repeatable; same key OR, across keys AND)
- `--tag-ignore key=value`: drop samples matching (same semantics, inverted)
- `--tag-breakdown-on k1,k2` / `--tag-breakdown-top N` (analyze only): per-function flat distribution across label values

Use `tags` to discover available labels first. A filter matching 0 samples exits with an error.

> **BREAKING**: the `list` command's `--exclude` flag is renamed to `--ignore-function` (same semantics — regex-exclude functions). The new `--tag-ignore key=value` is a separate flag for label filtering.

## Profile Types

Supports CPU, heap (alloc/inuse × objects/space), goroutine, contention, and custom types. Use `--value-type` to select among multiple value types in heap profiles.

## JSON Output Example

```json
{
  "type": "cpu",
  "duration": "30s",
  "samples": 5,
  "sample_types": ["samples/count", "cpu/nanoseconds"],
  "top": [
    {
      "function": "runtime.mallocgc",
      "flat": 500,
      "flat_percent": 45.45,
      "cum": 500,
      "cum_percent": 45.45
    }
  ],
  "summary": "Profile type: cpu. Total samples: 5. Top hotspot: runtime.mallocgc (45.45%)."
}
```

When symbols are missing, hotspots fall back to address/module:

```json
{ "location_id": 9, "address": "0x430bac", "module": "/usr/bin/myapp", "flat": 85, "flat_percent": 40.28 }
```

## License

See LICENSE file for details.
