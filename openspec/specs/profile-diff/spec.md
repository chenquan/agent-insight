## Purpose

Compare two pprof profile files to identify performance regressions and improvements between baseline and target profiles, with filtering and multi-format output for AI-assisted analysis.

## Requirements

### Requirement: Compare two profile files
The system SHALL accept two profile files as input and compare their performance characteristics.

#### Scenario: Valid base and target profiles
- **WHEN** user provides two valid profile files (base.prof and target.prof)
- **THEN** system successfully parses both and generates a comparison report

#### Scenario: Profile type mismatch
- **WHEN** user provides profiles of different types (e.g., CPU vs heap)
- **THEN** system outputs a clear error indicating the type mismatch

#### Scenario: Missing or invalid profile files
- **WHEN** user provides non-existent or corrupted profile files
- **THEN** system outputs clear error messages indicating which file(s) failed to load

### Requirement: Calculate value differences
The system SHALL calculate the difference in values between base and target profiles for each function.

#### Scenario: Show delta values
- **WHEN** comparing two profiles
- **THEN** system shows the absolute and percentage change for flat and cumulative values

#### Scenario: Positive deltas
- **WHEN** a function's value increased from base to target
- **THEN** system displays the increase with a "+" prefix (e.g., "+120ms (+15.3%)")

#### Scenario: Negative deltas
- **WHEN** a function's value decreased from base to target
- **THEN** system displays the decrease with a "-" prefix (e.g., "-80ms (-12.1%)")

#### Scenario: Unchanged functions
- **WHEN** a function's value remains constant between profiles
- **THEN** system may omit or group unchanged functions depending on output format

### Requirement: Identify performance regressions
The system SHALL identify and highlight functions that show performance degradation in the target profile.

#### Scenario: Top regressions
- **WHEN** user runs diff without --top flag
- **THEN** system displays the top 15 functions with the largest percentage increases

#### Scenario: Custom regression count
- **WHEN** user specifies --top 20
- **THEN** system displays the top 20 functions with the largest percentage increases

#### Scenario: Filter by minimum change
- **WHEN** user specifies --min-delta 10
- **THEN** system only includes functions with percentage changes greater than 10%

### Requirement: Identify performance improvements
The system SHALL identify and highlight functions that show performance improvement in the target profile.

#### Scenario: Show improvements
- **WHEN** comparing profiles
- **THEN** system includes functions that decreased in value, indicating improvements

#### Scenario: Separate improvements from regressions
- **WHEN** system outputs diff results
- **THEN** improvements and regressions are clearly separated in the output

#### Scenario: Optional improvements-only view
- **WHEN** user specifies --improvements-only flag
- **THEN** system displays only functions that decreased in value

### Requirement: Support filtering by function patterns
The system SHALL support filtering the diff results using regular expression patterns.

#### Scenario: Focus on specific package
- **WHEN** user specifies --focus "encoding.*"
- **THEN** diff results only include functions matching the pattern

#### Scenario: Ignore noise functions
- **WHEN** user specifies --ignore "runtime.*"
- **THEN** diff results exclude functions matching the ignore pattern

#### Scenario: Combine filters
- **WHEN** user provides both --focus and --ignore patterns
- **THEN** system applies both filters to include matching and exclude ignored functions

### Requirement: Support multiple output formats
The system SHALL support text, JSON, and markdown output formats for diff results.

#### Scenario: Text format output
- **WHEN** user runs diff without --format flag
- **THEN** system outputs formatted text with columns for function, base, target, delta, and percentage

#### Scenario: JSON format output
- **WHEN** user specifies --format json
- **THEN** system outputs JSON with structure: {base, target, regressions[], improvements[], summary}

#### Scenario: Markdown format output
- **WHEN** user specifies --format markdown
- **THEN** system outputs formatted markdown with tables for regressions and improvements

### Requirement: Generate summary comparison
The system SHALL generate a summary comparing overall profile characteristics.

#### Scenario: Profile duration comparison
- **WHEN** both profiles contain duration information
- **THEN** summary shows the duration difference (e.g., "Duration: 30.1s → 25.2s (-16.2%)")

#### Scenario: Total samples comparison
- **WHEN** comparing any profile type
- **THEN** summary shows the total sample count change between base and target

#### Scenario: Overall performance assessment
- **WHEN** generating diff summary
- **THEN** system provides a brief assessment (e.g., "Overall performance improved by 12.3%")

### Requirement: Handle profile metadata differences
The system SHALL handle differences in metadata between the two profiles.

#### Scenario: Different profile types
- **WHEN** profiles have different sample types
- **THEN** system attempts to map compatible types or reports incompatibility

#### Scenario: Different sampling periods
- **WHEN** CPU profiles have different sampling durations
- **THEN** system normalizes values by time where appropriate

#### Scenario: Missing metadata
- **WHEN** one profile lacks metadata present in the other
- **THEN** system continues with available information and notes the discrepancy

### Requirement: Support relative and absolute deltas
The system SHALL support displaying both relative (percentage) and absolute value differences.

#### Scenario: Default percentage view
- **WHEN** user runs diff without specifying delta type
- **THEN** system shows both absolute and percentage changes

#### Scenario: Absolute values only
- **WHEN** user specifies --absolute-only flag
- **THEN** system displays only absolute value differences without percentages

#### Scenario: Percentage values only
- **WHEN** user specifies --percentage-only flag
- **THEN** system displays only percentage changes without absolute values

### Requirement: Handle new and deleted functions
The system SHALL identify functions that appear only in one profile.

#### Scenario: New functions in target
- **WHEN** target profile contains functions not present in base
- **THEN** system marks these as "NEW" and reports their values as 100% increase

#### Scenario: Deleted functions in target
- **WHEN** base profile contains functions not present in target
- **THEN** system marks these as "REMOVED" and reports their values as 100% decrease

#### Scenario: Optional new/deleted display
- **WHEN** user specifies --hide-new and --hide-deleted flags
- **THEN** system omits new or deleted functions from the output

### Requirement: Sort diff results by impact
The system SHALL sort diff results to show the most impactful changes first.

#### Scenario: Sort by percentage change
- **WHEN** user runs diff without --sort flag
- **THEN** system sorts by percentage change in descending order

#### Scenario: Sort by absolute change
- **WHEN** user specifies --sort absolute
- **THEN** system sorts by absolute value change in descending order

#### Scenario: Sort by function name
- **WHEN** user specifies --sort name
- **THEN** system sorts alphabetically by function name

### Requirement: Handle missing symbol information in diff
The system SHALL handle profiles with different levels of symbol information when comparing.

#### Scenario: Both profiles have symbols
- **WHEN** both base and target profiles contain function symbols
- **THEN** diff output uses function names for comparison and display

#### Scenario: Only one profile has symbols
- **WHEN** one profile has symbols while the other only has Location IDs
- **THEN** system attempts to match by Location ID first
- **AND** falls back to position-based matching if Location IDs don't align
- **AND** output clearly indicates which entries use Location IDs vs function names

#### Scenario: Both profiles lack symbols
- **WHEN** both profiles only contain Location IDs
- **THEN** diff output uses Location IDs and memory addresses for comparison
- **AND** includes module information from Mapping to provide context

### Requirement: Handle multiple value types in diff
The system SHALL correctly handle profiles with multiple value types when comparing differences.

#### Scenario: Matching value types
- **WHEN** both profiles have the same value types (e.g., both have alloc_objects and alloc_space)
- **THEN** system compares all corresponding value types
- **AND** allows user to specify which value type to focus on via --value-type

#### Scenario: Different value type structures
- **WHEN** profiles have different value type structures (e.g., standard heap vs Go heap)
- **THEN** system compares only the common value types
- **AND** warns user about non-comparable value types

#### Scenario: Value type selection for comparison
- **WHEN** user specifies --value-type inuse_bytes for diff
- **THEN** system calculates and displays deltas only for the specified value type
- **AND** clearly indicates which value type is being compared

### Requirement: Diff compares two profiles
系统 SHALL 比较两个 pprof profile 文件并识别性能变化。

#### Scenario: Compare two CPU profiles
- **WHEN** 用户对比两个相同 PeriodType 的 profile
- **THEN** 系统正常输出对比结果

#### Scenario: Compare mixed profile types
- **WHEN** 用户对比两个不同 PeriodType 的 profile（如 cpu 和 heap）
- **THEN** 系统报错并提示不能对比不同类型的 profile

### Requirement: Diff text output hides zero cum values
系统 SHALL 在 text 格式输出中隐藏 cum 值为 0 的列，减少视觉噪声。

#### Scenario: Text output with zero cum
- **WHEN** diff 结果中某函数的 cum 值为 0
- **THEN** text 格式输出中不显示该函数的 cum 列
