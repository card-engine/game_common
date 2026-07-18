package cache

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// Options Manager 配置项。
type Options struct {
	// Channel Redis Pub/Sub 频道，默认 cache:notify。
	Channel string
	// RefreshInterval 定时全量刷新间隔；0 表示默认 5m，负值表示禁用定时刷新。
	RefreshInterval time.Duration
}

// Manager 管理多个本地缓存 Store：启动全量加载、订阅 Redis 通知、定时全量兜底。
type Manager struct {
	rdb     *redis.Client
	db      *gorm.DB
	opts    Options
	stores  map[string]Store
	mu      sync.Mutex
	loadMu  sync.Map // per-store sync.Mutex，避免并发 LoadAll/LoadOne 互相踩踏
	cancel  context.CancelFunc
	wg      sync.WaitGroup
	started bool
}

// NewManager 创建缓存管理器。db 由各 Store 自行持有时仍可传入供扩展；rdb 用于订阅。
func NewManager(rdb *redis.Client, db *gorm.DB, opts Options) *Manager {
	if opts.Channel == "" {
		opts.Channel = DefaultNotifyChannel
	}
	if opts.RefreshInterval == 0 {
		opts.RefreshInterval = 5 * time.Minute
	}
	return &Manager{
		rdb:    rdb,
		db:     db,
		opts:   opts,
		stores: make(map[string]Store),
	}
}

// Register 注册一个 Store；同名重复注册会覆盖。
func (m *Manager) Register(store Store) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.stores[store.Name()] = store
}

// Start 全量加载所有 Store，然后启动 Pub/Sub 监听与定时全量刷新。
func (m *Manager) Start(ctx context.Context) error {
	m.mu.Lock()
	if m.started {
		m.mu.Unlock()
		return fmt.Errorf("cache: manager already started")
	}
	if m.rdb == nil {
		m.mu.Unlock()
		return fmt.Errorf("cache: redis client is nil")
	}
	stores := make([]Store, 0, len(m.stores))
	for _, s := range m.stores {
		stores = append(stores, s)
	}
	m.started = true
	runCtx, cancel := context.WithCancel(ctx)
	m.cancel = cancel
	m.mu.Unlock()

	for _, s := range stores {
		if err := m.safeLoadAll(runCtx, s); err != nil {
			cancel()
			m.mu.Lock()
			m.started = false
			m.cancel = nil
			m.mu.Unlock()
			return fmt.Errorf("cache: load all %s: %w", s.Name(), err)
		}
	}

	m.wg.Add(1)
	go m.subscribeLoop(runCtx)

	if m.opts.RefreshInterval > 0 {
		m.wg.Add(1)
		go m.refreshLoop(runCtx)
	}
	return nil
}

// Stop 停止 Pub/Sub 与定时刷新，并等待后台 goroutine 退出。
func (m *Manager) Stop() {
	m.mu.Lock()
	cancel := m.cancel
	m.started = false
	m.cancel = nil
	m.mu.Unlock()
	if cancel != nil {
		cancel()
	}
	m.wg.Wait()
}

// Refresh 手动触发刷新：key 非空按 key，否则全量。
func (m *Manager) Refresh(ctx context.Context, cacheType, key string) error {
	m.mu.Lock()
	store, ok := m.stores[cacheType]
	m.mu.Unlock()
	if !ok {
		return fmt.Errorf("cache: unknown type %q", cacheType)
	}
	return m.applyReload(ctx, store, key)
}

func (m *Manager) applyReload(ctx context.Context, store Store, key string) error {
	if key == "" {
		return m.safeLoadAll(ctx, store)
	}
	return m.safeLoadOne(ctx, store, key)
}

func (m *Manager) storeLock(name string) *sync.Mutex {
	v, _ := m.loadMu.LoadOrStore(name, &sync.Mutex{})
	return v.(*sync.Mutex)
}

func (m *Manager) safeLoadAll(ctx context.Context, store Store) error {
	lk := m.storeLock(store.Name())
	lk.Lock()
	defer lk.Unlock()
	return store.LoadAll(ctx)
}

func (m *Manager) safeLoadOne(ctx context.Context, store Store, key string) error {
	lk := m.storeLock(store.Name())
	lk.Lock()
	defer lk.Unlock()
	return store.LoadOne(ctx, key)
}

func (m *Manager) subscribeLoop(ctx context.Context) {
	defer m.wg.Done()
	pubsub := m.rdb.Subscribe(ctx, m.opts.Channel)
	defer pubsub.Close()

	ch := pubsub.Channel()
	for {
		select {
		case <-ctx.Done():
			return
		case msg, ok := <-ch:
			if !ok {
				return
			}
			if msg == nil {
				continue
			}
			notify, err := parseNotifyMessage(msg.Payload)
			if err != nil {
				log.Printf("cache: parse notify: %v", err)
				continue
			}
			if notify.Action != ActionReload {
				log.Printf("cache: ignore unknown action %q", notify.Action)
				continue
			}
			m.mu.Lock()
			store, ok := m.stores[notify.Type]
			m.mu.Unlock()
			if !ok {
				log.Printf("cache: ignore unknown type %q", notify.Type)
				continue
			}
			if err := m.applyReload(ctx, store, notify.Key); err != nil {
				log.Printf("cache: reload %s key=%q: %v", notify.Type, notify.Key, err)
			}
		}
	}
}

func (m *Manager) refreshLoop(ctx context.Context) {
	defer m.wg.Done()
	ticker := time.NewTicker(m.opts.RefreshInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			m.mu.Lock()
			stores := make([]Store, 0, len(m.stores))
			for _, s := range m.stores {
				stores = append(stores, s)
			}
			m.mu.Unlock()
			for _, s := range stores {
				if err := m.safeLoadAll(ctx, s); err != nil {
					log.Printf("cache: scheduled reload %s: %v", s.Name(), err)
				}
			}
		}
	}
}
