package cache

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/redis/go-redis/v9"
)

const (
	DefaultNotifyChannel = "cache:notify"

	ActionReload = "reload"

	TypeAppInfo = "appinfo"
	TypeAppGame = "appgame"
)

// NotifyMessage Redis Pub/Sub 通知消息。
// Key 非空时按 key 刷新；Key 为空时全量刷新。
// appinfo 的 key 为 appId；appgame 的 key 也为 appId（按整个商户刷新其下全部游戏）。
type NotifyMessage struct {
	Type   string `json:"type"`
	Action string `json:"action"`
	Key    string `json:"key,omitempty"`
}

// Publish 向 Redis 发布缓存刷新通知。
// key 为空表示全量刷新；非空表示按 key 刷新（appinfo/appgame 均为 appId）。
func Publish(ctx context.Context, rdb *redis.Client, channel, cacheType, key string) error {
	if channel == "" {
		channel = DefaultNotifyChannel
	}
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
	return rdb.Publish(ctx, channel, payload).Err()
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
