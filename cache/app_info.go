package cache

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/card-engine/game_common/models"
	"gorm.io/gorm"
)

// AppInfoStore AppInfo 本地内存缓存。
type AppInfoStore struct {
	db          *gorm.DB
	mu          sync.RWMutex
	byAppID     map[string]*models.AppInfo
	byAccessKey map[string]*models.AppInfo
}

// NewAppInfoStore 创建 AppInfo 本地缓存，并设为包级默认 Store。
func NewAppInfoStore(db *gorm.DB) *AppInfoStore {
	s := &AppInfoStore{
		db:          db,
		byAppID:     make(map[string]*models.AppInfo),
		byAccessKey: make(map[string]*models.AppInfo),
	}
	defaultAppInfoStore = s
	return s
}

func (s *AppInfoStore) Name() string {
	return TypeAppInfo
}

func (s *AppInfoStore) RefreshInterval() time.Duration {
	return 5 * time.Minute
}

// LoadAll 全量从 DB 加载 AppInfo。
func (s *AppInfoStore) LoadAll(ctx context.Context) error {
	var list []models.AppInfo
	if err := s.db.WithContext(ctx).Find(&list).Error; err != nil {
		return err
	}
	byAppID := make(map[string]*models.AppInfo, len(list))
	byAccessKey := make(map[string]*models.AppInfo, len(list))
	for i := range list {
		item := list[i]
		cp := item
		byAppID[cp.AppId] = &cp
		if cp.AccessKeyId != "" {
			byAccessKey[cp.AccessKeyId] = &cp
		}
	}
	s.mu.Lock()
	s.byAppID = byAppID
	s.byAccessKey = byAccessKey
	s.mu.Unlock()
	return nil
}

// LoadOne 按 appId 加载单条；不存在则删除本地缓存。
func (s *AppInfoStore) LoadOne(ctx context.Context, key string) error {
	if key == "" {
		return fmt.Errorf("cache: appinfo LoadOne key is empty")
	}
	var item models.AppInfo
	err := s.db.WithContext(ctx).Where("app_id = ?", key).First(&item).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		s.remove(key)
		return nil
	}
	if err != nil {
		return err
	}
	s.put(&item)
	return nil
}

// GetByAppID 从本地缓存按 appId 获取。
func (s *AppInfoStore) GetByAppID(appID string) (*models.AppInfo, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	v, ok := s.byAppID[appID]
	if !ok || v == nil {
		return nil, false
	}
	cp := *v
	return &cp, true
}

// GetByAccessKeyID 从本地缓存按 accessKeyId 获取。
func (s *AppInfoStore) GetByAccessKeyID(accessKeyID string) (*models.AppInfo, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	v, ok := s.byAccessKey[accessKeyID]
	if !ok || v == nil {
		return nil, false
	}
	cp := *v
	return &cp, true
}

func (s *AppInfoStore) put(item *models.AppInfo) {
	cp := *item
	s.mu.Lock()
	defer s.mu.Unlock()
	if old, ok := s.byAppID[cp.AppId]; ok && old != nil && old.AccessKeyId != "" && old.AccessKeyId != cp.AccessKeyId {
		delete(s.byAccessKey, old.AccessKeyId)
	}
	s.byAppID[cp.AppId] = &cp
	if cp.AccessKeyId != "" {
		s.byAccessKey[cp.AccessKeyId] = &cp
	}
}

func (s *AppInfoStore) remove(appID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if old, ok := s.byAppID[appID]; ok && old != nil && old.AccessKeyId != "" {
		delete(s.byAccessKey, old.AccessKeyId)
	}
	delete(s.byAppID, appID)
}

// 包级默认 Store，供 Manager 注册后业务侧便捷读取。
var defaultAppInfoStore *AppInfoStore

// SetDefaultAppInfoStore 设置包级 AppInfoStore（通常在 Register 后调用）。
func SetDefaultAppInfoStore(s *AppInfoStore) {
	defaultAppInfoStore = s
}

// GetAppInfo 从默认 AppInfoStore 按 appId 读取。
func GetAppInfo(appID string) (*models.AppInfo, bool) {
	if defaultAppInfoStore == nil {
		return nil, false
	}
	return defaultAppInfoStore.GetByAppID(appID)
}

// GetAppInfoByAccessKey 从默认 AppInfoStore 按 accessKeyId 读取。
func GetAppInfoByAccessKey(accessKeyID string) (*models.AppInfo, bool) {
	if defaultAppInfoStore == nil {
		return nil, false
	}
	return defaultAppInfoStore.GetByAccessKeyID(accessKeyID)
}
