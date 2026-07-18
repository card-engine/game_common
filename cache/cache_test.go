package cache

import (
	"context"
	"encoding/json"
	"sync"
	"testing"

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
	mgr := NewManager(nil, nil, Options{RefreshInterval: -1})
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
	mgr := NewManager(nil, nil, Options{RefreshInterval: -1})
	if err := mgr.Refresh(context.Background(), "unknown", ""); err == nil {
		t.Fatal("expected error for unknown type")
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

func TestGetPackageDefaults(t *testing.T) {
	_ = NewAppInfoStore(nil)
	_ = NewAppGameStore(nil)
	_ = NewGameInfoStore(nil)
	defaultAppInfoStore.put(&models.AppInfo{AppId: "x", AccessKeyId: "y"})
	defaultAppGameStore.put(&models.AppGame{AppId: "a", GameBrand: "b", GameId: "c"})
	defaultGameInfoStore.put(&models.GameInfo{GameBrand: "jili", GameId: "99", Status: "ENABLE"})

	if got, ok := GetAppInfo("x"); !ok || got.AccessKeyId != "y" {
		t.Fatalf("GetAppInfo: %+v ok=%v", got, ok)
	}
	if got, ok := GetAppInfoByAccessKey("y"); !ok || got.AppId != "x" {
		t.Fatalf("GetAppInfoByAccessKey: %+v ok=%v", got, ok)
	}
	if got, ok := GetAppGame("a", "b", "c"); !ok || got.GameBrand != "b" {
		t.Fatalf("GetAppGame: %+v ok=%v", got, ok)
	}
	if got, ok := GetGameInfo("jili", "99"); !ok || got.Status != "ENABLE" {
		t.Fatalf("GetGameInfo: %+v ok=%v", got, ok)
	}
}

func TestManagerStartRequiresRedis(t *testing.T) {
	mgr := NewManager(nil, nil, Options{RefreshInterval: -1})
	mgr.Register(&mockStore{name: "x"})
	if err := mgr.Start(context.Background()); err == nil {
		t.Fatal("expected error when redis is nil")
	}
}
