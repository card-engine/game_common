package utils

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

func GenRoundId(rdb *redis.Client, appId string, gameBrand string, gameId string) (int64, error) {
	key := fmt.Sprintf("%s:%s:%s:roundid", gameBrand, gameId, appId)
	// 使用 INCR 命令，如果 key 不存在会自动创建并设为 1
	roundid, err := rdb.Incr(context.Background(), key).Result()
	if err != nil {
		return 0, fmt.Errorf("生成 roundid 失败: %w", err)
	}
	return roundid, nil
}
