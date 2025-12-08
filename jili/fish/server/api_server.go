package server

import (
	"context"
	"fmt"
	"strconv"

	v1 "cn.qingdou.server/game_common/api/game/v1"
	api_rpc_client "cn.qingdou.server/game_common/api/game/v1/client"
	rtp_rpc_v1 "cn.qingdou.server/game_common/api/rtp/v1"
	rtp_rpc_client "cn.qingdou.server/game_common/api/rtp/v1/client"
	"cn.qingdou.server/game_common/jili/fish/message"
	"cn.qingdou.server/game_common/player"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/registry"
	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
	google_grpc "google.golang.org/grpc"
)

type JiliFishApiServer struct {
	app         *fiber.App
	log         *log.Helper
	serverName  string //服务器名
	ds          registry.Discovery
	roomManager *RoomManager

	// 配桌算法
	tableMatcherType TableMatcherType

	rdb         *redis.Client
	apiGrpcConn *google_grpc.ClientConn
	rtpGrpcConn *google_grpc.ClientConn

	gameName   string // 游戏名称
	serverAddr string //服务器绑定的地址
}

// InitJiliFishApiServer 初始化JiliFish API服务器
// 参数:
//   - serverName: 服务器名称
//   - gameName: 游戏名称
//   - serverAddr: 服务器绑定的地址
//   - ds: 服务发现实例
//   - logger: 日志记录器
//
// 返回:
//   - *JiliFishApiServer: 初始化后的JiliFish API服务器实例
func InitJiliFishApiServer(
	serverName string, gameName string, serverAddr string,
	tableMatcherType TableMatcherType,
	roomCreator RoomCreator, rdb *redis.Client,
	apiGrpcConn *google_grpc.ClientConn, rtpGrpcConn *google_grpc.ClientConn,
	logger log.Logger) *JiliFishApiServer {
	app := fiber.New()

	s := &JiliFishApiServer{
		app: app,

		log: log.NewHelper(logger),

		roomManager: NewRoomManager(roomCreator),
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

func (s *JiliFishApiServer) route() {
	app := s.app

	routPath := fmt.Sprintf("/%s/ws/:token", s.gameName)

	app.Get(routPath, func(c *fiber.Ctx) error {
		// 验证是否为WebSocket升级请求
		if !websocket.IsWebSocketUpgrade(c) {
			return fiber.ErrUpgradeRequired
		}

		token := c.Params("token")
		retryCount := c.Query("r", "0") // 默认值为 "0"

		return websocket.New(func(conn *websocket.Conn) {
			s.OnWebSocketHandler(token, retryCount, conn)
		})(c)
	})
}

func (s *JiliFishApiServer) OnWebSocketHandler(token, retryCount string, c *websocket.Conn) {
	retryCountNum, err := strconv.Atoi(retryCount)
	if err != nil {
		s.log.Errorf("Invalid retry count parameter: %s", retryCount)
		retryCountNum = 0
	}

	tokenInfo, err := player.DecodedSSOKeyV3(token)
	if err != nil {
		s.log.Errorf("DecodedSSOKeyV3 failed: %v", err)
		return
	}

	playerInfo, err := player.GetPlayerByAppAndPlayerId(s.rdb, tokenInfo.AppId, tokenInfo.PlayerId)
	if err != nil {
		s.log.Errorf("GetPlayerInfoByToken failed: %v", err)
		return
	}

	rtp, err := rtp_rpc_client.GetPlayerRtp(context.Background(), s.rtpGrpcConn, &rtp_rpc_v1.GetPlayerRtpRequest{
		PlayerId:  playerInfo.PlayerID,
		AppId:     playerInfo.AppID,
		GameBrand: "jili",
		GameId:    playerInfo.GameID,
	})

	if err != nil {
		s.log.Errorf("GetPlayerRtp(%v) failed. err: %v", playerInfo.PlayerID, err)
		return
	}

	jiliPlayer := &JiliPlayer{conn: c,
		roomManager: s.roomManager,
		PlayerId:    playerInfo.PlayerID,
		AppId:       playerInfo.AppID,
		Rtp:         rtp.Rtp,
		Token:       token,
		Player:      playerInfo,
	}

	if retryCountNum == 0 {
		balanceRsp, err := api_rpc_client.Balance(context.Background(), s.apiGrpcConn, playerInfo.AppID, &v1.BalanceRequest{
			PlayerId: playerInfo.PlayerID,
			Currency: playerInfo.Currency,
		})
		if err != nil {
			s.log.Errorf("Balance failed: %v", err)
			return
		}

		if err := player.UpdateBalance(s.rdb, playerInfo.AppID, playerInfo.PlayerID, balanceRsp.Balance); err != nil {
			s.log.Errorf("UpdateBalance failed: %v", err)
			return
		}
		s.roomManager.OnLogin(jiliPlayer)
	} else {
		s.roomManager.OnReConnect(jiliPlayer)
	}

	defer s.onDisconnect(jiliPlayer)

	for {
		_, msg, err := c.ReadMessage()

		if err != nil {
			s.log.Errorf("read websocket message error: %v", err)
			break
		}

		data, err := message.UnPack(msg)
		if err != nil {
			s.log.Errorf("UnPack message error: %v", err)
			continue
		}

		if err := s.roomManager.OnMessage(jiliPlayer, int32(data.Type), data.Data); err != nil {
			s.log.Errorf("OnMessage error: %v", err)
			continue
		}
	}
}

func (s *JiliFishApiServer) onDisconnect(player *JiliPlayer) error {
	return s.roomManager.OnDisConnect(player)
}

func (s *JiliFishApiServer) Start(ctx context.Context) error {
	go func() {
		if err := s.app.Listen(s.serverAddr); err != nil {
			log.Fatalf("Listen failed: %v", err)
		}
	}()

	return nil
}

func (s *JiliFishApiServer) Stop(ctx context.Context) error {
	if s.app != nil {
		s.app.Shutdown()
		s.app = nil
	}

	return nil
}
