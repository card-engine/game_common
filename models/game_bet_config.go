package models

import "time"

// GameBetConfig 对应 game.game_bet_config 表。
type GameBetConfig struct {
	ID             int64     `gorm:"column:id;type:bigint(20);primaryKey;autoIncrement:true" json:"id"`
	Region         string    `gorm:"column:region;type:varchar(32);not null" json:"region"`
	Currency       string    `gorm:"column:currency;type:varchar(32);not null" json:"currency"`
	AppId          string    `gorm:"column:app_id;type:varchar(32);not null" json:"appId"`
	GameBrand      string    `gorm:"column:game_brand;type:varchar(32);not null" json:"gameBrand"`
	GameId         string    `gorm:"column:game_id;type:varchar(64);not null" json:"gameId"`
	GameType       string    `gorm:"column:game_type;type:varchar(32);default:NULL" json:"gameType"`
	Status         string    `gorm:"column:status;type:varchar(16);default:ENABLE" json:"status"`
	BetConfig      string    `gorm:"column:bet_config;type:longtext;default:NULL" json:"betConfig"`
	ControlMaxRate int64     `gorm:"column:control_max_rate;type:bigint(20);not null;default:-1" json:"controlMaxRate"`
	ControlMaxCash float64   `gorm:"column:control_max_cash;type:decimal(20,4);not null;default:-1" json:"controlMaxCash"`
	Remark         string    `gorm:"column:remark;type:varchar(512);default:NULL" json:"remark"`
	CreatedBy      string    `gorm:"column:created_by;type:varchar(32);default:NULL" json:"createdBy"`
	UpdatedBy      string    `gorm:"column:updated_by;type:varchar(32);default:NULL" json:"updatedBy"`
	CreatedAt      time.Time `gorm:"column:created_at;type:datetime;default:CURRENT_TIMESTAMP" json:"createdAt"`
	UpdatedAt      time.Time `gorm:"column:updated_at;type:datetime;default:CURRENT_TIMESTAMP" json:"updatedAt"`
}

func (GameBetConfig) TableName() string {
	return "game.game_bet_config"
}
