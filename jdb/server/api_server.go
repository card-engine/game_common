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

type JdbGameApiServer struct {
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

func InitJiliFishApiServer(
	serverName string, gameName string, serverAddr string,
	tableMatcherType TableMatcherType,
	roomCreator RoomCreator, rdb *redis.Client,
	apiGrpcConn *google_grpc.ClientConn, rtpGrpcConn *google_grpc.ClientConn,
	logger log.Logger) *JdbGameApiServer {
	app := fiber.New()

	s := &JdbGameApiServer{
		app: app,

		log: log.NewHelper(logger),

		roomManager: NewRoomManager(roomCreator, logger),
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

func (s *JdbGameApiServer) route() {
	app := s.app

	routPath := fmt.Sprintf("/%s/websocket", s.gameName)

	app.Get(routPath, func(c *fiber.Ctx) error {
		// 验证是否为WebSocket升级请求
		if !websocket.IsWebSocketUpgrade(c) {
			return fiber.ErrUpgradeRequired
		}

		return websocket.New(func(conn *websocket.Conn) {
			s.OnWebSocketHandler(conn)
		})(c)
	})
}

func (s *JdbGameApiServer) OnWebSocketHandler(c *websocket.Conn) error {
	var player *JdbPlayer = nil
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

	// 第一阶段：握手
	_, msg, err := c.ReadMessage()
	if err != nil {
		s.log.Errorf("read websocket message error: %v", err)
		return nil
	}
	if err := s.onHandshake(c, msg); err != nil {
		s.log.Errorf("handshake error: %v", err)
		return nil
	}

	// 第二阶段：登录
	_, msg, err = c.ReadMessage()
	if err != nil {
		s.log.Errorf("read websocket message error: %v", err)
		return nil
	}
	if err := s.onLogin(c, msg); err != nil {
		s.log.Errorf("login error: %v", err)
		return nil
	}

	// 第三阶段：正常消息处理
	for {
		_, msg, err := c.ReadMessage()
		if err != nil {
			s.log.Errorf("read websocket message error: %v", err)
			break
		}

		action, controller, data, err := utils.Unpack(msg)
		if err != nil {
			return err
		}

		// ping
		if action == 29 && controller == 0 {
			c.WriteMessage(websocket.BinaryMessage, msg)
		} else if action == 13 && controller == 1 {
			cmd, cmdOk := data["c"].(string)
			p, pOk := data["p"].(sfs.SFSObject)
			if !cmdOk || !pOk {
				return fmt.Errorf("onMessage error, cmd: %s, p: %v", cmd, p)
			}

			if player == nil {
				if cmd != "gameLogin" {
					return fmt.Errorf("onMessage error, cmd: %s, p: %v", cmd, p)
				}

				// 进入房间赋值player
				if _player, err := s.onGameLogin(c, p); err != nil {
					return err
				} else {
					player = _player
				}
			} else if err := s.onMessage(player, cmd, p); err != nil {
				s.log.Errorf("handle message error: %v", err)
				break
			}
		} else {
			s.log.Debugf("unhandled message, action: %d, controller: %d", action, controller)
		}
	}
	return nil
}

func (s *JdbGameApiServer) onHandshake(c *websocket.Conn, buff []byte) error {
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
func (s *JdbGameApiServer) onLogin(c *websocket.Conn, buff []byte) error {
	action, controller, data, err := utils.Unpack(buff)
	if err != nil {
		return err
	}
	if action != 1 || controller != 0 {
		return fmt.Errorf("onLogin error, action: %d, controller: %d", action, controller)
	}

	// 解析请求参数
	var req protocol.LoginRequest
	if err := sfs.Unmarshal(data, &req); err != nil {
		return err
	}

	rsp, err := sfs.Marshal(protocol.LoginRespond{
		Id: 2406821,
		Pi: 0,
		Rl: []interface{}{},
		Rs: 0,
		Un: req.UserName,
		Zn: "JDB_ZONE_GAME",
	})

	if err != nil {
		s.log.Errorf("Marshal(%v) failed. err: %v", rsp, err)
		return err
	}

	if err := s.Send(c, 0, 1, rsp); err != nil {
		return err
	}

	return nil
}

func (s *JdbGameApiServer) onGameLogin(c *websocket.Conn, p sfs.SFSObject) (*JdbPlayer, error) {
	return s.roomManager.onGameLogin(c, p)
}

func (s *JdbGameApiServer) onMessage(player *JdbPlayer, cmd string, p sfs.SFSObject) error {
	return s.roomManager.OnMessage(player, cmd, p)
}

func (s *JdbGameApiServer) Send(c *websocket.Conn, controller int16, action uint8, payload sfs.SFSObject) error {
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

func (s *JdbGameApiServer) onDisconnect(player *JdbPlayer) error {
	return s.roomManager.OnDisConnect(player)
}

func (s *JdbGameApiServer) Start(ctx context.Context) error {
	go func() {
		if err := s.app.Listen(s.serverAddr); err != nil {
			log.Fatalf("Listen failed: %v", err)
		}
	}()

	return nil
}

func (s *JdbGameApiServer) Stop(ctx context.Context) error {
	if s.app != nil {
		s.app.Shutdown()
		s.app = nil
	}

	return nil
}
