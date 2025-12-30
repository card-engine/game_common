package types

import (
	v1 "github.com/card-engine/game_common/api/game/v1"
	"github.com/card-engine/game_common/player"
	"github.com/gofiber/contrib/websocket"
)

const DefaultMsgId = "42"

type GameBrand string

const (
	GameBrand_Inout  GameBrand = "inout"
	GameBrand_Spribe GameBrand = "spribe"
	GameBrand_jdb    GameBrand = "jdb"
)

type RoomManagerImp interface {

	// 玩家退出房间
	ExitRoom(player PlayerImp, isDisconnect bool)

	// 玩家登录房间
	OnLogin(player PlayerImp) error

	// 玩家收到了消息了
	OnMessage(player PlayerImp, msg interface{}) error

	OnDisConnect(player PlayerImp) error
}

type PlayerImp interface {
	SetConn(conn *websocket.Conn)
	GetConn() *websocket.Conn
	CloseConn()
	IsConnect() bool

	SetRoom(room RoomImp)
	GetRoom() RoomImp
	ExitRoom(isDisconnect bool) error

	GetRoomManager() RoomManagerImp
	SetRoomManager(roomManager RoomManagerImp)

	// 获取金额
	GetBalance() float64
	SetBalanceByBalanceReply(balanceReply *v1.BalanceReply) error
	SetBalanceByWinReply(winReply *v1.WinReply) error
	SetBalanceByBetReply(betReply *v1.BetReply) error
	SetBalanceByRefundReply(refundReply *v1.RefundReply) error

	// 获取用户唯一标识
	GetPlayerIdent() string
	// 获取用户信息
	GetPlayerInfo() *player.PlayerInfo
	GetPlayerId() string
	GetAppId() string
	GetCurrency() string
	GetLang() string

	GetRtpStr() string
	GetRtp() float64

	SendString(msg string) error
	SendBinary(data []byte) error
}

// 定义一个房间的概念
type RoomImp interface {
	// 获取当前玩家的数量
	GetPlayerNum() int32
	// 玩家进入房间
	OnLogin(player PlayerImp) error
	// 玩家重连房间
	OnReConnect(player PlayerImp) error
	// 玩家退出房间
	OnDisConnect(player PlayerImp) error
	// 玩家收到了消息了
	OnMessage(player PlayerImp, data interface{}) error
	// 房间销毁时间的调用
	OnDispose()
}

type InoutMsgData struct {
	MsgId   string
	Action  string
	Payload string
}

// 配桌算法
type TableMatcherType int

const (
	TableMatcherType_RTP    TableMatcherType = iota //通过rtp进行配桌，适合小飞机类游戏
	TableMatcherType_SINGLE                         //单桌配桌，使用玩后就释放的那种
)

// rtp类型的房间
type RtpRoomArgs struct {
	Appid    string
	Rtp      string
	Currency string
}

type SingleRoomArgs struct {
	Appid string
}

type RoomCreator interface {
	// RtpRoomArgs、SingleArgs 都可以创建房间
	CreateRoom(args interface{}) RoomImp
}

type Router interface {
	Route()
}
