package server

import "github.com/qd2ss/sfs"

// 定义一个房间的概念
type RoomImp interface {
	// 获取当前玩家的数量
	GetPlayerNum() int32
	// 玩家进入房间
	OnLogin(player *JdbPlayer) error
	// 玩家重连房间
	OnReConnect(player *JdbPlayer) error
	// 玩家退出房间
	OnDisConnect(player *JdbPlayer) error
	// 玩家收到了消息了
	OnMessage(player *JdbPlayer, cmd string, data sfs.SFSObject) error
}

type RoomArgs struct {
	RoomType string //房间类型
	// Appid    string
	// Rtp      string
}

type RoomCreator interface {
	CreateRoom(args *RoomArgs) RoomImp
}

// 配桌算法
type TableMatcherType int

const (
	TableMatcherType_RTP TableMatcherType = iota //通过rtp进行配桌，适合小飞机类游戏
)
