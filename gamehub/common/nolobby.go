package common

import (
	"fmt"

	"github.com/card-engine/game_common/gamehub/types"
)

type NoLobby struct {
	roomManager types.RoomManagerImp

	tableMatcherType types.TableMatcherType
	gameBrand        types.GameBrand

	commonMessageHandler types.CommonMessageHandler
}

func NewNoLobby(roomManager types.RoomManagerImp, tableMatcherType types.TableMatcherType, gameBrand types.GameBrand) *NoLobby {
	l := &NoLobby{
		roomManager:      roomManager,
		tableMatcherType: tableMatcherType,
		gameBrand:        gameBrand,
	}

	if gameBrand == types.GameBrand_Jili {
		l.commonMessageHandler = &JiliCommonMessageHandler{}
	}

	return l
}

func (l *NoLobby) OnMessage(player types.PlayerImp, data interface{}) error {
	if l.commonMessageHandler != nil {
		return l.commonMessageHandler.OnMessage(player, data)
	}
	return nil
}

func (l *NoLobby) OnLogin(player types.PlayerImp) error {
	// 尝试重连游戏
	err, ok := l.roomManager.TryReConnectGame(player)
	if ok { //有重连
		return err
	} else { //无重连
		switch l.tableMatcherType {
		case types.TableMatcherType_RTP:
			return l.roomManager.OnJoin(player,
				fmt.Sprintf("%v-%v", player.GetPlayerInfo().AppID, player.GetRtpStr()),
				&types.RtpRoomArgs{
					Appid:    player.GetPlayerInfo().AppID,
					Rtp:      player.GetRtpStr(),
					Currency: player.GetPlayerInfo().Currency,
				})

		case types.TableMatcherType_SINGLE:
			return l.roomManager.OnJoin(player, "", &types.SingleRoomArgs{
				Appid: player.GetPlayerInfo().AppID,
			})
		}

		return fmt.Errorf("tableMatcherType not support")
	}
}
