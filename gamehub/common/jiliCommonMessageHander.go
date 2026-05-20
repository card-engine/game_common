package common

import (
	"github.com/card-engine/game_common/gamehub/jili/annin_protocol"
	"github.com/card-engine/game_common/gamehub/types"
	"google.golang.org/protobuf/proto"
)

type JiliCommonMessageHandler struct {
}

// jili公共的协议

func (h *JiliCommonMessageHandler) OnMessage(player types.PlayerImp, data interface{}) error {
	command := data.(*annin_protocol.Command)
	// ping
	if command.Type == uint32(annin_protocol.UserToServer_U2S_HEART_CHECK_REQ) {
		return h.send(player, uint32(annin_protocol.ServerToUser_S2U_HEART_CHECK_ACK), nil)
	} else if command.Type == uint32(annin_protocol.UserToServer_U2S_CONFIG_INFO) { //120
		return h.onConfigInfo(player)
	} else if command.Type == uint32(annin_protocol.UserToServer_U2S_VIP_INFO) { //92
		return h.onVipInfo(player)
	} else if command.Type == uint32(annin_protocol.UserToServer_U2S_PROMOTION_INFO) { //95
		return h.onPromotionInfo(player)
	} else if command.Type == uint32(annin_protocol.UserToServer_U2S_MAIL_INFO) { //115
		return h.onMailInfo(player)
	} else if command.Type == uint32(annin_protocol.UserToServer_U2S_CARD_INFO) { //82
		return h.onCardInfo(player)
	} else if command.Type == uint32(annin_protocol.UserToServer_U2S_LOGIN_REQ) { // 0 (主动请求的登陆协议)
		//return l.OnLogin(player)
	}
	return nil
}

func (h *JiliCommonMessageHandler) onConfigInfo(player types.PlayerImp) error {
	rsp := &annin_protocol.AgentConfig{
		BetList: []float64{0.1, 0.2, 0.3, 0.4, 0.5, 0.6, 0.7, 0.8, 0.9, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 20, 30, 40, 50, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 20, 30, 40, 50, 60, 70, 80, 90, 100},
	}

	if err := h.send(player, uint32(annin_protocol.ServerToUser_S2U_CONFIG_INFO), rsp); err != nil {
		return err
	}

	return nil
}

func (h *JiliCommonMessageHandler) onVipInfo(player types.PlayerImp) error {
	rsp := &annin_protocol.Vip{
		Exp:      0.8,
		Treasure: &annin_protocol.Vip_Treasure{Exp: 0.8},
	}
	if err := h.send(player, uint32(annin_protocol.ServerToUser_S2U_VIP_INFO), rsp); err != nil {
		return err
	}

	return nil
}

func (h *JiliCommonMessageHandler) onPromotionInfo(player types.PlayerImp) error {
	rsp := &annin_protocol.Promotion{
		Enabled: true,
		List: []*annin_protocol.Promotion_Info{
			{
				Id:     32,
				Labels: []uint32{1},
				Dau:    25013,
			},
		},
	}

	if err := h.send(player, uint32(annin_protocol.ServerToUser_S2U_PROMOTION_INFO), rsp); err != nil {
		return err
	}

	return nil
}

func (h *JiliCommonMessageHandler) onMailInfo(player types.PlayerImp) error {
	return nil
}

func (h *JiliCommonMessageHandler) onCardInfo(player types.PlayerImp) error {
	rsp := &annin_protocol.CardUseAck{}
	if err := h.send(player, uint32(annin_protocol.ServerToUser_S2U_CARD_INFO), rsp); err != nil {
		return err
	}

	return nil
}

func (h *JiliCommonMessageHandler) send(player types.PlayerImp, cmd uint32, msg proto.Message) error {
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
