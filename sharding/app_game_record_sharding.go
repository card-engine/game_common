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
	OnlyMain       ShardMode = iota // 0 只有主表（小商户）
	MainAndHistory                  // 1 主表 + 月历史表（大商户）
	HotAndDaily                     // 2 热表 + 日冷表（超大商户）
)

// DefaultRecentColdDays 结算/幂等查单时默认回看近几日冷表天数。
const DefaultRecentColdDays = 2

// SettleOpt 结算/幂等查单候选表选项（仅 HotAndDaily 生效）。
type SettleOpt struct {
	FallbackMain   bool      // 是否包含旧主表兜底
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
	case 2:
		return HotAndDaily
	default:
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

// GetMainTable 获取旧主表名（state 0/1 写表；state=2 过渡期兜底）
func (r *AppGameRecordRouter) GetMainTable(appID string) string {
	return fmt.Sprintf("%s_%s", r.TablePrefix, appID)
}

// GetHotTable 获取热表名（未完结）
func (r *AppGameRecordRouter) GetHotTable(appID string) string {
	return fmt.Sprintf("%s_%s_hot", r.TablePrefix, appID)
}

// GetColdTable 获取日冷表名（UTC 日期 YYYYMMDD）
func (r *AppGameRecordRouter) GetColdTable(appID string, t time.Time) string {
	return fmt.Sprintf("%s_%s_%s", r.TablePrefix, appID, t.UTC().Format("20060102"))
}

// GetWriteTable 获取写入表：HotAndDaily 且 writeHot 时写热表，否则写主表。
func (r *AppGameRecordRouter) GetWriteTable(appID string, writeHot bool) string {
	if writeHot && r.GetMode(appID) == HotAndDaily {
		return r.GetHotTable(appID)
	}
	return r.GetMainTable(appID)
}

// GetSettleTables 结算/幂等候选表有序列表。
// state=2: [hot] + 近 N 日冷表 + (FallbackMain ? [main] : [])
// 其它: [main]
func (r *AppGameRecordRouter) GetSettleTables(appID string, opt SettleOpt) []string {
	if r.GetMode(appID) != HotAndDaily {
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
	tables := make([]string, 0, 1+days+1)
	tables = append(tables, r.GetHotTable(appID))
	for i := 0; i <= days; i++ {
		tables = append(tables, r.GetColdTable(appID, now.AddDate(0, 0, -i)))
	}
	if opt.FallbackMain {
		tables = append(tables, r.GetMainTable(appID))
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
	if rule.Mode == HotAndDaily {
		return r.getTablesHotAndDaily(appID, rule, start, end)
	}
	return r.getTablesWithHistory(appID, rule, start, end)
}

// getTablesHotAndDaily：热表 + 日冷表（按查询区间）+ 过渡主表 + 存量月表。
func (r *AppGameRecordRouter) getTablesHotAndDaily(appID string, rule *AppGameRecordShardRule, start, end time.Time) []string {
	var tables []string
	seen := make(map[string]struct{})
	add := func(name string) {
		if _, ok := seen[name]; ok {
			return
		}
		seen[name] = struct{}{}
		tables = append(tables, name)
	}

	add(r.GetHotTable(appID))
	add(r.GetMainTable(appID)) // 过渡期旧主表存量

	actualStart := start
	if rule.shardingStartDate != nil && rule.shardingStartDate.After(start) {
		actualStart = *rule.shardingStartDate
	}
	startDay := time.Date(actualStart.UTC().Year(), actualStart.UTC().Month(), actualStart.UTC().Day(), 0, 0, 0, 0, time.UTC)
	endDay := time.Date(end.UTC().Year(), end.UTC().Month(), end.UTC().Day(), 0, 0, 0, 0, time.UTC)
	for d := startDay; !d.After(endDay); d = d.AddDate(0, 0, 1) {
		add(r.GetColdTable(appID, d))
	}

	// 存量月表：覆盖查询区间内、且早于近 30 天窗口的月份（与 MainAndHistory 一致）
	now := time.Now()
	startMonth := time.Date(actualStart.Year(), actualStart.Month(), 1, 0, 0, 0, 0, time.UTC)
	endMonth := time.Date(end.Year(), end.Month(), 1, 0, 0, 0, 0, time.UTC)
	for d := startMonth; !d.After(endMonth); d = d.AddDate(0, 1, 0) {
		if d.Before(now.AddDate(0, 0, -30)) {
			add(fmt.Sprintf("%s_%s_%s", r.TablePrefix, appID, d.Format("200601")))
		}
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

// IsHotTable 判断表名是否为热表。
func IsHotTable(table string) bool {
	return strings.HasSuffix(table, "_hot")
}
