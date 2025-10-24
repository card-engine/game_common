package models

// 玩家的RTP信息

import "time"

type GameRtp struct {
	Id uint64 `gorm:"primaryKey;comment:id;"`

	AppId  string `gorm:"column:app_id;comment:应用ID;"`
	Brand  string `gorm:"column:brand;comment:品牌;"`
	GameId string `gorm:"column:game_id;comment:游戏ID;"`

	Rtp string `gorm:"column:rtp;comment:默认rtp;"`

	CreateTime time.Time `gorm:"autoCreateTime;comment:创建时间;"`
	UpdateTime time.Time `gorm:"autoCreateTime;comment:创建时间;"`
}

func (GameRtp) TableName() string {
	return "game_rtp"
}
