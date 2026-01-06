package common

import (
	"sync"
	"time"

	"github.com/card-engine/game_common/gamehub/const_val"
	"github.com/card-engine/game_common/gamehub/event"
	"github.com/card-engine/game_common/gamehub/types"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/ouqiang/timewheel"
)

type RoomManager struct {
	gameBrand types.GameBrand

	roomCreator types.RoomCreator

	roomMap   map[string][]types.RoomImp
	roomMapMu sync.RWMutex

	playerRoomMap   map[string]types.RoomImp
	playerRoomMapMu sync.RWMutex

	tableMatcherType types.TableMatcherType

	players sync.Map // 存储玩家

	log *log.Helper

	tw *timewheel.TimeWheel //时间轮
}

func NewRoomManager(
	gameBrand types.GameBrand,
	roomCreator types.RoomCreator,
	tableMatcherType types.TableMatcherType,
	logger log.Logger) *RoomManager {
	rm := &RoomManager{
		gameBrand:        gameBrand,
		roomCreator:      roomCreator,
		log:              log.NewHelper(logger),
		tableMatcherType: tableMatcherType,
		roomMap:          make(map[string][]types.RoomImp),
		playerRoomMap:    make(map[string]types.RoomImp),
	}

	//==================================================仅和inout有关系===========================================================
	// inout需要使用定时器发送心跳
	if gameBrand == types.GameBrand_Inout {
		tw := timewheel.New(1*time.Second, 3600, func(data interface{}) {
			rm.onTimer(data)
		})

		// 启动时间轮
		tw.Start()

		tw.AddTimer(time.Duration(const_val.InoutPingTime)*time.Second, const_val.InoutPingTimeWheelKey, &event.PingPongEvent{})

		rm.tw = tw
	}
	//============================================================================================================================

	return rm
}

func (r *RoomManager) ExitRoom(player types.PlayerImp, isDisconnect bool) {
	room := player.GetRoom()

	defer func() {
		player.SetRoom(nil)
		player.SetRoomManager(nil)
	}()

	//防止执行两次，使用LoadAndDelete会安全点
	r.players.Delete(player.GetPlayerIdent())

	r.playerRoomMapMu.Lock()
	playerIdent := player.GetPlayerIdent()
	_, ok := r.playerRoomMap[playerIdent]
	if ok {
		delete(r.playerRoomMap, playerIdent)
	}
	r.playerRoomMapMu.Unlock()

	player.CloseConn()

	// 如果是一次性房间，那么通知room也释放内存
	if r.tableMatcherType == types.TableMatcherType_SINGLE {
		room.OnDispose()
	}
}

// 尝试重连游戏， 返回的第一个参数是错误信息，第二个参数是是否进行了重连
func (r *RoomManager) TryReConnectGame(player types.PlayerImp) (error, bool) {
	//========================================================
	//判断是不是断线重连回来的
	r.playerRoomMapMu.Lock()
	playerIdent := player.GetPlayerIdent()
	room, ok := r.playerRoomMap[playerIdent]
	r.playerRoomMapMu.Unlock()
	player.SetRoom(room)

	if ok {
		if value, ok := r.players.Load(playerIdent); ok {
			if oldPlayer, ok := value.(*Player); ok && oldPlayer.conn != nil {
				// 以防止，旧的客户端没有完全处理干净
				oldPlayer.conn.Close()
			}
		}

		// 更新管理器记录的新的玩家对像
		r.players.Store(playerIdent, player)
		player.SetRoomManager(r)

		return room.OnReConnect(player), ok
	}
	//=========================配房逻辑==================================
	return nil, ok
}

func (r *RoomManager) OnJoin(player types.PlayerImp, roomType string, roomArgs interface{}) error {
	player.SetRoomManager(r)

	r.roomMapMu.Lock()
	if roomType != "" {
		if rooms, ok := r.roomMap[roomType]; ok {
			for _, room := range rooms {
				// 找到了一个可以登陆的房间
				if err := room.OnJoin(player); err == nil {
					player.SetRoom(room)
					break
				}
			}
		}
	}

	if player.GetRoom() == nil {
		room := r.roomCreator.CreateRoom(roomArgs)
		if err := room.OnJoin(player); err == nil {
			player.SetRoom(room)
		} else {
			r.log.Errorf("create room %s failed, err: %v", roomType, err)
			r.roomMapMu.Unlock()
			return err
		}

		if roomType != "" {
			r.roomMap[roomType] = append(r.roomMap[roomType], room)
		}
	}
	r.roomMapMu.Unlock()

	playerIdent := player.GetPlayerIdent()

	// 保存登陆信息，以便后续判断是否是重连回来的用户
	r.playerRoomMapMu.Lock()
	r.playerRoomMap[playerIdent] = player.GetRoom()
	r.players.Store(playerIdent, player)
	r.playerRoomMapMu.Unlock()
	return nil
}

func (r *RoomManager) OnMessage(player types.PlayerImp, msg interface{}) error {
	room := player.GetRoom()
	if room == nil {
		return nil
	}
	return room.OnMessage(player, msg)
}

func (r *RoomManager) OnDisConnect(player types.PlayerImp) error {
	room := player.GetRoom()
	if room != nil {
		return room.OnDisConnect(player)
	}
	return nil
}

// 定时器处理
func (r *RoomManager) onTimer(data interface{}) {
	if r.gameBrand == types.GameBrand_Inout {
		r.onInoutTimer(data)
	}

}

// =================================================仅和inout有关系=======================================================================
func (r *RoomManager) onInoutTimer(data interface{}) {
	switch data.(type) {
	case *event.PingPongEvent:
		r.tw.AddTimer(time.Duration(const_val.InoutPingTime)*time.Second, const_val.InoutPingTimeWheelKey, &event.PingPongEvent{})
		r.broadInoutPing()
	}
}

func (r *RoomManager) broadInoutPing() {
	players := []types.PlayerImp{}

	r.players.Range(func(key, value interface{}) bool {
		if player, ok := value.(types.PlayerImp); ok && player.IsConnect() {
			players = append(players, player)
		}
		return true // 继续迭代
	})

	for _, play := range players {
		play.SendString("2")
	}
}

//============================================================================================================================
