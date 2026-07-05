package models

import "time"

const (
	WalletBizTypeTransferIn     = "transfer_in"
	WalletBizTypeTransferOut    = "transfer_out"
	WalletBizTypeTransferOutAll = "transfer_out_all"
	WalletBizTypeBet            = "bet"
	WalletBizTypeWin            = "win"
	WalletBizTypeRefund         = "refund"

	WalletTransStatusSuccess = "SUCCESS"
	WalletTransStatusFailed  = "FAILED"
	WalletTransStatusPending = "PENDING"

	WalletDirectionDebit  int8 = -1
	WalletDirectionCredit int8 = 1
)

type PlayerWalletTrans struct {
	ID            int64     `gorm:"primaryKey;comment:主键"`
	AppId         string    `gorm:"column:app_id;size:32;not null;comment:商户AppId"`
	PlayerId      string    `gorm:"column:player_id;size:64;not null;comment:玩家ID"`
	Currency      string    `gorm:"column:currency;size:32;not null;comment:货币类型"`
	Tid           string    `gorm:"column:tid;size:64;not null;comment:幂等键"`
	BizType       string    `gorm:"column:biz_type;size:16;not null;comment:业务类型"`
	Direction     int8      `gorm:"column:direction;not null;comment:方向:-1扣款+1加款"`
	Amount        float64   `gorm:"column:amount;type:decimal(16,2);not null;comment:变动金额绝对值"`
	BalanceBefore float64   `gorm:"column:balance_before;type:decimal(16,2);not null;comment:变动前余额"`
	BalanceAfter  float64   `gorm:"column:balance_after;type:decimal(16,2);not null;comment:变动后余额"`
	TransStatus   string    `gorm:"column:trans_status;size:16;not null;comment:流水状态"`
	GameBrand     string    `gorm:"column:game_brand;size:32;comment:游戏厂商"`
	GameId        string    `gorm:"column:game_id;size:32;comment:游戏ID"`
	RoundId       string    `gorm:"column:round_id;size:64;comment:局号"`
	BetTid        string    `gorm:"column:bet_tid;size:64;comment:关联下注tid"`
	GameRecordId  int64     `gorm:"column:game_record_id;comment:关联app_game_record.id"`
	ErrorMsg      string    `gorm:"column:error_msg;size:256;comment:失败原因"`
	CreatedAt     time.Time `gorm:"column:created_at;comment:创建时间"`
}

func (PlayerWalletTrans) TableName() string {
	return "player_wallet_trans"
}
