package inout

import (
	"encoding/json"
	"fmt"

	"github.com/card-engine/game_common/gamehub/types"
)

const DefaultMsgId = "42"

func SendErrorMessage(player types.PlayerImp, msgId, errMsg string) error {
	if msgId == "" {
		msgId = DefaultMsgId
	}
	errorMsg := fmt.Sprintf(`%s[{"error":{"message":"%s"}}]`, msgId, errMsg)
	return player.SendString(errorMsg)
}

// 发送与服务业务相关的游戏数据
func SendGameServiceData(player types.PlayerImp, msgId, action string, response interface{}) error {
	responseData, err := json.Marshal([]interface{}{action, response})
	if err != nil {
		return err
	}
	if msgId == "" {
		msgId = DefaultMsgId
	}
	// 构造自定义格式响应: 数字[响应数据]
	responseMsg := fmt.Sprintf("%s%s", msgId, string(responseData))
	return player.SendString(responseMsg)
}

// 不带action的游戏数据响应
func SendData(player types.PlayerImp, msgId string, response interface{}) error {
	responseData, err := json.Marshal([]interface{}{response})
	if err != nil {
		return err
	}
	if msgId == "" {
		msgId = DefaultMsgId
	}
	// 构造自定义格式响应: 数字[响应数据]
	responseMsg := fmt.Sprintf("%s%s", msgId, string(responseData))
	return player.SendString(responseMsg)
}
