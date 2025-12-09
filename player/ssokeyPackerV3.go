package player

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	minitoken "github.com/card-engine/game_common/player/mini_token"
	"google.golang.org/protobuf/proto"
)

// 验证 HMAC 的函数
func VerifyHMACSHA256(text, key string, expectedHMAC string) bool {
	result := HMACSHA256Encrypt(text, key)
	return result == expectedHMAC
}

// HMACSHA256Encrypt 使用 HMAC-SHA256 加密文本，直接返回前6后6的hex字符
func HMACSHA256Encrypt(text, key string) string {
	// 将字符串转换为字节切片
	textBytes := []byte(text)
	keyBytes := []byte(key)

	// 创建 HMAC 对象
	h := hmac.New(sha256.New, keyBytes)

	// 写入数据
	h.Write(textBytes)

	// 计算 HMAC
	hmacResult := h.Sum(nil)

	// 转换为hex格式
	hexResult := hex.EncodeToString(hmacResult)

	// 直接返回前6和后6字符
	return hexResult[:6] + hexResult[len(hexResult)-6:]
}

const (
	SSOKeyV3SignKey = "RandomStringWithSpecialChars"
)

func generateTokenToString(params *minitoken.TokenPayload) string {
	var builder strings.Builder

	// 按 protobuf 字段编号顺序
	builder.WriteString(params.AppId)
	builder.WriteString(params.PlayerId)
	builder.WriteString(params.GameBrand)
	builder.WriteString(params.GameId)
	builder.WriteString(fmt.Sprintf("%d", params.Expire))

	return builder.String()
}

func EncodedSSOKeyV3(params *minitoken.TokenPayload) (string, error) {
	params.Expire = time.Now().Add(24 * time.Hour).Unix()
	params.Sign = HMACSHA256Encrypt(generateTokenToString(params), SSOKeyV3SignKey)
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

	// 验证签名
	if !VerifyHMACSHA256(generateTokenToString(minitoken), SSOKeyV3SignKey, minitoken.Sign) {
		return nil, errors.New("invalid sign")
	}

	// 验证过期时间
	if minitoken.Expire < time.Now().Unix() {
		return nil, errors.New("expired")
	}

	return minitoken, nil
}

// 判断一个游戏有没有多地登陆
func IsMultiLogin(tokenInfo *minitoken.TokenPayload, playerInfo *PlayerInfo) bool {
	if playerInfo.Brand != tokenInfo.GameBrand {
		return true
	}

	if playerInfo.GameID != tokenInfo.GameId {
		return true
	}
	return false
}
