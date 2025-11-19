# Project Context

## Purpose

A cross-platform Tauri plugin that provides unified storage and synchronization capabilities powered by AnySync/AnyStore. The plugin enables local-first storage with optional P2P synchronization across desktop (Windows, macOS, Linux) and mobile (iOS, Android) platforms through a single TypeScript API.

**Core Objectives:**
- Platform Abstraction: Hide platform-specific implementation details behind a unified TypeScript API
- Code Reuse: Maximize Go backend code sharing across all platforms (target: 95%+)
- Developer Experience: Allow application developers to work primarily in TypeScript
- Data Sovereignty: Enable local-first storage with optional P2P synchronization

## Tech Stack

**Frontend/Plugin Layer:**
- TypeScript (application API surface)
- Rust (Tauri plugin core - required by Tauri framework)
- Tauri v2.0+ (cross-platform framework with mobile support)

**Backend:**
- Go 1.21+ (AnySync/AnyStore integration)
- gRPC with Protocol Buffers (type-safe IPC/communication)

**Mobile Native:**
- Kotlin (Android plugin wrapper)
- Swift (iOS plugin wrapper)
- gomobile (Go→mobile FFI bindings)

**Build Tools:**
- gomobile bind (for Android .aar and iOS .xcframework)
- Standard Go build tools (for desktop sidecar executables)

## Project Conventions

### Code Style

**Rust:**
- Follow standard Rust conventions and rustfmt
- Use `#[cfg(target_os)]` for platform-specific code
- Organize into modules: `commands.rs`, `desktop.rs`, `mobile.rs`, `error.rs`

**Go:**
- Standard Go formatting with gofmt
- Internal packages use unrestricted Go features
- API boundary (`api/` layer) enforces gomobile compatibility constraints
- Use Protocol Buffers for all external API definitions

**TypeScript:**
- Async/Promise-based API design (all operations asynchronous)
- Single unified interface regardless of platform
- Event-driven patterns for synchronization updates
- Error types map to TypeScript classes

**Mobile Native (Kotlin/Swift):**
- Follow platform-specific conventions
- Thin wrapper layer around gomobile bindings
- Minimal logic - delegate to Go backend

### Architecture Patterns

**Desktop: Sidecar Process Pattern**
- Go backend compiled as standalone executable
- Bundled in `binaries/any-sync-<target-triple>`
- Tauri spawns sidecar on app startup
- Communication via gRPC over localhost
- No FFI complexity or toolchain conflicts

**Mobile: gomobile Embedded Library Pattern**
- Go backend compiled via `gomobile bind`
- Android: `.aar` library consumed by Kotlin plugin
- iOS: `.xcframework` consumed by Swift plugin
- Direct FFI integration (sandboxing prevents child processes)

**API Boundary Design:**
- gRPC/HTTP API layer is the natural boundary
- Complex internal Go types marshaled to simple, gomobile-compatible types
- Internal packages (`internal/`) use any Go features
- External API (`api/`) restricted to gomobile-compatible types:
  - Primitives: int, float, string, bool
  - Byte slices (read-only)
  - Exported structs with compatible fields
  - Complex structures JSON-serialized at boundary

**Project Structure:**
```
tauri-plugin-any-sync/
├── src/                  # Rust plugin core
│   ├── commands.rs       # Unified TypeScript API
│   ├── desktop.rs        # Sidecar management
│   ├── mobile.rs         # Platform dispatch
│   └── error.rs          # Error mapping
├── android/              # Kotlin plugin wrapper
├── ios/                  # Swift plugin wrapper
├── binaries/             # Pre-built Go sidecars
└── go-backend/
    ├── internal/         # Unrestricted Go code
    │   ├── sync/         # AnySync integration
    │   ├── storage/      # AnyStore logic
    │   └── models/       # Internal domain models
    ├── api/              # gomobile-compatible layer
    │   ├── proto/        # gRPC definitions
    │   └── server/       # gRPC server
    └── cmd/
        ├── server/       # Desktop sidecar entrypoint
        └── mobile/       # gomobile bind entrypoint
```

### Testing Strategy

- Platform-specific integration test suites (desktop, Android, iOS)
- Test Go backend independently before plugin integration
- Mock gRPC endpoints for unit testing Rust layer
- End-to-end tests on all 5 target platforms
- Focus on gomobile boundary and type marshaling edge cases

### Git Workflow

- Feature branches for platform-specific work
- Tag releases with platform compatibility notes
- CI/CD builds for all platforms
- Pre-commit hooks for formatting (rustfmt, gofmt)

## Domain Context

**AnySync/AnyStore:**
- AnySync: Decentralized synchronization protocol
- AnyStore: Local storage layer with sync capabilities
- Supports local-first architecture with optional P2P sync
- Written entirely in Go (hence the Go backend requirement)

**Tauri Plugin Architecture:**
- Plugins bridge Rust core with platform-specific capabilities
- Desktop: Rust can spawn processes and communicate via IPC
- Mobile: Requires native plugin layer (Kotlin/Swift) for platform APIs
- Commands are invoked from TypeScript frontend

**gomobile Constraints:**
- Limited to specific Go types for FFI compatibility
- No generics, no channels, no goroutines in exported APIs
- Must use interfaces with simple method signatures
- Complex data passed as JSON strings or byte slices

## Important Constraints

**Technical:**
- Must support Tauri v2.0+ (mobile support requirement)
- Go 1.21+ required for gomobile compatibility
- Android API 24+ (Android 7.0+)
- iOS 13.0+
- Desktop: No FFI - sidecar pattern only
- Mobile: Must use gomobile bindings (FFI required due to sandboxing)
- API must be gomobile-compatible at boundary layer

**Design:**
- Single TypeScript codebase must work on all 5 platforms
- >90% Go backend code sharing target
- Application developers should not need Rust, Kotlin, or Swift knowledge
- Plugin is generic and reusable (not application-specific)

**Out of Scope:**
- Web platform support (WebAssembly deferred)
- Direct Rust↔Go FFI on desktop
- Alternative sync protocols beyond AnySync
- UI components (data layer only)
- Built-in conflict resolution (application responsibility)

## External Dependencies

**Core Dependencies:**
- [AnySync](https://github.com/anyproto/any-sync) - Synchronization protocol
- [AnyStore](https://github.com/anyproto/any-store) - Storage layer
- [Tauri Framework](https://tauri.app/) - Cross-platform application framework
- gomobile - Go mobile bindings generator

**Communication:**
- gRPC - Type-safe IPC protocol
- Protocol Buffers - Serialization format

**References:**
- [AnySync Overview](https://tech.anytype.io/any-sync/overview)
- [Tauri Plugin Development](https://tauri.app/develop/plugins)
- [Tauri Mobile Plugin Development](https://tauri.app/develop/plugins/develop-mobile)

## Development Phases

**Phase 0**: Scaffold minimal project with all wiring but no functionality  
**Phase 1**: Desktop-only plugin with sidecar pattern  
**Phase 2**: Android support via gomobile  
**Phase 3**: iOS support via gomobile  
**Phase 4**: Advanced features (streaming sync, conflict resolution)

## Success Criteria

1. ✅ Single TypeScript codebase works on all 5 platforms
2. ✅ Go backend code >90% shared between desktop and mobile
3. ✅ Application developers don't need to write Rust, Kotlin, or Swift
4. ✅ All AnySync/AnyStore features accessible via plugin
5. ✅ Zero FFI code in desktop builds
6. ✅ Plugin is generic and reusable by other projects
