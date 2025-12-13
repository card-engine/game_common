package server

const DefaultMsgId = "42"

// 定义一个房间的概念
type RoomImp interface {
	// 获取当前玩家的数量
	GetPlayerNum() int32
	// 玩家进入房间
	OnLogin(player *InoutPlayer) error
	// 玩家重连房间
	OnReConnect(player *InoutPlayer) error
	// 玩家退出房间
	OnDisConnect(player *InoutPlayer) error
	// 玩家收到了消息了
	OnMessage(player *InoutPlayer, msgId, action, payload string) error
	// 房间销毁时间的调用
	OnDispose()
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
