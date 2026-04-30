package common

import (
	"context"
	"fmt"
	"strconv"
	"sync"

	v1 "github.com/card-engine/game_common/api/game/v1"
	rtp_rpc_v1 "github.com/card-engine/game_common/api/rtp/v1"
	rtp_rpc_client "github.com/card-engine/game_common/api/rtp/v1/client"
	"github.com/card-engine/game_common/gamehub/types"
	"github.com/card-engine/game_common/player"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/gofiber/contrib/websocket"

	google_grpc "google.golang.org/grpc"
)

type Player struct {
	gameBrand types.GameBrand
	conn      *websocket.Conn
	mu        sync.Mutex // 新增互斥锁

	room        types.RoomImp
	roomManager types.RoomManagerImp

	PlayerInfo *player.PlayerInfo
	Rtp        string

	rtpGrpcConn *google_grpc.ClientConn
	log         *log.Helper
}

func NewPlayer(gameBrand types.GameBrand, conn *websocket.Conn, PlayerInfo *player.PlayerInfo, rtpGrpcConn *google_grpc.ClientConn, log *log.Helper) *Player {
	return &Player{
		gameBrand:   gameBrand,
		conn:        conn,
		PlayerInfo:  PlayerInfo,
		rtpGrpcConn: rtpGrpcConn,

		log: log,
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

func (p *Player) CloseConn() {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.conn != nil {
		p.conn.Close()
		p.conn = nil
	}
}

func (p *Player) SetRoom(room types.RoomImp) {
	p.room = room
}

func (p *Player) GetRoom() types.RoomImp {
	return p.room
}

func (p *Player) GetRoomManager() types.RoomManagerImp {
	return p.roomManager
}

func (p *Player) SetRoomManager(roomManager types.RoomManagerImp) {
	p.roomManager = roomManager
}

func (p *Player) SendString(msg string) error {
	return p.send(websocket.TextMessage, []byte(msg))
}

func (p *Player) SendBinary(data []byte) error {
	return p.send(websocket.BinaryMessage, data)
}

// 直接发送原始文本消息, messageType有websocket.TextMessage和websocket.BinaryMessage
func (p *Player) send(messageType int, data []byte) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.conn != nil {
		if err := p.conn.WriteMessage(messageType, data); err != nil {
			p.conn.Close()
		}
	}
	return nil
}

func (p *Player) IsConnect() bool {
	return p.conn != nil
}

// 从房间移出去
func (p *Player) ExitRoom(isDisconnect bool) error {
	if p.roomManager != nil {
		p.roomManager.ExitRoom(p, isDisconnect)
	}
	return nil
}

func (p *Player) GetPlayerId() string {
	return p.PlayerInfo.PlayerID
}

func (p *Player) GetAppId() string {
	return p.PlayerInfo.AppID
}

func (p *Player) GetCurrency() string {
	return p.PlayerInfo.Currency
}

// ========================================================================================
// 设置玩家的余额，设置成不直接使用，通过下方的场景来更新玩家的余额
func (p *Player) setBalance(balance float64) error {
	p.PlayerInfo.Balance = balance
	if p.gameBrand == types.GameBrand_Inout {
		balanceMsg := fmt.Sprintf(`42["onBalanceChange",{"currency":"%s","balance":"%.2f"}]`, p.PlayerInfo.Currency, balance)
		return p.SendString(balanceMsg)
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

// ================================
func (p *Player) GetLang() string {
	if p.PlayerInfo == nil || p.PlayerInfo.Lang == "" {
		return "en"
	}
	return p.PlayerInfo.Lang
}

// =================================
func (p *Player) RefreshRtp() (string, error) {
	rtp, err := rtp_rpc_client.GetPlayerRtp(context.Background(), p.rtpGrpcConn, &rtp_rpc_v1.GetPlayerRtpRequest{
		PlayerId:  p.PlayerInfo.PlayerID,
		AppId:     p.PlayerInfo.AppID,
		GameBrand: string(p.gameBrand),
		GameId:    p.PlayerInfo.GameID,
	})
	if err != nil {
		p.log.Errorf("GetPlayerRtp failed: %v", err)
		// 这里只是兜底
		if p.Rtp == "" {
			p.Rtp = "95"
		}
		return p.Rtp, err
	}
	p.Rtp = rtp.Rtp
	return p.Rtp, nil
}

// 获取玩家的RTP
func (p *Player) GetRtpStr() string {
	// 每次获取RTP都刷新一下，避免缓存过期
	p.RefreshRtp()
	return p.Rtp
}

func (p *Player) GetRtp() float64 {
	rtpStr := p.GetRtpStr()
	rtp, err := strconv.ParseFloat(rtpStr, 64)
	if err != nil {
		return 97
	}
	return rtp
}

// ==============================
func (p *Player) GetPlayerInfo() *player.PlayerInfo {
	return p.PlayerInfo
}
