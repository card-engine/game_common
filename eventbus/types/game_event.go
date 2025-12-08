package types

const GAME_EVENT_TOPIC = "ApiGameEvent"

// 派奖事件
type WinEvent struct {
	AppId            string  `json:"appId,omitempty"`     //商户Id
	GameBrand        string  `json:"gameBrand,omitempty"` //游戏品牌
	GameType         string  `json:"gameType"`            // 游戏类型
	GameId           string  `json:"gameId,omitempty"`    //游戏id
	PlayerId         string  `json:"playerId,omitempty"`  //账号id
	RoundId          string  `json:"roundId,omitempty"`   //游戏回合, 长度64以內。
	Currency         string  `json:"currency,omitempty"`
	Bet              float64 `json:"bet,omitempty"`              //下注
	Win              float64 `json:"win"`                        //赢钱
	BetTransactionId string  `json:"betTransactionId,omitempty"` // 下注交易id(必填，用来查关联下注记录)
	TransactionId    string  `json:"transactionId,omitempty"`    //派奖交易id
	Rtp              string  `json:"rtp"`                        // 玩家rtp
}
