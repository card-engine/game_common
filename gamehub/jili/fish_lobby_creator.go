package jili

import (
	"github.com/card-engine/game_common/gamehub/types"
	"github.com/go-kratos/kratos/v2/log"
)

type FishLobbyCreator struct {
	log *log.Helper
}

func NewFishLobbyCreator(logger log.Logger) *FishLobbyCreator {
	return &FishLobbyCreator{log: log.NewHelper(logger)}
}

// 大厅创建器
func (c *FishLobbyCreator) CreateLobby(roomManager types.RoomManagerImp) types.LobbyImp {
	return NewFishLobby(roomManager, c.log)
}
