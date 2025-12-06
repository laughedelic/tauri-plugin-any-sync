# Example Tauri App

Reference implementation showing tauri-plugin-any-sync usage.

## Quick Start

```bash
task app:dev          # Run development server
task app:dev:android  # Run on Android emulator
task app:build        # Build for production
```

## Domain Service Pattern

The app demonstrates recommended architecture using a domain service layer (`src/services/notes.ts`) that wraps the SyncSpace client:

```typescript
// Domain service encapsulates SyncSpace operations
class NotesService {
  private client = new SyncSpaceClient()
  private spaceId?: string

  async createNote(title: string, content: string) {
    const data = new TextEncoder().encode(JSON.stringify({ title, content }))
    return await this.client.createDocument({
      spaceId: this.spaceId!,
      title,
      data
    })
  }
}
```

**Why domain services?**
- Encapsulate data encoding/decoding (opaque bytes ↔ app-specific types)
- Provide app-specific API (createNote vs createDocument)
- Handle space/collection management
- Centralize error handling

See `src/services/notes.ts` and `src/App.svelte` for complete implementation.

## Structure

```
example-app/
├── src/
│   ├── App.svelte           # UI
│   └── services/notes.ts    # Domain service (recommended pattern)
├── src-tauri/
│   ├── src/lib.rs           # Plugin initialization
│   └── capabilities/        # Permissions
```
