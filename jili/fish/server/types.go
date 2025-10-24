package server

// 定义一个房间的概念
type RoomImp interface {
	// 玩家进入房间
	OnLogin(player *JiliPlayer) error
	// 玩家重连房间
	OnReConnect(player *JiliPlayer) error
	// 玩家退出房间
	OnDisConnect(player *JiliPlayer) error
	// 玩家收到了消息了
	OnMessage(player *JiliPlayer, cmd int32, msg []byte) error
}

type RoomArgs struct {
	Appid string
	Rtp   string
}

type RoomCreator interface {
	CreateRoom(args *RoomArgs) RoomImp
}

// 配桌算法
type TableMatcherType int

const (
	TableMatcherType_RTP TableMatcherType = iota //通过rtp进行配桌，适合小飞机类游戏
)
