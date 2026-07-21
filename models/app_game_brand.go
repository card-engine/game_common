package models

import "time"

// AppGameBrand 对应 app_game_brand 表，商户游戏厂商配置。
type AppGameBrand struct {
	ID        int64     `gorm:"column:id;type:bigint(20);primaryKey;autoIncrement" json:"id"`
	AppId     string    `gorm:"column:app_id;type:varchar(32);not null;comment:'应用ID'" json:"app_id"`
	GameBrand string    `gorm:"column:game_brand;type:varchar(32);not null;comment:'游戏厂商:jili,pg'" json:"game_brand"`
	GameType  string    `gorm:"column:game_type;type:varchar(32);not null;comment:'游戏类型:slot'" json:"game_type"`
	GameGgr   float64   `gorm:"column:game_ggr;type:decimal(10,4);not null;default:0;comment:'GGR分成比例'" json:"game_ggr"`
	Status    string    `gorm:"column:status;type:varchar(8);not null;default:ENABLE;comment:'状态：ENABLE,DISABLE'" json:"status"`
	CreatedAt time.Time `gorm:"column:created_at;type:datetime;default:CURRENT_TIMESTAMP;comment:'创建时间'" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at;type:datetime;not null;default:CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP;comment:'更新时间'" json:"updated_at"`
}

func (AppGameBrand) TableName() string {
	return "app_game_brand"
}
