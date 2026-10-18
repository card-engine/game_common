package sharding

import (
	"testing"
	"time"

	"github.com/card-engine/game_common/models"
)

func TestModeFromState(t *testing.T) {
	if ModeFromState(0) != OnlyMain {
		t.Fatalf("state0 want OnlyMain")
	}
	if ModeFromState(1) != MainAndHistory {
		t.Fatalf("state1 want MainAndHistory")
	}
	if ModeFromState(2) != OnlyMain {
		t.Fatalf("state2 deprecated, want OnlyMain")
	}
	if ModeFromState(3) != MainAndDaily {
		t.Fatalf("state3 want MainAndDaily")
	}
}

func TestInitializeRulesModes(t *testing.T) {
	router := NewAppGameRecordRouter()
	router.InitializeRules([]*models.AppInfo{
		{AppId: "a0", ShardingState: 0},
		{AppId: "a1", ShardingState: 1},
		{AppId: "a2", ShardingState: 2},
		{AppId: "a3", ShardingState: 3},
	})
	if router.GetMode("a0") != OnlyMain || router.GetMode("a1") != MainAndHistory ||
		router.GetMode("a2") != OnlyMain || router.GetMode("a3") != MainAndDaily {
		t.Fatalf("unexpected modes: %v %v %v %v",
			router.GetMode("a0"), router.GetMode("a1"), router.GetMode("a2"), router.GetMode("a3"))
	}
	if router.GetMode("missing") != OnlyMain {
		t.Fatalf("missing should default OnlyMain")
	}
}

func TestGetWriteTable(t *testing.T) {
	router := NewAppGameRecordRouter()
	router.UpdateRule("18038", MainAndDaily, nil)
	router.UpdateRule("1001", MainAndHistory, nil)

	if got := router.GetWriteTable("18038"); got != "app_game_record_18038" {
		t.Fatalf("state3 write got %s", got)
	}
	if got := router.GetWriteTable("1001"); got != "app_game_record_1001" {
		t.Fatalf("state1 write got %s", got)
	}
}

func TestGetSettleTables(t *testing.T) {
	router := NewAppGameRecordRouter()
	now := time.Date(2026, 9, 8, 12, 0, 0, 0, time.UTC)

	onlyMain := router.GetSettleTables("1001", SettleOpt{})
	if len(onlyMain) != 1 || onlyMain[0] != "app_game_record_1001" {
		t.Fatalf("state0/1 settle tables=%v", onlyMain)
	}

	router.UpdateRule("18039", MainAndDaily, nil)
	mainDaily := router.GetSettleTables("18039", SettleOpt{
		RecentColdDays: 1,
		Now:            now,
	})
	wantMainDaily := []string{
		"app_game_record_18039",
		"app_game_record_18039_20260908",
		"app_game_record_18039_20260907",
	}
	if len(mainDaily) != len(wantMainDaily) {
		t.Fatalf("state3 settle len=%d want %d tables=%v", len(mainDaily), len(wantMainDaily), mainDaily)
	}
	for i := range wantMainDaily {
		if mainDaily[i] != wantMainDaily[i] {
			t.Fatalf("state3 settle idx %d got %s want %s", i, mainDaily[i], wantMainDaily[i])
		}
	}
}

func TestGetColdTable(t *testing.T) {
	router := NewAppGameRecordRouter()
	day := time.Date(2026, 9, 8, 1, 2, 3, 0, time.FixedZone("CST", 8*3600))
	if router.GetColdTable("18038", day) != "app_game_record_18038_20260907" {
		// CST 01:02 → UTC 前一日 17:02 → 20260907
		t.Fatalf("cold utc day got %s", router.GetColdTable("18038", day))
	}
}

func TestGetQueryTablesMainAndDaily(t *testing.T) {
	router := NewAppGameRecordRouter()
	start := time.Date(2026, 9, 7, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 9, 8, 23, 0, 0, 0, time.UTC)
	router.UpdateRule("18038", MainAndDaily, &start)

	tables := router.GetQueryTables("18038", start, end)
	mustContain := []string{
		"app_game_record_18038",
		"app_game_record_18038_20260906", // start-1 pad
		"app_game_record_18038_20260907",
		"app_game_record_18038_20260908",
		"app_game_record_18038_20260909", // end+1 pad
	}
	set := map[string]struct{}{}
	for _, tb := range tables {
		set[tb] = struct{}{}
	}
	for _, w := range mustContain {
		if _, ok := set[w]; !ok {
			t.Fatalf("missing %s in %v", w, tables)
		}
	}
	if got := router.GetWriteTable("18038"); got != "app_game_record_18038" {
		t.Fatalf("state3 write must be main, got %s", got)
	}
}

func TestGetQueryTablesOnlyMainAndHistoryUnchanged(t *testing.T) {
	router := NewAppGameRecordRouter()
	router.Rules["2001"] = &AppGameRecordShardRule{AppID: "2001", Mode: OnlyMain}
	start, _ := time.Parse("2006-01-02", "2025-06-01")
	end, _ := time.Parse("2006-01-02", "2025-09-30")

	tables := router.GetQueryTables("2001", start, end)
	if len(tables) != 1 || tables[0] != "app_game_record_2001" {
		t.Fatalf("only main got %v", tables)
	}

	router.Rules["1001"] = &AppGameRecordShardRule{AppID: "1001", Mode: MainAndHistory}
	tables = router.GetQueryTables("1001", start, end)
	if len(tables) == 0 {
		t.Fatal("main+history empty")
	}
	// 近 30 天应含主表（测试日期相对「今天」可能全是历史月表，但至少有月表或主表）
	hasMainOrMonth := false
	for _, tb := range tables {
		if tb == "app_game_record_1001" || len(tb) > len("app_game_record_1001_") {
			hasMainOrMonth = true
			break
		}
	}
	if !hasMainOrMonth {
		t.Fatalf("unexpected tables %v", tables)
	}
}
