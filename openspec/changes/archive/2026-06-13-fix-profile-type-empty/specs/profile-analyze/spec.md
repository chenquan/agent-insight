## ADDED Requirements

### Requirement: Profile type fallback when PeriodType is missing
The system MUST infer profile type from `SampleType` when `PeriodType` is nil or its type is empty, instead of returning an empty string.

#### Scenario: Heap profile with missing PeriodType reports heap
- **WHEN** user analyzes a heap profile where `PeriodType` is nil but `SampleType` contains `inuse_space` and `alloc_objects`
- **THEN** JSON output's `type` field is `heap` (not empty string)

#### Scenario: Goroutine profile reports goroutine
- **WHEN** user analyzes a goroutine profile
- **THEN** JSON output's `type` field is `goroutine`

#### Scenario: Unrecognized profile reports unknown
- **WHEN** user analyzes a profile with SampleType that contains no recognizable keywords
- **THEN** JSON output's `type` field is `unknown` (not empty string)

#### Scenario: PeriodType present still wins
- **WHEN** user analyzes a profile with `PeriodType.Type = "cpu"` and `SampleType` containing `inuse_space`
- **THEN** JSON output's `type` field is `cpu` (PeriodType takes precedence over inference)