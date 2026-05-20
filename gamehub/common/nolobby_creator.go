package common

import "github.com/card-engine/game_common/gamehub/types"

// 没有游戏大厅的创建者

type NoLobbyCreator struct {
	tableMatcherType types.TableMatcherType
	gameBrand        types.GameBrand
}

func NewNoLobbyCreator(tableMatcherType types.TableMatcherType, gameBrand types.GameBrand) *NoLobbyCreator {
	return &NoLobbyCreator{
		tableMatcherType: tableMatcherType,
		gameBrand:        gameBrand,
	}
}

// 大厅创建器
func (c *NoLobbyCreator) CreateLobby(roomManager types.RoomManagerImp) types.LobbyImp {
	return NewNoLobby(roomManager, c.tableMatcherType, c.gameBrand)
}
