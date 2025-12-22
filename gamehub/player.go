package gamehub

import (
	"fmt"
	"sync"

	v1 "github.com/card-engine/game_common/api/game/v1"
	"github.com/card-engine/game_common/player"
	"github.com/gofiber/contrib/websocket"
)

type Player struct {
	gameBrand GameBrand
	conn      *websocket.Conn
	mu        sync.Mutex // 新增互斥锁

	room        RoomImp
	roomManager *RoomManager

	PlayerInfo *player.PlayerInfo
	Rtp        string
}

func NewPlayer(gameBrand GameBrand, conn *websocket.Conn, PlayerInfo *player.PlayerInfo, Rtp string) *Player {
	return &Player{
		gameBrand: gameBrand,
	}
}

func (p *Player) SetConn(conn *websocket.Conn) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.conn = conn
}

func (p *Player) GetConn() *websocket.Conn {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.conn
}

// 直接发送原始文本消息
func (p *Player) Send(data string) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.conn != nil {
		if err := p.conn.WriteMessage(websocket.TextMessage, []byte(data)); err != nil {
			p.conn.Close()
		}
	}
	return nil
}

func (p *Player) IsConnect() bool {
	return p.conn != nil
}

// 从房间移出去
func (p *Player) ExitRoom(isDisconnect bool) {
	if p.roomManager != nil {
		p.roomManager.ExitRoom(p, isDisconnect)
	}
}

func (p *Player) GetPlayerId() string {
	return p.PlayerInfo.PlayerID
}

func (p *Player) GetAppId() string {
	return p.PlayerInfo.AppID
}

// ========================================================================================
// 设置玩家的余额，设置成不直接使用，通过下方的场景来更新玩家的余额
func (p *Player) setBalance(balance float64) error {
	p.PlayerInfo.Balance = balance
	if p.gameBrand == GameBrand_Inout {
		balanceMsg := fmt.Sprintf(`42["onBalanceChange",{"currency":"%s","balance":"%.2f"}]`, p.PlayerInfo.Currency, balance)
		return p.Send(balanceMsg)
	}

	return nil
}

func (p *Player) SetBalanceByBalanceReply(balanceReply *v1.BalanceReply) error {
	return p.setBalance(balanceReply.Balance)
}

func (p *Player) SetBalanceByWinReply(winReply *v1.WinReply) error {
	if winReply.HashBalance {
		return p.setBalance(winReply.Balance)
	}

	return nil
}

func (p *Player) SetBalanceByBetReply(betReply *v1.BetReply) error {
	return p.setBalance(betReply.Balance)
}

func (p *Player) SetBalanceByRefundReply(refundReply *v1.RefundReply) error {
	return p.setBalance(refundReply.Balance)
}

// ========================================================================================
// 获取玩家当前的余额
func (p *Player) GetBalance() float64 {
	return p.PlayerInfo.Balance
}

// 用户的唯一标识
func (p *Player) GetPlayerIdent() string {
	return fmt.Sprintf("%s-%s", p.GetAppId(), p.GetPlayerId())
}
