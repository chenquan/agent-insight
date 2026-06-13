## ADDED Requirements

### Requirement: Mapping file paths are normalized to basename
The system SHALL normalize mapping `File` fields to `filepath.Base()` form so that the same binary produces consistent output across profiles collected at different times or locations.

#### Scenario: Absolute path is reduced to basename
- **WHEN** profile mapping has `File = "/home/user/cppbench_server_main"`
- **THEN** system outputs `cppbench_server_main`

#### Scenario: Bracketed identifiers are preserved
- **WHEN** profile mapping has `File = "[vdso]"` or `"[vsyscall]"`
- **THEN** system outputs the same bracketed identifier (unchanged)

#### Scenario: Empty file field stays empty
- **WHEN** profile mapping has empty `File`
- **THEN** system outputs empty string (no placeholder)

#### Scenario: Module field in analyze hotspot uses normalized file
- **WHEN** analyze output includes `module` field for a hotspot without symbols
- **THEN** `module` value is the basename of the mapping file

#### Scenario: Mappings list in info output uses normalized file
- **WHEN** info command outputs `mappings[]` array
- **THEN** each `file` value is the basename of the original mapping file