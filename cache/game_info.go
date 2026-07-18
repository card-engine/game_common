package cache

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/card-engine/game_common/models"
	"gorm.io/gorm"
)

// GameInfoStore GameInfo 本地内存缓存。
type GameInfoStore struct {
	db   *gorm.DB
	mu   sync.RWMutex
	data map[string]*models.GameInfo // key: gameBrand:gameId
}

// NewGameInfoStore 创建 GameInfo 本地缓存。
func NewGameInfoStore(db *gorm.DB) *GameInfoStore {
	return &GameInfoStore{
		db:   db,
		data: make(map[string]*models.GameInfo),
	}
}

func (s *GameInfoStore) Name() string {
	return TypeGameInfo
}

func (s *GameInfoStore) RefreshInterval() time.Duration {
	return 5 * time.Minute
}

// GameInfoKey 生成本地缓存查找 key。
func GameInfoKey(gameBrand, gameID string) string {
	return fmt.Sprintf("%s:%s", gameBrand, gameID)
}

// LoadAll 全量从 DB 加载 GameInfo。
func (s *GameInfoStore) LoadAll(ctx context.Context) error {
	var list []models.GameInfo
	if err := s.db.WithContext(ctx).Find(&list).Error; err != nil {
		return err
	}
	next := make(map[string]*models.GameInfo, len(list))
	for i := range list {
		item := list[i]
		cp := item
		next[GameInfoKey(cp.GameBrand, cp.GameId)] = &cp
	}
	s.mu.Lock()
	s.data = next
	s.mu.Unlock()
	return nil
}

// LoadOne 按 gameBrand 刷新该厂商下全部 GameInfo。
// key 即为 gameBrand；DB 无记录时清除本地该厂商的所有条目。
func (s *GameInfoStore) LoadOne(ctx context.Context, key string) error {
	gameBrand := strings.TrimSpace(key)
	if gameBrand == "" {
		return fmt.Errorf("cache: gameinfo LoadOne key(gameBrand) is empty")
	}
	var list []models.GameInfo
	if err := s.db.WithContext(ctx).Where("game_brand = ?", gameBrand).Find(&list).Error; err != nil {
		return err
	}
	s.replaceByBrand(gameBrand, list)
	return nil
}

// Get 从本地缓存获取 GameInfo。
func (s *GameInfoStore) Get(gameBrand, gameID string) (*models.GameInfo, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	v, ok := s.data[GameInfoKey(gameBrand, gameID)]
	if !ok || v == nil {
		return nil, false
	}
	cp := *v
	return &cp, true
}

// Len 返回本地缓存条数。
func (s *GameInfoStore) Len() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.data)
}

func (s *GameInfoStore) replaceByBrand(gameBrand string, list []models.GameInfo) {
	s.mu.Lock()
	defer s.mu.Unlock()
	prefix := gameBrand + ":"
	for k := range s.data {
		if strings.HasPrefix(k, prefix) {
			delete(s.data, k)
		}
	}
	for i := range list {
		item := list[i]
		cp := item
		s.data[GameInfoKey(cp.GameBrand, cp.GameId)] = &cp
	}
}

func (s *GameInfoStore) put(item *models.GameInfo) {
	cp := *item
	key := GameInfoKey(cp.GameBrand, cp.GameId)
	s.mu.Lock()
	s.data[key] = &cp
	s.mu.Unlock()
}

func (s *GameInfoStore) remove(key string) {
	s.mu.Lock()
	delete(s.data, key)
	s.mu.Unlock()
}
