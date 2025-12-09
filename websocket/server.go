package websocket

import (
	"go.uber.org/zap"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
)

type Server struct {
	Addr string
	app  *fiber.App
}

func NewServer(addr string) *Server {
	return &Server{
		app:  fiber.New(),
		Addr: addr, // 默认端口
	}
}

func (server *Server) GetRouter() *fiber.App {
	return server.app
}

func (server *Server) SetupRoutes(url string, callBack OnConnectCallBack) {
	server.app.Use(url, func(c *fiber.Ctx) error {
		if websocket.IsWebSocketUpgrade(c) {
			c.Locals("allowed", true)
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	})

	server.app.Get(url, websocket.New(func(c *websocket.Conn) {
		conn := &Connect{
			Conn:     c,
			RecvChan: make(chan *WebSocketData, 10),
		}

		callBack(conn)

		var (
			mt  int
			msg []byte
			err error
		)
		for {
			if mt, msg, err = c.ReadMessage(); err != nil {
				conn.RecvChan <- &WebSocketData{
					Mt:  mt,
					Msg: msg,
					Err: err,
				}
				break
			}
			conn.RecvChan <- &WebSocketData{
				Mt:  mt,
				Msg: msg,
				Err: err,
			}
		}
	}))
}

func (server *Server) Start() {
	if err := server.app.Listen(server.Addr); err != nil {
		zap.L().Error("Failed to start WebSocket server", zap.String("addr", server.Addr), zap.Error(err))
	}
}

func (server *Server) Stop() {
	if server.app != nil {
		if err := server.app.Shutdown(); err != nil {
			zap.L().Error("Failed to shutdown WebSocket server", zap.Error(err))
		}
	}
}
