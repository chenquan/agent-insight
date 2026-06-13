## MODIFIED Requirements

### Requirement: Diagnose command accepts pprof file and outputs diagnostic prompt

The `diagnose` command SHALL accept a pprof file path as a positional argument and output a structured diagnostic prompt to stdout.

The command SHALL support the following flags:
- `--top N`: control the number of hotspot functions included (default: 10)
- `--context <string>`: user-provided application context to embed in the prompt
- `--format <format>`: output format, one of `text` (default), `markdown`, `json`

The command SHALL fail with a descriptive error when:
- no file argument is provided
- the file cannot be parsed as a valid pprof profile
- an invalid format value is specified

The SKILL.md template SHALL document the `diagnose` trigger condition as: the user expresses an exploratory or uncertain intent about a profile (e.g., "帮我看看这个 profile", "诊断一下性能", "这个 profile 有什么问题") without specifying a concrete analysis target.

The SKILL.md template SHALL document scenarios where `diagnose` SHOULD NOT be used, directing the AI to use specific commands instead:
- User asks about a specific function's performance → use `analyze` + `list`
- User asks "who calls X" → use `list`
- User wants to compare versions → use `diff`
- User suspects memory leak → use `analyze --value-type alloc_objects`
- User suspects goroutine leak → use `traces --focus`
- User wants a flame graph → use `flame`

#### Scenario: Basic diagnose output
- **WHEN** user runs `agent-insight diagnose cpu.pb.gz`
- **THEN** the command outputs a diagnostic prompt containing: role definition, profile metadata, analysis data, and diagnostic guidance to stdout

#### Scenario: Diagnose with context
- **WHEN** user runs `agent-insight diagnose cpu.pb.gz --context "HTTP microservice processing JSON API requests"`
- **THEN** the output prompt includes the user-provided context in the prompt

#### Scenario: Diagnose with custom top N
- **WHEN** user runs `agent-insight diagnose cpu.pb.gz --top 5`
- **THEN** the analysis data section contains at most 5 hotspot functions

#### Scenario: Diagnose with JSON format
- **WHEN** user runs `agent-insight diagnose cpu.pb.gz --format json`
- **THEN** the command outputs a JSON object containing `prompt` (string) and `data` (object with raw analysis results)

#### Scenario: Invalid profile file
- **WHEN** user runs `agent-insight diagnose not_a_profile.txt`
- **THEN** the command exits with an error message indicating the file could not be parsed as a pprof profile

#### Scenario: SKILL.md documents exploratory trigger
- **WHEN** the SKILL.md template is generated from the skill template
- **THEN** the trigger table entry for `diagnose` describes exploratory/uncertain intent scenarios (e.g., "用户不确定问题在哪，需要全面概览")

#### Scenario: SKILL.md documents when not to use diagnose
- **WHEN** the SKILL.md template is generated from the skill template
- **THEN** the diagnose section includes a comparison list showing specific user intents that should use other commands instead
