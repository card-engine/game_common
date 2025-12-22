package server

import (
	"fmt"
	"sync"

	v1 "github.com/card-engine/game_common/api/game/v1"
	"github.com/card-engine/game_common/player"
	"github.com/gofiber/contrib/websocket"
)

type SpribePlayer struct {
	conn *websocket.Conn
	mu   sync.Mutex // 新增互斥锁

	room        RoomImp
	roomManager *RoomManager

	PlayerInfo *player.PlayerInfo
	Rtp        string
}

// 直接发送原始文本消息
func (p *SpribePlayer) Send(data string) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.conn != nil {
		if err := p.conn.WriteMessage(websocket.TextMessage, []byte(data)); err != nil {
			p.conn.Close()
		}
	}
	return nil
}

// 发送与服务业务相关的游戏数据
// func (p *SpribePlayer) SendGameServiceData(msgId, action string, response interface{}) error {
// 	responseData, err := json.Marshal([]interface{}{action, response})
// 	if err != nil {
// 		return err
// 	}
// 	if msgId == "" {
// 		msgId = DefaultMsgId
// 	}
// 	// 构造自定义格式响应: 数字[响应数据]
// 	responseMsg := fmt.Sprintf("%s%s", msgId, string(responseData))
// 	return p.Send(responseMsg)
// }

// // 不带action的游戏数据响应
// func (p *SpribePlayer) SendData(msgId string, response interface{}) error {
// 	responseData, err := json.Marshal([]interface{}{response})
// 	if err != nil {
// 		return err
// 	}
// 	if msgId == "" {
// 		msgId = DefaultMsgId
// 	}
// 	// 构造自定义格式响应: 数字[响应数据]
// 	responseMsg := fmt.Sprintf("%s%s", msgId, string(responseData))
// 	return p.Send(responseMsg)
// }

// // 发送游戏的错误信息
// func (p *SpribePlayer) SendErrorMessage(msgId, errMsg string) error {
// 	if msgId == "" {
// 		msgId = DefaultMsgId
// 	}
// 	errorMsg := fmt.Sprintf(`%s[{"error":{"message":"%s"}}]`, msgId, errMsg)
// 	return p.Send(errorMsg)
// }

// 客户端是不是断开了？
func (p *SpribePlayer) IsConnect() bool {
	return p.conn != nil
}

// 从房间移出去
func (p *SpribePlayer) ExitRoom(isDisconnect bool) {
	if p.roomManager != nil {
		p.roomManager.ExitRoom(p, isDisconnect)
	}
}

func (p *SpribePlayer) GetPlayerId() string {
	return p.PlayerInfo.PlayerID
}

func (p *SpribePlayer) GetAppId() string {
	return p.PlayerInfo.AppID
}

// ========================================================================================
// 设置玩家的余额，设置成不直接使用，通过下方的场景来更新玩家的余额
func (p *SpribePlayer) setBalance(balance float64) error {
	p.PlayerInfo.Balance = balance
	balanceMsg := fmt.Sprintf(`42["onBalanceChange",{"currency":"%s","balance":"%.2f"}]`, p.PlayerInfo.Currency, balance)
	return p.Send(balanceMsg)
}

func (p *SpribePlayer) SetBalanceByBalanceReply(balanceReply *v1.BalanceReply) error {
	return p.setBalance(balanceReply.Balance)
}

func (p *SpribePlayer) SetBalanceByWinReply(winReply *v1.WinReply) error {
	if winReply.HashBalance {
		return p.setBalance(winReply.Balance)
	}

	return nil
}

func (p *SpribePlayer) SetBalanceByBetReply(betReply *v1.BetReply) error {
	return p.setBalance(betReply.Balance)
}

func (p *SpribePlayer) SetBalanceByRefundReply(refundReply *v1.RefundReply) error {
	return p.setBalance(refundReply.Balance)
}

// ========================================================================================
// 获取玩家当前的余额
func (p *SpribePlayer) GetBalance() float64 {
	return p.PlayerInfo.Balance
}

// 用户的唯一标识
func (p *SpribePlayer) GetPlayerIdent() string {
	return fmt.Sprintf("%s-%s", p.GetAppId(), p.GetPlayerId())
}
