package inout

import (
	"context"
	cryptorand "crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/bitly/go-simplejson"
	v1 "github.com/card-engine/game_common/api/game/v1"
	client_utils "github.com/card-engine/game_common/api/game/v1/client"
	rtp_rpc_v1 "github.com/card-engine/game_common/api/rtp/v1"
	rtp_rpc_client "github.com/card-engine/game_common/api/rtp/v1/client"
	"github.com/card-engine/game_common/gamehub/common"
	"github.com/card-engine/game_common/gamehub/types"

	inout_utils "github.com/card-engine/game_common/inout/utils"
	"github.com/card-engine/game_common/player"
	"github.com/card-engine/game_common/utils"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
	google_grpc "google.golang.org/grpc"
)

type InoutRouter struct {
	log      *log.Helper
	app      *fiber.App
	gameName string

	rdb         *redis.Client
	apiGrpcConn *google_grpc.ClientConn
	rtpGrpcConn *google_grpc.ClientConn
	roomManager *common.RoomManager
	lobby       types.LobbyImp
	logger      log.Logger
}

func NewInoutRouter(
	gameName string,
	app *fiber.App,
	rdb *redis.Client,
	apiGrpcConn *google_grpc.ClientConn,
	rtpGrpcConn *google_grpc.ClientConn,
	roomManager *common.RoomManager,
	lobby types.LobbyImp,
	logger log.Logger) *InoutRouter {
	return &InoutRouter{
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

func (r *InoutRouter) Route() {
	app := r.app

	routPath := fmt.Sprintf("/%s/io", r.gameName)

	app.Get(routPath, func(c *fiber.Ctx) error {
		// 验证是否为WebSocket升级请求
		if !websocket.IsWebSocketUpgrade(c) {
			return fiber.ErrUpgradeRequired
		}

		operatorId := c.Query("operatorId")
		token := c.Query("Authorization", "") // 默认值为 "0"

		return websocket.New(func(conn *websocket.Conn) {
			r.OnWebSocketHandler(conn, operatorId, token)
		})(c)
	})
}

func (r *InoutRouter) OnWebSocketHandler(c *websocket.Conn, operatorId string, token string) error {
	var inoutPlayer types.PlayerImp = nil
	defer func() {
		c.Close()
		if inoutPlayer != nil {
			r.roomManager.OnDisConnect(inoutPlayer)
			inoutPlayer.SetConn(nil)
		}
	}()

	tokenInfo, err := player.DecodedSSOKeyV3(token)
	if err != nil {
		r.log.Errorf("DecodedSSOKeyV3 failed: %v", err)
		return err
	}

	playerInfo, err := player.GetPlayerByAppAndPlayerId(r.rdb, tokenInfo.AppId, tokenInfo.PlayerId)
	if err != nil {
		r.log.Errorf("GetPlayerByAppAndPlayerId failed: %v", err)
		return err
	}

	rtp, err := rtp_rpc_client.GetPlayerRtp(context.Background(), r.rtpGrpcConn, &rtp_rpc_v1.GetPlayerRtpRequest{
		PlayerId:  tokenInfo.PlayerId,
		AppId:     tokenInfo.AppId,
		GameBrand: "inout",
		GameId:    tokenInfo.GameId,
	})
	if err != nil {
		r.log.Errorf("GetPlayerRtp failed: %v", err)
		return err
	}

	inoutPlayer = common.NewPlayer(types.GameBrand_Inout, c, playerInfo, rtp.Rtp)

	// inoutPlayer = &Player{
	// 	gameBrand:  GameBrand_Inout,
	// 	conn:       c,
	// 	PlayerInfo: playerInfo,
	// 	Rtp:        rtp.Rtp,
	// }
	// 发送握手响应
	if err := r.startHandshake(inoutPlayer); err != nil {
		r.log.Errorf("startHandshake failed: %v", err)
		return err
	}

	for {
		messageType, msg, err := c.ReadMessage()
		if err != nil {
			r.log.Errorf("read websocket message error: %v", err)
			break
		}

		if messageType != websocket.TextMessage {
			r.log.Error("recv error websocket message type")
			break
		}

		if err := r.OnMessage(inoutPlayer, msg); err != nil {
			r.log.Errorf("OnMessage failed: %v", err)
			break
		}
	}

	return nil
}

// generateSessionID 生成会话ID
func (r *InoutRouter) generateSessionID() string {
	bytes := make([]byte, 16)
	_, _ = cryptorand.Read(bytes)
	return hex.EncodeToString(bytes)[:20]
}

func (r *InoutRouter) OnMessage(player types.PlayerImp, msg []byte) error {
	msgType, payload, err := inout_utils.ParseCustomMessage(string(msg))
	if err != nil {
		r.log.Errorf("ParseCustomMessage failed: %v", err)
		return err
	}

	switch msgType {
	case "0": // Engine.IO握手请求，发送握手响应
		return nil
		//return s.onHandleSocketConnect(player)
	case "2": //ping
		return r.onPing(player)
	case "3": // pong
		return nil
	case "40":
		return r.onInitData(player)
	default:
		// 自定义消息格式处理
		return r.onCustomMessage(player, msgType, payload)
	}
}

func (r *InoutRouter) startHandshake(player types.PlayerImp) error {
	handshakeMsg := fmt.Sprintf(`0{"sid":"%s","upgrades":[],"pingInterval":25000,"pingTimeout":20000,"maxPayload":1000000}`, r.generateSessionID())
	return player.SendString(handshakeMsg)
}

func (r *InoutRouter) onPing(player types.PlayerImp) error {
	return player.SendString("3")
}

// 处理所有42xx消息并返回43xx响应
func (r *InoutRouter) onCustomMessage(player types.PlayerImp, msgType string, payload string) error {
	responseType := "43" + msgType[2:]
	if strings.HasPrefix(msgType, "42") {
		simpleJson, err := simplejson.NewJson([]byte(payload))
		if err != nil {
			r.log.Errorf("NewJson failed: %v", err)
			return err
		}

		if len(simpleJson.MustArray()) < 1 {
			r.log.Errorf("invalid custom message payload: %s", payload)
			return errors.New("invalid custom message payload")
		}

		// 提取 action
		action, err := simpleJson.GetIndex(0).String()
		if err != nil {
			r.log.Errorf("Get action failed: %v", err)
			return err
		}

		var dataJson *simplejson.Json = nil
		if len(simpleJson.MustArray()) > 1 {
			dataJson = simpleJson.GetIndex(1)
		}

		switch action {
		case "gameService-latencyTest":
			responseMsg := fmt.Sprintf(`%s[{"date":%v}]`, responseType, time.Now().UnixMilli())
			return player.SendString(responseMsg)
		case "gameService":
			return r.onHandleGameServiceMessage(player, dataJson, responseType)
		case "gameService-get-my-bets-history":
			// 交给房间处理
			return r.roomManager.OnMessage(player, &types.InoutMsgData{
				MsgId:   responseType,
				Action:  action,
				Payload: "{}",
			})
		case "changeGameAvatar":
			// 处理修改游戏头像消息
			return r.onHandleChangeGameAvatar(player, dataJson, responseType)
		}

		return nil
	}

	// return nil

	// responseTypeStr := "43" + msgType[2:] // 将42xx改为43xx
	// responseType, _ := strconv.Atoi(responseTypeStr)

	// // 解析载荷，根据事件名称判断消息类型
	// if payload != "" {
	// 	var eventData []interface{}
	// 	if err := json.Unmarshal([]byte(payload), &eventData); err == nil && len(eventData) > 0 {
	// 		if eventName, ok := eventData[0].(string); ok {
	// 			// wr.logger.Infof("处理事件: %s, 消息类型: %d", eventName, msgTypeInt)
	// 			switch eventName {
	// 			case "gameService":
	// 				// 处理游戏服务消息：bet, step, withdraw, get-game-seeds, get-game-state等
	// 				if len(eventData) > 1 {
	// 					return s.onHandleGameServiceMessage(player, eventData[1], responseType)
	// 				}
	// 			case "gameService-get-my-bets-history":
	// 				// 处理获取投注历史消息
	// 				return s.onHandleBetsHistory(player, responseType)
	// 			case "changeGameAvatar":
	// 				// 处理修改游戏头像消息
	// 				if len(eventData) > 1 {
	// 					err := s.onHandleChangeGameAvatar(player, eventData[1], responseType)
	// 					if err != nil {
	// 						r.log.Errorf("onHandleChangeGameAvatar failed: %v", err)
	// 						return err
	// 					}
	// 				}
	// 			default:
	// 				// 对于其他事件名称，返回通用响应
	// 				// wr.logger.Infof("未知事件类型: %s", eventName)
	// 				return nil
	// 			}
	// 		}
	// 	}
	// }

	// 默认返回通用的43x[null]响应
	responseMsg := fmt.Sprintf("%s[null]", responseType)
	return player.SendString(responseMsg)
}

// handleChangeGameAvatar 处理修改游戏头像
func (r *InoutRouter) onHandleChangeGameAvatar(player types.PlayerImp, data interface{}, responseType string) error {
	// reqData, ok := data.(map[string]interface{})
	// if !ok {
	// 	r.log.Errorf("无效的修改头像请求数据")
	// 	return
	// }

	// gameAvatarFloat, ok := reqData["gameAvatar"].(float64)
	// if !ok {
	// 	wr.logger.Errorf("缺少gameAvatar字段或类型错误")
	// 	return
	// }

	// gameAvatar := int(gameAvatarFloat)
	// if gameAvatar < 0 || gameAvatar > 11 {
	// 	wr.logger.Errorf("无效的头像ID: %d，必须在0-11之间", gameAvatar)
	// 	return
	// }

	// // 更新连接中保存的头像
	// wsConn.gameAvatar = gameAvatar
	// // wr.logger.Infof("用户 %s 修改头像为: %d", wsConn.userID, gameAvatar)

	// // 发送成功响应
	// responseData := map[string]interface{}{
	// 	"success":    true,
	// 	"gameAvatar": gameAvatar,
	// }
	// responseJSON, _ := json.Marshal([]interface{}{responseData})
	// responseMsg := fmt.Sprintf("%d%s", responseType, string(responseJSON))
	// _ = wsConn.SendMessage(responseMsg)
	return nil
}

func (r *InoutRouter) onHandleGameServiceMessage(player types.PlayerImp, data *simplejson.Json, responseType string) error {
	action := data.Get("action").MustString("")
	if action == "" {
		r.log.Errorf("缺少action字段")
		return errors.New("缺少action字段")
	}

	payload := ""
	if payloadJson, exists := data.CheckGet("payload"); exists {
		if payloadData, err := payloadJson.MarshalJSON(); err == nil {
			payload = string(payloadData)
		}
	}

	return r.roomManager.OnMessage(player, &types.InoutMsgData{
		MsgId:   responseType,
		Action:  action,
		Payload: payload,
	})
}

func (r *InoutRouter) onInitData(player types.PlayerImp) error {
	// 发送连接成功消息
	connectMsg := fmt.Sprintf(`40{"sid":"%s"}`, r.generateSessionID())
	if err := player.SendString(connectMsg); err != nil {
		r.log.Errorf("Send connectMsg failed: %v", err)
		return err
	}

	// 初使化金币
	balanceRsp, err := client_utils.Balance(context.Background(), r.apiGrpcConn, player.GetAppId(), &v1.BalanceRequest{
		PlayerId: player.GetPlayerId(),
		Currency: player.GetCurrency(),
	})

	if err != nil {
		r.log.Errorf("Balance failed: %v", err)
		return err
	}

	if err := player.SetBalanceByBalanceReply(balanceRsp); err != nil {
		r.log.Errorf("SetBalanceByBalanceReply failed: %v", err)
		return err
	}

	// 剩下的交给房间处理
	if err := r.lobby.OnLogin(player); err != nil {
		r.log.Errorf("OnLogin failed: %v", err)
		return err
	}

	nickname := player.GetPlayerId()
	if len(nickname) > 15 {
		nickname = nickname[:15] // 截断到前15个字符
	}

	err = player.SendString(fmt.Sprintf(`42["myData",{"userId":"%s","nickname":"%s","gameAvatar":null}]`, player.GetPlayerId(), nickname))
	if err != nil {
		r.log.Errorf("Send myData failed: %v", err)
		return err
	}

	jsonData, err := json.MarshalIndent(utils.ExchangeRates, "", "  ")
	if err != nil {
		log.Fatalf("ExchangeRates json.MarshalIndent failed: %v", err)
		return err
	}

	err = player.SendString(fmt.Sprintf(`42["currencies",%s]`, string(jsonData)))
	if err != nil {
		return err
	}

	return nil
}
