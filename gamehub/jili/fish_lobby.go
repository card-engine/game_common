package jili

import (
	"github.com/card-engine/game_common/gamehub/jili/annin_protocol"
	pb "github.com/card-engine/game_common/gamehub/jili/annin_protocol"
	"github.com/card-engine/game_common/gamehub/types"
	"github.com/go-kratos/kratos/v2/log"
	"google.golang.org/protobuf/proto"
)

type Lobby struct {
	roomManager types.RoomManagerImp
	log         *log.Helper
}

func NewFishLobby(roomManager types.RoomManagerImp, log *log.Helper) *Lobby {
	return &Lobby{
		roomManager: roomManager,
		log:         log,
	}
}

func (l *Lobby) OnMessage(player types.PlayerImp, data interface{}) error {
	command := data.(*pb.Command)

	// ping
	if command.Type == uint32(pb.UserToServer_U2S_HEART_CHECK_REQ) {
		return l.send(player, uint32(pb.ServerToUser_S2U_HEART_CHECK_ACK), nil)
	} else if command.Type == uint32(pb.UserToServer_U2S_CONFIG_INFO) { //120
		return l.onConfigInfo(player)
	} else if command.Type == uint32(pb.UserToServer_U2S_VIP_INFO) { //92
		return l.onVipInfo(player)
	} else if command.Type == uint32(pb.UserToServer_U2S_PROMOTION_INFO) { //95
		return l.onPromotionInfo(player)
	} else if command.Type == uint32(pb.UserToServer_U2S_MAIL_INFO) { //115
		return l.onMailInfo(player)
	} else if command.Type == uint32(pb.UserToServer_U2S_CARD_INFO) { //82
		return l.onCardInfo(player)
		// } else if command.Type == uint32(protocol.UserToServer_U2S_JOIN_TABLE_REQ) { //11
		// 	return l.onJoinTableReq(player, command.Data)
		// } else if command.Type == uint32(protocol.UserToServer_U2S_LOGIN_REQ) { // 0 (主动请求的登陆协议)
		// 	// return l.OnLogin(player)
	}
	return nil
}

func (l *Lobby) onConfigInfo(player types.PlayerImp) error {
	rsp := &pb.AgentConfig{
		BetList: []float64{0.1, 0.2, 0.3, 0.4, 0.5, 0.6, 0.7, 0.8, 0.9, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 20, 30, 40, 50, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 20, 30, 40, 50, 60, 70, 80, 90, 100},
	}

	if err := l.send(player, uint32(pb.ServerToUser_S2U_CONFIG_INFO), rsp); err != nil {
		return err
	}

	return nil
}

func (l *Lobby) onVipInfo(player types.PlayerImp) error {
	rsp := &pb.Vip{
		Exp:      0.8,
		Treasure: &pb.Vip_Treasure{Exp: 0.8},
	}
	if err := l.send(player, uint32(pb.ServerToUser_S2U_VIP_INFO), rsp); err != nil {
		return err
	}

	return nil
}

func (l *Lobby) onPromotionInfo(player types.PlayerImp) error {
	rsp := &pb.Promotion{
		Enabled: true,
		List: []*pb.Promotion_Info{
			{
				Id:     32,
				Labels: []uint32{1},
				Dau:    25013,
			},
		},
	}

	if err := l.send(player, uint32(pb.ServerToUser_S2U_PROMOTION_INFO), rsp); err != nil {
		return err
	}

	return nil
}

func (l *Lobby) onMailInfo(player types.PlayerImp) error {

	return nil
}

func (l *Lobby) onCardInfo(player types.PlayerImp) error {
	rsp := &pb.CardUseAck{}
	if err := l.send(player, uint32(pb.ServerToUser_S2U_CARD_INFO), rsp); err != nil {
		return err
	}

	return nil
}

func (l *Lobby) onJoinTableReq(player types.PlayerImp, data []byte) error {
	// err, isReconnect := l.roomManager.TryReConnectGame(player)
	// if err != nil {
	// 	return err
	// }

	// if !isReconnect {
	// 	req := &protocol.JoinTableReq{}
	// 	if err := proto.Unmarshal(data, req); err != nil {
	// 		return err
	// 	}

	// 	// 1是欢乐场， 2是富豪场
	// 	roomType := fmt.Sprintf("room_type:%v:%v", req.Theme, req.Room)

	// 	if err := l.roomManager.OnJoin(player, roomType, req); err != nil {
	// 		return err
	// 	}
	// }

	return nil
}

func (l *Lobby) OnLogin(player types.PlayerImp) error {
	// var currency int32
	// if v, ok := game_pb.Currency_value[player.GetCurrency()]; ok {
	// 	currency = v
	// }

	// rsp := &protocol.LoginDataAck{
	// 	Uid: 2200154868,
	// 	Player: &protocol.PlayerInfo{
	// 		Uid:  2200154868,
	// 		Name: player.GetPlayerInfo().PlayerID,
	// 		Coin: player.GetBalance(),
	// 		Seat: 99,
	// 	},
	// 	Version: 1747879611,

	// 	Reward: &protocol.Rewardv2{
	// 		Level: &protocol.Rewardv2_Level{
	// 			Table: []*protocol.Rewardv2_ExpTable{
	// 				{Level: 1},
	// 				{Level: 2, Exp: 1875},
	// 				{Level: 3, Exp: 3750},
	// 				{Level: 4, Exp: 5625},
	// 				{Level: 5, Exp: 7500, Reward: 30},
	// 				{Level: 6, Exp: 10000},
	// 				{Level: 7, Exp: 12500},
	// 				{Level: 8, Exp: 15000},
	// 				{Level: 9, Exp: 17500},
	// 				{Level: 10, Exp: 20000, Reward: 50},
	// 				{Level: 11, Exp: 28000},
	// 				{Level: 12, Exp: 36000},
	// 				{Level: 13, Exp: 44000},
	// 				{Level: 14, Exp: 52000},
	// 				{Level: 15, Exp: 60000, Reward: 100},
	// 				{Level: 16, Exp: 76000},
	// 				{Level: 17, Exp: 92000},
	// 				{Level: 18, Exp: 108000},
	// 				{Level: 19, Exp: 124000},
	// 				{Level: 20, Exp: 140000, Reward: 200},
	// 				{Level: 21, Exp: 155000},
	// 				{Level: 22, Exp: 170000},
	// 				{Level: 23, Exp: 185000},
	// 				{Level: 24, Exp: 200000},
	// 				{Level: 25, Exp: 215000},
	// 				{Level: 26, Exp: 235000},
	// 				{Level: 27, Exp: 255000},
	// 				{Level: 28, Exp: 275000},
	// 				{Level: 29, Exp: 295000},
	// 				{Level: 30, Exp: 315000, Reward: 300},
	// 				{Level: 31, Exp: 340000},
	// 				{Level: 32, Exp: 365000},
	// 				{Level: 33, Exp: 390000},
	// 				{Level: 34, Exp: 415000},
	// 				{Level: 35, Exp: 440000},
	// 				{Level: 36, Exp: 465000},
	// 				{Level: 37, Exp: 490000},
	// 				{Level: 38, Exp: 515000},
	// 				{Level: 39, Exp: 540000},
	// 				{Level: 40, Exp: 565000, Reward: 400},
	// 				{Level: 41, Exp: 590000},
	// 				{Level: 42, Exp: 615000},
	// 				{Level: 43, Exp: 640000},
	// 				{Level: 44, Exp: 665000},
	// 				{Level: 45, Exp: 690000},
	// 				{Level: 46, Exp: 715000},
	// 				{Level: 47, Exp: 740000},
	// 				{Level: 48, Exp: 765000},
	// 				{Level: 49, Exp: 790000},
	// 				{Level: 50, Exp: 815000, Reward: 500},
	// 				{Level: 51, Exp: 865000},
	// 				{Level: 52, Exp: 915000},
	// 				{Level: 53, Exp: 965000},
	// 				{Level: 54, Exp: 1015000},
	// 				{Level: 55, Exp: 1065000},
	// 				{Level: 56, Exp: 1115000},
	// 				{Level: 57, Exp: 1165000},
	// 				{Level: 58, Exp: 1215000},
	// 				{Level: 59, Exp: 1265000},
	// 				{Level: 60, Exp: 1315000, Reward: 800},
	// 				{Level: 61, Exp: 1377500},
	// 				{Level: 62, Exp: 1440000},
	// 				{Level: 63, Exp: 1502500},
	// 				{Level: 64, Exp: 1565000},
	// 				{Level: 65, Exp: 1627500},
	// 				{Level: 66, Exp: 1690000},
	// 				{Level: 67, Exp: 1752500},
	// 				{Level: 68, Exp: 1815000},
	// 				{Level: 69, Exp: 1877500},
	// 				{Level: 70, Exp: 1940000, Reward: 1000},
	// 				{Level: 71, Exp: 2033750},
	// 				{Level: 72, Exp: 2127500},
	// 				{Level: 73, Exp: 2221250},
	// 				{Level: 74, Exp: 2315000},
	// 				{Level: 75, Exp: 2408750},
	// 				{Level: 76, Exp: 2502500},
	// 				{Level: 77, Exp: 2596250},
	// 				{Level: 78, Exp: 2690000},
	// 				{Level: 79, Exp: 2783750},
	// 				{Level: 80, Exp: 2877500, Reward: 1500},
	// 				{Level: 81, Exp: 3077500},
	// 				{Level: 82, Exp: 3277500},
	// 				{Level: 83, Exp: 3477500},
	// 				{Level: 84, Exp: 3677500},
	// 				{Level: 85, Exp: 3877500},
	// 				{Level: 86, Exp: 4077500},
	// 				{Level: 87, Exp: 4277500},
	// 				{Level: 88, Exp: 4477500},
	// 				{Level: 89, Exp: 4677500},
	// 				{Level: 90, Exp: 4877500, Reward: 2000},
	// 				{Level: 91, Exp: 5127500},
	// 				{Level: 92, Exp: 5377500},
	// 				{Level: 93, Exp: 5627500},
	// 				{Level: 94, Exp: 5877500},
	// 				{Level: 95, Exp: 6127500},
	// 				{Level: 96, Exp: 6377500},
	// 				{Level: 97, Exp: 6627500},
	// 				{Level: 98, Exp: 6877500},
	// 				{Level: 99, Exp: 7127500},
	// 				{Level: 100, Exp: 7377500, Reward: 2500},
	// 				{Level: 101, Exp: 7877500},
	// 				{Level: 102, Exp: 8377500},
	// 				{Level: 103, Exp: 8877500},
	// 				{Level: 104, Exp: 9377500},
	// 				{Level: 105, Exp: 9877500},
	// 				{Level: 106, Exp: 10377500},
	// 				{Level: 107, Exp: 10877500},
	// 				{Level: 108, Exp: 11377500},
	// 				{Level: 109, Exp: 11877500},
	// 				{Level: 110, Exp: 12377500, Reward: 5000},
	// 				{Level: 111, Exp: 13177500},
	// 				{Level: 112, Exp: 13977500},
	// 				{Level: 113, Exp: 14777500},
	// 				{Level: 114, Exp: 15577500},
	// 				{Level: 115, Exp: 16377500},
	// 				{Level: 116, Exp: 17177500},
	// 				{Level: 117, Exp: 17977500},
	// 				{Level: 118, Exp: 18777500},
	// 				{Level: 119, Exp: 19577500},
	// 				{Level: 120, Exp: 20377500, Reward: 8000},
	// 				{Level: 121, Exp: 21377500},
	// 				{Level: 122, Exp: 22377500},
	// 				{Level: 123, Exp: 23377500},
	// 				{Level: 124, Exp: 24377500},
	// 				{Level: 125, Exp: 25377500},
	// 				{Level: 126, Exp: 26377500},
	// 				{Level: 127, Exp: 27377500},
	// 				{Level: 128, Exp: 28377500},
	// 				{Level: 129, Exp: 29377500},
	// 				{Level: 130, Exp: 30377500, Reward: 10000},
	// 				{Level: 131, Exp: 31877500},
	// 				{Level: 132, Exp: 33377500},
	// 				{Level: 133, Exp: 34877500},
	// 				{Level: 134, Exp: 36377500},
	// 				{Level: 135, Exp: 37877500},
	// 				{Level: 136, Exp: 39377500},
	// 				{Level: 137, Exp: 40877500},
	// 				{Level: 138, Exp: 42377500},
	// 				{Level: 139, Exp: 43877500},
	// 				{Level: 140, Exp: 45377500, Reward: 15000},
	// 				{Level: 141, Exp: 47377500},
	// 				{Level: 142, Exp: 49377500},
	// 				{Level: 143, Exp: 51377500},
	// 				{Level: 144, Exp: 53377500},
	// 				{Level: 145, Exp: 55377500},
	// 				{Level: 146, Exp: 57377500},
	// 				{Level: 147, Exp: 59377500},
	// 				{Level: 148, Exp: 61377500},
	// 				{Level: 149, Exp: 63377500},
	// 				{Level: 150, Exp: 65377500, Reward: 20000},
	// 				{Level: 151, Exp: 69377500},
	// 				{Level: 152, Exp: 73377500},
	// 				{Level: 153, Exp: 77377500},
	// 				{Level: 154, Exp: 81377500},
	// 				{Level: 155, Exp: 85377500, Reward: 20000},
	// 				{Level: 156, Exp: 89377500},
	// 				{Level: 157, Exp: 93377500},
	// 				{Level: 158, Exp: 97377500},
	// 				{Level: 159, Exp: 101377500},
	// 				{Level: 160, Exp: 105377500, Reward: 20000},
	// 				{Level: 161, Exp: 109377500},
	// 				{Level: 162, Exp: 113377500},
	// 				{Level: 163, Exp: 117377500},
	// 				{Level: 164, Exp: 121377500},
	// 				{Level: 165, Exp: 125377500, Reward: 20000},
	// 				{Level: 166, Exp: 129377500},
	// 				{Level: 167, Exp: 133377500},
	// 				{Level: 168, Exp: 137377500},
	// 				{Level: 169, Exp: 141377500},
	// 				{Level: 170, Exp: 145377500, Reward: 20000},
	// 				{Level: 171, Exp: 149377500},
	// 				{Level: 172, Exp: 153377500},
	// 				{Level: 173, Exp: 157377500},
	// 				{Level: 174, Exp: 161377500},
	// 				{Level: 175, Exp: 165377500, Reward: 20000},
	// 				{Level: 176, Exp: 169377500},
	// 				{Level: 177, Exp: 173377500},
	// 				{Level: 178, Exp: 177377500},
	// 				{Level: 179, Exp: 181377500},
	// 				{Level: 180, Exp: 185377500, Reward: 20000},
	// 				{Level: 181, Exp: 189377500},
	// 				{Level: 182, Exp: 193377500},
	// 				{Level: 183, Exp: 197377500},
	// 				{Level: 184, Exp: 201377500},
	// 				{Level: 185, Exp: 205377500, Reward: 20000},
	// 				{Level: 186, Exp: 209377500},
	// 				{Level: 187, Exp: 213377500},
	// 				{Level: 188, Exp: 217377500},
	// 				{Level: 189, Exp: 221377500},
	// 				{Level: 190, Exp: 225377500, Reward: 20000},
	// 				{Level: 191, Exp: 229377500},
	// 				{Level: 192, Exp: 233377500},
	// 				{Level: 193, Exp: 237377500},
	// 				{Level: 194, Exp: 241377500},
	// 				{Level: 195, Exp: 245377500, Reward: 20000},
	// 				{Level: 196, Exp: 249377500},
	// 				{Level: 197, Exp: 253377500},
	// 				{Level: 198, Exp: 257377500},
	// 				{Level: 199, Exp: 261377500},
	// 				{Level: 200, Exp: 265377500, Reward: 20000},
	// 				{Level: 201, Exp: 269377500},
	// 				{Level: 202, Exp: 273377500},
	// 				{Level: 203, Exp: 277377500},
	// 				{Level: 204, Exp: 281377500},
	// 				{Level: 205, Exp: 285377500, Reward: 20000},
	// 				{Level: 206, Exp: 289377500},
	// 				{Level: 207, Exp: 293377500},
	// 				{Level: 208, Exp: 297377500},
	// 				{Level: 209, Exp: 301377500},
	// 				{Level: 210, Exp: 305377500, Reward: 20000},
	// 				{Level: 211, Exp: 309377500},
	// 				{Level: 212, Exp: 313377500},
	// 				{Level: 213, Exp: 317377500},
	// 				{Level: 214, Exp: 321377500},
	// 				{Level: 215, Exp: 325377500, Reward: 20000},
	// 				{Level: 216, Exp: 329377500},
	// 				{Level: 217, Exp: 333377500},
	// 				{Level: 218, Exp: 337377500},
	// 				{Level: 219, Exp: 341377500},
	// 				{Level: 220, Exp: 345377500, Reward: 20000},
	// 				{Level: 221, Exp: 349377500},
	// 				{Level: 222, Exp: 353377500},
	// 				{Level: 223, Exp: 357377500},
	// 				{Level: 224, Exp: 361377500},
	// 				{Level: 225, Exp: 365377500, Reward: 20000},
	// 				{Level: 226, Exp: 369377500},
	// 				{Level: 227, Exp: 373377500},
	// 				{Level: 228, Exp: 377377500},
	// 				{Level: 229, Exp: 381377500},
	// 				{Level: 230, Exp: 385377500, Reward: 20000},
	// 				{Level: 231, Exp: 389377500},
	// 				{Level: 232, Exp: 393377500},
	// 				{Level: 233, Exp: 397377500},
	// 				{Level: 234, Exp: 401377500},
	// 				{Level: 235, Exp: 405377500, Reward: 20000},
	// 				{Level: 236, Exp: 409377500},
	// 				{Level: 237, Exp: 413377500},
	// 				{Level: 238, Exp: 417377500},
	// 				{Level: 239, Exp: 421377500},
	// 				{Level: 240, Exp: 425377500, Reward: 20000},
	// 				{Level: 241, Exp: 429377500},
	// 				{Level: 242, Exp: 433377500},
	// 				{Level: 243, Exp: 437377500},
	// 				{Level: 244, Exp: 441377500},
	// 				{Level: 245, Exp: 445377500, Reward: 20000},
	// 				{Level: 246, Exp: 449377500},
	// 				{Level: 247, Exp: 453377500},
	// 				{Level: 248, Exp: 457377500},
	// 				{Level: 249, Exp: 461377500},
	// 				{Level: 250, Exp: 465377500, Reward: 20000},
	// 				{Level: 251, Exp: 469377500},
	// 				{Level: 252, Exp: 473377500},
	// 				{Level: 253, Exp: 477377500},
	// 				{Level: 254, Exp: 481377500},
	// 				{Level: 255, Exp: 485377500, Reward: 20000},
	// 			},
	// 		},
	// 		Week: &protocol.Rewardv2_Week{
	// 			Table: []*protocol.Rewardv2_ExpTable{{Level: 1, Exp: 2000, Reward: 5},
	// 				{Level: 2, Exp: 5000, Reward: 10},
	// 				{Level: 3, Exp: 10000, Reward: 15},
	// 				{Level: 4, Exp: 20000, Reward: 30},
	// 				{Level: 5, Exp: 35000, Reward: 45},
	// 				{Level: 6, Exp: 60000, Reward: 75},
	// 				{Level: 7, Exp: 100000, Reward: 120},
	// 				{Level: 8, Exp: 150000, Reward: 150},
	// 				{Level: 9, Exp: 250000, Reward: 300},
	// 				{Level: 10, Exp: 500000, Reward: 750},
	// 				{Level: 11, Exp: 1000000, Reward: 1500},
	// 				{Level: 12, Exp: 2000000, Reward: 3000},
	// 				{Level: 13, Exp: 3500000, Reward: 4500},
	// 				{Level: 14, Exp: 6000000, Reward: 7500},
	// 				{Level: 15, Exp: 10000000, Reward: 12000},
	// 				{Level: 16, Exp: 15000000, Reward: 15000},
	// 				{Level: 17, Exp: 25000000, Reward: 30000},
	// 				{Level: 18, Exp: 50000000, Reward: 75000},
	// 				{Level: 19, Exp: 100000000, Reward: 150000},
	// 				{Level: 20, Exp: 200000000, Reward: 300000},
	// 				{Level: 21, Exp: 350000000, Reward: 450000},
	// 				{Level: 22, Exp: 600000000, Reward: 750000},
	// 				{Level: 23, Exp: 1000000000, Reward: 1200000},
	// 			},
	// 			Next: 1769479200,
	// 		},
	// 		Exp: []float64{4.6, 0.0032, 4.6},
	// 	},
	// 	Wallet: &protocol.Wallet{
	// 		Currency: currency, //这里记得要改 '{"NONE":0,"CNY":1,"USD":2,"THB":3,"VND_k":4,"MMK":5,"INR":6,"IDR":7,"MYR":8,"VND":9,"IDR_k":10,"kVND":11,"kIDR":12,"JPY":13,"BND":14,"SGD":15,"HKD":16,"PHP":17,"RUB":18,"THB_01":19,"MYR_01":20,"BDT":21,"kMMK":22,"AUD":23,"KZT":24,"CLP":25,"mmyr":26,"LKR":27,"NGN":28,"BRL":29,"TND":30,"MXN":31,"hMMK":32,"PKR":33,"kVND_h":34,"kKHR":35,"kLAK":36,"KRW":37,"EUR":38}'
	// 		Coin:     player.GetBalance(),
	// 		Ratio:    1.0,
	// 		Rate:     0.43,
	// 		Symbol:   game_common_utils.GetCurrencySymbol(player.GetCurrency()), //"₹",
	// 	},
	// }

	// if err := l.send(player, uint32(protocol.ServerToUser_S2U_LOGIN_ACK), rsp); err != nil {
	// 	return err
	// }

	return nil
	// 尝试重连游戏
	//err, ok := l.roomManager.TryReConnectGame(player)
}

func (l *Lobby) send(player types.PlayerImp, cmd uint32, msg proto.Message) error {
	command := &annin_protocol.Command{
		Type: cmd,
	}

	if msg != nil {
		data, err := proto.Marshal(msg)
		if err != nil {
			return err
		}
		command.Data = data
	}

	buff, err := proto.Marshal(command)
	if err != nil {
		return err
	}

	if err := player.SendBinary(buff); err != nil {
		return err
	}
	return nil
}
