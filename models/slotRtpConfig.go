/* slot类的游戏rtp配置 */
package models

import "time"

type SlotRtpConfig struct {
	Id uint64 `gorm:"primaryKey;comment:id;"`

	Name string `gorm:"column:table_name;comment:对应的爬取的数据表名;"`

	Rtp string `gorm:"column:rtp;comment:默认rtp;default:95"`

	Data string `gorm:"column:data;comment:配置数据的本体;"`

	CreateTime time.Time `gorm:"autoCreateTime;comment:创建时间;"`
	UpdateTime time.Time `gorm:"autoCreateTime;comment:创建时间;"`
}

func (SlotRtpConfig) TableName() string {
	return "slot_rtp_config"
}
