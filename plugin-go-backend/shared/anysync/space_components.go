package anysync

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	anystore "github.com/anyproto/any-store"
	"github.com/anyproto/any-sync/accountservice"
	"github.com/anyproto/any-sync/app"
	"github.com/anyproto/any-sync/commonspace/config"
	"github.com/anyproto/any-sync/commonspace/object/accountdata"
	"github.com/anyproto/any-sync/commonspace/object/tree/objecttree"
	"github.com/anyproto/any-sync/commonspace/object/tree/treestorage"
	"github.com/anyproto/any-sync/commonspace/object/treemanager"
	"github.com/anyproto/any-sync/commonspace/peermanager"
	"github.com/anyproto/any-sync/commonspace/spacestorage"
	"github.com/anyproto/any-sync/net/peer"
	"github.com/anyproto/any-sync/net/pool"
	"github.com/anyproto/any-sync/nodeconf"
	"github.com/anyproto/go-chash"
	"storj.io/drpc"
)

// localSpaceStorageProvider implements SpaceStorageProvider for local-only operation.
type localSpaceStorageProvider struct {
	mu         sync.Mutex
	storageDir string
	databases  map[string]anystore.DB
	storages   map[string]spacestorage.SpaceStorage
}

func newLocalSpaceStorageProvider(storageDir string) *localSpaceStorageProvider {
	return &localSpaceStorageProvider{
		storageDir: storageDir,
		databases:  make(map[string]anystore.DB),
		storages:   make(map[string]spacestorage.SpaceStorage),
	}
}

func (p *localSpaceStorageProvider) Name() string                  { return spacestorage.CName }
func (p *localSpaceStorageProvider) Init(a *app.App) error         { return nil }
func (p *localSpaceStorageProvider) Run(ctx context.Context) error { return nil }
func (p *localSpaceStorageProvider) Close(ctx context.Context) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Close all storages
	for _, storage := range p.storages {
		storage.Close(ctx)
	}

	// Close all databases
	for _, db := range p.databases {
		db.Close()
	}

	p.databases = make(map[string]anystore.DB)
	p.storages = make(map[string]spacestorage.SpaceStorage)
	return nil
}

func (p *localSpaceStorageProvider) SpaceExists(id string) bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	_, exists := p.storages[id]
	return exists
}

func (p *localSpaceStorageProvider) CreateSpaceStorage(ctx context.Context, payload spacestorage.SpaceStorageCreatePayload) (spacestorage.SpaceStorage, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	spaceID := payload.SpaceHeaderWithId.Id

	// Check if already exists
	if _, exists := p.storages[spaceID]; exists {
		return nil, fmt.Errorf("space storage already exists: %s", spaceID)
	}

	// Open database for this space
	dbPath := filepath.Join(p.storageDir, spaceID+".db")
	db, err := anystore.Open(ctx, dbPath, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Create space storage
	storage, err := spacestorage.Create(ctx, db, payload)
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to create space storage: %w", err)
	}

	// Register
	p.databases[spaceID] = db
	p.storages[spaceID] = storage

	return storage, nil
}

func (p *localSpaceStorageProvider) WaitSpaceStorage(ctx context.Context, id string) (spacestorage.SpaceStorage, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Check if already loaded
	if storage, exists := p.storages[id]; exists {
		return storage, nil
	}

	// Open database
	dbPath := filepath.Join(p.storageDir, id+".db")
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("space storage not found: %s", id)
	}

	db, err := anystore.Open(ctx, dbPath, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Open storage
	storage, err := spacestorage.New(ctx, id, db)
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to open space storage: %w", err)
	}

	// Register
	p.databases[id] = db
	p.storages[id] = storage

	return storage, nil
}

// localAccountService implements accountservice.Service for local-only operation.
type localAccountService struct {
	keys *accountdata.AccountKeys
}

func newLocalAccountService(keys *accountdata.AccountKeys) *localAccountService {
	return &localAccountService{keys: keys}
}

func (s *localAccountService) Name() string                    { return accountservice.CName }
func (s *localAccountService) Init(a *app.App) error           { return nil }
func (s *localAccountService) Run(ctx context.Context) error   { return nil }
func (s *localAccountService) Close(ctx context.Context) error { return nil }

func (s *localAccountService) Account() *accountdata.AccountKeys {
	return s.keys
}

func (s *localAccountService) SignData(data []byte) (signature []byte, err error) {
	return s.keys.SignKey.Sign(data)
}

// noOpTreeManager is a minimal TreeManager implementation.
type noOpTreeManager struct{}

func newNoOpTreeManager() *noOpTreeManager { return &noOpTreeManager{} }

func (m *noOpTreeManager) Name() string                    { return treemanager.CName }
func (m *noOpTreeManager) Init(a *app.App) error           { return nil }
func (m *noOpTreeManager) Run(ctx context.Context) error   { return nil }
func (m *noOpTreeManager) Close(ctx context.Context) error { return nil }
func (m *noOpTreeManager) GetTree(ctx context.Context, spaceId, treeId string) (objecttree.ObjectTree, error) {
	return nil, fmt.Errorf("tree not found (local-only)")
}
func (m *noOpTreeManager) ValidateAndPutTree(ctx context.Context, spaceId string, payload treestorage.TreeStorageCreatePayload) error {
	return nil
}
func (m *noOpTreeManager) MarkTreeDeleted(ctx context.Context, spaceId, treeId string) error {
	return nil
}
func (m *noOpTreeManager) DeleteTree(ctx context.Context, spaceId, treeId string) error {
	return nil
}

// noOpNodeConf is a minimal NodeConf implementation.
type noOpNodeConf struct{}

func newNoOpNodeConf() *noOpNodeConf { return &noOpNodeConf{} }

// app.Component methods
func (c *noOpNodeConf) Name() string                    { return nodeconf.CName }
func (c *noOpNodeConf) Init(a *app.App) error           { return nil }
func (c *noOpNodeConf) Run(ctx context.Context) error   { return nil }
func (c *noOpNodeConf) Close(ctx context.Context) error { return nil }

// nodeconf.NodeConf methods
func (c *noOpNodeConf) Id() string                                   { return "local-node" }
func (c *noOpNodeConf) Configuration() nodeconf.Configuration        { return nodeconf.Configuration{} }
func (c *noOpNodeConf) NodeIds(spaceId string) []string              { return []string{} }
func (c *noOpNodeConf) IsResponsible(spaceId string) bool            { return true }
func (c *noOpNodeConf) FilePeers() []string                          { return []string{} }
func (c *noOpNodeConf) ConsensusPeers() []string                     { return []string{} }
func (c *noOpNodeConf) CoordinatorPeers() []string                   { return []string{} }
func (c *noOpNodeConf) NamingNodePeers() []string                    { return []string{} }
func (c *noOpNodeConf) PaymentProcessingNodePeers() []string         { return []string{} }
func (c *noOpNodeConf) PeerAddresses(peerId string) ([]string, bool) { return []string{}, false }
func (c *noOpNodeConf) CHash() chash.CHash                           { return &noOpCHash{} }
func (c *noOpNodeConf) Partition(spaceId string) int                 { return 0 }
func (c *noOpNodeConf) NodeTypes(nodeId string) []nodeconf.NodeType  { return []nodeconf.NodeType{} }
func (c *noOpNodeConf) NetworkCompatibilityStatus() nodeconf.NetworkCompatibilityStatus {
	return nodeconf.NetworkCompatibilityStatusOk
}

// noOpPeerManagerProvider is a minimal PeerManagerProvider implementation.
type noOpPeerManagerProvider struct{}

func newNoOpPeerManagerProvider() *noOpPeerManagerProvider { return &noOpPeerManagerProvider{} }

func (p *noOpPeerManagerProvider) Name() string                    { return peermanager.CName }
func (p *noOpPeerManagerProvider) Init(a *app.App) error           { return nil }
func (p *noOpPeerManagerProvider) Run(ctx context.Context) error   { return nil }
func (p *noOpPeerManagerProvider) Close(ctx context.Context) error { return nil }
func (p *noOpPeerManagerProvider) NewPeerManager(ctx context.Context, spaceId string) (peermanager.PeerManager, error) {
	return &noOpPeerManager{}, nil
}

// noOpPeerManager is a minimal PeerManager implementation.
type noOpPeerManager struct{}

func newNoOpPeerManager() *noOpPeerManager { return &noOpPeerManager{} }

func (m *noOpPeerManager) Name() string                    { return peermanager.CName }
func (m *noOpPeerManager) Init(a *app.App) error           { return nil }
func (m *noOpPeerManager) Run(ctx context.Context) error   { return nil }
func (m *noOpPeerManager) Close(ctx context.Context) error { return nil }
func (m *noOpPeerManager) GetResponsiblePeers(ctx context.Context) ([]peer.Peer, error) {
	return []peer.Peer{}, nil
}
func (m *noOpPeerManager) GetNodePeers(ctx context.Context) ([]peer.Peer, error) {
	return []peer.Peer{}, nil
}
func (m *noOpPeerManager) BroadcastMessage(ctx context.Context, msg drpc.Message) error { return nil }
func (m *noOpPeerManager) SendMessage(ctx context.Context, peerId string, msg drpc.Message) error {
	return nil
}
func (m *noOpPeerManager) KeepAlive(ctx context.Context) {}

// noOpPool is a minimal Pool implementation.
type noOpPool struct{}

func newNoOpPool() *noOpPool { return &noOpPool{} }

func (p *noOpPool) Name() string                    { return pool.CName }
func (p *noOpPool) Init(a *app.App) error           { return nil }
func (p *noOpPool) Run(ctx context.Context) error   { return nil }
func (p *noOpPool) Close(ctx context.Context) error { return nil }
func (p *noOpPool) Get(ctx context.Context, id string) (peer.Peer, error) {
	return nil, fmt.Errorf("no peers in local-only mode")
}
func (p *noOpPool) GetOneOf(ctx context.Context, peerIds []string) (peer.Peer, error) {
	return nil, fmt.Errorf("no peers in local-only mode")
}
func (p *noOpPool) AddPeer(ctx context.Context, pr peer.Peer) error {
	return nil // No-op for local-only
}
func (p *noOpPool) Pick(ctx context.Context, id string) (peer.Peer, error) {
	return nil, fmt.Errorf("no peers in local-only mode")
}
func (p *noOpPool) Flush(ctx context.Context) error {
	return nil
}

// noOpConfig is a minimal Config implementation that satisfies config.ConfigGetter.
type noOpConfig struct{}

func newNoOpConfig() *noOpConfig { return &noOpConfig{} }

func (c *noOpConfig) Name() string                    { return "config" }
func (c *noOpConfig) Init(a *app.App) error           { return nil }
func (c *noOpConfig) Run(ctx context.Context) error   { return nil }
func (c *noOpConfig) Close(ctx context.Context) error { return nil }

// GetSpace returns minimal space configuration for local-only operation.
func (c *noOpConfig) GetSpace() config.Config {
	return config.Config{
		GCTTL:                60,
		SyncPeriod:           5,
		KeepTreeDataInMemory: true,
	}
}

// noOpTreeSyncer is a minimal TreeSyncer implementation for local-only operation.
// It implements the treesyncer.TreeSyncer interface.
type noOpTreeSyncer struct{}

func newNoOpTreeSyncer() *noOpTreeSyncer { return &noOpTreeSyncer{} }

func (t *noOpTreeSyncer) Name() string                    { return "common.object.treesyncer" }
func (t *noOpTreeSyncer) Init(a *app.App) error           { return nil }
func (t *noOpTreeSyncer) Run(ctx context.Context) error   { return nil }
func (t *noOpTreeSyncer) Close(ctx context.Context) error { return nil }
func (t *noOpTreeSyncer) StartSync()                      {}
func (t *noOpTreeSyncer) StopSync()                       {}
func (t *noOpTreeSyncer) ShouldSync(peerId string) bool   { return false }
func (t *noOpTreeSyncer) SyncAll(ctx context.Context, p peer.Peer, existing, missing []string) error {
	return nil
}

// noOpCHash is a no-op implementation of chash.CHash for local-only operation.
type noOpCHash struct{}

func (c *noOpCHash) AddMembers(members ...chash.Member) error               { return nil }
func (c *noOpCHash) RemoveMembers(memberIds ...string) error                { return nil }
func (c *noOpCHash) Reconfigure(members []chash.Member) error               { return nil }
func (c *noOpCHash) GetMembers(key string) []chash.Member                   { return nil }
func (c *noOpCHash) GetPartition(key string) int                            { return 0 }
func (c *noOpCHash) GetPartitionMembers(partId int) ([]chash.Member, error) { return nil, nil }
func (c *noOpCHash) Distribute()                                            {}
func (c *noOpCHash) PartitionCount() int                                    { return 1 }
