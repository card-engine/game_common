package player

import (
	"fmt"
	"testing"

	minitoken "cn.qingdou.server/game_common/player/mini_token"
)

func TestEncodedAndDecodedSSOKeyV3(t *testing.T) {
	// 构造一个 TokenPayload 测试数据
	payload := &minitoken.TokenPayload{
		PlayerId:  "11254455",
		AppId:     "11254aa5dadsdadfasdfasdfasdfasdfs5",
		GameBrand: "jili",
		GameId:    "1125",
		// 其他字段根据你的结构体补充
	}

	// 编码
	encoded, err := EncodedSSOKeyV3(payload)
	if err != nil {
		t.Errorf("EncodedSSOKeyV3 should not return error: %v", err)
	}
	fmt.Printf("t: %v, encoded: %s\n", t, encoded)

	// 解码
	decoded, err := DecodedSSOKeyV3(encoded)
	if err != nil {
		t.Errorf("DecodedSSOKeyV3 should not return error: %v", err)
	}
	fmt.Printf("t: %v, decoded: %v\n", t, decoded)
}
