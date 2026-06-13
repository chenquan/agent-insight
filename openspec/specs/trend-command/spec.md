## Purpose

Analyze multiple pprof profile files as a time series to detect performance trends, identify regressing and improving functions via linear regression, and surface new hotspots and volatile functions.

## Requirements

### Requirement: Trend command accepts multiple profile files
The `trend` command SHALL accept either a directory path (scanning for .pb and .pb.gz files recursively) or an explicit list of profile file paths. The system SHALL require at least 3 profile files; with fewer, it MUST print an error message suggesting the `diff` command.

#### Scenario: Directory with sufficient profiles
- **WHEN** user runs `agent-insight trend ./profiles/` and the directory contains 5 valid .pb.gz files
- **THEN** the system loads all 5 profiles sorted by file modification time and performs trend analysis

#### Scenario: Explicit file list
- **WHEN** user runs `agent-insight trend p1.pb.gz p2.pb.gz p3.pb.gz`
- **THEN** the system loads the 3 profiles in the order provided (or sorted by --sort-by if specified)

#### Scenario: Fewer than 3 profiles
- **WHEN** user runs `agent-insight trend p1.pb.gz p2.pb.gz`
- **THEN** the system prints an error and suggests using the `diff` command instead

#### Scenario: Directory with fewer than 3 profile files
- **WHEN** user runs `agent-insight trend ./profiles/` and the directory contains only 2 valid profile files
- **THEN** the system prints an error indicating insufficient profiles and suggests using the `diff` command

#### Scenario: Directory with no profile files
- **WHEN** user runs `agent-insight trend ./profiles/` and the directory contains no .pb or .pb.gz files
- **THEN** the system prints an error indicating no profile files were found

### Requirement: Profile time ordering
The system SHALL sort profiles by file modification time (mtime) by default. The `--sort-by` flag (values: `mtime`, `name`) SHALL override the default sort order. When `--sort-by name`, profiles are sorted by file path lexicographically.

#### Scenario: Default mtime sorting
- **WHEN** user runs `agent-insight trend ./profiles/` without `--sort-by`
- **THEN** profiles are ordered by file modification time ascending (oldest first)

#### Scenario: Sort by name
- **WHEN** user runs `agent-insight trend ./profiles/ --sort-by name`
- **THEN** profiles are ordered by file path lexicographically

### Requirement: Profile type consistency validation
The system SHALL validate that all loaded profiles have the same sample type. If profiles have inconsistent types, the system MUST print an error identifying the conflicting types.

#### Scenario: Mixed profile types
- **WHEN** user provides a CPU profile and a heap profile
- **THEN** the system prints an error indicating type mismatch and exits

### Requirement: Function time series extraction
For each function observed across all profiles, the system SHALL extract flat and cumulative value series across all time points. When a function is absent from a profile, the series entry SHALL be null (JSON) or "-" (text/markdown).

#### Scenario: Function present in all profiles
- **WHEN** function `runtime.gcBgMarkWorker` appears in all 5 profiles
- **THEN** both flat_series and cum_series have 5 numeric values

#### Scenario: Function absent from some profiles
- **WHEN** function `main.newWorker` appears only in profiles 3, 4, and 5
- **THEN** flat_series is [null, null, val3, val4, val5] in JSON output

### Requirement: Linear regression trend detection
The system SHALL compute the linear regression slope for each function's flat value series (skipping null entries). The classification formula is: `|slope / average_value| * 100 > threshold`. When `slope / average_value * 100 > threshold`, the function SHALL be classified as `regressing`. When `slope / average_value * 100 < -threshold`, it SHALL be `improving`. Otherwise it SHALL be `stable`. The `--threshold` flag (default: 5) represents a percentage of the average value.

#### Scenario: Regressing function
- **WHEN** a function's average flat value is 100, slope is 8, and threshold is 5
- **THEN** (8 / 100) * 100 = 8 > 5, so the function is classified as `regressing`

#### Scenario: Improving function
- **WHEN** a function's average flat value is 200, slope is -15, and threshold is 5
- **THEN** (-15 / 200) * 100 = -7.5 < -5, so the function is classified as `improving`

#### Scenario: Stable function
- **WHEN** a function's average flat value is 100, slope is 3, and threshold is 5
- **THEN** (3 / 100) * 100 = 3, which is within +/- 5, so the function is classified as `stable`

#### Scenario: Function with zero average value
- **WHEN** a function's average flat value is 0 (all non-null entries are 0)
- **THEN** the function is classified as `stable` (division by zero avoided)

### Requirement: Four-layer filtering
The system SHALL apply four layers of filtering in order: (1) `--focus` / `--ignore` regex patterns on function names, (2) `--min-impact` threshold (default: 1) filtering functions whose `max(flat_value_i / total_samples_at_timepoint_i * 100)` across all time points is below the threshold, (3) `--threshold` for trend classification, (4) `--top N` (default: 10) limiting output count per category.

#### Scenario: Focus pattern
- **WHEN** user provides `--focus "pkg/server.*"`
- **THEN** only functions matching the pattern are included in trend analysis

#### Scenario: Min-impact filter
- **WHEN** a function's flat proportion never exceeds 1% at any time point and `--min-impact` is default (1)
- **THEN** the function is excluded from output

#### Scenario: Min-impact set to zero
- **WHEN** user provides `--min-impact 0`
- **THEN** all functions pass the min-impact filter (no filtering applied at this layer)

#### Scenario: Top N limit
- **WHEN** there are 20 regressing functions and `--top` is 10
- **THEN** only the 10 with the highest slope are included in regressions output

### Requirement: Overall trend summary
The system SHALL output a summary section containing: time range (first to last profile label), number of profiles analyzed, value type used, total sample series across time points, overall slope, and counts of regressing/improving/stable functions.

#### Scenario: Summary output
- **WHEN** trend analysis completes successfully
- **THEN** output begins with a summary showing time range, profile count, value type, overall slope direction, and function category counts

### Requirement: Top regressions and improvements output
The system SHALL output two ranked sections: Top Regressions (sorted by slope descending) and Top Improvements (sorted by slope ascending). Each entry SHALL include: function name, slope, trend direction, flat series, cumulative series, start and end values, and change percentage.

#### Scenario: Regressions section
- **WHEN** 3 functions are classified as regressing and `--top` is 10
- **THEN** all 3 appear in regressions section ordered by slope descending, each with full time series data

### Requirement: New hotspots detection
When `--include-new` flag is provided, the system SHALL detect and output functions whose first appearance is after 30% of the time range and whose final flat proportion exceeds `--min-impact`. These are sorted by final flat proportion descending.

#### Scenario: New hotspot detected
- **WHEN** a function first appears at time point 4 of 5 and its final flat proportion is 5% with `--include-new` enabled
- **THEN** the function appears in the New Hotspots section

#### Scenario: New hotspots disabled by default
- **WHEN** trend analysis runs without `--include-new`
- **THEN** no New Hotspots section appears in output

### Requirement: Volatile functions detection
When `--include-volatile` flag is provided, the system SHALL detect and output functions classified as stable but with coefficient of variation > 0.3 in their flat series. These are sorted by coefficient of variation descending.

#### Scenario: Volatile function detected
- **WHEN** a function has slope near zero but flat values fluctuate significantly (CV > 0.3) and `--include-volatile` is enabled
- **THEN** the function appears in the Volatile section

### Requirement: Multi-format output
The trend command SHALL support `--format text|json|markdown` (default: text). JSON output SHALL use null for missing values. Text and markdown output SHALL use "-" for missing values.

#### Scenario: JSON format
- **WHEN** user runs `agent-insight trend ./profiles/ --format json`
- **THEN** output is valid JSON with null for missing series entries

#### Scenario: Text format
- **WHEN** user runs `agent-insight trend ./profiles/ --format text`
- **THEN** output is human-readable with "-" for missing series entries

#### Scenario: Markdown format
- **WHEN** user runs `agent-insight trend ./profiles/ --format markdown`
- **THEN** output is markdown-formatted with "-" for missing series entries

### Requirement: Value type selection
The system SHALL support `--value-type` flag to override automatic value type selection. For multi-value profiles (e.g., heap), the system SHALL use the same default logic as the `analyze` command when `--value-type` is not specified.

#### Scenario: Override value type
- **WHEN** user runs `agent-insight trend ./heap/ --value-type inuse_space`
- **THEN** analysis uses the specified value type

### Requirement: Three-layer architecture compliance
The trend command SHALL follow the project's three-layer architecture: `pkg/commands/trend.go` handles CLI argument parsing and output format selection only; `pkg/profile/trend.go` performs all computation and returns a `TrendResult` struct; `pkg/output/formatter.go` renders `TrendResult` without knowledge of CLI.

#### Scenario: Separation of concerns
- **WHEN** the trend command executes
- **THEN** no computation logic exists in commands layer and no formatting logic exists in profile layer
