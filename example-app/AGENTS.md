# AnySync Notes - Development Guide

This guide covers development principles and testing for the example notes application.

## UI Design Principles

### Philosophy
The app demonstrates plugin functionality through a **minimal, Apple Notes-inspired interface**. The design prioritizes clarity, responsiveness, and ease of use.

### Key Principles

1. **Responsive Layout**
   - Two-column layout on desktop (sidebar + editor)
   - Single-column with slide-out sidebar on mobile
   - Smooth animations using Framer Motion

2. **Component Architecture**
   - React 18 with functional components
   - CSS Modules for scoped styles
   - Zustand for state management
   - Clear separation: components, store, services, hooks

3. **Visual Hierarchy**
   - Sidebar for note list and navigation
   - Main content area for editing
   - Status indicators for save state

4. **Interaction Patterns**
   - **Click to select** - Notes load automatically when clicked
   - **Auto-save** - Changes saved 500ms after typing stops
   - **Optimistic updates** - UI updates immediately, rolls back on error
   - **Inline feedback** - Save status shown in editor header

5. **Feedback & Messaging**
   - Save status indicator: "Saving...", "Saved", or error state
   - Visual distinction via color (green for saved, red for error)
   - Loading spinner during initialization

6. **Styling Approach**
   - **CSS Modules**: Scoped styles per component
   - **CSS Variables**: Design tokens in `index.css`
   - **System fonts**: -apple-system stack for native feel
   - **Apple-inspired colors**: Subtle grays, blue accent

7. **State Management**
   - Zustand store for global state
   - Local component state for form inputs
   - Custom hooks for auto-save and media queries

8. **User Experience**
   - Auto-initialize on mount
   - Empty states for no notes / no selection
   - Mobile-first responsive behavior

### Code Organization

```
src/
├── components/           # React components with CSS Modules
│   ├── Layout/          # Responsive layout container
│   ├── Sidebar/         # Notes list + new button
│   ├── NoteList/        # Animated list with items
│   ├── Editor/          # Note editor with auto-save
│   └── EmptyState/      # Placeholder states
├── store/               # Zustand state management
│   └── useNotesStore.ts
├── services/            # Plugin integration layer
│   └── notes.ts
├── hooks/               # Custom React hooks
│   ├── useAutoSave.ts
│   └── useMediaQuery.ts
├── lib/                 # Utilities
│   └── formatDate.ts
├── App.tsx              # Root component
├── main.tsx             # Entry point
└── index.css            # CSS reset + design tokens
```

## Quick Start

```bash
cd example-app
bun install
task app:dev
```

## Testing Workflow

### 1. Plugin Integration Testing

The app demonstrates complete any-sync plugin integration:

```typescript
import { syncspace } from 'tauri-plugin-any-sync-api';

// Initialize
await syncspace.init({ dataDir, networkId: "local", deviceId: "example-app" });

// CRUD operations
const { documentId } = await syncspace.createDocument({ spaceId, data, metadata });
const { document } = await syncspace.getDocument({ spaceId, documentId });
await syncspace.updateDocument({ spaceId, documentId, data, metadata });
await syncspace.deleteDocument({ spaceId, documentId });
```

### 2. End-to-End Testing

Test the complete communication flow:

1. **UI Layer**: React component triggers action
2. **Store Layer**: Zustand calls NotesService
3. **Service Layer**: NotesService calls syncspace API
4. **Rust Layer**: Plugin routes to Go backend
5. **Go Backend**: Processes request and persists data

### 3. Expected Behavior

When creating a note:
1. Click "+" button in sidebar
2. New empty note appears and is selected
3. Type title and content
4. "Saving..." appears in header
5. "Saved" appears after 500ms debounce

### 4. Testing Checklist

#### Basic Functionality
- [ ] App starts successfully
- [ ] Plugin initializes without errors
- [ ] Create new note works
- [ ] Edit note auto-saves
- [ ] Delete note works
- [ ] Notes persist across restarts

#### Responsive Design
- [ ] Desktop: sidebar always visible
- [ ] Mobile (<768px): sidebar slides out
- [ ] Animations smooth on both

#### State Management
- [ ] Save status updates correctly
- [ ] Optimistic updates work
- [ ] Error rollback works

#### Performance
- [ ] Auto-save debounced properly
- [ ] No memory leaks
- [ ] Animations smooth (60fps)

## Development Commands

```bash
# Development mode
task app:dev

# Production build
task app:build

# Type checking
bunx tsc --noEmit

# Run on Android
task app:dev:android
```

## Troubleshooting

### Common Issues

1. **TypeScript errors**
   ```bash
   bunx tsc --noEmit
   ```
   Check for type issues in components or store.

2. **CSS Modules not working**
   Ensure `.module.css` extension and proper import:
   ```typescript
   import styles from "./Component.module.css";
   ```

3. **Plugin not initializing**
   Check Tauri permissions in `src-tauri/capabilities/`.

4. **Mobile sidebar not working**
   Verify `useMediaQuery` hook and Framer Motion setup.
