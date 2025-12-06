# Implementation Tasks

## Phase 1: Network Configuration

- [ ] 1.1 Create `plugin-go-backend/shared/anysync/network.go` with NetworkConfig struct and NetworkMode enum
- [ ] 1.2 Update `syncspace.proto` InitRequest to include optional network_config field
- [ ] 1.3 Update Init handler to parse and validate network configuration
- [ ] 1.4 Write unit tests for network config (4 tests: valid parse, invalid config, default to LocalOnly, required fields)

## Phase 2: Coordinator Integration

- [ ] 2.1 Create `plugin-go-backend/shared/anysync/coordinator.go` with coordinator client initialization
- [ ] 2.2 Implement space registration with coordinator
- [ ] 2.3 Update SpaceManager to register new spaces when NetworkEnabled
- [ ] 2.4 Implement offline mode (queue registration attempts)
- [ ] 2.5 Write unit tests for coordinator (6 tests: registration success/failure, offline mode, re-registration on reconnect)

## Phase 3: Peer Management

- [ ] 3.1 Create `plugin-go-backend/shared/anysync/peers.go` with peer manager initialization
- [ ] 3.2 Implement peer discovery via coordinator
- [ ] 3.3 Implement peer connection lifecycle (connect/disconnect/reconnect)
- [ ] 3.4 Write unit tests for peers (6 tests: discovery, connection, disconnect handling, reconnection)

## Phase 4: Sync Protocols

- [ ] 4.1 Create `plugin-go-backend/shared/anysync/sync.go` with SyncTree wrapper
- [ ] 4.2 Wrap existing ObjectTrees with SyncTree on space open
- [ ] 4.3 Implement HeadSync integration (exchange heads, detect missing changes)
- [ ] 4.4 Implement ObjectSync integration (stream changes, merge remote, handle conflicts as LWW)
- [ ] 4.5 Write unit tests for sync (8 tests: SyncTree wraps ObjectTree, local changes trigger sync, remote changes received, conflict resolution, offline/online transitions)

## Phase 5: Network Space Operations

- [ ] 5.1 Update `syncspace.proto` with JoinSpace/LeaveSpace RPCs and messages
- [ ] 5.2 Create `plugin-go-backend/shared/handlers/join_space.go` (accept invite, fetch from peers, init local storage, join ACL)
- [ ] 5.3 Create `plugin-go-backend/shared/handlers/leave_space.go` (notify coordinator, remove from ACL, optionally delete local)
- [ ] 5.4 Register JoinSpace/LeaveSpace handlers in registry
- [ ] 5.5 Write unit tests for space ops (8 tests: valid/invalid invite, leave success, fetch from peers, peer unavailability)

## Phase 6: Sync Control Operations

- [ ] 6.1 Update `syncspace.proto` with StartSync/PauseSync/GetSyncStatus RPCs and messages
- [ ] 6.2 Create `plugin-go-backend/shared/handlers/start_sync.go` (enable sync loops, emit sync.started)
- [ ] 6.3 Create `plugin-go-backend/shared/handlers/pause_sync.go` (disable sync, emit sync.paused)
- [ ] 6.4 Create `plugin-go-backend/shared/handlers/sync_status.go` (query state, pending count, last sync, errors)
- [ ] 6.5 Register sync control handlers in registry
- [ ] 6.6 Write unit tests for sync control (8 tests: start/pause behavior, status accuracy, per-space status, error reporting)

## Phase 7: Network Events

- [ ] 7.1 Extend EventManager with network event types (sync.*, peer.*)
- [ ] 7.2 Hook sync state change events in sync.go
- [ ] 7.3 Hook peer connection events in peers.go
- [ ] 7.4 Hook conflict detection events in sync.go (ObjectSync merge conflicts)
- [ ] 7.5 Write unit tests for network events (6 tests: sync events emitted, peer events emitted, conflict events, filtering)

## Phase 8: Integration Testing

- [ ] 8.1 Write two-node sync tests (12 tests: create space on A syncs to B, document sync, concurrent edits, offline catchup, multiple spaces, join/leave, pause/resume)
- [ ] 8.2 Write network failure tests (6 tests: coordinator unavailable, peer unavailable, intermittent network, sync error recovery)
- [ ] 8.3 Validate full network stack builds on all platforms
- [ ] 8.4 Verify no regression in local-only mode

## Dependencies

**Sequential:**
- Phase 1 → Phase 2 (config needed for coordinator)
- Phase 2 → Phase 3 (coordinator needed for peer discovery)
- Phase 3 → Phase 4 (peers needed for sync)
- Phase 4 → Phase 5 (sync wrapper needed for JoinSpace)
- Phase 5,6,7 can be parallel after Phase 4
- Phase 8 requires all phases complete

**Test Coverage Target:**
- Network config: 4 tests
- Coordinator: 6 tests
- Peers: 6 tests
- Sync protocols: 8 tests
- Space ops: 8 tests
- Sync control: 8 tests
- Events: 6 tests
- Integration: 18 tests
- **Total: ~64 new tests**
