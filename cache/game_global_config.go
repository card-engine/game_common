package cache

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/card-engine/game_common/models"
	"github.com/go-kratos/kratos/v2/log"
	"gorm.io/gorm"
)

// GameGlobalConfigStore GameGlobalConfig 本地内存缓存。
type GameGlobalConfigStore struct {
	db   *gorm.DB
	mu   sync.RWMutex
	data map[string]*models.GameGlobalConfig // key: gkey
}

// NewGameGlobalConfigStore 创建 GameGlobalConfig 本地缓存。
func NewGameGlobalConfigStore(db *gorm.DB) *GameGlobalConfigStore {
	return &GameGlobalConfigStore{
		db:   db,
		data: make(map[string]*models.GameGlobalConfig),
	}
}

func (s *GameGlobalConfigStore) Name() string {
	return TypeGameGlobalConfig
}

func (s *GameGlobalConfigStore) RefreshInterval() time.Duration {
	return 60 * time.Minute
}

// LoadAll 全量从 DB 加载 GameGlobalConfig。
func (s *GameGlobalConfigStore) LoadAll(ctx context.Context) error {
	var list []models.GameGlobalConfig
	if err := s.db.WithContext(ctx).Find(&list).Error; err != nil {
		return err
	}
	next := make(map[string]*models.GameGlobalConfig, len(list))
	for i := range list {
		item := list[i]
		cp := item
		next[cp.GKey] = &cp
	}
	s.mu.Lock()
	s.data = next
	s.mu.Unlock()
	return nil
}

// LoadOne 按 gkey 加载单条；不存在则删除本地缓存。
func (s *GameGlobalConfigStore) LoadOne(ctx context.Context, key string) error {
	gkey := key
	if gkey == "" {
		return fmt.Errorf("cache: gameglobalconfig LoadOne key(gkey) is empty")
	}
	var item models.GameGlobalConfig
	err := s.db.WithContext(ctx).Where("gkey = ?", gkey).First(&item).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		s.remove(gkey)
		return nil
	}
	if err != nil {
		return err
	}
	s.put(&item)
	return nil
}

// Get 按 gkey 获取：先读本地缓存，未命中则查 DB 并回填本地。
func (s *GameGlobalConfigStore) Get(gkey string) (*models.GameGlobalConfig, bool) {
	s.mu.RLock()
	v, ok := s.data[gkey]
	if ok && v != nil {
		cp := *v
		s.mu.RUnlock()
		return &cp, true
	}
	s.mu.RUnlock()

	if s.db == nil {
		return nil, false
	}
	var item models.GameGlobalConfig
	err := s.db.Where("gkey = ?", gkey).First(&item).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, false
	}
	if err != nil {
		log.Errorf("[cache] gameglobalconfig get from db failed gkey=%s err=%v", gkey, err)
		return nil, false
	}
	s.put(&item)
	cp := item
	return &cp, true
}

func (s *GameGlobalConfigStore) put(item *models.GameGlobalConfig) {
	cp := *item
	s.mu.Lock()
	s.data[cp.GKey] = &cp
	s.mu.Unlock()
}

func (s *GameGlobalConfigStore) remove(gkey string) {
	s.mu.Lock()
	delete(s.data, gkey)
	s.mu.Unlock()
}
