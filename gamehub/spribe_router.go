package gamehub

import (
	"context"
	"fmt"

	v1 "github.com/card-engine/game_common/api/game/v1"
	client_utils "github.com/card-engine/game_common/api/game/v1/client"
	rtp_rpc_v1 "github.com/card-engine/game_common/api/rtp/v1"
	rtp_rpc_client "github.com/card-engine/game_common/api/rtp/v1/client"
	"github.com/card-engine/game_common/player"
	"github.com/card-engine/game_common/sfs/protocol"
	"github.com/card-engine/game_common/sfs/utils"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/qd2ss/sfs"
	"github.com/redis/go-redis/v9"
	google_grpc "google.golang.org/grpc"
)

type SpribeRouter struct {
	log      *log.Helper
	app      *fiber.App
	gameName string

	rdb         *redis.Client
	apiGrpcConn *google_grpc.ClientConn
	rtpGrpcConn *google_grpc.ClientConn
	roomManager *RoomManager
	logger      log.Logger
}

func NewSpribeRouter(
	gameName string,
	app *fiber.App,
	rdb *redis.Client,
	apiGrpcConn *google_grpc.ClientConn,
	rtpGrpcConn *google_grpc.ClientConn,
	roomManager *RoomManager,
	logger log.Logger) *SpribeRouter {
	return &SpribeRouter{
		app:         app,
		gameName:    gameName,
		log:         log.NewHelper(logger),
		rdb:         rdb,
		apiGrpcConn: apiGrpcConn,
		rtpGrpcConn: rtpGrpcConn,
		roomManager: roomManager,
		logger:      logger,
	}
}

func (r *SpribeRouter) Route() {
	app := r.app

	routPath := fmt.Sprintf("/%s/websocket", r.gameName)
	// 只允许WebSocket升级的中间件
	app.Use(routPath, func(c *fiber.Ctx) error {
		if websocket.IsWebSocketUpgrade(c) {
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	})

	app.Get(routPath, websocket.New(func(c *websocket.Conn) {
		step := 0

		var player *Player = nil

		defer func() {
			if player != nil {
				if err := r.onDisconnect(player); err != nil {
					r.log.Errorf("onDisconnect error: %v", err)
				}
			}

			if err := c.Close(); err != nil {
				r.log.Errorf("close websocket error: %v", err)
			}
		}()

		for {
			_, msg, err := c.ReadMessage()

			if err != nil {
				r.log.Errorf("read websocket message error: %v", err)
				break
			}

			if step == 0 {
				if err := r.onHandshake(c, msg); err != nil {
					r.log.Errorf("handshake error: %v", err)
					break
				}
				step += 1
			} else if step == 1 {
				if player, err = r.onLogin(c, msg); err != nil {
					r.log.Errorf("login error: %v", err)
					break
				}
				step += 1
			} else {
				if err := r.onMessage(player, msg); err != nil {
					r.log.Errorf("handle message error: %v", err)
					break
				}
			}
		}
	}))
}

func (s *SpribeRouter) onHandshake(c *websocket.Conn, buff []byte) error {
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
func (r *SpribeRouter) onLogin(c *websocket.Conn, buff []byte) (*Player, error) {
	action, controller, data, err := utils.Unpack(buff)
	if err != nil {
		return nil, err
	}

	// 判断如果不是登陆协议
	if !(controller == 0 && action == 1) {
		return nil, fmt.Errorf("login action error")
	}

	// 解析请求参数
	var req protocol.LoginRequest
	if err := sfs.Unmarshal(data, &req); err != nil {
		return nil, err
	}

	// appid := req.Payload.Jurisdiction
	ssokey := req.Payload.Token

	params, err := player.DecodedSSOKeyV3(ssokey)
	if err != nil {
		r.log.Errorf("DecodeSSOKeyParams(%v) failed. err: %v", ssokey, err)
		return nil, err
	}

	if err := player.UpdateGameId(r.rdb, params.AppId, params.PlayerId, "spribe", r.gameName); err != nil {
		r.log.Errorf("UpdateGameId(%v) failed. err: %v", params.PlayerId, err)
		return nil, err
	}

	playerInfo, err := player.GetPlayerByAppAndPlayerId(r.rdb, params.AppId, params.PlayerId)
	if err != nil {
		r.log.Errorf("GetPlayerByAppAndPlayerId(%v) failed. err: %v", params.PlayerId, err)
		return nil, err
	}

	rtp, err := rtp_rpc_client.GetPlayerRtp(context.Background(), r.rtpGrpcConn, &rtp_rpc_v1.GetPlayerRtpRequest{
		PlayerId:  params.PlayerId,
		AppId:     params.AppId,
		GameBrand: "spribe",
		GameId:    r.gameName,
	})

	if err != nil {
		r.log.Errorf("GetPlayerRtp(%v) failed. err: %v", params.PlayerId, err)
		return nil, err
	}

	player := NewPlayer(GameBrand_Spribe, c, playerInfo, rtp.Rtp)

	// 初使化金币
	balanceRsp, err := client_utils.Balance(context.Background(), r.apiGrpcConn, playerInfo.AppID, &v1.BalanceRequest{
		PlayerId: playerInfo.PlayerID,
		Currency: playerInfo.Currency,
	})

	if err != nil {
		r.log.Errorf("Balance failed: %v", err)
		return nil, err
	}

	if err := player.SetBalanceByBalanceReply(balanceRsp); err != nil {
		r.log.Errorf("SetBalanceByBalanceReply failed: %v", err)
		return nil, err
	}

	// 注：这个时区要改成商户时区，待修正
	// room := r.GetRoom(playerInfo.AppID, playerInfo.Currency, rtp.Rtp, "Asia/Shanghai")
	// spribePlayer := game.NewSpribePlayer(conn, room)
	// spribePlayer.AppId = params.AppId
	// spribePlayer.PlayerId = params.PlayerId
	// spribePlayer.Lang = strings.ToLower(req.Payload.Lang) //语言
	// if err := room.OnLogin(spribePlayer); err != nil {
	// 	conn.Close()
	// 	return nil, err
	// }
	if err := r.roomManager.OnLogin(player); err != nil {
		return nil, err
	}
	return player, nil
}

func (s *SpribeRouter) onMessage(player *Player, buff []byte) error {
	return s.roomManager.OnMessage(player, buff)
}

func (s *SpribeRouter) onDisconnect(player *Player) error {
	return s.roomManager.OnDisConnect(player)
}

func (s *SpribeRouter) Send(c *websocket.Conn, controller int16, action uint8, payload sfs.SFSObject) error {
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
