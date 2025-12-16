package server

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	cryptorand "crypto/rand"

	"github.com/bitly/go-simplejson"
	inout_utils "github.com/card-engine/game_common/inout/utils"
	"github.com/card-engine/game_common/player"
	"github.com/card-engine/game_common/utils"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"

	v1 "github.com/card-engine/game_common/api/game/v1"
	client_utils "github.com/card-engine/game_common/api/game/v1/client"
	rtp_rpc_v1 "github.com/card-engine/game_common/api/rtp/v1"
	rtp_rpc_client "github.com/card-engine/game_common/api/rtp/v1/client"
	google_grpc "google.golang.org/grpc"
)

type InoutGameApiServer struct {
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

func InitInoutGameApiServer(
	serverName string, gameName string, serverAddr string,
	tableMatcherType TableMatcherType,
	roomCreator RoomCreator, rdb *redis.Client,
	apiGrpcConn *google_grpc.ClientConn, rtpGrpcConn *google_grpc.ClientConn,
	logger log.Logger) *InoutGameApiServer {
	app := fiber.New()

	s := &InoutGameApiServer{
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

func (s *InoutGameApiServer) route() {
	app := s.app

	routPath := fmt.Sprintf("/%s/io", s.gameName)

	app.Get(routPath, func(c *fiber.Ctx) error {
		// 验证是否为WebSocket升级请求
		if !websocket.IsWebSocketUpgrade(c) {
			return fiber.ErrUpgradeRequired
		}

		operatorId := c.Query("operatorId")
		token := c.Query("Authorization", "") // 默认值为 "0"

		return websocket.New(func(conn *websocket.Conn) {
			s.OnWebSocketHandler(conn, operatorId, token)
		})(c)
	})
}

func (s *InoutGameApiServer) OnWebSocketHandler(c *websocket.Conn, operatorId string, token string) error {
	var inoutPlayer *InoutPlayer = nil
	defer func() {
		c.Close()
		if inoutPlayer != nil {
			s.roomManager.OnDisConnect(inoutPlayer)
			inoutPlayer.conn = nil
		}
	}()

	tokenInfo, err := player.DecodedSSOKeyV3(token)
	if err != nil {
		s.log.Errorf("DecodedSSOKeyV3 failed: %v", err)
		return err
	}

	playerInfo, err := player.GetPlayerByAppAndPlayerId(s.rdb, tokenInfo.AppId, tokenInfo.PlayerId)

	// token, err := player.GetTokenBySSOKey(s.rdb, token, operatorId)
	// if err != nil {
	// 	s.log.Errorf("GetTokenBySSOKey failed: %v", err)
	// 	return err
	// }

	// playerInfo, err := player.GetPlayerInfoByToken(s.rdb, token)

	// if err != nil {
	// 	s.log.Errorf("DecodedSSOKeyV3 failed: %v", err)
	// 	return err
	// }

	rtp, err := rtp_rpc_client.GetPlayerRtp(context.Background(), s.rtpGrpcConn, &rtp_rpc_v1.GetPlayerRtpRequest{
		PlayerId:  tokenInfo.PlayerId,
		AppId:     tokenInfo.AppId,
		GameBrand: "inout",
		GameId:    tokenInfo.GameId,
	})
	if err != nil {
		s.log.Errorf("GetPlayerRtp failed: %v", err)
		return err
	}

	inoutPlayer = &InoutPlayer{
		conn:       c,
		PlayerInfo: playerInfo,
		Rtp:        rtp.Rtp,
	}
	// 发送握手响应
	if err := s.startHandshake(inoutPlayer); err != nil {
		s.log.Errorf("startHandshake failed: %v", err)
		return err
	}

	for {
		messageType, msg, err := c.ReadMessage()
		if err != nil {
			s.log.Errorf("read websocket message error: %v", err)
			break
		}

		if messageType != websocket.TextMessage {
			s.log.Error("recv error websocket message type")
			break
		}

		if err := s.OnMessage(inoutPlayer, msg); err != nil {
			s.log.Errorf("OnMessage failed: %v", err)
			break
		}
	}

	return nil
}

// generateSessionID 生成会话ID
func (s *InoutGameApiServer) generateSessionID() string {
	bytes := make([]byte, 16)
	_, _ = cryptorand.Read(bytes)
	return hex.EncodeToString(bytes)[:20]
}

func (s *InoutGameApiServer) OnMessage(player *InoutPlayer, msg []byte) error {
	msgType, payload, err := inout_utils.ParseCustomMessage(string(msg))
	if err != nil {
		s.log.Errorf("ParseCustomMessage failed: %v", err)
		return err
	}

	switch msgType {
	case "0": // Engine.IO握手请求，发送握手响应
		return nil
		//return s.onHandleSocketConnect(player)
	case "2": //ping
		return s.onPing(player)
	case "3": // pong
		return nil
	case "40":
		return s.onInitData(player)
	default:
		// 自定义消息格式处理
		return s.onCustomMessage(player, msgType, payload)
	}
}

func (s *InoutGameApiServer) startHandshake(player *InoutPlayer) error {
	handshakeMsg := fmt.Sprintf(`0{"sid":"%s","upgrades":[],"pingInterval":25000,"pingTimeout":20000,"maxPayload":1000000}`, s.generateSessionID())
	return player.Send(handshakeMsg)
}

func (s *InoutGameApiServer) onPing(player *InoutPlayer) error {
	return player.Send("3")
}

// 处理所有42xx消息并返回43xx响应
func (s *InoutGameApiServer) onCustomMessage(player *InoutPlayer, msgType string, payload string) error {
	responseType := "43" + msgType[2:]
	if strings.HasPrefix(msgType, "42") {
		simpleJson, err := simplejson.NewJson([]byte(payload))
		if err != nil {
			s.log.Errorf("NewJson failed: %v", err)
			return err
		}

		if len(simpleJson.MustArray()) < 1 {
			s.log.Errorf("invalid custom message payload: %s", payload)
			return errors.New("invalid custom message payload")
		}

		// 提取 action
		action, err := simpleJson.GetIndex(0).String()
		if err != nil {
			s.log.Errorf("Get action failed: %v", err)
			return err
		}

		var dataJson *simplejson.Json = nil
		if len(simpleJson.MustArray()) > 1 {
			dataJson = simpleJson.GetIndex(1)
		}

		switch action {
		case "gameService-latencyTest":
			responseMsg := fmt.Sprintf(`%s[{"date":%v}]`, responseType, time.Now().UnixMilli())
			return player.Send(responseMsg)
		case "gameService":
			return s.onHandleGameServiceMessage(player, dataJson, responseType)
		case "gameService-get-my-bets-history":
			// 交给房间处理
			return s.roomManager.OnMessage(player, responseType, action, "{}")
		case "changeGameAvatar":
			// 处理修改游戏头像消息
			return s.onHandleChangeGameAvatar(player, dataJson, responseType)
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
	// 			// ws.logger.Infof("处理事件: %s, 消息类型: %d", eventName, msgTypeInt)
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
	// 						s.log.Errorf("onHandleChangeGameAvatar failed: %v", err)
	// 						return err
	// 					}
	// 				}
	// 			default:
	// 				// 对于其他事件名称，返回通用响应
	// 				// ws.logger.Infof("未知事件类型: %s", eventName)
	// 				return nil
	// 			}
	// 		}
	// 	}
	// }

	// 默认返回通用的43x[null]响应
	responseMsg := fmt.Sprintf("%s[null]", responseType)
	return player.Send(responseMsg)
}

// handleChangeGameAvatar 处理修改游戏头像
func (s *InoutGameApiServer) onHandleChangeGameAvatar(player *InoutPlayer, data interface{}, responseType string) error {
	// reqData, ok := data.(map[string]interface{})
	// if !ok {
	// 	s.log.Errorf("无效的修改头像请求数据")
	// 	return
	// }

	// gameAvatarFloat, ok := reqData["gameAvatar"].(float64)
	// if !ok {
	// 	ws.logger.Errorf("缺少gameAvatar字段或类型错误")
	// 	return
	// }

	// gameAvatar := int(gameAvatarFloat)
	// if gameAvatar < 0 || gameAvatar > 11 {
	// 	ws.logger.Errorf("无效的头像ID: %d，必须在0-11之间", gameAvatar)
	// 	return
	// }

	// // 更新连接中保存的头像
	// wsConn.gameAvatar = gameAvatar
	// // ws.logger.Infof("用户 %s 修改头像为: %d", wsConn.userID, gameAvatar)

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

func (s *InoutGameApiServer) onHandleGameServiceMessage(player *InoutPlayer, data *simplejson.Json, responseType string) error {
	action := data.Get("action").MustString("")
	if action == "" {
		s.log.Errorf("缺少action字段")
		return errors.New("缺少action字段")
	}

	payload := ""
	if payloadJson, exists := data.CheckGet("payload"); exists {
		if payloadData, err := payloadJson.MarshalJSON(); err == nil {
			payload = string(payloadData)
		}
	}

	return s.roomManager.OnMessage(player, responseType, action, payload)
}

func (s *InoutGameApiServer) onInitData(player *InoutPlayer) error {
	// 发送连接成功消息
	connectMsg := fmt.Sprintf(`40{"sid":"%s"}`, s.generateSessionID())
	if err := player.Send(connectMsg); err != nil {
		s.log.Errorf("Send connectMsg failed: %v", err)
		return err
	}

	// 初使化金币
	balanceRsp, err := client_utils.Balance(context.Background(), s.apiGrpcConn, player.GetAppId(), &v1.BalanceRequest{
		PlayerId: player.GetPlayerId(),
		Currency: player.PlayerInfo.Currency,
	})
	if err != nil {
		s.log.Errorf("Balance failed: %v", err)
		return err
	}

	if err := player.SetBalanceByBalanceReply(balanceRsp); err != nil {
		s.log.Errorf("SetBalanceByBalanceReply failed: %v", err)
		return err
	}

	// 剩下的交给房间处理
	if err := s.roomManager.OnLogin(player); err != nil {
		s.log.Errorf("OnLogin failed: %v", err)
		return err
	}

	nickname := player.GetPlayerId()
	if len(nickname) > 15 {
		nickname = nickname[:15] // 截断到前15个字符
	}

	err = player.Send(fmt.Sprintf(`42["myData",{"userId":"%s","nickname":"%s","gameAvatar":null}]`, player.GetPlayerId(), nickname))
	if err != nil {
		s.log.Errorf("Send myData failed: %v", err)
		return err
	}

	jsonData, err := json.MarshalIndent(utils.ExchangeRates, "", "  ")
	if err != nil {
		log.Fatalf("ExchangeRates json.MarshalIndent failed: %v", err)
		return err
	}

	err = player.Send(fmt.Sprintf(`42["currencies",%s]`, string(jsonData)))
	if err != nil {
		return err
	}

	return nil
}

func (s *InoutGameApiServer) Start(ctx context.Context) error {
	go func() {
		if err := s.app.Listen(s.serverAddr); err != nil {
			log.Fatalf("Listen failed: %v", err)
		}
	}()

	return nil
}

func (s *InoutGameApiServer) Stop(ctx context.Context) error {
	if s.app != nil {
		s.app.Shutdown()
		s.app = nil
	}

	return nil
}
