package player

import (
	"encoding/hex"

	minitoken "cn.qingdou.server/game_common/player/mini_token"
	"google.golang.org/protobuf/proto"
)

func EncodedSSOKeyV3(params *minitoken.TokenPayload) (string, error) {
	protoBytes, err := proto.Marshal(params)
	if err != nil {
		return "", err
	}
	hexStr := hex.EncodeToString(protoBytes)
	return hexStr, nil
}

func DecodedSSOKeyV3(encodedSSOKey string) (*minitoken.TokenPayload, error) {
	protoBytes, err := hex.DecodeString(encodedSSOKey)
	if err != nil {
		return nil, err
	}
	minitoken := &minitoken.TokenPayload{}
	err = proto.Unmarshal(protoBytes, minitoken)
	if err != nil {
		return nil, err
	}
	return minitoken, nil
}
