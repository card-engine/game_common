package models

import (
	"time"
)

const (
	ProxyModel_Local = "Local" // 本地模式
)

type AppGame struct {
	ID           int64     `gorm:"column:id;type:bigint(20);primaryKey;" json:"id"`                                               // 主键ID
	AppId        string    `gorm:"column:app_id;type:varchar(32);not null;comment:'appId'" json:"app_id"`                         // 应用ID
	GameId       string    `gorm:"column:game_id;type:varchar(32);not null;comment:'游戏id'" json:"game_id"`                        // 游戏ID
	GameName     string    `gorm:"column:game_name;type:varchar(64);default:NULL;comment:'游戏名称'" json:"game_name"`                // 游戏名称
	GameFullName string    `gorm:"column:game_full_name;type:varchar(512);default:NULL;comment:'游戏Full名称'" json:"game_full_name"` // 游戏Full名称
	GameIcon     string    `gorm:"column:game_icon;type:varchar(512);default:NULL;comment:'游戏icon'" json:"game_icon"`             // 游戏icon
	GameType     string    `gorm:"column:game_type;type:varchar(32);default:NULL;comment:'游戏类型:slot'" json:"game_type"`           // 游戏类型
	GameBrand    string    `gorm:"column:game_brand;type:varchar(32);default:NULL;comment:'游戏厂商:jili,pg'" json:"game_brand"`      // 游戏厂商
	Status       string    `gorm:"column:status;type:varchar(8);default:ENABLE;comment:'状态：ENABLE,DISABLE'" json:"status"`        // 状态
	Rtp          string    `gorm:"column:rtp;type:varchar(32);default:NULL" json:"rtp"`
	ProxyModel   string    `gorm:"column:proxy_model;type:varchar(8);default:Local;comment:'代理模式：空值或Local为本地'" json:"proxyModel"`                                     // 代理模式                                // 状态
	CreatedAt    time.Time `gorm:"column:created_at;type:datetime;default:CURRENT_TIMESTAMP;comment:'创建时间'" json:"created_at"`                                        // 创建时间
	UpdatedAt    time.Time `gorm:"column:updated_at;type:datetime;not null;default:CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP;comment:'db更新时间'" json:"updated_at"` // 数据库更新时间
}

// TableName 指定表名
// GORM默认会使用结构体名的复数形式作为表名，这里显式指定为数据库中的实际表名
func (a *AppGame) TableName() string {
	return "app_game"
}

// AppGameRecord 商家游戏记录
type AppGameRecord struct {
	ID                  int64      `gorm:"column:id;primaryKey;" json:"id"`
	AppID               string     `gorm:"column:app_id;type:varchar(32);not null" json:"app_id"`
	PlayerID            string     `gorm:"column:player_id;type:varchar(64);not null" json:"player_id"`
	GameID              string     `gorm:"column:game_id;type:varchar(32);not null" json:"game_id"`
	GameBrand           string     `gorm:"column:game_brand;type:varchar(32);not null" json:"game_brand"`
	GameType            string     `gorm:"column:game_type;type:varchar(32)" json:"game_type"`
	RoundID             string     `gorm:"column:round_id;type:varchar(64)" json:"round_id"`
	PreRoundID          string     `gorm:"column:pre_round_id;type:varchar(64)" json:"pre_round_id"`
	TransactionID       string     `gorm:"column:transaction_id;type:varchar(64)" json:"transaction_id"` // 下注交易id
	Currency            string     `gorm:"column:currency;type:varchar(32)" json:"currency"`
	Rtp                 string     `gorm:"column:rtp;type:varchar(32)" json:"rtp"`
	Bet                 float64    `gorm:"column:bet;type:decimal(16,2)" json:"bet"`
	Win                 float64    `gorm:"column:win;type:decimal(16,2)" json:"win"`
	PostBalance         float64    `gorm:"column:post_balance;type:decimal(16,2)" json:"post_balance"` // 下注后余额
	Status              string     `gorm:"column:status;type:varchar(10);default:ENABLE" json:"status"`
	BetTime             *time.Time `gorm:"column:bet_time" json:"bet_time"`
	WinTime             *time.Time `gorm:"column:win_time" json:"win_time"`
	StatusTime          *time.Time `gorm:"column:status_time" json:"status_time"`
	TraceID             string     `gorm:"column:trace_id;type:varchar(64)" json:"trace_id"`
	CreatedAt           time.Time  `gorm:"column:created_at;type:datetime;comment:'创建时间'" json:"created_at"`
	UpdatedAt           time.Time  `gorm:"column:updated_at;type:datetime;autoUpdateTime;comment:'db更新时间'" json:"updated_at"`
	GameData            []byte     `gorm:"column:game_data;type:longblob" json:"game_data"`                                         // 完整game数据
	RoundModel          string     `gorm:"column:round_model;type:varchar(32);default:NULL" json:"round_model"`                     // 模式
	WinTransactionId    string     `gorm:"column:win_transaction_id;type:varchar(64);default:NULL" json:"win_transaction_id"`       // 派奖交易id
	RefundTransactionId string     `gorm:"column:refund_transaction_id;type:varchar(64);default:NULL" json:"refund_transaction_id"` // 取消交易id
	IsFree              bool       `gorm:"default:false;comment:'是否是免费下注'" json:"is_free"`
	PreTransactionID    string     `gorm:"column:pre_transaction_id;type:varchar(64)" json:"pre_transaction_id"` // 上一次下注交易id
	Note                string     `gorm:"column:note;type:varchar(256)" json:"note"`                            // 备注
	PreBalance          float64    `gorm:"column:pre_balance;type:decimal(16,2)" json:"pre_balance"`             // 下注前余额
	WinBalance          float64    `gorm:"column:win_balance;type:decimal(16,2)" json:"win_balance"`             // 派奖后余额
	RefundBalance       float64    `gorm:"column:refund_balance;type:decimal(16,2)" json:"refund_balance"`       // 取消后余额
}

// AppGameRecord 状态常量定义
const (
	AppGameRecordStatusInit     = "INIT"     // 初始化
	AppGameRecordStatusBet      = "BET"      // 下注完成
	AppGameRecordStatusSettled  = "SETTLED"  // 结算完成
	AppGameRecordStatusCanceled = "CANCELED" // 撤销完成
	AppGameRecordStatusError    = "ERROR"    // 故障
)

// 可选：定义状态描述映射
var AppGameRecordStatusDesc = map[string]string{
	AppGameRecordStatusInit:     "初始化",
	AppGameRecordStatusBet:      "下注完成",
	AppGameRecordStatusSettled:  "结算完成",
	AppGameRecordStatusCanceled: "撤销完成",
	AppGameRecordStatusError:    "故障",
}

// TableName 设置表名
//func (AppGameRecord) TableName() string {
//	return "app_game_record"
//}
