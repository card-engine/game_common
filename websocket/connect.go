package websocket

import (
	"errors"
	"sync"

	"github.com/gofiber/websocket/v2"
)

type WebSocketData struct {
	Mt  int
	Msg []byte
	Err error
}

type OnConnRecvCallBack func() []byte
type OnDisConnCallBack func()
type OnConnectCallBack func(conn *Connect)

type Connect struct {
	Conn               *websocket.Conn
	OnConnRecvCallBack OnConnRecvCallBack
	OnDisConnCallBack  OnDisConnCallBack
	RecvChan           chan *WebSocketData

	messageType int
	mu          sync.RWMutex
}

func (conn *Connect) Send(buff []byte) error {
	conn.mu.Lock()
	defer conn.mu.Unlock()

	if conn.Conn == nil {
		return errors.New("this conn is disconnect")
	}
	return conn.Conn.WriteMessage(websocket.BinaryMessage, buff)
}

func (conn *Connect) Recv() ([]byte, error) {
	conn.mu.RLock()
	defer conn.mu.RUnlock()

	data := <-conn.RecvChan
	conn.messageType = data.Mt
	if data.Err != nil {
		conn.Conn = nil
	}
	return data.Msg, data.Err
}

func (conn *Connect) Close() error {
	conn.mu.Lock()
	defer conn.mu.Unlock()

	if conn.Conn != nil {
		return conn.Conn.Close()
	}
	return nil
}

func (conn *Connect) OnRecv(callBack OnConnRecvCallBack) {
	conn.OnConnRecvCallBack = callBack
}

func (conn *Connect) OnDisconnect(callBack OnDisConnCallBack) {
	conn.OnDisConnCallBack = callBack
}
