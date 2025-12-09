package cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/card-engine/game_common/models"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// 初始化所有AppInfo缓存
func InitApp(rdb *redis.Client, db *gorm.DB) error {
	var dbAppInfos []models.AppInfo
	result := db.Find(&dbAppInfos)
	if result.Error != nil {
		return result.Error
	}
	// 遍历所有 AppInfo 并缓存到 Redis
	for _, appInfo := range dbAppInfos {
		// 序列化对象为 JSON 字符串再存储到 Redis
		appInfoBytes, err := json.Marshal(appInfo)
		if err != nil {
			return err
		}
		cacheKey := getAppCacheKey(appInfo.AppId)
		err = rdb.Set(context.Background(), cacheKey, appInfoBytes, 0).Err()
		if err != nil {
			return err
		}
	}
	return nil
}

// 刷新单个AppInfo
func RefreshApp(rdb *redis.Client, db *gorm.DB, appId string) (*models.AppInfo, error) {
	var dbAppInfo models.AppInfo
	result := db.Where("app_id = ?", appId).First(&dbAppInfo)
	if result.Error != nil {
		return nil, result.Error
	}
	// 序列化对象为 JSON 字符串再存储到 Redis
	appInfoBytes, err := json.Marshal(dbAppInfo)
	if err != nil {
		return nil, err
	}
	cacheKey := getAppCacheKey(appId)
	err = rdb.Set(context.Background(), cacheKey, appInfoBytes, 0).Err()
	if err != nil {
		return nil, err
	}
	return &dbAppInfo, nil
}

// 获取AppInfo
func GetApp(rdb *redis.Client, db *gorm.DB, appId string) (*models.AppInfo, error) {
	cacheKey := getAppCacheKey(appId)
	// 尝试从 Redis 获取缓存数据
	val, err := rdb.Get(context.Background(), cacheKey).Result()
	if errors.Is(err, redis.Nil) {
		// 缓存中没有数据，调用 RefreshAppInfo 刷新缓存
		if appInfo, err := RefreshApp(rdb, db, appId); err != nil {
			return nil, err
		} else {
			return appInfo, nil
		}
	} else if err != nil {
		return nil, err
	}
	// 反序列化 Redis 中的数据
	var appInfo models.AppInfo
	if err := json.Unmarshal([]byte(val), &appInfo); err != nil {
		return nil, err
	}
	return &appInfo, nil
}

func getAppCacheKey(appId string) string {
	return fmt.Sprintf("cache:app:%s", appId)
}
