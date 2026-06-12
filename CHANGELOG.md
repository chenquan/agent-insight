# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/).

## [Unreleased]

### Added

- **init command**: Generate Claude Code skill file for agent-insight
  - Generates `.claude/skills/agent-insight/SKILL.md` with embedded usage guide
  - Includes trigger conditions, command reference, workflows, and output interpretation
  - Supports `--force` flag to overwrite existing skill file

## [0.1.0] - 2026-06-13

### Added

- **analyze command**: Parse and analyze pprof files, output performance hotspots
  - Top N hotspots by flat or cumulative value
  - Call stack path extraction with configurable depth
  - Natural language summary generation
  - Regex-based focus/ignore filtering
  - Optional collapsed stack output (`--collapse`)
- **list command**: Query function call relationships
  - Caller and callee information display
  - Regex pattern matching for function queries
  - Depth control and direction filtering
  - Leaf, recursive, and inline function handling
- **flame command**: Generate folded stack format for flame graphs
  - Stack aggregation and deduplication
  - Regex-based filtering (focus/ignore)
  - Depth limitation
  - Multiple value type support
  - Output statistics
- **diff command**: Compare two profile files
  - Regression and improvement detection
  - New and deleted function identification
  - Minimum delta threshold filtering
  - Overall performance assessment
- **Output formats**: text, JSON, markdown
- **Symbol fallback**: Gracefully handles profiles without debug symbols
- **Multi-value type support**: Smart defaults for CPU, heap, goroutine profiles
- **CI/CD**: GitHub Actions workflow with multi-version Go testing
- **Tests**: 27 unit, integration, error scenario, and benchmark tests
