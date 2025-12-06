# TypeScript API

Generated TypeScript client for tauri-plugin-any-sync from protobuf definitions.

## Usage

```typescript
import { SyncSpaceClient } from 'tauri-plugin-any-sync-api'

const client = new SyncSpaceClient()
await client.init({ dataDir: './data' })

const { spaceId } = await client.createSpace({ name: 'Notes' })
const { documentId } = await client.createDocument({
  spaceId,
  title: 'Note',
  data: new TextEncoder().encode(JSON.stringify({ content: 'Hello' }))
})
```

**Binary transport**: Protobuf bytes sent directly via Tauri IPC (no JSON).

**Code generation**: Client auto-generated from `buf/proto/syncspace-api/syncspace/v1/syncspace.proto` by a custom buf plugin script in `plugin-js-api/scripts/generate_api.ts` (runs via `task buf:generate-syncspace`).

See [root README](../README.md) and [example-app](../example-app/) for complete usage.
