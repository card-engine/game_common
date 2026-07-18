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

// Get 获取 GameInfo：先读本地缓存，未命中则查 DB 并回填本地。
func (s *GameInfoStore) Get(gameBrand, gameID string) (*models.GameInfo, bool) {
	key := GameInfoKey(gameBrand, gameID)
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
	var item models.GameInfo
	err := s.db.Where("game_brand = ? AND game_id = ?", gameBrand, gameID).First(&item).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, false
	}
	if err != nil {
		log.Errorf("[cache] gameinfo get from db failed key=%s err=%v", key, err)
		return nil, false
	}
	s.put(&item)
	cp := item
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
