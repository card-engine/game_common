package jili

import (
	"context"
	"fmt"
	"strconv"

	v1 "github.com/card-engine/game_common/api/game/v1"
	client_utils "github.com/card-engine/game_common/api/game/v1/client"
	rtp_rpc_v1 "github.com/card-engine/game_common/api/rtp/v1"
	rtp_rpc_client "github.com/card-engine/game_common/api/rtp/v1/client"
	"github.com/card-engine/game_common/gamehub/common"
	"github.com/card-engine/game_common/gamehub/types"
	"github.com/card-engine/game_common/jili/fish/message"
	"github.com/card-engine/game_common/player"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
	google_grpc "google.golang.org/grpc"
)

type JiliRouter struct {
	log      *log.Helper
	app      *fiber.App
	gameName string

	rdb         *redis.Client
	apiGrpcConn *google_grpc.ClientConn
	rtpGrpcConn *google_grpc.ClientConn
	roomManager *common.RoomManager

	lobby types.LobbyImp

	logger log.Logger
}

func NewJiliRouter(
	gameName string,
	app *fiber.App,
	rdb *redis.Client,
	apiGrpcConn *google_grpc.ClientConn,
	rtpGrpcConn *google_grpc.ClientConn,
	roomManager *common.RoomManager,
	lobby types.LobbyImp,
	logger log.Logger) *JiliRouter {
	return &JiliRouter{
		app:         app,
		gameName:    gameName,
		log:         log.NewHelper(logger),
		rdb:         rdb,
		apiGrpcConn: apiGrpcConn,
		rtpGrpcConn: rtpGrpcConn,
		roomManager: roomManager,
		lobby:       lobby,
		logger:      logger,
	}
}

func (r *JiliRouter) Route() {
	app := r.app

	routPath := fmt.Sprintf("/%s/ws/:token", r.gameName)

	app.Get(routPath, func(c *fiber.Ctx) error {
		// 验证是否为WebSocket升级请求
		if !websocket.IsWebSocketUpgrade(c) {
			return fiber.ErrUpgradeRequired
		}

		token := c.Params("token")
		retryCount := c.Query("r", "0") // 默认值为 "0"

		return websocket.New(func(conn *websocket.Conn) {
			r.onWebSocketHandler(token, retryCount, conn)
		})(c)
	})
}

func (s *JiliRouter) onWebSocketHandler(token, retryCount string, c *websocket.Conn) error {
	retryCountNum, err := strconv.Atoi(retryCount)
	if err != nil {
		s.log.Errorf("Invalid retry count parameter: %s", retryCount)
		retryCountNum = 0
	}

	s.log.Infof("onWebSocketHandler token: %s, retryCount: %d", token, retryCountNum)

	tokenInfo, err := player.DecodedSSOKeyV3(token)
	if err != nil {
		s.log.Errorf("DecodedSSOKeyV3 failed: %v", err)
		return err
	}

	playerInfo, err := player.GetPlayerByAppAndPlayerId(s.rdb, tokenInfo.AppId, tokenInfo.PlayerId)
	if err != nil {
		s.log.Errorf("GetPlayerInfoByToken failed: %v", err)
		return err
	}

	rtp, err := rtp_rpc_client.GetPlayerRtp(context.Background(), s.rtpGrpcConn, &rtp_rpc_v1.GetPlayerRtpRequest{
		PlayerId:  playerInfo.PlayerID,
		AppId:     playerInfo.AppID,
		GameBrand: "jili",
		GameId:    playerInfo.GameID,
	})

	if err != nil {
		s.log.Errorf("GetPlayerRtp(%v) failed. err: %v", playerInfo.PlayerID, err)
		return err
	}

	jiliPlayer := common.NewPlayer(types.GameBrand_Spribe, c, playerInfo, rtp.Rtp)

	defer func() {
		if jiliPlayer != nil {
			if err := s.onDisconnect(jiliPlayer); err != nil {
				s.log.Errorf("onDisconnect error: %v", err)
			}
		}

		if err := c.Close(); err != nil {
			s.log.Errorf("close websocket error: %v", err)
		}
	}()

	// 初使化金币
	balanceRsp, err := client_utils.Balance(context.Background(), s.apiGrpcConn, playerInfo.AppID, &v1.BalanceRequest{
		PlayerId: playerInfo.PlayerID,
		Currency: playerInfo.Currency,
	})

	if err != nil {
		s.log.Errorf("Balance failed: %v", err)
		return err
	}

	if err := jiliPlayer.SetBalanceByBalanceReply(balanceRsp); err != nil {
		s.log.Errorf("SetBalanceByBalanceReply failed: %v", err)
		return err
	}

	if err := s.lobby.OnLogin(jiliPlayer); err != nil {
		return err
	}

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

		// 如果有大厅的话，将消息转发至大厅
		if s.lobby != nil {
			if err := s.lobby.OnMessage(jiliPlayer, data); err != nil {
				return err
			}
		}

		if err := s.roomManager.OnMessage(jiliPlayer, data); err != nil {
			return err
		}
	}

	return nil
}

func (s *JiliRouter) onDisconnect(player types.PlayerImp) error {
	return s.roomManager.OnDisConnect(player)
}
