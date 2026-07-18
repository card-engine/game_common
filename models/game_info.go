package models

import (
	"strings"
	"time"
)

const (
	GameStatusEnable  = "ENABLE"
	GameStatusDisable = "DISABLE"

	GameRtpModelDynamic = "Dynamic"

	SpinDataModelGroup = "Group"
)

// GameInfo 对应 game.game_info 表。
type GameInfo struct {
	ID              int64     `gorm:"column:id;type:bigint;primaryKey;autoIncrement" json:"id"`
	GameCode        string    `gorm:"column:game_code;type:varchar(32)" json:"game_code"`
	GameId          string    `gorm:"column:game_id;type:varchar(32);not null" json:"game_id"`
	GameName        string    `gorm:"column:game_name;type:varchar(64)" json:"game_name"`
	GameFullName    string    `gorm:"column:game_full_name;type:varchar(512)" json:"game_full_name"`
	GameIcon        string    `gorm:"column:game_icon;type:varchar(512)" json:"game_icon"`
	GameType        string    `gorm:"column:game_type;type:varchar(32)" json:"game_type"`
	GameBrand       string    `gorm:"column:game_brand;type:varchar(32)" json:"game_brand"`
	Status          string    `gorm:"column:status;type:varchar(8);default:ENABLE" json:"status"`
	Rtp             string    `gorm:"column:rtp;type:varchar(32)" json:"rtp"`
	RtpModel        string    `gorm:"column:rtp_model;type:varchar(32)" json:"rtp_model"`
	RtpSupportLevel string    `gorm:"column:rtp_support_level;type:varchar(64)" json:"rtp_support_level"`
	ProxyModel      string    `gorm:"column:proxy_model;type:varchar(16)" json:"proxy_model"`
	SpinDataModel   string    `gorm:"column:spin_data_model;type:varchar(16)" json:"spin_data_model"`
	ResType         string    `gorm:"column:res_type;type:varchar(16)" json:"res_type"`
	CreatedAt       time.Time `gorm:"column:created_at;type:datetime;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt       time.Time `gorm:"column:updated_at;type:datetime;not null;default:CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP" json:"updated_at"`
}

func (GameInfo) TableName() string {
	return "game_info"
}

func (g *GameInfo) IsEnabled() bool {
	if g == nil {
		return false
	}
	if g.Status == "" {
		return true
	}
	return strings.EqualFold(g.Status, GameStatusEnable)
}

func (g *GameInfo) IsDynamicRtp() bool {
	if g == nil {
		return false
	}
	return strings.EqualFold(g.RtpModel, GameRtpModelDynamic)
}

// IsGroupSpinData 结果集按 group_id 分组，一局中奖对应多行，有效倍率为组内 MAX(rate)。
func (g *GameInfo) IsGroupSpinData() bool {
	if g == nil {
		return false
	}
	return strings.EqualFold(g.SpinDataModel, SpinDataModelGroup)
}

func IsRtpInSupportLevel(rtp, supportLevel string) bool {
	if supportLevel == "" {
		return false
	}
	for _, level := range strings.Split(supportLevel, ",") {
		if strings.TrimSpace(level) == rtp {
			return true
		}
	}
	return false
}
