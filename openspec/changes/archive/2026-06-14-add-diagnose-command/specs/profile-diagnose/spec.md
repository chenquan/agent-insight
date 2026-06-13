## ADDED Requirements

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

### Requirement: Language detection from profile data

The system SHALL analyze Function.Name and Function.Filename from the pprof profile to detect the programming language of the profiled program.

The system SHALL support detection of the following languages: Go, C++, Rust, Java, C, and Unknown.

Detection SHALL use pattern matching on function names and file extensions, selecting the language with the most matches across all functions.

When the profile has no function symbols (all locations are address-only), the system SHALL report language as Unknown.

The detected language SHALL be included in the prompt output as part of the profile metadata section.

#### Scenario: Go program detection
- **WHEN** a profile contains functions like `runtime.mallocgc`, `main.main`, `encoding/json.Marshal`
- **THEN** the system detects language as "Go" and includes "Go 性能诊断专家" in the role definition

#### Scenario: C++ program detection
- **WHEN** a profile contains functions like `std::vector<int>::push_back`, `MyClass::process`
- **THEN** the system detects language as "C++"

#### Scenario: Rust program detection
- **WHEN** a profile contains functions like `core::alloc::alloc`, `<T as std::fmt::Display>::fmt`
- **THEN** the system detects language as "Rust"

#### Scenario: Java program detection
- **WHEN** a profile contains functions like `java.lang.String.getBytes`, `com.example.Service.process`
- **THEN** the system detects language as "Java"

#### Scenario: Profile without symbols
- **WHEN** a profile contains only address-based locations with no function names
- **THEN** the system detects language as "Unknown" and uses generic diagnostic guidance

### Requirement: Profile-type-specific diagnostic guidance

The system SHALL generate different diagnostic guidance based on the profile type (cpu, heap, goroutine, contentions, thread, or unknown).

The guidance SHALL instruct the AI to focus on dimensions relevant to the specific profile type:

- **CPU**: computational hotspots, algorithm efficiency, GC pressure, concurrency efficiency
- **Heap**: memory allocation patterns, leak indicators, allocation optimization, GC impact
- **Goroutine**: goroutine count anomalies, blocking patterns, leak indicators, concurrency structure
- **Contentions**: lock contention hotspots, hold time, alternative concurrency primitives
- **Thread**: thread creation patterns, excessive thread count
- **Unknown**: general analysis based on function names and call paths

#### Scenario: CPU profile guidance
- **WHEN** the profile type is "cpu"
- **THEN** the diagnostic guidance instructs analysis of computational hotspots, algorithm efficiency, GC pressure, and concurrency efficiency

#### Scenario: Heap profile guidance
- **WHEN** the profile type is "heap"
- **THEN** the diagnostic guidance instructs analysis of memory allocation patterns, leak indicators, allocation optimization, and GC impact

#### Scenario: Goroutine profile guidance
- **WHEN** the profile type is "goroutine"
- **THEN** the diagnostic guidance instructs analysis of goroutine count, blocking patterns, leak indicators, and concurrency structure

#### Scenario: Unknown profile type guidance
- **WHEN** the profile type does not match any known type
- **THEN** the diagnostic guidance uses generic analysis instructions

### Requirement: Language-specific diagnostic additions

The system SHALL append language-specific diagnostic guidance to the profile-type-specific base guidance.

The language addition SHALL highlight optimization techniques and patterns specific to the detected language:

- **Go**: sync.Pool, slice pre-allocation, GC pressure signals from runtime functions, string/[]byte conversion, goroutine concurrency
- **C++**: virtual function overhead, cache locality, SIMD/OpenMP opportunities, custom allocators, RAII patterns
- **Rust**: clone overhead vs borrow, unsafe code hotspots, iterator vs manual loop, async runtime overhead, Arc/Rc allocation
- **Java**: JIT compilation status, GC strategy impact (G1/ZGC/Shenandoah), object lifecycle, thread blocking, static collection leaks
- **C**: malloc/free patterns, buffer management, struct layout optimization

When language is Unknown, no language-specific addition SHALL be appended.

#### Scenario: Go language addition for CPU profile
- **WHEN** the profile is CPU type and language is detected as Go
- **THEN** the prompt includes Go-specific guidance mentioning sync.Pool, slice pre-allocation, and runtime function analysis

#### Scenario: Unknown language produces no language addition
- **WHEN** the profile language is detected as Unknown
- **THEN** the prompt contains only the base profile-type guidance without language-specific additions

### Requirement: Prompt output structure

The diagnostic prompt SHALL follow this structure:
1. Role definition — dynamically generated based on detected language
2. Profile metadata — type, sample count, duration, value types, function count, symbol status, detected language
3. Analysis data — top N hotspot functions (flat/cum/percent/file), call tree summary (cum >= 1% paths), top 5 traces by value
4. Diagnostic guidance — base guidance (profile type) + language addition
5. Output format requirements — structured output instructions for the AI
6. User context — only when `--context` flag is provided

#### Scenario: Complete prompt structure for CPU profile of Go program
- **WHEN** user runs `agent-insight diagnose cpu.pb.gz --context "HTTP API server"`
- **THEN** the output contains all 6 sections with Go-specific role, CPU-specific guidance with Go language addition, and user context

#### Scenario: Prompt without user context
- **WHEN** user runs `agent-insight diagnose cpu.pb.gz` without `--context`
- **THEN** the output contains 5 sections (sections 1-5), omitting the user context section

### Requirement: Reuse existing profile analysis capabilities

The `diagnose` command SHALL reuse the existing `profile/` package functions for extracting analysis data:
- `Analyze()` for hotspot functions and call stack paths
- `BuildTree()` for call tree structure
- `GetTopTraces()` for individual sample traces

The diagnose logic SHALL NOT duplicate analysis computation already provided by these functions.

#### Scenario: Diagnose uses existing analysis
- **WHEN** the diagnose command processes a profile
- **THEN** it calls profile.Analyze(), profile.BuildTree(), and profile.GetTopTraces() to extract data, then formats the results into a prompt
