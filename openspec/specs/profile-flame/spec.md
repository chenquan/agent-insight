## Purpose

Generate collapsed stack format from pprof profiles for flame graph visualization and AI-friendly structured output.
## Requirements
### Requirement: Generate collapsed stack format
The system SHALL convert pprof profile data into collapsed stack format compatible with flame graph visualization tools.

#### Scenario: Generate basic collapsed stacks
- **WHEN** user runs flame command on a profile
- **THEN** system outputs lines in format "func1;func2;func3 count" where count is the sample value

#### Scenario: Handle multiple stack traces
- **WHEN** profile contains multiple distinct stack traces
- **THEN** system outputs each unique stack trace on a separate line with aggregated counts

#### Scenario: Sort output by count
- **WHEN** generating collapsed stacks
- **THEN** system sorts output lines by count in descending order (highest first)

### Requirement: Collapse inline frames
The system SHALL handle inlined function calls appropriately in the collapsed stack output.

#### Scenario: Include inlined functions
- **WHEN** profile contains inlined function calls
- **THEN** system includes inlined functions in the stack trace separated by semicolons

#### Scenario: Distinguish inline calls
- **WHEN** displaying inlined functions
- **THEN** system uses clear naming (e.g., "function_name [inlined]") to indicate inline nature

### Requirement: Handle missing symbol information in collapsed stacks
The system SHALL gracefully handle profiles lacking function symbol information when generating collapsed stack format.

#### Scenario: Profile with function symbols
- **WHEN** profile contains function information
- **THEN** collapsed stacks use function names (e.g., "main;handleRequest;json.Marshal 15")

#### Scenario: Profile without function symbols
- **WHEN** profile lacks function information
- **THEN** collapsed stacks use Location IDs and memory addresses (e.g., "0x1234;0x5678;0x9abc 15")
- **AND** optionally includes module information when available (e.g., "[libc];0x5678;0x9abc 15")

#### Scenario: Mixed symbol availability
- **WHEN** profile has partial symbol information
- **THEN** collapsed stacks use function names where available and Location IDs where missing
- **AND** maintains consistent separator format regardless of information type

#### Scenario: User preference for information type
- **WHEN** user specifies --use-location-ids flag
- **THEN** system uses Location IDs even when function symbols are available
- **AND** this is useful for profiles with unreliable symbolization

### Requirement: Filter stacks by focus pattern
The system SHALL support filtering stack traces using regular expression patterns.

#### Scenario: Focus on specific package
- **WHEN** user specifies --focus "encoding/json"
- **THEN** system only includes stack traces containing functions matching the pattern

#### Scenario: Exclude ignored patterns
- **WHEN** user specifies --ignore "runtime.*"
- **THEN** system excludes stack traces where any function matches the ignore pattern

#### Scenario: Combine focus and ignore
- **WHEN** user specifies both --focus and --ignore patterns
- **THEN** system applies both filters, including only stacks that match focus and don't match ignore

### Requirement: Control stack truncation depth
The system SHALL allow users to limit the depth of stack traces in the output.

#### Scenario: Default depth limit
- **WHEN** user runs flame without --depth flag
- **THEN** system outputs full stack traces without truncation

#### Scenario: Custom depth limit
- **WHEN** user specifies --depth 10
- **THEN** system truncates stack traces at 10 frames, keeping only the top N frames

#### Scenario: Truncate from bottom
- **WHEN** depth limit is applied
- **THEN** system keeps leaf functions (bottom of stack) and removes from the root (top)

### Requirement: Support different sample value types
The system SHALL support using different value types from the profile for the collapsed stack counts.

#### Scenario: Use samples for CPU profiles
- **WHEN** processing a CPU profile
- **THEN** system uses sample count as the value in collapsed stacks

#### Scenario: Use bytes for heap profiles
- **WHEN** processing a heap profile
- **THEN** system uses allocated bytes as the value in collapsed stacks

#### Scenario: Specify custom value type
- **WHEN** user specifies --value-type inuse_objects
- **THEN** system uses the specified value type for count aggregation

### Requirement: Handle leaf and root aggregation
The system SHALL properly aggregate samples for common stack prefixes and suffixes.

#### Scenario: Aggregate common prefixes
- **WHEN** multiple stacks share the same root functions
- **THEN** system correctly aggregates counts for shared prefixes

#### Scenario: Aggregate common suffixes
- **WHEN** multiple stacks end with the same leaf function
- **THEN** system correctly aggregates counts for shared suffixes

#### Scenario: Preserve unique paths
- **WHEN** stacks differ at any point in the trace
- **THEN** system treats them as distinct entries in the output

### Requirement: Output to stdout for redirection
The system SHALL output collapsed stacks to stdout for easy redirection to files or pipes.

#### Scenario: Direct output to file
- **WHEN** user runs "agent-insight flame profile.pb.gz > stacks.folded"
- **THEN** system writes collapsed stacks to the file without additional metadata or formatting

#### Scenario: Pipe to flamegraph tool
- **WHEN** user pipes output to flamegraph.pl or similar tool
- **THEN** output format is compatible with the flame graph generator's input expectations

### Requirement: Support compressed input files
The system SHALL handle gzip-compressed profile files (.pb.gz) transparently.

#### Scenario: Parse compressed profile
- **WHEN** user provides a .pb.gz file
- **THEN** system automatically decompresses and parses the file

#### Scenario: Handle uncompressed input
- **WHEN** user provides an uncompressed .pb file
- **THEN** system parses the file without requiring compression

### Requirement: Report output statistics
The system SHALL optionally report statistics about the collapsed stack generation.

#### Scenario: Show total stacks
- **WHEN** user specifies --stats flag
- **THEN** system outputs the total number of unique stack traces generated

#### Scenario: Show filtered stacks count
- **WHEN** filters are applied with --focus or --ignore
- **THEN** stats output includes both total and filtered stack counts

#### Scenario: Show top stacks
- **WHEN** user specifies --top 5 with flame command
- **THEN** system outputs only the top 5 stacks by count in collapsed format

### Requirement: Flame command supports format flag
The system MUST support `--format` flag on the `flame` command with values `text`, `json`, or `markdown`. Default value MUST be `text`.

#### Scenario: Default text output unchanged
- **WHEN** user runs `agent-insight flame profile.pb.gz` without `--format`
- **THEN** output is byte-identical to previous text behavior (folded stack lines)

#### Scenario: JSON output structure
- **WHEN** user runs `agent-insight flame profile.pb.gz --format json`
- **THEN** output is valid JSON with fields `{total_stacks, filtered_stacks, unique_stacks, stacks: [{stack, value}], config}`
- **AND** `stack` is an array of function name strings (root → leaf)

#### Scenario: Markdown output includes table and code block
- **WHEN** user runs `agent-insight flame profile.pb.gz --format markdown`
- **THEN** output contains a markdown table with top 20 stacks (path + value)
- **AND** output contains a fenced code block with full folded stacks in flame graph format

#### Scenario: Invalid format value rejected
- **WHEN** user runs `agent-insight flame profile.pb.gz --format yaml`
- **THEN** system exits with non-zero status and clear error listing valid values

#### Scenario: Format flag composes with existing flags
- **WHEN** user runs `agent-insight flame profile.pb.gz --format json --focus "main\." --top 10`
- **THEN** JSON output reflects both focus filtering and top-10 limit

