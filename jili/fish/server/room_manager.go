package server

import (
	"fmt"
	"sync"

	"github.com/card-engine/game_common/jili/fish/message"
	"github.com/card-engine/game_common/jili/fish/pb"
)

type RoomManager struct {
	roomCreator RoomCreator
	roomMap     sync.Map
}

func NewRoomManager(roomCreator RoomCreator) *RoomManager {
	return &RoomManager{roomCreator: roomCreator}
}

func (r *RoomManager) GetRoom(appid, rtp string) RoomImp {
	roomKey := fmt.Sprintf("%s_%s", appid, rtp)
	if room, ok := r.roomMap.Load(roomKey); ok {
		return room.(RoomImp)
	}

	room := r.roomCreator.CreateRoom(&RoomArgs{Appid: appid, Rtp: rtp})
	actual, loaded := r.roomMap.LoadOrStore(roomKey, room)
	if loaded {
		return actual.(RoomImp)
	}

	return room
}

// 玩家从房间移除
func (r *RoomManager) ExitRoom(player *JiliPlayer) {

}

func (r *RoomManager) OnLogin(player *JiliPlayer) {
	room := r.GetRoom(player.AppId, player.Rtp)
	player.room = room
	room.OnLogin(player)
}

func (r *RoomManager) OnReConnect(player *JiliPlayer) {

}

func (rm *RoomManager) OnDisConnect(player *JiliPlayer) error {
	room := player.room
	if room != nil {
		return room.OnDisConnect(player)
	}
	player.room = nil
	return nil
}

func (r *RoomManager) OnMessage(player *JiliPlayer, cmd int32, msg []byte) error {
	room := player.room
	if room == nil {
		return nil
	}

	// 心跳包
	if cmd == int32(pb.UserToServer_U2S_HEART_CHECK_REQ) {
		if sendData, err := message.MsgPack(uint32(pb.ServerToUser_S2U_HEART_CHECK_ACK), []byte{}); err == nil {
			return player.Send(sendData)
		} else {
			return err
		}
	}

	return room.OnMessage(player, cmd, msg)
}
