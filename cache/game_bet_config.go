package cache

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/card-engine/game_common/models"
	"github.com/go-kratos/kratos/v2/log"
	"gorm.io/gorm"
)

// GameBetConfigStore GameBetConfig 本地内存缓存。
type GameBetConfigStore struct {
	db   *gorm.DB
	mu   sync.RWMutex
	data map[string]*models.GameBetConfig
}

// NewGameBetConfigStore 创建 GameBetConfig 本地缓存。
func NewGameBetConfigStore(db *gorm.DB) *GameBetConfigStore {
	return &GameBetConfigStore{
		db:   db,
		data: make(map[string]*models.GameBetConfig),
	}
}

func (s *GameBetConfigStore) Name() string {
	return TypeGameBetConfig
}

func (s *GameBetConfigStore) RefreshInterval() time.Duration {
	return 5 * time.Minute
}

// GameBetConfigKey 生成商户级配置缓存 key。
func GameBetConfigKey(appID, gameBrand, gameID string) string {
	return fmt.Sprintf("%s:%s:%s", appID, gameBrand, gameID)
}

// GameBetConfigGlobalKey 生成全局配置缓存 key（app_id=all）。
func GameBetConfigGlobalKey(currency, gameBrand, gameID string) string {
	return fmt.Sprintf("all:%s:%s:%s", currency, gameBrand, gameID)
}

func gameBetConfigCacheKey(item *models.GameBetConfig) string {
	if item.AppId == "all" {
		return GameBetConfigGlobalKey(item.Currency, item.GameBrand, item.GameId)
	}
	return GameBetConfigKey(item.AppId, item.GameBrand, item.GameId)
}

// LoadAll 全量从 DB 加载 GameBetConfig。
func (s *GameBetConfigStore) LoadAll(ctx context.Context) error {
	var list []models.GameBetConfig
	if err := s.db.WithContext(ctx).Find(&list).Error; err != nil {
		return err
	}
	next := make(map[string]*models.GameBetConfig, len(list))
	for i := range list {
		item := list[i]
		cp := item
		next[gameBetConfigCacheKey(&cp)] = &cp
	}
	s.mu.Lock()
	s.data = next
	s.mu.Unlock()
	return nil
}

// LoadOne 按 appId 刷新该商户下全部 GameBetConfig；key 为 "all" 时刷新全局配置。
func (s *GameBetConfigStore) LoadOne(ctx context.Context, key string) error {
	appID := strings.TrimSpace(key)
	if appID == "" {
		return fmt.Errorf("cache: gamebetconfig LoadOne key(appId) is empty")
	}
	var list []models.GameBetConfig
	if err := s.db.WithContext(ctx).Where("app_id = ?", appID).Find(&list).Error; err != nil {
		return err
	}
	s.replaceByAppID(appID, list)
	return nil
}

// Get 获取商户级 GameBetConfig：先读本地缓存，未命中则查 DB 并回填本地。
func (s *GameBetConfigStore) Get(appID, gameBrand, gameID string) (*models.GameBetConfig, bool) {
	key := GameBetConfigKey(appID, gameBrand, gameID)
	s.mu.RLock()
	v, ok := s.data[key]
	if ok && v != nil {
		cp := *v
		s.mu.RUnlock()
		return &cp, true
	}
	s.mu.RUnlock()

	if s.db == nil {
		return nil, false
	}
	var item models.GameBetConfig
	err := s.db.Where("app_id = ? AND game_brand = ? AND game_id = ?", appID, gameBrand, gameID).First(&item).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, false
	}
	if err != nil {
		log.Errorf("[cache] gamebetconfig get from db failed key=%s err=%v", key, err)
		return nil, false
	}
	s.put(&item)
	cp := item
	return &cp, true
}

// GetGlobal 获取全局 GameBetConfig（app_id=all）：先读本地缓存，未命中则查 DB 并回填本地。
func (s *GameBetConfigStore) GetGlobal(currency, gameBrand, gameID string) (*models.GameBetConfig, bool) {
	key := GameBetConfigGlobalKey(currency, gameBrand, gameID)
	s.mu.RLock()
	v, ok := s.data[key]
	if ok && v != nil {
		cp := *v
		s.mu.RUnlock()
		return &cp, true
	}
	s.mu.RUnlock()

	if s.db == nil {
		return nil, false
	}
	var item models.GameBetConfig
	err := s.db.Where("app_id = ? AND currency = ? AND game_brand = ? AND game_id = ?", "all", currency, gameBrand, gameID).First(&item).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, false
	}
	if err != nil {
		log.Errorf("[cache] gamebetconfig get global from db failed key=%s err=%v", key, err)
		return nil, false
	}
	s.put(&item)
	cp := item
	return &cp, true
}

func (s *GameBetConfigStore) replaceByAppID(appID string, list []models.GameBetConfig) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if appID == "all" {
		for k := range s.data {
			if strings.HasPrefix(k, "all:") {
				delete(s.data, k)
			}
		}
	} else {
		prefix := appID + ":"
		for k := range s.data {
			if strings.HasPrefix(k, prefix) {
				delete(s.data, k)
			}
		}
	}
	for i := range list {
		item := list[i]
		cp := item
		s.data[gameBetConfigCacheKey(&cp)] = &cp
	}
}

func (s *GameBetConfigStore) put(item *models.GameBetConfig) {
	cp := *item
	key := gameBetConfigCacheKey(&cp)
	s.mu.Lock()
	s.data[key] = &cp
	s.mu.Unlock()
}
