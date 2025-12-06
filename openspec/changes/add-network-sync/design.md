# Design: Network Synchronization

## Context

Local-first foundation complete (97 tests passing). All data uses Any-Sync structures ready for network sync. Need to add coordinator integration, peer management, and sync protocols without breaking local-only usage.

**Constraints:**
- Must remain backwards compatible (local-only mode default)
- Must handle offline/online transitions gracefully
- Must work within existing single-dispatch architecture
- gomobile compatible (no complex Go types in FFI)

**Stakeholders:**
- Users wanting multi-device sync
- Users wanting collaboration/sharing
- Plugin maintainers (keep complexity in Go layer)

## Goals / Non-Goals

**Goals:**
- Enable peer-to-peer synchronization via Any-Sync coordinator
- Support join/leave shared spaces
- Provide sync control (start/pause/status)
- Graceful offline/online handling
- Basic conflict resolution (last-write-wins)

**Non-Goals:**
- Advanced conflict resolution UI (app responsibility)
- Custom sync protocols (Any-Sync only)
- Zero-config p2p (require coordinator address)
- Breaking local-only mode

## Decisions

### Network Mode Default

**Decision:** Default to LocalOnly, require explicit NetworkEnabled config

**Why:**
- Backwards compatible with existing apps
- No surprise network connections
- Users must provide coordinator address explicitly
- Clear security model

**Alternatives:**
- Auto-enable network: Breaks existing apps, unclear where to get coordinator address
- Always require mode: Verbose for local-only use case

### Conflict Resolution Strategy

**Decision:** Start with last-write-wins (LWW), emit conflict events

**Why:**
- Simple to implement and reason about
- Sufficient for single-user multi-device
- Any-Sync supports operational transformation for future enhancement
- Apps can handle conflicts via events

**Alternatives:**
- Full OT: Complex, can defer until multi-user collaboration needed
- Manual resolution: Requires UI work, breaks automatic sync

### Coordinator Availability Handling

**Decision:** Queue operations when coordinator unavailable, don't fail

**Why:**
- Users can work offline indefinitely
- Sync resumes automatically when connectivity returns
- Matches mobile app expectations
- Avoids "must be online" failure modes

**Alternatives:**
- Fail hard: Poor offline UX
- Time-bound queue: Unclear what timeout is appropriate

### SyncTree Integration

**Decision:** Wrap existing ObjectTrees with SyncTree on space open

**Why:**
- Minimal code changes to existing handlers
- Automatic sync on document changes
- Any-Sync's proven pattern (used in Anytype)
- Can toggle sync on/off per space

**Alternatives:**
- Separate sync API calls: More control but requires manual sync triggering
- Always-on sync: No per-space control, wastes bandwidth

## Risks / Trade-offs

**Risks:**
- **Coordinator dependency**: Need valid coordinator endpoint
  - Mitigation: Provide test/dev coordinator, queue operations offline
- **NAT traversal**: Some peer connections may fail
  - Mitigation: Use coordinator relaying (Any-Sync supports this)
- **Sync conflicts**: Must handle gracefully
  - Mitigation: Start with LWW, emit conflict events for app handling
- **Testing complexity**: Need multi-node infrastructure
  - Mitigation: Integration tests with mock coordinator, defer E2E

**Trade-offs:**
- Network overhead (battery/bandwidth) vs automatic sync
  - Mitigation: Pause sync API, per-space control
- Complexity in Go layer vs thin clients
  - Chosen: Keep complexity in Go (matches architecture)

## Migration Plan

**No data migration required.** All existing data uses Any-Sync structures.

**For apps not using network:**
- No changes required (LocalOnly is default)

**For apps adding network:**
1. Add network config to Init call
2. Handle network events if desired
3. Optionally add sync control UI

**Rollback:**
- Remove network config from Init
- Plugin reverts to local-only mode
- All local data remains intact

## Open Questions

- Should we expose Any-Sync network ID in config? Or hardcode to "anytype" network?
- Do we need invite token expiration/revocation in MVP?
- Should GetSyncStatus return per-document status or just per-space aggregate?
