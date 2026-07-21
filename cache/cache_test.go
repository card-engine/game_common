package cache

import (
	"context"
	"encoding/json"
	"sync"
	"testing"
	"time"

	"github.com/card-engine/game_common/models"
)

type mockStore struct {
	name       string
	mu         sync.Mutex
	loadAllN   int
	loadOneN   int
	lastOneKey string
	loadAllErr error
	loadOneErr error
}

func (m *mockStore) Name() string { return m.name }

func (m *mockStore) RefreshInterval() time.Duration { return DefaultRefreshInterval }

func (m *mockStore) LoadAll(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.loadAllN++
	return m.loadAllErr
}

func (m *mockStore) LoadOne(ctx context.Context, key string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.loadOneN++
	m.lastOneKey = key
	return m.loadOneErr
}

func (m *mockStore) stats() (allN, oneN int, lastKey string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.loadAllN, m.loadOneN, m.lastOneKey
}

func TestParseNotifyMessage(t *testing.T) {
	msg, err := parseNotifyMessage(`{"type":"appinfo","action":"reload","key":"a1"}`)
	if err != nil {
		t.Fatal(err)
	}
	if msg.Type != TypeAppInfo || msg.Key != "a1" || msg.Action != ActionReload {
		t.Fatalf("unexpected message: %+v", msg)
	}

	msg, err = parseNotifyMessage(`{"type":"appgame"}`)
	if err != nil {
		t.Fatal(err)
	}
	if msg.Action != ActionReload || msg.Key != "" {
		t.Fatalf("expected default action and empty key, got %+v", msg)
	}

	if _, err := parseNotifyMessage(`{}`); err == nil {
		t.Fatal("expected error for missing type")
	}
}

func TestNotifyMessageJSON(t *testing.T) {
	raw, err := json.Marshal(NotifyMessage{Type: TypeAppGame, Action: ActionReload})
	if err != nil {
		t.Fatal(err)
	}
	var msg NotifyMessage
	if err := json.Unmarshal(raw, &msg); err != nil {
		t.Fatal(err)
	}
	if msg.Key != "" {
		t.Fatalf("empty key should omit or be empty, got %q", msg.Key)
	}
}

func TestManagerRefreshKeyVsAll(t *testing.T) {
	store := &mockStore{name: TypeAppInfo}
	mgr := NewManager(nil, nil, Options{})
	mgr.Register(store)

	if err := mgr.Refresh(context.Background(), TypeAppInfo, "app-1"); err != nil {
		t.Fatal(err)
	}
	allN, oneN, lastKey := store.stats()
	if allN != 0 || oneN != 1 || lastKey != "app-1" {
		t.Fatalf("key refresh: all=%d one=%d key=%q", allN, oneN, lastKey)
	}

	if err := mgr.Refresh(context.Background(), TypeAppInfo, ""); err != nil {
		t.Fatal(err)
	}
	allN, oneN, _ = store.stats()
	if allN != 1 || oneN != 1 {
		t.Fatalf("full refresh: all=%d one=%d", allN, oneN)
	}
}

func TestManagerRefreshUnknownType(t *testing.T) {
	mgr := NewManager(nil, nil, Options{})
	if err := mgr.Refresh(context.Background(), "unknown", ""); err == nil {
		t.Fatal("expected error for unknown type")
	}
}

func TestStoreRefreshIntervals(t *testing.T) {
	if got := NewAppInfoStore(nil).RefreshInterval(); got != 5*time.Minute {
		t.Fatalf("appinfo want 5m, got %s", got)
	}
	if got := NewAppGameStore(nil).RefreshInterval(); got != 10*time.Minute {
		t.Fatalf("appgame want 10m, got %s", got)
	}
	if got := NewGameInfoStore(nil).RefreshInterval(); got != 5*time.Minute {
		t.Fatalf("gameinfo want 5m, got %s", got)
	}
	if got := NewAppGameBrandStore(nil).RefreshInterval(); got != 10*time.Minute {
		t.Fatalf("appgamebrand want 10m, got %s", got)
	}
}

func TestAppGameReplaceByAppID(t *testing.T) {
	s := NewAppGameStore(nil)
	s.put(&models.AppGame{AppId: "app1", GameBrand: "jili", GameId: "1001"})
	s.put(&models.AppGame{AppId: "app1", GameBrand: "pg", GameId: "2001"})
	s.put(&models.AppGame{AppId: "app2", GameBrand: "jili", GameId: "3001"})

	// 按 appId 替换：删除旧条目，写入新列表
	s.replaceByAppID("app1", []models.AppGame{
		{AppId: "app1", GameBrand: "jili", GameId: "1001"},
		{AppId: "app1", GameBrand: "jili", GameId: "1002"},
	})

	if _, ok := s.Get("app1", "pg", "2001"); ok {
		t.Fatal("old app1/pg/2001 should be removed")
	}
	if got, ok := s.Get("app1", "jili", "1002"); !ok || got.GameId != "1002" {
		t.Fatalf("new app1 game missing: %+v ok=%v", got, ok)
	}
	if _, ok := s.Get("app2", "jili", "3001"); !ok {
		t.Fatal("app2 entry should remain")
	}

	// 空列表：清除该 appId 全部缓存
	s.replaceByAppID("app1", nil)
	if _, ok := s.Get("app1", "jili", "1001"); ok {
		t.Fatal("app1 should be fully cleared")
	}
	if _, ok := s.Get("app2", "jili", "3001"); !ok {
		t.Fatal("app2 entry should remain after clearing app1")
	}
}

func TestAppInfoStoreGetPutRemove(t *testing.T) {
	s := NewAppInfoStore(nil)
	s.put(&models.AppInfo{AppId: "a1", AccessKeyId: "ak1"})
	got, ok := s.GetByAppID("a1")
	if !ok || got.AccessKeyId != "ak1" {
		t.Fatalf("GetByAppID failed: %+v ok=%v", got, ok)
	}
	got, ok = s.GetByAccessKeyID("ak1")
	if !ok || got.AppId != "a1" {
		t.Fatalf("GetByAccessKeyID failed: %+v ok=%v", got, ok)
	}

	s.put(&models.AppInfo{AppId: "a1", AccessKeyId: "ak2"})
	if _, ok := s.GetByAccessKeyID("ak1"); ok {
		t.Fatal("old access key should be removed")
	}
	if got, ok := s.GetByAccessKeyID("ak2"); !ok || got.AppId != "a1" {
		t.Fatalf("new access key missing: %+v", got)
	}

	s.remove("a1")
	if _, ok := s.GetByAppID("a1"); ok {
		t.Fatal("should be removed")
	}
	if _, ok := s.GetByAccessKeyID("ak2"); ok {
		t.Fatal("access key should be removed")
	}
}

func TestAppGameStoreGetPutRemove(t *testing.T) {
	s := NewAppGameStore(nil)
	s.put(&models.AppGame{AppId: "app1", GameBrand: "jili", GameId: "1001"})
	got, ok := s.Get("app1", "jili", "1001")
	if !ok || got.GameId != "1001" {
		t.Fatalf("Get failed: %+v ok=%v", got, ok)
	}
	s.remove(AppGameKey("app1", "jili", "1001"))
	if _, ok := s.Get("app1", "jili", "1001"); ok {
		t.Fatal("should be removed")
	}
}

func TestGameInfoReplaceByBrand(t *testing.T) {
	s := NewGameInfoStore(nil)
	s.put(&models.GameInfo{GameBrand: "jili", GameId: "1001"})
	s.put(&models.GameInfo{GameBrand: "jili", GameId: "1002"})
	s.put(&models.GameInfo{GameBrand: "pg", GameId: "2001"})

	s.replaceByBrand("jili", []models.GameInfo{
		{GameBrand: "jili", GameId: "1001"},
		{GameBrand: "jili", GameId: "1003"},
	})

	if _, ok := s.Get("jili", "1002"); ok {
		t.Fatal("old jili/1002 should be removed")
	}
	if got, ok := s.Get("jili", "1003"); !ok || got.GameId != "1003" {
		t.Fatalf("new jili game missing: %+v ok=%v", got, ok)
	}
	if _, ok := s.Get("pg", "2001"); !ok {
		t.Fatal("pg entry should remain")
	}

	s.replaceByBrand("jili", nil)
	if _, ok := s.Get("jili", "1001"); ok {
		t.Fatal("jili should be fully cleared")
	}
	if _, ok := s.Get("pg", "2001"); !ok {
		t.Fatal("pg entry should remain after clearing jili")
	}
}

func TestGameInfoStoreGetPutRemove(t *testing.T) {
	s := NewGameInfoStore(nil)
	s.put(&models.GameInfo{GameBrand: "jili", GameId: "1001", GameName: "demo"})
	got, ok := s.Get("jili", "1001")
	if !ok || got.GameName != "demo" {
		t.Fatalf("Get failed: %+v ok=%v", got, ok)
	}
	s.remove(GameInfoKey("jili", "1001"))
	if _, ok := s.Get("jili", "1001"); ok {
		t.Fatal("should be removed")
	}
}

func TestAppGameBrandReplaceByAppID(t *testing.T) {
	s := NewAppGameBrandStore(nil)
	s.put(&models.AppGameBrand{AppId: "app1", GameBrand: "jili", GameType: "slot"})
	s.put(&models.AppGameBrand{AppId: "app1", GameBrand: "pg", GameType: "slot"})
	s.put(&models.AppGameBrand{AppId: "app2", GameBrand: "jili", GameType: "slot"})

	s.replaceByAppID("app1", []models.AppGameBrand{
		{AppId: "app1", GameBrand: "jili", GameType: "slot", GameGgr: 0.15},
		{AppId: "app1", GameBrand: "jili", GameType: "fish", GameGgr: 0.20},
	})

	if _, ok := s.Get("app1", "pg", "slot"); ok {
		t.Fatal("old app1/pg/slot should be removed")
	}
	if got, ok := s.Get("app1", "jili", "fish"); !ok || got.GameGgr != 0.20 {
		t.Fatalf("new app1 config missing: %+v ok=%v", got, ok)
	}
	if _, ok := s.Get("app2", "jili", "slot"); !ok {
		t.Fatal("app2 entry should remain")
	}

	s.replaceByAppID("app1", nil)
	if _, ok := s.Get("app1", "jili", "slot"); ok {
		t.Fatal("app1 should be fully cleared")
	}
}

func TestAppGameBrandStoreGetPutRemove(t *testing.T) {
	s := NewAppGameBrandStore(nil)
	s.put(&models.AppGameBrand{AppId: "app1", GameBrand: "jili", GameType: "slot", GameGgr: 0.12})
	got, ok := s.Get("app1", "jili", "slot")
	if !ok || got.GameGgr != 0.12 {
		t.Fatalf("Get failed: %+v ok=%v", got, ok)
	}
	s.remove(AppGameBrandKey("app1", "jili", "slot"))
	if _, ok := s.Get("app1", "jili", "slot"); ok {
		t.Fatal("should be removed")
	}
}

func TestManagerRegisterTypedAccessors(t *testing.T) {
	mgr := NewManager(nil, nil, Options{})
	appInfo := NewAppInfoStore(nil)
	appGame := NewAppGameStore(nil)
	gameInfo := NewGameInfoStore(nil)
	appGameBrand := NewAppGameBrandStore(nil)
	mgr.Register(appInfo)
	mgr.Register(appGame)
	mgr.Register(gameInfo)
	mgr.Register(appGameBrand)

	if mgr.AppInfo() != appInfo {
		t.Fatal("AppInfo accessor mismatch")
	}
	if mgr.AppGame() != appGame {
		t.Fatal("AppGame accessor mismatch")
	}
	if mgr.GameInfo() != gameInfo {
		t.Fatal("GameInfo accessor mismatch")
	}
	if mgr.AppGameBrand() != appGameBrand {
		t.Fatal("AppGameBrand accessor mismatch")
	}

	appInfo.put(&models.AppInfo{AppId: "x", AccessKeyId: "y"})
	if got, ok := mgr.AppInfo().GetByAppID("x"); !ok || got.AccessKeyId != "y" {
		t.Fatalf("GetByAppID via manager: %+v ok=%v", got, ok)
	}
}

func TestManagerStartRequiresRedis(t *testing.T) {
	mgr := NewManager(nil, nil, Options{})
	mgr.Register(&mockStore{name: "x"})
	if err := mgr.Start(context.Background()); err == nil {
		t.Fatal("expected error when redis is nil")
	}
}

func TestInitRequiresDB(t *testing.T) {
	if _, err := Init(context.Background(), nil, nil, Options{}); err == nil {
		t.Fatal("expected error when db is nil")
	}
}
