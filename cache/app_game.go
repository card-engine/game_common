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

// AppGameStore AppGame 本地内存缓存。
type AppGameStore struct {
	db   *gorm.DB
	mu   sync.RWMutex
	data map[string]*models.AppGame // key: appId:gameBrand:gameId
}

// NewAppGameStore 创建 AppGame 本地缓存。
func NewAppGameStore(db *gorm.DB) *AppGameStore {
	return &AppGameStore{
		db:   db,
		data: make(map[string]*models.AppGame),
	}
}

func (s *AppGameStore) Name() string {
	return TypeAppGame
}

func (s *AppGameStore) RefreshInterval() time.Duration {
	return 10 * time.Minute
}

// AppGameKey 生成本地缓存查找 key。
func AppGameKey(appID, gameBrand, gameID string) string {
	return fmt.Sprintf("%s:%s:%s", appID, gameBrand, gameID)
}

const appGameLoadConcurrency = 16

// LoadAll 按 appId 并发从 DB 加载全部 AppGame。
func (s *AppGameStore) LoadAll(ctx context.Context) error {
	var appIDs []string
	if err := s.db.WithContext(ctx).Model(&models.AppGame{}).
		Distinct("app_id").
		Pluck("app_id", &appIDs).Error; err != nil {
		return err
	}
	if len(appIDs) == 0 {
		s.mu.Lock()
		s.data = make(map[string]*models.AppGame)
		s.mu.Unlock()
		log.Infof("[cache] appgame LoadAll done size=0 apps=0")
		return nil
	}

	next := make(map[string]*models.AppGame)
	var nextMu sync.Mutex
	var loadedApps int64

	sem := make(chan struct{}, appGameLoadConcurrency)
	var wg sync.WaitGroup
	errCh := make(chan error, len(appIDs))

	log.Infof("[cache] appgame LoadAll start apps=%d concurrency=%d", len(appIDs), appGameLoadConcurrency)
	for _, id := range appIDs {
		wg.Add(1)
		go func(appID string) {
			defer wg.Done()
			select {
			case <-ctx.Done():
				errCh <- ctx.Err()
				return
			case sem <- struct{}{}:
			}
			defer func() { <-sem }()

			var list []models.AppGame
			if err := s.db.WithContext(ctx).Where("app_id = ?", appID).Find(&list).Error; err != nil {
				errCh <- fmt.Errorf("cache: appgame load appId=%s: %w", appID, err)
				return
			}

			local := make(map[string]*models.AppGame, len(list))
			for i := range list {
				item := list[i]
				cp := item
				local[AppGameKey(cp.AppId, cp.GameBrand, cp.GameId)] = &cp
			}

			nextMu.Lock()
			for k, v := range local {
				next[k] = v
			}
			loadedApps++
			done := loadedApps
			nextMu.Unlock()

			if done%50 == 0 || int(done) == len(appIDs) {
				log.Infof("[cache] appgame LoadAll progress apps=%d/%d", done, len(appIDs))
			}
		}(id)
	}
	wg.Wait()
	close(errCh)
	for err := range errCh {
		return err
	}

	s.mu.Lock()
	s.data = next
	s.mu.Unlock()
	log.Infof("[cache] appgame LoadAll done size=%d apps=%d", len(next), len(appIDs))
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

// Get 获取 AppGame：先读本地缓存，未命中则查 DB 并回填本地。
func (s *AppGameStore) Get(appID, gameBrand, gameID string) (*models.AppGame, bool) {
	key := AppGameKey(appID, gameBrand, gameID)
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
	var item models.AppGame
	err := s.db.Where("app_id = ? AND game_brand = ? AND game_id = ?", appID, gameBrand, gameID).First(&item).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, false
	}
	if err != nil {
		log.Errorf("[cache] appgame get from db failed key=%s err=%v", key, err)
		return nil, false
	}
	s.put(&item)
	cp := item
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
