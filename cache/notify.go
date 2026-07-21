package cache

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/redis/go-redis/v9"
)

const (
	DefaultNotifyChannel = "cache:notify"

	ActionReload = "reload"

	TypeAppInfo      = "appinfo"
	TypeAppGame      = "appgame"
	TypeGameInfo     = "gameinfo"
	TypeAppGameBrand = "appgamebrand"
)

// NotifyMessage Redis Pub/Sub 通知消息。
// Key 非空时按 key 刷新；Key 为空时全量刷新。
// appinfo/appgame/appgamebrand 的 key 为 appId；gameinfo 的 key 为 gameBrand（按整个厂商刷新）。
type NotifyMessage struct {
	Type   string `json:"type"`
	Action string `json:"action"`
	Key    string `json:"key,omitempty"`
}

// Publish 向 Redis 发布缓存刷新通知，固定使用 DefaultNotifyChannel（cache:notify）。
//
// 参数说明：
//   - cacheType: 缓存类型，见 TypeAppInfo / TypeAppGame / TypeGameInfo / TypeAppGameBrand
//   - key: 刷新范围；空表示该类型全量刷新，非空按类型含义刷新：
//
// 各类型传参示例：
//
//	// AppInfo：key 为空全量；key 为 appId 刷新单个商户
//	Publish(ctx, rdb, TypeAppInfo, "")
//	Publish(ctx, rdb, TypeAppInfo, "appId")
//
//	// AppGame：key 为空全量；key 为 appId 刷新该商户下全部游戏配置
//	Publish(ctx, rdb, TypeAppGame, "")
//	Publish(ctx, rdb, TypeAppGame, "appId")
//
//	// GameInfo：key 为空全量；key 为 gameBrand 刷新该厂商下全部游戏
//	Publish(ctx, rdb, TypeGameInfo, "")
//	Publish(ctx, rdb, TypeGameInfo, "jili")
//
//	// AppGameBrand：key 为空全量；key 为 appId 刷新该商户下全部厂商配置
//	Publish(ctx, rdb, TypeAppGameBrand, "")
//	Publish(ctx, rdb, TypeAppGameBrand, "appId")
func Publish(ctx context.Context, rdb *redis.Client, cacheType, key string) error {
	if cacheType == "" {
		return fmt.Errorf("cache: publish type is required")
	}
	msg := NotifyMessage{
		Type:   cacheType,
		Action: ActionReload,
		Key:    key,
	}
	payload, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	channel := DefaultNotifyChannel
	if err := rdb.Publish(ctx, channel, payload).Err(); err != nil {
		log.Errorf("[cache] publish failed channel=%s type=%s key=%q err=%v", channel, cacheType, key, err)
		return err
	}
	log.Infof("[cache] publish notify channel=%s type=%s key=%q", channel, cacheType, key)
	return nil
}

func parseNotifyMessage(payload string) (*NotifyMessage, error) {
	var msg NotifyMessage
	if err := json.Unmarshal([]byte(payload), &msg); err != nil {
		return nil, fmt.Errorf("cache: invalid notify payload: %w", err)
	}
	if msg.Type == "" {
		return nil, fmt.Errorf("cache: notify type is required")
	}
	if msg.Action == "" {
		msg.Action = ActionReload
	}
	return &msg, nil
}
