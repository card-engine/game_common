package cache

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/card-engine/game_common/models"
	"gorm.io/gorm"
)

// AppGameStore AppGame 本地内存缓存。
type AppGameStore struct {
	db   *gorm.DB
	mu   sync.RWMutex
	data map[string]*models.AppGame // key: appId:gameBrand:gameId
}

// NewAppGameStore 创建 AppGame 本地缓存，并设为包级默认 Store。
func NewAppGameStore(db *gorm.DB) *AppGameStore {
	s := &AppGameStore{
		db:   db,
		data: make(map[string]*models.AppGame),
	}
	defaultAppGameStore = s
	return s
}

func (s *AppGameStore) Name() string {
	return TypeAppGame
}

// AppGameKey 生成本地缓存查找 key。
func AppGameKey(appID, gameBrand, gameID string) string {
	return fmt.Sprintf("%s:%s:%s", appID, gameBrand, gameID)
}

// LoadAll 全量从 DB 加载 AppGame。
func (s *AppGameStore) LoadAll(ctx context.Context) error {
	var list []models.AppGame
	if err := s.db.WithContext(ctx).Find(&list).Error; err != nil {
		return err
	}
	next := make(map[string]*models.AppGame, len(list))
	for i := range list {
		item := list[i]
		cp := item
		next[AppGameKey(cp.AppId, cp.GameBrand, cp.GameId)] = &cp
	}
	s.mu.Lock()
	s.data = next
	s.mu.Unlock()
	return nil
}

// LoadOne 按 appId 刷新该商户下全部 AppGame。
// key 即为 appId；DB 无记录时清除本地该 appId 的所有条目。
func (s *AppGameStore) LoadOne(ctx context.Context, key string) error {
	appID := strings.TrimSpace(key)
	if appID == "" {
		return fmt.Errorf("cache: appgame LoadOne key(appId) is empty")
	}
	var list []models.AppGame
	if err := s.db.WithContext(ctx).Where("app_id = ?", appID).Find(&list).Error; err != nil {
		return err
	}
	s.replaceByAppID(appID, list)
	return nil
}

// Get 从本地缓存获取 AppGame。
func (s *AppGameStore) Get(appID, gameBrand, gameID string) (*models.AppGame, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	v, ok := s.data[AppGameKey(appID, gameBrand, gameID)]
	if !ok || v == nil {
		return nil, false
	}
	cp := *v
	return &cp, true
}

func (s *AppGameStore) replaceByAppID(appID string, list []models.AppGame) {
	s.mu.Lock()
	defer s.mu.Unlock()
	prefix := appID + ":"
	for k := range s.data {
		if strings.HasPrefix(k, prefix) {
			delete(s.data, k)
		}
	}
	for i := range list {
		item := list[i]
		cp := item
		s.data[AppGameKey(cp.AppId, cp.GameBrand, cp.GameId)] = &cp
	}
}

func (s *AppGameStore) put(item *models.AppGame) {
	cp := *item
	key := AppGameKey(cp.AppId, cp.GameBrand, cp.GameId)
	s.mu.Lock()
	s.data[key] = &cp
	s.mu.Unlock()
}

func (s *AppGameStore) remove(key string) {
	s.mu.Lock()
	delete(s.data, key)
	s.mu.Unlock()
}

var defaultAppGameStore *AppGameStore

// SetDefaultAppGameStore 设置包级 AppGameStore。
func SetDefaultAppGameStore(s *AppGameStore) {
	defaultAppGameStore = s
}

// GetAppGame 从默认 AppGameStore 读取。
func GetAppGame(appID, gameBrand, gameID string) (*models.AppGame, bool) {
	if defaultAppGameStore == nil {
		return nil, false
	}
	return defaultAppGameStore.Get(appID, gameBrand, gameID)
}
