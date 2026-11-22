package cache

import (
	"context"
	"strings"
	"time"
)

const DefaultRefreshInterval = 5 * time.Minute

// cacheKeyPart 缓存 map key 分段：去空白并小写，避免 PG/pg 这类大小写导致永远 miss。
func cacheKeyPart(s string) string {
	return strings.ToLower(strings.TrimSpace(s))
}

// Store 本地缓存存储接口。
// Manager 负责调度 LoadAll / LoadOne；具体 map 读写由各实现自行维护。
// 各 Store 的 Get* 方法采用 cache-aside：本地未命中时回源 DB，并回填本地缓存。
type Store interface {
	// Name 返回缓存类型名，与通知消息 type 字段对应，如 "appinfo" / "appgame"。
	Name() string
	// LoadAll 从 DB 全量加载到本地内存。
	LoadAll(ctx context.Context) error
	// LoadOne 按 key 从 DB 加载单条；记录不存在时应从本地删除该 key。
	LoadOne(ctx context.Context, key string) error
	// RefreshInterval 该缓存定时全量刷新间隔；<=0 时 Manager 回退为 DefaultRefreshInterval（5m）。
	RefreshInterval() time.Duration
}
