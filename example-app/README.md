# Example Tauri App

Example application demonstrating tauri-plugin-any-sync usage.

## Quick Start

```bash
# Run development server
task app:dev

# Build for production
task app:build
```

## Development

The app uses the plugin from the workspace:
- Rust plugin: `path = "../../plugin-rust-core"`
- TypeScript API: `workspace:*`

## Structure

```
example-app/
├── src/                # Svelte frontend
│   └── App.svelte
├── src-tauri/          # Tauri backend
│   ├── gen/android/    # (git-ignored) Android native plugin (generated)
│   ├── src/lib.rs      # Plugin initialization
│   ├── Cargo.toml      # Plugin dependency
│   └── capabilities/   # Permissions
├── package.json
└── Taskfile.yml
```

See [root README.md](../README.md) for full documentation.
