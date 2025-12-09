package models

import "time"

type Player struct {
	AId       uint64     `gorm:"primaryKey;comment:我方平台的用户id;"`
	AppId     string     `gorm:"size:64;comment:商用id;uniqueIndex:idx_app_account"`
	AccountId string     `gorm:"size:64;comment:第三方平台的用户id;uniqueIndex:idx_app_account"`
	NickName  string     `gorm:"comment:昵称;"`
	Balance   float64    `gorm:"comment:余额;"`
	Rtp       string     `gorm:"comment:游戏RTP;"`
	HasSetRtp bool       `gorm:"comment:是否设置了rtp;default:false"`
	RtpTime   *time.Time `gorm:"column:rtp_time;comment:rtp更新时间;"`

	CreateTime time.Time `gorm:"autoCreateTime;comment:创建时间;"`
	UpdateTime time.Time `gorm:"autoUpdateTime;comment:创建时间;"`
}

func (Player) TableName() string {
	return "player"
}

type PlayerLoginLog struct {
	ID        int64     `gorm:"column:id;primaryKey;" json:"id"`                               // id
	AppID     string    `gorm:"column:app_id;type:varchar(32);not null" json:"app_id"`         // appId
	PlayerID  string    `gorm:"column:player_id;type:varchar(64);not null" json:"player_id"`   // 玩家id
	GameID    string    `gorm:"column:game_id;type:varchar(32);not null" json:"game_id"`       // 游戏id
	GameBrand string    `gorm:"column:game_brand;type:varchar(32);not null" json:"game_brand"` // 游戏厂商:jili,pg
	GameType  string    `gorm:"column:game_type;type:varchar(32)" json:"game_type"`            // 游戏类型:slot
	LoginIp   string    `gorm:"column:login_ip;type:varchar(64)" json:"login_ip"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime;" json:"created_at"` // 创建时间
	UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`  // db更新时间
}

// TableName 指定表名
func (PlayerLoginLog) TableName() string {
	return "player_login_log"
}
