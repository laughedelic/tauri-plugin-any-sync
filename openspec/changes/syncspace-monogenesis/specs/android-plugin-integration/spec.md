# Android Plugin Integration Specification

## REMOVED Requirements

### Requirement: Per-Operation Native Methods

~~The Kotlin plugin SHALL define native methods for each storage operation.~~

**Reason:** Replaced by single command method.

## MODIFIED Requirements

### Requirement: Native Library Loading

The Kotlin plugin SHALL load the gomobile-generated .aar library at initialization.

**Changes:**
- No structural change, but the native API surface is reduced to 4 functions
- Library loading mechanism remains the same

#### Scenario: .aar library is loaded on plugin initialization

- **GIVEN** the Android application initializes the plugin
- **WHEN** the plugin's initialization code runs
- **THEN** the any-sync-android.aar library is loaded successfully

### Requirement: JNI Bridging

The Kotlin plugin SHALL bridge Tauri plugin calls to Go JNI functions.

**Changes:**
- Bridge only 4 functions instead of N per-operation functions
- All operations go through single command bridge

#### Scenario: Command method bridges to JNI

- **GIVEN** the Kotlin plugin's command method is called
- **WHEN** the method executes
- **THEN** it invokes the Go JNI Command function

## ADDED Requirements

### Requirement: Minimal Kotlin Implementation

The Kotlin plugin implementation SHALL be minimal passthrough code under 50 lines.

#### Scenario: Kotlin plugin initializes native library

- **GIVEN** the plugin initialization
- **WHEN** the plugin is created
- **THEN** it loads the .aar and initializes the Go backend via Init JNI call

#### Scenario: Kotlin plugin forwards command calls

- **GIVEN** a command method call from Rust
- **WHEN** the Kotlin plugin processes it
- **THEN** it directly forwards cmd and data to Go Command JNI function with no additional logic

#### Scenario: Kotlin plugin forwards event handler

- **GIVEN** event handler registration from Rust
- **WHEN** the Kotlin plugin receives the registration
- **THEN** it forwards the handler callback to Go SetEventHandler JNI function

#### Scenario: Kotlin plugin contains no business logic

- **GIVEN** the Kotlin plugin implementation
- **WHEN** the code is reviewed
- **THEN** it contains only JNI bridging code, no business logic

#### Scenario: Kotlin plugin is under 50 lines

- **GIVEN** the Kotlin plugin implementation
- **WHEN** lines of code are counted (excluding comments and whitespace)
- **THEN** the implementation is under 50 lines
