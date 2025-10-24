package models

import (
	"time"
)

/**
* 游戏的数据
**/
type JdbGameInfo struct {
	Id uint `gorm:"primaryKey;comment:id;"`

	TypeId   string `gorm:"ize:5;comment:厂商游戏类型id;"`
	GameId   string `gorm:"size:10;comment:厂商游戏id;uniqueIndex;"`
	GameName string `gorm:"comment:游戏的名称;"`
	GameRes  string `gorm:"comment:游戏的资源名;"`

	GameType string `gorm:"comment:游戏类型:slot,fish,table,crash;"` // 游戏类型:slot,fish,table,crash

	CreateTime time.Time `gorm:"autoCreateTime;comment:创建时间;"`
	UpdateTime time.Time `gorm:"autoCreateTime;comment:创建时间;"`
}

// 备忘录
func (JdbGameInfo) TableName() string {
	return "jdb_info"
}
