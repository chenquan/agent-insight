## ADDED Requirements

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