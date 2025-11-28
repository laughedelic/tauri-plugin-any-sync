# TypeScript API

Promise-based TypeScript API for tauri-plugin-any-sync.

## Usage

```typescript
import {
  ping,
  storagePut,
  storageGet,
  storageDelete,
  storageList
} from 'tauri-plugin-any-sync-api'

// Health check
const response = await ping('Hello')

// CRUD operations
await storagePut('users', 'user123', { name: 'Alice' })
const user = await storageGet('users', 'user123')
const deleted = await storageDelete('users', 'user123')
const ids = await storageList('users')
```

## Structure

```
plugin-js-api/
├── src/
│   └── index.ts        # API functions
├── dist/               # Build output (git-ignored)
├── package.json
```

See [root README.md](../README.md) for API documentation.
