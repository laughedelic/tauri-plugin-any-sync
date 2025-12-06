# android-plugin-integration Specification

## Purpose
Defines the Android Kotlin plugin layer that bridges Tauri to the gomobile-compiled Go backend. Provides minimal passthrough code that loads the native .aar library and forwards all plugin commands through the single-dispatch interface.
## Requirements
### Requirement: gomobile Library Loading

The Android plugin SHALL load the gomobile-generated native library at initialization.

**Changes:**
- No structural change, but the native API surface is reduced to 4 functions (Init, Command, SetEventHandler, Shutdown)
- Library loading mechanism remains the same

#### Scenario: Native library initialization

- **GIVEN** an Android app using the plugin
- **WHEN** the plugin class is first loaded
- **THEN** the plugin loads the `gojni` native library via `System.loadLibrary("gojni")`
- **AND** loading happens in a static initializer block before any instance methods
- **AND** loading failure throws UnsatisfiedLinkError with descriptive message
- **AND** the error is logged to Android logcat for debugging

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

