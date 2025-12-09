package sharding

import (
	"fmt"
	"testing"
	"time"
)

func TestAppGameRecordRouter(t *testing.T) {
	router := NewAppGameRecordRouter()

	// 初始化商户规则
	router.Rules["1001"] = &AppGameRecordShardRule{AppID: "1001", Mode: MainAndHistory} // 大商户
	router.Rules["2001"] = &AppGameRecordShardRule{AppID: "2001", Mode: OnlyMain}       // 小商户

	fmt.Println("Main Table:", router.GetMainTable("1001"))

	start, _ := time.Parse("2006-01-02", "2025-06-01")
	end, _ := time.Parse("2006-01-02", "2025-09-30")

	tables := router.GetQueryTables("1002", start, end)
	fmt.Println("查询历史数据，只会查当前表 Tables:", tables)

	shardStart, _ := time.Parse("2006-01-02", "2025-07-01")
	router.UpdateRule("1002", MainAndHistory, &shardStart)
	tables = router.GetQueryTables("1002", start, end)
	fmt.Println("更改模式后，查询历史数据，查表 Tables:", tables)

	tables = router.GetQueryTables("1001", start, end)
	fmt.Println("查询历史数据，不包含当前表 Tables:", tables)

	end, _ = time.Parse("2006-01-02", "2025-09-30")
	tables = router.GetQueryTables("1001", start, end)
	fmt.Println("查询历史数据，包含当前表 Tables:", tables)

	sql := BuildAppGameRecordUnionSQL(tables, "id, user_id, score", "user_id=123")
	fmt.Println("Generated SQL:")
	fmt.Println(sql)
	/*
	   SELECT id, user_id, score FROM app_game_record_1001_202506 WHERE user_id=123
	*/
}
