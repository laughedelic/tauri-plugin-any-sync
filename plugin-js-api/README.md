# TypeScript API

Promise-based TypeScript API for tauri-plugin-any-sync, generated from protobuf definitions.

## Usage

```typescript
import { syncspace } from 'tauri-plugin-any-sync-api'

// Initialize
await syncspace.init({ dataDir: '/path/to/data', networkId: 'local' })

// Space operations
const space = await syncspace.createSpace({ name: 'My Space' })
const spaces = await syncspace.listSpaces({})

// Document operations
const doc = await syncspace.createDocument({
  spaceId: space.spaceId,
  title: 'Note',
  data: new TextEncoder().encode(JSON.stringify({ content: 'Hello' }))
})
```

## Implementation

**Binary transport**: Protobuf messages serialized to `Uint8Array` and passed directly to Tauri via `ipc::Request` (no JSON conversion).

**Generated code**: Client methods auto-generated from `syncspace.proto` using `scripts/generate_api.ts`.

## Structure

```
plugin-js-api/
├── src/
│   └── index.ts        # API functions
├── dist/               # Build output (git-ignored)
├── package.json
```

See [root README.md](../README.md) for API documentation.
