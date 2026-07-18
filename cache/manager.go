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
	// Logger 可选；为空时使用全局默认 logger。
	Logger log.Logger
}

// Manager 管理多个本地缓存 Store：启动全量预加载、订阅 Redis 通知、定时全量兜底。
type Manager struct {
	rdb      *redis.Client
	db       *gorm.DB
	opts     Options
	log      *log.Helper
	stores   map[string]Store
	appInfo  *AppInfoStore
	appGame  *AppGameStore
	gameInfo *GameInfoStore
	mu       sync.Mutex
	loadMu   sync.Map // per-store sync.Mutex，避免并发 LoadAll/LoadOne 互相踩踏
	cancel   context.CancelFunc
	wg       sync.WaitGroup
	started  bool
	warmedUp bool
}

// NewManager 创建缓存管理器。
func NewManager(rdb *redis.Client, db *gorm.DB, opts Options) *Manager {
	if opts.Channel == "" {
		opts.Channel = DefaultNotifyChannel
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

// Init 启动初始化：注册 AppInfo/AppGame/GameInfo，全量预加载后启动订阅与定时刷新。
// 业务侧通过返回的 Manager 访问各 Store，例如 mgr.AppInfo().GetByAppID(appId)。
func Init(ctx context.Context, rdb *redis.Client, db *gorm.DB, opts Options) (*Manager, error) {
	if db == nil {
		return nil, fmt.Errorf("cache: db is nil")
	}
	mgr := NewManager(rdb, db, opts)
	mgr.Register(NewAppInfoStore(db))
	mgr.Register(NewAppGameStore(db))
	mgr.Register(NewGameInfoStore(db))
	if err := mgr.Start(ctx); err != nil {
		return nil, err
	}
	return mgr, nil
}

// Register 注册一个 Store；同名重复注册会覆盖。
func (m *Manager) Register(store Store) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.stores[store.Name()] = store
	switch s := store.(type) {
	case *AppInfoStore:
		m.appInfo = s
	case *AppGameStore:
		m.appGame = s
	case *GameInfoStore:
		m.gameInfo = s
	}
	m.log.Infof("[cache] register store type=%s", store.Name())
}

// AppInfo 返回已注册的 AppInfoStore，未注册则为 nil。
func (m *Manager) AppInfo() *AppInfoStore {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.appInfo
}

// AppGame 返回已注册的 AppGameStore，未注册则为 nil。
func (m *Manager) AppGame() *AppGameStore {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.appGame
}

// GameInfo 返回已注册的 GameInfoStore，未注册则为 nil。
func (m *Manager) GameInfo() *GameInfoStore {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.gameInfo
}

// WarmUp 对已注册 Store 执行全量预加载。Start 会自动调用；也可单独调用。
func (m *Manager) WarmUp(ctx context.Context) error {
	m.mu.Lock()
	stores := make([]Store, 0, len(m.stores))
	for _, s := range m.stores {
		stores = append(stores, s)
	}
	m.mu.Unlock()

	m.log.Infof("[cache] warmup start stores=%d", len(stores))
	start := time.Now()
	for _, s := range stores {
		m.log.Infof("[cache] warmup load start type=%s", s.Name())
		t0 := time.Now()
		if err := m.safeLoadAll(ctx, s); err != nil {
			m.log.Errorf("[cache] warmup load failed type=%s err=%v", s.Name(), err)
			return fmt.Errorf("cache: warmup %s: %w", s.Name(), err)
		}
		m.log.Infof("[cache] warmup load done type=%s cost=%s", s.Name(), time.Since(t0))
	}
	m.mu.Lock()
	m.warmedUp = true
	m.mu.Unlock()
	m.log.Infof("[cache] warmup done stores=%d cost=%s", len(stores), time.Since(start))
	return nil
}

// Start 全量预加载所有 Store，然后启动 Pub/Sub 监听与定时全量刷新。
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

	m.log.Infof("[cache] starting: stores=%d channel=%s", len(stores), m.opts.Channel)

	if err := m.WarmUp(runCtx); err != nil {
		cancel()
		m.mu.Lock()
		m.started = false
		m.cancel = nil
		m.warmedUp = false
		m.mu.Unlock()
		return err
	}

	m.wg.Add(1)
	go m.subscribeLoop(runCtx)

	for _, s := range stores {
		interval := s.RefreshInterval()
		if interval <= 0 {
			interval = DefaultRefreshInterval
		}
		m.wg.Add(1)
		go m.refreshLoopOne(runCtx, s, interval)
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

// WarmedUp 是否已完成全量预加载。
func (m *Manager) WarmedUp() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.warmedUp
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

// refreshLoopOne 单个 Store 独立定时全量刷新，互不影响。
func (m *Manager) refreshLoopOne(ctx context.Context, store Store, interval time.Duration) {
	defer m.wg.Done()
	m.log.Infof("[cache] scheduled refresh enabled type=%s interval=%s", store.Name(), interval)
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			m.log.Infof("[cache] scheduled refresh stopped type=%s", store.Name())
			return
		case <-ticker.C:
			if err := m.applyReload(ctx, store, ""); err != nil {
				m.log.Errorf("[cache] scheduled reload failed type=%s err=%v", store.Name(), err)
			}
		}
	}
}
