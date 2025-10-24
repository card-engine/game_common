package server

import (
	"fmt"
	"sync"

	"github.com/gofiber/contrib/websocket"
)

type JdbPlayer struct {
	conn *websocket.Conn
	mu   sync.Mutex // 新增互斥锁

	room        RoomImp
	roomManager *RoomManager

	PlayerId string
	AppId    string
	Rtp      string
}

func (p *JdbPlayer) Send(data []byte) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.conn.WriteMessage(websocket.BinaryMessage, data)
}

// 从房间移出去
func (p *JdbPlayer) ExitRoom() {
	p.roomManager.ExitRoom(p)
}

func (p *JdbPlayer) GetPlayerIdent() string {
	return fmt.Sprintf("%s-%s", p.AppId, p.PlayerId)
}
