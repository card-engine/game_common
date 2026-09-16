package types

// GameSettlementFactTopic 结算系统专用 Topic（与 ApiGameEvent 解耦）。
const GameSettlementFactTopic = "GameSettlementFact"

// 结算事实事件类型。
const (
	SettlementEventBet    = "BET"
	SettlementEventWin    = "WIN"
	SettlementEventRefund = "REFUND"
)

// 钱包 / 流水模式。
const (
	SettlementWalletSingle   = "SINGLE"
	SettlementWalletTransfer = "TRANSFER"

	SettlementFlowClassic     = "CLASSIC"
	SettlementFlowTransaction = "TRANSACTION"
	SettlementFlowSession     = "SESSION"
)

// GameSettlementFact 游戏资金结算事实：一笔资金动作一条消息，金额恒 >= 0。
type GameSettlementFact struct {
	FactID           string  `json:"factId"`
	EventType        string  `json:"eventType"` // BET | WIN | REFUND
	AppID            string  `json:"appId"`
	PlayerID         string  `json:"playerId"`
	GameBrand        string  `json:"gameBrand"`
	GameType         string  `json:"gameType"`
	GameID           string  `json:"gameId"`
	RoundID          string  `json:"roundId"`
	Currency         string  `json:"currency"`
	Amount           float64 `json:"amount"`
	TransactionID    string  `json:"transactionId"`
	BetTransactionID string  `json:"betTransactionId,omitempty"`
	Rtp              string  `json:"rtp"`
	IsFree           bool    `json:"isFree"`
	WalletMode       string  `json:"walletMode"`
	FlowMode         string  `json:"flowMode"`
	OccurredAtMs     int64   `json:"occurredAtMs"`
	SettleDay        string  `json:"settleDay"` // YYYY-MM-DD，商户时区
	MerchantType     string  `json:"merchantType,omitempty"`
}
