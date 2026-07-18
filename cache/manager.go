package cache

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// Options Manager 配置项。
type Options struct {
	// Channel Redis Pub/Sub 频道，默认 cache:notify。
	Channel string
	// RefreshInterval 定时全量刷新间隔；0 表示默认 5m，负值表示禁用定时刷新。
	RefreshInterval time.Duration
	// Logger 可选；为空时使用全局默认 logger。
	Logger log.Logger
}

// Manager 管理多个本地缓存 Store：启动全量加载、订阅 Redis 通知、定时全量兜底。
type Manager struct {
	rdb     *redis.Client
	db      *gorm.DB
	opts    Options
	log     *log.Helper
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
	logger := opts.Logger
	if logger == nil {
		logger = log.GetLogger()
	}
	return &Manager{
		rdb:    rdb,
		db:     db,
		opts:   opts,
		log:    log.NewHelper(log.With(logger, "module", "cache")),
		stores: make(map[string]Store),
	}
}

// Register 注册一个 Store；同名重复注册会覆盖。
func (m *Manager) Register(store Store) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.stores[store.Name()] = store
	m.log.Infof("[cache] register store type=%s", store.Name())
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

	m.log.Infof("[cache] starting: stores=%d channel=%s refreshInterval=%s",
		len(stores), m.opts.Channel, m.opts.RefreshInterval)

	for _, s := range stores {
		m.log.Infof("[cache] initial load start type=%s", s.Name())
		if err := m.safeLoadAll(runCtx, s); err != nil {
			cancel()
			m.mu.Lock()
			m.started = false
			m.cancel = nil
			m.mu.Unlock()
			m.log.Errorf("[cache] initial load failed type=%s err=%v", s.Name(), err)
			return fmt.Errorf("cache: load all %s: %w", s.Name(), err)
		}
		m.log.Infof("[cache] initial load done type=%s", s.Name())
	}

	m.wg.Add(1)
	go m.subscribeLoop(runCtx)

	if m.opts.RefreshInterval > 0 {
		m.wg.Add(1)
		go m.refreshLoop(runCtx)
		m.log.Infof("[cache] scheduled refresh enabled interval=%s", m.opts.RefreshInterval)
	} else {
		m.log.Info("[cache] scheduled refresh disabled")
	}
	m.log.Info("[cache] manager started")
	return nil
}

// Stop 停止 Pub/Sub 与定时刷新，并等待后台 goroutine 退出。
func (m *Manager) Stop() {
	m.log.Info("[cache] manager stopping")
	m.mu.Lock()
	cancel := m.cancel
	m.started = false
	m.cancel = nil
	m.mu.Unlock()
	if cancel != nil {
		cancel()
	}
	m.wg.Wait()
	m.log.Info("[cache] manager stopped")
}

// Refresh 手动触发刷新：key 非空按 key，否则全量。
func (m *Manager) Refresh(ctx context.Context, cacheType, key string) error {
	m.mu.Lock()
	store, ok := m.stores[cacheType]
	m.mu.Unlock()
	if !ok {
		return fmt.Errorf("cache: unknown type %q", cacheType)
	}
	m.log.Infof("[cache] manual refresh type=%s key=%q", cacheType, key)
	return m.applyReload(ctx, store, key)
}

func (m *Manager) applyReload(ctx context.Context, store Store, key string) error {
	start := time.Now()
	var err error
	if key == "" {
		m.log.Infof("[cache] reload all start type=%s", store.Name())
		err = m.safeLoadAll(ctx, store)
		if err != nil {
			m.log.Errorf("[cache] reload all failed type=%s cost=%s err=%v", store.Name(), time.Since(start), err)
			return err
		}
		m.log.Infof("[cache] reload all done type=%s cost=%s", store.Name(), time.Since(start))
		return nil
	}
	m.log.Infof("[cache] reload one start type=%s key=%q", store.Name(), key)
	err = m.safeLoadOne(ctx, store, key)
	if err != nil {
		m.log.Errorf("[cache] reload one failed type=%s key=%q cost=%s err=%v", store.Name(), key, time.Since(start), err)
		return err
	}
	m.log.Infof("[cache] reload one done type=%s key=%q cost=%s", store.Name(), key, time.Since(start))
	return nil
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
	m.log.Infof("[cache] subscribe start channel=%s", m.opts.Channel)
	pubsub := m.rdb.Subscribe(ctx, m.opts.Channel)
	defer pubsub.Close()

	ch := pubsub.Channel()
	for {
		select {
		case <-ctx.Done():
			m.log.Infof("[cache] subscribe stopped channel=%s", m.opts.Channel)
			return
		case msg, ok := <-ch:
			if !ok {
				m.log.Warnf("[cache] subscribe channel closed channel=%s", m.opts.Channel)
				return
			}
			if msg == nil {
				continue
			}
			m.log.Infof("[cache] notify received channel=%s payload=%s", m.opts.Channel, msg.Payload)
			notify, err := parseNotifyMessage(msg.Payload)
			if err != nil {
				m.log.Errorf("[cache] parse notify failed payload=%s err=%v", msg.Payload, err)
				continue
			}
			if notify.Action != ActionReload {
				m.log.Warnf("[cache] ignore unknown action=%q type=%s key=%q", notify.Action, notify.Type, notify.Key)
				continue
			}
			m.mu.Lock()
			store, ok := m.stores[notify.Type]
			m.mu.Unlock()
			if !ok {
				m.log.Warnf("[cache] ignore unknown type=%q key=%q", notify.Type, notify.Key)
				continue
			}
			m.log.Infof("[cache] notify reload type=%s key=%q", notify.Type, notify.Key)
			if err := m.applyReload(ctx, store, notify.Key); err != nil {
				m.log.Errorf("[cache] notify reload failed type=%s key=%q err=%v", notify.Type, notify.Key, err)
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
			m.log.Info("[cache] scheduled refresh stopped")
			return
		case <-ticker.C:
			m.mu.Lock()
			stores := make([]Store, 0, len(m.stores))
			for _, s := range m.stores {
				stores = append(stores, s)
			}
			m.mu.Unlock()
			m.log.Infof("[cache] scheduled refresh tick stores=%d", len(stores))
			for _, s := range stores {
				if err := m.applyReload(ctx, s, ""); err != nil {
					m.log.Errorf("[cache] scheduled reload failed type=%s err=%v", s.Name(), err)
				}
			}
		}
	}
}
