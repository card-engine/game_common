package cache

import (
	"cn.qingdou.server/game_common/models"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// 初始化所有AppGame缓存
func InitAppGame(rdb *redis.Client, db *gorm.DB) error {
	var dbAppInfos []models.AppGame
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
		cacheKey := getAppGameCacheKey(appInfo.AppId, appInfo.GameBrand, appInfo.GameId)
		err = rdb.Set(context.Background(), cacheKey, appInfoBytes, 0).Err()
		if err != nil {
			return err
		}
	}
	return nil
}

// 刷新单个AppGame
func RefreshAppGame(rdb *redis.Client, db *gorm.DB, appId, gameBrand, gameId string) (*models.AppGame, error) {
	var dbAppInfo models.AppGame
	result := db.Where("app_id = ? and game_id = ?", appId, gameId).First(&dbAppInfo)
	if result.Error != nil {
		return nil, result.Error
	}
	// 序列化对象为 JSON 字符串再存储到 Redis
	appInfoBytes, err := json.Marshal(dbAppInfo)
	if err != nil {
		return nil, err
	}
	cacheKey := getAppGameCacheKey(appId, gameBrand, gameId)
	err = rdb.Set(context.Background(), cacheKey, appInfoBytes, 0).Err()
	if err != nil {
		return nil, err
	}
	return &dbAppInfo, nil
}

// 获取AppGame
func GetAppGame(rdb *redis.Client, db *gorm.DB, appId, gameBrand, gameId string) (*models.AppGame, error) {
	cacheKey := getAppGameCacheKey(appId, gameBrand, gameId)
	// 尝试从 Redis 获取缓存数据
	val, err := rdb.Get(context.Background(), cacheKey).Result()
	if errors.Is(err, redis.Nil) {
		// 缓存中没有数据，调用 RefreshAppInfo 刷新缓存
		if appInfo, err := RefreshAppGame(rdb, db, appId, gameBrand, gameId); err != nil {
			return nil, err
		} else {
			return appInfo, nil
		}
	} else if err != nil {
		return nil, err
	}
	// 反序列化 Redis 中的数据
	var appInfo models.AppGame
	if err := json.Unmarshal([]byte(val), &appInfo); err != nil {
		return nil, err
	}
	return &appInfo, nil
}

func getAppGameCacheKey(appId, gameBrand, gameId string) string {
	return fmt.Sprintf("cache:appgame:%s:%s:%s", appId, gameBrand, gameId)
}
