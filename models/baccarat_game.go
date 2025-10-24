package models

import (
	"time"
)

// AppBaccaratRecord 百家乐游戏记录模型
type AppBaccaratRecord struct {
	ID             int64     `gorm:"column:id;type:bigint;primaryKey;autoIncrement" json:"id"`
	AppID          string    `gorm:"column:app_id;type:varchar(32);not null" json:"app_id"`
	GameID         string    `gorm:"column:game_id;type:varchar(32);not null" json:"game_id"`
	GameBrand      string    `gorm:"column:game_brand;type:varchar(32);not null" json:"game_brand"`
	GameType       string    `gorm:"column:game_type;type:varchar(32)" json:"game_type"`
	RoundID        string    `gorm:"column:round_id;type:varchar(64)" json:"round_id"`
	PreRoundID     string    `gorm:"column:pre_round_id;type:varchar(64)" json:"pre_round_id"`
	RoomName       string    `gorm:"column:room_name;type:varchar(128)" json:"room_name"`
	Currency       string    `gorm:"column:currency;type:varchar(32)" json:"currency"`
	Rtp            string    `gorm:"column:rtp;type:varchar(32)" json:"rtp"`
	TotalPlayerBet float64   `gorm:"column:total_player_bet;type:decimal(16,2)" json:"total_player_bet"`
	TotalPlayerWin float64   `gorm:"column:total_player_win;type:decimal(16,2)" json:"total_player_win"`
	Status         string    `gorm:"column:status;type:varchar(10);default:INIT" json:"status"`
	StartTime      time.Time `gorm:"column:start_time;type:datetime" json:"start_time"`
	EndTime        time.Time `gorm:"column:end_time;type:datetime" json:"end_time"`
	TraceID        string    `gorm:"column:trace_id;type:varchar(64)" json:"trace_id"`
	CreatedAt      time.Time `gorm:"column:created_at;type:datetime" json:"created_at"`
	UpdatedAt      time.Time `gorm:"column:updated_at;type:datetime" json:"updated_at"`
	GameData       []byte    `gorm:"column:game_data;type:longblob" json:"game_data"`
	RoundModel     string    `gorm:"column:round_model;type:varchar(32)" json:"round_model"`
	IsFree         bool      `gorm:"column:is_free;type:tinyint(1);default:0" json:"is_free"`
}

// TableName 指定表名
// GORM默认会使用结构体名的复数形式作为表名，这里显式指定为数据库中的实际表名
func (a *AppBaccaratRecord) TableName() string {
	return "app_baccarat_record"
}
