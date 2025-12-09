package server

import (
	"strconv"
	"sync"

	"github.com/card-engine/game_common/sfs/utils"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/gofiber/contrib/websocket"
	"github.com/qd2ss/sfs"
)

type RoomManager struct {
	roomCreator RoomCreator

	roomMap   map[string][]RoomImp
	roomMapMu sync.RWMutex

	log *log.Helper
}

func NewRoomManager(roomCreator RoomCreator, logger log.Logger) *RoomManager {
	return &RoomManager{
		roomCreator: roomCreator,
		log:         log.NewHelper(logger),
		roomMap:     make(map[string][]RoomImp),
	}
}

func (r *RoomManager) GetRoom(appid, rtp string) RoomImp {
	// roomKey := fmt.Sprintf("%s_%s", appid, rtp)
	// if room, ok := r.roomMap.Load(roomKey); ok {
	// 	return room.(RoomImp)
	// }

	// room := r.roomCreator.CreateRoom(&RoomArgs{Appid: appid, Rtp: rtp})
	// actual, loaded := r.roomMap.LoadOrStore(roomKey, room)
	// if loaded {
	// 	return actual.(RoomImp)
	// }

	// return room
	return nil
}

// 玩家从房间移除
func (r *RoomManager) ExitRoom(player *JdbPlayer) {

}

func (r *RoomManager) OnLogin(player *JdbPlayer) {
	room := r.GetRoom(player.AppId, player.Rtp)
	player.room = room
	room.OnLogin(player)
}

func (r *RoomManager) OnReConnect(player *JdbPlayer) {

}

func (rm *RoomManager) OnDisConnect(player *JdbPlayer) error {
	room := player.room
	if room != nil {
		return room.OnDisConnect(player)
	}
	player.room = nil
	return nil
}

func (r *RoomManager) onGameLogin(c *websocket.Conn, p sfs.SFSObject) (*JdbPlayer, error) {
	player := &JdbPlayer{
		// AppId: player.AppId,
		// Rtp:   player.Rtp,
		conn:        c,
		roomManager: r,
	}

	if data, err := utils.PackCustomData("gameLoginReturn", sfs.SFSObject{
		"GroupId":   "7003_RS",
		"balance":   float64(1000),
		"data":      true,
		"loginRoom": "7003_RS",
		"serverId":  "02",
		"testMode":  false,
		"ts":        int64(1761058708274),
	}); err != nil {
		return nil, err
	} else {
		if err := player.Send(data); err != nil {
			return nil, err
		}
	}

	if data, err := utils.PackCustomData("EV_SG_USERINFO", sfs.SFSObject{
		"iDenom": []int32{
			1,
			1,
			1,
		},
		"iMaxBet": []int32{
			100000,
			100000,
			10000,
		},
		"iMinBet": []int32{
			10000,
			1000,
			100,
		},
		"iRoomCount": int32(3),
		"iRoomType": []int32{
			2,
			1,
			0,
		},
		"llCent":     int64(1000000),
		"showName":   "tprich142956",
		"singleMode": false,
	}); err != nil {
		return nil, err
	} else {
		if err := player.Send(data); err != nil {
			return nil, err
		}
	}

	return player, nil

}

func (r *RoomManager) OnMessage(player *JdbPlayer, cmd string, data sfs.SFSObject) error {
	if cmd == "EV_GS_QUICK_LOGIN" {
		iRoomType := data["iRoomType"].(int32)
		r.roomMapMu.Lock()
		roomTypeStr := strconv.FormatInt(int64(iRoomType), 10)
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
			room := r.roomCreator.CreateRoom(&RoomArgs{RoomType: roomTypeStr})
			if err := room.OnLogin(player); err == nil {
				player.room = room
			} else {
				r.log.Errorf("create room %s failed, err: %v", roomTypeStr, err)
				return err
			}

			if _, ok := r.roomMap[roomTypeStr]; ok {
				r.roomMap[roomTypeStr] = append(r.roomMap[roomTypeStr], room)
			} else {
				r.roomMap[roomTypeStr] = []RoomImp{room}
			}
		}

		r.roomMapMu.Unlock()
		return nil
	}

	room := player.room
	if room == nil {
		return nil
	}

	// 玩家收到了消息了
	if err := room.OnMessage(player, cmd, data); err != nil {
		return err
	}

	return nil
}
