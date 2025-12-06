# Change: Add Network Synchronization

## Why

The plugin operates in local-first mode only. All spaces and documents use full Any-Sync cryptographic structures (keys, ACLs, ObjectTree) but peer-to-peer synchronization is not implemented.

Users need to sync across devices, share spaces via invites, and collaborate with conflict resolution.

## What Changes

**Network Infrastructure:**
- Add network configuration to Init (coordinator address, mode: LocalOnly/NetworkEnabled)
- Initialize coordinator client for peer discovery
- Initialize peer manager for connection lifecycle
- Wrap ObjectTrees with SyncTree for automatic HeadSync/ObjectSync

**New Operations:**
- `JoinSpace(inviteToken)` - Join shared space from invite
- `LeaveSpace(spaceId, deleteLocal)` - Leave shared space
- `StartSync(spaceIds)` - Enable sync for spaces
- `PauseSync(spaceIds)` - Pause sync temporarily
- `GetSyncStatus(spaceIds)` - Query sync state/pending changes

**Network Events:**
- `sync.{started,completed,paused,error,conflict}`
- `peer.{connected,disconnected}`

## Impact

**Affected Specs:**
- `syncspace-api` (MODIFIED) - Add network operations
- `any-sync-integration` (MODIFIED) - Add coordinator, peer manager, sync protocols
- `go-backend-scaffolding` (MODIFIED) - Add network config parsing

**New Specs:**
- `network-sync` - Network synchronization requirements

**Code Changes:**
- ~2500 new lines (network components + handlers + tests)
- ~64 new tests (network config, coordinator, peers, sync, integration)
- Modify: `syncspace.proto`, `init.go`, `spaces.go`, `events.go`
- No changes to: TypeScript client (auto-generated), Rust plugin, mobile FFI

**Not Affected:**
- Single-dispatch architecture
- Local-only mode (default, backwards compatible)
- Build system and binary distribution
