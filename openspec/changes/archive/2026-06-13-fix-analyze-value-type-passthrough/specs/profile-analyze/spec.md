## ADDED Requirements

### Requirement: Value-type flag must actually change analysis output
The system MUST honor the `--value-type` flag by analyzing samples using the user-specified value type instead of falling back to the default.

#### Scenario: Different value types produce different hotspot rankings
- **WHEN** user runs `agent-insight analyze heap.pb.gz --value-type alloc_objects` and again with `--value-type inuse_space`
- **THEN** the two outputs show different flat values for the same hotspots (because alloc_objects counts objects while inuse_space counts bytes)
- **AND** the JSON output's `analyzed_type` field reflects the user-specified value type

#### Scenario: Invalid value-type returns clear error
- **WHEN** user runs `agent-insight analyze heap.pb.gz --value-type nonexistent_metric`
- **THEN** system exits with non-zero status
- **AND** error message lists the available value types from the profile