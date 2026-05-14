package server

// 这套 server 包看起来是早期/独立实现的 Spribe WebSocket 服务端。
// 为了保证该包在全仓 `go test ./...` 下可编译，这里补齐其内部依赖的最小类型定义。

type TableMatcherType int

const (
	TableMatcherType_RTP TableMatcherType = iota
	TableMatcherType_SINGLE
	TableMatcherType_CUSTOM
)

type RtpRoomArgs struct {
	Appid    string
	Rtp      string
	Currency string
}

type SingleRoomArgs struct {
	Appid string
}

// RoomImp 定义了 RoomManager/Player 会用到的最小房间能力集合。
// 具体房间实现由上层注入（RoomCreator.CreateRoom）提供。
type RoomImp interface {
	GetPlayerNum() int32
	OnLogin(player *SpribePlayer) error
	OnReConnect(player *SpribePlayer) error
	OnDisConnect(player *SpribePlayer) error
	OnMessage(player *SpribePlayer, msg []byte) error
	OnDispose()
}

type RoomCreator interface {
	CreateRoom(args interface{}) RoomImp
}

