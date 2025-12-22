package server

import (
	"context"
	"fmt"

	"github.com/card-engine/game_common/sfs/protocol"
	"github.com/card-engine/game_common/sfs/utils"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/qd2ss/sfs"
	"github.com/redis/go-redis/v9"
	google_grpc "google.golang.org/grpc"
)

type SpribeGameApiServer struct {
	app         *fiber.App
	log         *log.Helper
	serverName  string //服务器名
	roomManager *RoomManager

	// 配桌算法
	tableMatcherType TableMatcherType

	rdb         *redis.Client
	apiGrpcConn *google_grpc.ClientConn
	rtpGrpcConn *google_grpc.ClientConn

	gameName   string // 游戏名称
	serverAddr string //服务器绑定的地址
}

func InitSpribeGameApiServer(
	serverName string,
	gameName string,
	serverAddr string,
	tableMatcherType TableMatcherType,
	roomCreator RoomCreator,
	rdb *redis.Client,
	apiGrpcConn *google_grpc.ClientConn,
	rtpGrpcConn *google_grpc.ClientConn,
	logger log.Logger) *SpribeGameApiServer {
	app := fiber.New()

	s := &SpribeGameApiServer{
		app: app,

		log: log.NewHelper(logger),

		roomManager: NewRoomManager(roomCreator, tableMatcherType, logger),
		serverName:  serverName,

		tableMatcherType: tableMatcherType,
		rdb:              rdb,
		apiGrpcConn:      apiGrpcConn,
		rtpGrpcConn:      rtpGrpcConn,

		gameName:   gameName,
		serverAddr: serverAddr,
	}
	s.route()
	return s
}

func (s *SpribeGameApiServer) route() {
	app := s.app

	routPath := fmt.Sprintf("/%s/websocket", s.gameName)

	// 只允许WebSocket升级的中间件
	app.Use(routPath, func(c *fiber.Ctx) error {
		if websocket.IsWebSocketUpgrade(c) {
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	})

	app.Get(routPath, websocket.New(func(c *websocket.Conn) {
		step := 0

		var player *SpribePlayer = nil

		defer func() {
			if player != nil {
				if err := s.onDisconnect(player); err != nil {
					s.log.Errorf("onDisconnect error: %v", err)
				}
			}

			if err := c.Close(); err != nil {
				s.log.Errorf("close websocket error: %v", err)
			}
		}()

		for {
			_, msg, err := c.ReadMessage()

			if err != nil {
				s.log.Errorf("read websocket message error: %v", err)
				break
			}

			if step == 0 {
				if err := s.onHandshake(c, msg); err != nil {
					s.log.Errorf("handshake error: %v", err)
					break
				}
				step += 1
			} else if step == 1 {
				if player, err = s.onLogin(c, msg); err != nil {
					s.log.Errorf("login error: %v", err)
					break
				}
				step += 1
			} else {
				if err := s.onMessage(player, msg); err != nil {
					s.log.Errorf("handle message error: %v", err)
					break
				}
			}
		}
	}))
}

func (s *SpribeGameApiServer) onHandshake(c *websocket.Conn, buff []byte) error {
	action, controller, data, err := utils.Unpack(buff)
	if err != nil {
		return err
	}
	if action != 0 || controller != 0 {
		return fmt.Errorf("handshake error, action: %d, controller: %d", action, controller)
	}

	var req protocol.PreLoginRequest
	if err := sfs.Unmarshal(data, &req); err != nil {
		return err
	}
	rsp, err := sfs.Marshal(protocol.PreLoginRespond{
		Ct: 1024,
		Ms: 500000,
		Tk: "38345f8ddea9855b9aaa83d06d3b2a01",
	})
	if err != nil {
		return err
	}

	return s.Send(c, 0, 0, rsp)
}

// 登陆
func (s *SpribeGameApiServer) onLogin(c *websocket.Conn, buff []byte) (*SpribePlayer, error) {
	return s.roomManager.OnLogin(c, buff)
}

func (s *SpribeGameApiServer) onMessage(player *SpribePlayer, buff []byte) error {
	return s.roomManager.OnMessage(player, buff)
}

func (s *SpribeGameApiServer) onDisconnect(player *SpribePlayer) error {
	return s.roomManager.OnDisConnect(player)
}

func (s *SpribeGameApiServer) Send(c *websocket.Conn, controller int16, action uint8, payload sfs.SFSObject) error {
	sendData := sfs.SFSObject{
		"a": action,
		"c": controller,
		"p": payload,
	}

	packer := sfs.NewPacker()
	buff, err := packer.Pack(sendData, false)
	if err != nil {
		return err
	}

	return c.WriteMessage(websocket.BinaryMessage, buff)
}

// ============================================================
func (s *SpribeGameApiServer) Start(ctx context.Context) error {
	go func() {
		if err := s.app.Listen(s.serverAddr); err != nil {
			log.Fatalf("Listen failed: %v", err)
		}
	}()

	return nil
}

func (s *SpribeGameApiServer) Stop(ctx context.Context) error {
	if s.app != nil {
		s.app.Shutdown()
		s.app = nil
	}

	return nil
}

// ============================================================
