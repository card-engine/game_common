package models

import "time"

// GameGlobalConfig 对应 game.game_global_config 表。
type GameGlobalConfig struct {
	ID        int64     `gorm:"column:id;type:bigint(20);primaryKey;autoIncrement:true" json:"id"`
	GKey      string    `gorm:"column:gkey;type:varchar(64);not null;uniqueIndex" json:"gkey"`
	GValue    string    `gorm:"column:gvalue;type:longtext;default:NULL" json:"gvalue"`
	CreatedAt time.Time `gorm:"column:created_at;type:datetime;default:CURRENT_TIMESTAMP" json:"createdAt"`
	UpdatedAt time.Time `gorm:"column:updated_at;type:datetime;default:CURRENT_TIMESTAMP" json:"updatedAt"`
	UpdateOps string    `gorm:"column:update_ops;type:varchar(32);default:NULL" json:"updateOps"`
}

func (GameGlobalConfig) TableName() string {
	return "game.game_global_config"
}
