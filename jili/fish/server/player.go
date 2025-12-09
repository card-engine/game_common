package server

import (
	"sync"

	"github.com/card-engine/game_common/player"

	"github.com/gofiber/contrib/websocket"
)

type JiliPlayer struct {
	conn *websocket.Conn
	mu   sync.Mutex // 新增互斥锁

	room        RoomImp
	roomManager *RoomManager

	PlayerId string
	AppId    string
	Rtp      string

	Player *player.PlayerInfo
	Token  string
}

func (p *JiliPlayer) Send(data []byte) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.conn.WriteMessage(websocket.BinaryMessage, data)
}

// 从房间移出去
func (p *JiliPlayer) ExitRoom() {
	p.roomManager.ExitRoom(p)
}
