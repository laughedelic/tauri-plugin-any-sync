# AnySync Notes

A minimalist notes app showcasing tauri-plugin-any-sync, inspired by Apple Notes.

## Features

- Clean, adaptive UI (desktop sidebar + mobile slide-out)
- Auto-save with status indicator
- Create, edit, and delete notes
- Notes persist via any-sync storage backend

## Quick Start

```bash
task app:dev          # Run development server
task app:dev:android  # Run on Android emulator
task app:build        # Build for production
```

## Architecture

### Tech Stack

- **React 18** - UI framework
- **Zustand** - State management
- **Framer Motion** - Animations
- **CSS Modules** - Scoped styling
- **Vite** - Build tool

### Structure

```
src/
├── components/           # React components
│   ├── Layout/          # Responsive layout container
│   ├── Sidebar/         # Notes list sidebar
│   ├── NoteList/        # Animated note list
│   ├── Editor/          # Note editor with auto-save
│   └── EmptyState/      # Placeholder states
├── store/               # Zustand state management
│   └── useNotesStore.ts
├── services/            # Plugin integration
│   └── notes.ts         # Domain service layer
├── hooks/               # Custom React hooks
└── lib/                 # Utilities
```

### Domain Service Pattern

The app uses a domain service layer (`src/services/notes.ts`) to wrap the SyncSpace client:

```typescript
// NotesService encapsulates SyncSpace operations
class NotesService {
  async createNote(note: Note): Promise<string> {
    const data = new TextEncoder().encode(JSON.stringify(note));
    return await syncspace.createDocument({
      spaceId: this.spaceId,
      collection: "notes",
      data,
      metadata: { title: note.title }
    });
  }
}
```

**Benefits:**
- Encapsulates data encoding/decoding (opaque bytes ↔ typed Note objects)
- Provides app-specific API (createNote vs createDocument)
- Handles space/collection management
- Centralizes error handling

### State Management

Zustand store (`src/store/useNotesStore.ts`) manages:
- Notes list with optimistic updates
- Selection state
- Save status (idle/saving/saved/error)
- Plugin initialization

### Auto-save

The `useAutoSave` hook debounces changes (500ms delay) and automatically saves:
- Updates saveStatus in store for UI feedback
- Handles errors gracefully
