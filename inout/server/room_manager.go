package server

import (
	"fmt"
	"sync"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/ouqiang/timewheel"
)

type RoomManager struct {
	roomCreator RoomCreator

	roomMap   map[string][]RoomImp
	roomMapMu sync.RWMutex

	playerRoomMap   map[string]RoomImp
	playerRoomMapMu sync.RWMutex

	tableMatcherType TableMatcherType

	players sync.Map // 存储玩家

	tw  *timewheel.TimeWheel //时间轮
	log *log.Helper
}

const PingTimeWheelKey = "PingTimeWheelKey"
const PingTime = 25

type PingPongEvent struct{}

func NewRoomManager(roomCreator RoomCreator, tableMatcherType TableMatcherType, logger log.Logger) *RoomManager {
	rm := &RoomManager{
		roomCreator:      roomCreator,
		log:              log.NewHelper(logger),
		tableMatcherType: tableMatcherType,
		roomMap:          make(map[string][]RoomImp),
		playerRoomMap:    make(map[string]RoomImp),
	}

	tw := timewheel.New(1*time.Second, 3600, func(data interface{}) {
		rm.onTimer(data)
	})

	// 启动时间轮
	tw.Start()

	tw.AddTimer(time.Duration(PingTime)*time.Second, PingTimeWheelKey, &PingPongEvent{})

	rm.tw = tw
	return rm
}

func (r *RoomManager) ExitRoom(player *InoutPlayer, isDisconnect bool) {
	player.room = nil
	player.roomManager = nil
	//防止执行两次，使用LoadAndDelete会安全点
	r.players.Delete(player.GetPlayerIdent())

	r.playerRoomMapMu.Lock()
	defer r.playerRoomMapMu.Unlock()
	playerIdent := player.GetPlayerIdent()
	_, ok := r.playerRoomMap[playerIdent]
	if ok {
		delete(r.playerRoomMap, playerIdent)
	}

	if isDisconnect && player.conn != nil {
		player.conn.Close()
	}
}

func (r *RoomManager) OnLogin(player *InoutPlayer) error {
	//========================================================
	//判断是不是断线重连回来的
	r.playerRoomMapMu.Lock()
	playerIdent := player.GetPlayerIdent()
	room, ok := r.playerRoomMap[playerIdent]
	r.playerRoomMapMu.Unlock()
	player.room = room
	if ok {
		if value, ok := r.players.Load(playerIdent); ok {
			if oldPlayer, ok := value.(*InoutPlayer); ok && oldPlayer.conn != nil {
				// 以防止，旧的客户端没有完全处理干净
				oldPlayer.conn.Close()
			}
		}
		r.players.Store(playerIdent, player)

		return room.OnReConnect(player)
	}
	//============================================================
	if r.tableMatcherType == TableMatcherType_RTP {
		roomTypeStr := fmt.Sprintf("%v-%v", player.PlayerInfo.AppID, player.Rtp)
		r.roomMapMu.Lock()

		if rooms, ok := r.roomMap[roomTypeStr]; ok {
			for _, room := range rooms {
				// 找到了一个可以登陆的房间
				if err := room.OnLogin(player); err == nil {
					player.room = room
					break
				}
			}
		}

		if player.room == nil {
			var roomArgs interface{} = &RtpRoomArgs{
				Appid:    player.PlayerInfo.AppID,
				Rtp:      player.Rtp,
				Currency: player.PlayerInfo.Currency,
			}

			room := r.roomCreator.CreateRoom(roomArgs)
			if err := room.OnLogin(player); err == nil {
				player.room = room
			} else {
				r.log.Errorf("create room %s failed, err: %v", roomTypeStr, err)
				r.roomMapMu.Unlock()
				return err
			}

			r.roomMap[roomTypeStr] = append(r.roomMap[roomTypeStr], room)
		}
		r.roomMapMu.Unlock()

	} else if r.tableMatcherType == TableMatcherType_SINGLE {
		var roomArgs interface{} = &SingleRoomArgs{
			Appid: player.PlayerInfo.AppID,
		}

		room := r.roomCreator.CreateRoom(roomArgs)
		if err := room.OnLogin(player); err == nil {
			player.room = room
		} else {
			r.log.Errorf("create room %s failed, err: %v", err)
			r.roomMapMu.Unlock()
			return err
		}
	}

	// 保存登陆信息，以便后续判断是否是重连回来的用户
	r.playerRoomMapMu.Lock()
	r.playerRoomMap[playerIdent] = player.room
	r.players.Store(playerIdent, player)
	r.playerRoomMapMu.Unlock()
	return nil
}

func (r *RoomManager) broadPing() {
	players := []*InoutPlayer{}

	r.players.Range(func(key, value interface{}) bool {
		if player, ok := value.(*InoutPlayer); ok && player.IsConnect() {
			players = append(players, player)
		}
		return true // 继续迭代
	})

	for _, play := range players {
		play.Send("2")
	}
}

func (r *RoomManager) OnMessage(player *InoutPlayer, msgId, action, payload string) error {
	if player.room == nil {
		return nil
	}
	return player.room.OnMessage(player, msgId, action, payload)
}

func (r *RoomManager) onTimer(data interface{}) {
	switch data.(type) {
	case *PingPongEvent:
		r.tw.AddTimer(time.Duration(PingTime)*time.Second, PingTimeWheelKey, &PingPongEvent{})
		r.broadPing()
	}
}

func (r *RoomManager) OnDisConnect(player *InoutPlayer) error {
	room := player.room
	if room != nil {
		return room.OnDisConnect(player)
	}
	return nil
}
