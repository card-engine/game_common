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
	if ModeFromState(2) != HotAndDaily {
		t.Fatalf("state2 want HotAndDaily")
	}
}

func TestInitializeRulesModes(t *testing.T) {
	router := NewAppGameRecordRouter()
	router.InitializeRules([]*models.AppInfo{
		{AppId: "a0", ShardingState: 0},
		{AppId: "a1", ShardingState: 1},
		{AppId: "a2", ShardingState: 2},
	})
	if router.GetMode("a0") != OnlyMain || router.GetMode("a1") != MainAndHistory || router.GetMode("a2") != HotAndDaily {
		t.Fatalf("unexpected modes: %v %v %v", router.GetMode("a0"), router.GetMode("a1"), router.GetMode("a2"))
	}
	if router.GetMode("missing") != OnlyMain {
		t.Fatalf("missing should default OnlyMain")
	}
}

func TestGetWriteTable(t *testing.T) {
	router := NewAppGameRecordRouter()
	router.UpdateRule("18038", HotAndDaily, nil)
	router.UpdateRule("1001", MainAndHistory, nil)

	if got := router.GetWriteTable("18038", true); got != "app_game_record_18038_hot" {
		t.Fatalf("write hot got %s", got)
	}
	if got := router.GetWriteTable("18038", false); got != "app_game_record_18038" {
		t.Fatalf("write main got %s", got)
	}
	if got := router.GetWriteTable("1001", true); got != "app_game_record_1001" {
		t.Fatalf("state1 ignore writeHot, got %s", got)
	}
}

func TestGetSettleTables(t *testing.T) {
	router := NewAppGameRecordRouter()
	router.UpdateRule("18038", HotAndDaily, nil)
	now := time.Date(2026, 9, 8, 12, 0, 0, 0, time.UTC)

	tables := router.GetSettleTables("18038", SettleOpt{
		FallbackMain:   true,
		RecentColdDays: 2,
		Now:            now,
	})
	want := []string{
		"app_game_record_18038_hot",
		"app_game_record_18038_20260908",
		"app_game_record_18038_20260907",
		"app_game_record_18038_20260906",
		"app_game_record_18038",
	}
	if len(tables) != len(want) {
		t.Fatalf("len=%d want %d tables=%v", len(tables), len(want), tables)
	}
	for i := range want {
		if tables[i] != want[i] {
			t.Fatalf("idx %d got %s want %s", i, tables[i], want[i])
		}
	}

	onlyMain := router.GetSettleTables("1001", SettleOpt{FallbackMain: true})
	if len(onlyMain) != 1 || onlyMain[0] != "app_game_record_1001" {
		t.Fatalf("state0/1 settle tables=%v", onlyMain)
	}
}

func TestGetColdAndHotTable(t *testing.T) {
	router := NewAppGameRecordRouter()
	if router.GetHotTable("18038") != "app_game_record_18038_hot" {
		t.Fatal("hot")
	}
	day := time.Date(2026, 9, 8, 1, 2, 3, 0, time.FixedZone("CST", 8*3600))
	if router.GetColdTable("18038", day) != "app_game_record_18038_20260907" {
		// CST 01:02 → UTC 前一日 17:02 → 20260907
		t.Fatalf("cold utc day got %s", router.GetColdTable("18038", day))
	}
	if !IsHotTable("app_game_record_18038_hot") || IsHotTable("app_game_record_18038_20260908") {
		t.Fatal("IsHotTable")
	}
}

func TestGetQueryTablesHotAndDaily(t *testing.T) {
	router := NewAppGameRecordRouter()
	start := time.Date(2026, 9, 7, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 9, 8, 23, 0, 0, 0, time.UTC)
	router.UpdateRule("18038", HotAndDaily, &start)

	tables := router.GetQueryTables("18038", start, end)
	mustContain := []string{
		"app_game_record_18038_hot",
		"app_game_record_18038",
		"app_game_record_18038_20260907",
		"app_game_record_18038_20260908",
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
