package sharding

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/card-engine/game_common/models"
)

// ShardMode 表示分片模式
type ShardMode int

const (
	OnlyMain       ShardMode = 0 // 只有主表（小商户），sharding_state=0
	MainAndHistory ShardMode = 1 // 主表 + 月历史表（大商户），sharding_state=1
	MainAndDaily   ShardMode = 3 // 主表 + UTC 日冷表，sharding_state=3
	// sharding_state=2（热表+日冷）已废弃，不再映射，按 OnlyMain 处理。
)

// DefaultRecentColdDays 结算/幂等查单时默认回看近几日冷表天数。
const DefaultRecentColdDays = 2

// DefaultColdDayPad 历史查询日冷表相对 [start,end] 各扩的天数。
// 冷表按归档 UTC 日命名，可能与 bet_time 差 1 天。
const DefaultColdDayPad = 1

// SettleOpt 结算/幂等查单候选表选项（MainAndDaily 生效）。
type SettleOpt struct {
	RecentColdDays int       // 近 N 日冷表；<=0 时用 DefaultRecentColdDays
	Now            time.Time // 锚点时间；零值用 time.Now().UTC()
}

// AppGameRecordShardRule 商户分表规则
type AppGameRecordShardRule struct {
	AppID             string
	Mode              ShardMode
	shardingStartDate *time.Time
}

// AppGameRecordRouter 分表路由器
type AppGameRecordRouter struct {
	TablePrefix string
	Rules       map[string]*AppGameRecordShardRule
	mu          sync.RWMutex
}

// NewAppGameRecordRouter 初始化
func NewAppGameRecordRouter() *AppGameRecordRouter {
	return &AppGameRecordRouter{
		TablePrefix: "app_game_record",
		Rules:       make(map[string]*AppGameRecordShardRule),
	}
}

// ModeFromState 将 app_info.sharding_state 映射为 ShardMode。
func ModeFromState(state uint8) ShardMode {
	switch state {
	case 1:
		return MainAndHistory
	case 3:
		return MainAndDaily
	default:
		// 含已废弃的 sharding_state=2
		return OnlyMain
	}
}

// InitializeRules 根据 AppInfo 列表初始化分片规则
func (r *AppGameRecordRouter) InitializeRules(appInfos []*models.AppInfo) {
	tempRules := make(map[string]*AppGameRecordShardRule)
	for _, appInfo := range appInfos {
		tempRules[appInfo.AppId] = &AppGameRecordShardRule{
			AppID:             appInfo.AppId,
			Mode:              ModeFromState(appInfo.ShardingState),
			shardingStartDate: appInfo.ShardingStartDate,
		}
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.Rules = tempRules
}

// UpdateRule 更新或添加指定 appID 的分片规则
func (r *AppGameRecordRouter) UpdateRule(appID string, mode ShardMode, shardingStartDate *time.Time) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if rule, exists := r.Rules[appID]; exists {
		rule.Mode = mode
		rule.shardingStartDate = shardingStartDate
	} else {
		r.Rules[appID] = &AppGameRecordShardRule{
			AppID:             appID,
			Mode:              mode,
			shardingStartDate: shardingStartDate,
		}
	}
}

// GetMode 返回商户分片模式；无规则时默认 OnlyMain。
func (r *AppGameRecordRouter) GetMode(appID string) ShardMode {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if rule, ok := r.Rules[appID]; ok {
		return rule.Mode
	}
	return OnlyMain
}

// GetMainTable 获取主表名（state 0/1/3 写表）
func (r *AppGameRecordRouter) GetMainTable(appID string) string {
	return fmt.Sprintf("%s_%s", r.TablePrefix, appID)
}

// GetColdTable 获取日冷表名（UTC 日期 YYYYMMDD）
func (r *AppGameRecordRouter) GetColdTable(appID string, t time.Time) string {
	return fmt.Sprintf("%s_%s_%s", r.TablePrefix, appID, t.UTC().Format("20060102"))
}

// GetWriteTable 获取写入表，始终为主表。
func (r *AppGameRecordRouter) GetWriteTable(appID string) string {
	return r.GetMainTable(appID)
}

// GetSettleTables 结算/幂等候选表有序列表。
// state=3: [main] + 近 N 日冷表
// 其它: [main]
func (r *AppGameRecordRouter) GetSettleTables(appID string, opt SettleOpt) []string {
	if r.GetMode(appID) != MainAndDaily {
		return []string{r.GetMainTable(appID)}
	}
	now := opt.Now
	if now.IsZero() {
		now = time.Now().UTC()
	} else {
		now = now.UTC()
	}
	days := opt.RecentColdDays
	if days <= 0 {
		days = DefaultRecentColdDays
	}
	tables := make([]string, 0, 1+days)
	tables = append(tables, r.GetMainTable(appID))
	for i := 0; i <= days; i++ {
		tables = append(tables, r.GetColdTable(appID, now.AddDate(0, 0, -i)))
	}
	return tables
}

// GetQueryTables 获取历史查询表列表
func (r *AppGameRecordRouter) GetQueryTables(appID string, start, end time.Time) []string {
	r.mu.RLock()
	rule, ok := r.Rules[appID]
	r.mu.RUnlock()
	if !ok || rule.Mode == OnlyMain {
		return []string{r.GetMainTable(appID)}
	}
	if rule.Mode == MainAndDaily {
		return r.getTablesMainAndDaily(appID, rule, start, end)
	}
	return r.getTablesWithHistory(appID, rule, start, end)
}

// getTablesMainAndDaily：主表 + UTC 日冷表（按查询区间，各扩 DefaultColdDayPad 天）。
// 不含热表；不扫存量月表（state=3 终态只进日冷，月迁任务会跳过）。
func (r *AppGameRecordRouter) getTablesMainAndDaily(appID string, rule *AppGameRecordShardRule, start, end time.Time) []string {
	var tables []string
	seen := make(map[string]struct{})
	add := func(name string) {
		if _, ok := seen[name]; ok {
			return
		}
		seen[name] = struct{}{}
		tables = append(tables, name)
	}

	add(r.GetMainTable(appID))

	actualStart := start
	if rule.shardingStartDate != nil && rule.shardingStartDate.After(start) {
		actualStart = *rule.shardingStartDate
	}
	startDay := time.Date(actualStart.UTC().Year(), actualStart.UTC().Month(), actualStart.UTC().Day(), 0, 0, 0, 0, time.UTC)
	endDay := time.Date(end.UTC().Year(), end.UTC().Month(), end.UTC().Day(), 0, 0, 0, 0, time.UTC)
	startDay = startDay.AddDate(0, 0, -DefaultColdDayPad)
	endDay = endDay.AddDate(0, 0, DefaultColdDayPad)
	for d := startDay; !d.After(endDay); d = d.AddDate(0, 0, 1) {
		add(r.GetColdTable(appID, d))
	}
	return tables
}

// getTablesWithHistory 获取大商户查询表列表
func (r *AppGameRecordRouter) getTablesWithHistory(appID string, rule *AppGameRecordShardRule, start, end time.Time) []string {
	var tables []string
	now := time.Now()

	if end.After(now.AddDate(0, 0, -30)) {
		tables = append(tables, r.GetMainTable(appID))
	}
	actualStart := start
	if rule.shardingStartDate != nil && rule.shardingStartDate.After(start) {
		actualStart = *rule.shardingStartDate
	}
	startMonth := time.Date(actualStart.Year(), actualStart.Month(), 1, 0, 0, 0, 0, time.UTC)
	endMonth := time.Date(end.Year(), end.Month(), 1, 0, 0, 0, 0, time.UTC)

	for d := startMonth; !d.After(endMonth); d = d.AddDate(0, 1, 0) {
		if d.Before(now.AddDate(0, 0, -30)) {
			historyTable := fmt.Sprintf("%s_%s_%s", r.TablePrefix, appID, d.Format("200601"))
			tables = append(tables, historyTable)
		}
	}

	return tables
}

// BuildAppGameRecordUnionSQL 根据查询表列表生成 UNION ALL SQL
func BuildAppGameRecordUnionSQL(tables []string, fields, whereClause string) string {
	if len(tables) == 0 {
		return ""
	}
	var sqlParts []string
	for _, t := range tables {
		part := fmt.Sprintf("SELECT %s FROM %s WHERE %s", fields, t, whereClause)
		sqlParts = append(sqlParts, part)
	}
	return strings.Join(sqlParts, " UNION ALL ")
}
