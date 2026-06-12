# agent-insight

A lightweight pprof analysis CLI designed specifically for Claude Code and other AI coding assistants.

## Features

- **Zero Configuration**: Single binary, no external dependencies
- **AI-Friendly Output**: Structured JSON format optimized for LLM parsing
- **Symbol Fallback**: Gracefully handles production profiles without debug symbols
- **Multi-Value Support**: Smart defaults for complex profile types (Go heap, etc.)
- **Eight Core Commands**: analyze, list, flame, diff, info, traces, tree, init

## Installation

```bash
# From source (binary at ./agent-insight)
make build

# Or via go install (requires Go 1.26+)
go install github.com/chenquan/agent-insight@latest
```

## Usage

### info - Show profile metadata

```bash
agent-insight info profile.pb.gz
agent-insight info cpu.pb.gz --format json
```

Zero-computation overview: type, duration, sample count, value types, symbol status, mappings.

### analyze - Analyze performance hotspots

```bash
agent-insight analyze profile.pb.gz
agent-insight analyze profile.pb.gz --top 20 --cum
agent-insight analyze profile.pb.gz --format json --focus "runtime.*"
agent-insight analyze profile.pb.gz --collapse              # include folded stacks
```

### list - Query function call relationships

```bash
agent-insight list profile.pb.gz "main.*"
agent-insight list profile.pb.gz "runtime.mallocgc" --callers-only
agent-insight list profile.pb.gz "encoding.*" --depth 3 --format json
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
```

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
