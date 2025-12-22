package gamehub

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"

	google_grpc "google.golang.org/grpc"
)

type GameApiServer struct {
	app *fiber.App
	log *log.Helper

	serverAddr string //服务器绑定的地址
	router     Router
}

func InitGameApiServer(
	gameBrand GameBrand, // 游戏品牌
	serverName string, // 服务器名称
	gameName string, // 游戏名称
	serverAddr string, // 服务器绑定的地址
	tableMatcherType TableMatcherType, // 配桌算法
	roomCreator RoomCreator, // 房间创建器
	rdb *redis.Client, // redis 客户端
	apiGrpcConn *google_grpc.ClientConn,
	rtpGrpcConn *google_grpc.ClientConn, // rtp 客户端
	logger log.Logger) *GameApiServer {
	app := fiber.New()

	s := &GameApiServer{
		app: app,
		log: log.NewHelper(logger),

		serverAddr: serverAddr,
	}

	roomManager := NewRoomManager(gameBrand, roomCreator, tableMatcherType, logger)

	switch gameBrand {
	case GameBrand_Inout:
		s.router = NewInoutRouter(gameName, app, rdb, apiGrpcConn, rtpGrpcConn, roomManager, logger)
	case GameBrand_Spribe:
		s.router = NewSpribeRouter(gameName, app, rdb, apiGrpcConn, rtpGrpcConn, roomManager, logger)
	}

	s.route()
	return s
}

func (s *GameApiServer) route() {
	if s.router == nil {
		s.log.Fatalf("router is nil")
	}

	s.router.Route()
}

func (s *GameApiServer) Start(ctx context.Context) error {
	go func() {
		if err := s.app.Listen(s.serverAddr); err != nil {
			log.Fatalf("Listen failed: %v", err)
		}
	}()

	return nil
}

func (s *GameApiServer) Stop(ctx context.Context) error {
	if s.app != nil {
		s.app.Shutdown()
		s.app = nil
	}

	return nil
}
