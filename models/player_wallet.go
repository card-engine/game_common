package models

import "time"

type PlayerWallet struct {
	ID          int64     `gorm:"primaryKey;comment:主键"`
	AppId       string    `gorm:"column:app_id;size:32;not null;comment:商户AppId"`
	PlayerId    string    `gorm:"column:player_id;size:64;not null;comment:玩家ID"`
	Currency    string    `gorm:"column:currency;size:32;not null;comment:货币类型"`
	Balance     float64   `gorm:"column:balance;type:decimal(16,2);not null;default:0;comment:可用余额"`
	LockVersion int64     `gorm:"column:lock_version;not null;default:0;comment:乐观锁版本号"`
	CreatedAt   time.Time `gorm:"column:created_at;comment:创建时间"`
	UpdatedAt   time.Time `gorm:"column:updated_at;comment:更新时间"`
}

func (PlayerWallet) TableName() string {
	return "player_wallet"
}
