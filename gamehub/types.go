package gamehub

const DefaultMsgId = "42"

type GameBrand string

const (
	GameBrand_Inout  GameBrand = "inout"
	GameBrand_Spribe GameBrand = "spribe"
	GameBrand_jdb    GameBrand = "jdb"
)

// 定义一个房间的概念
type RoomImp interface {
	// 获取当前玩家的数量
	GetPlayerNum() int32
	// 玩家进入房间
	OnLogin(player *Player) error
	// 玩家重连房间
	OnReConnect(player *Player) error
	// 玩家退出房间
	OnDisConnect(player *Player) error
	// 玩家收到了消息了
	OnMessage(player *Player, data interface{}) error
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
