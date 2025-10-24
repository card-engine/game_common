package player

import (
	"bytes"
	"encoding/hex"
	"fmt"

	"github.com/golang/snappy"
)

// SSOKeyParams 定义SSO Key参数结构
type SSOKeyParams struct {
	AppID  string `json:"appId"`
	GameID string `json:"gameId"`
	SSOKey string `json:"ssoKey"`
}

// EncodedSSOKeyParams 使用结构体编码SSO Key参数
func EncodedSSOKeyParams(params SSOKeyParams) (string, error) {
	// 构建参数数组，按固定顺序：AppID, GameID, SSOKey
	// 注意：即使字段为空也要保留位置，确保向后兼容
	paramList := []string{
		params.AppID,
		params.GameID,
		params.SSOKey,
	}

	return EncodedSSOKeyV2(paramList...)
}

// EncodedSSOKeyV2 支持任意参数的编码方法
func EncodedSSOKeyV2(params ...string) (string, error) {
	if len(params) == 0 {
		return "", fmt.Errorf("至少需要一个参数")
	}

	// 检查所有字符串长度
	for i, param := range params {
		if len(param) > 255 {
			return "", fmt.Errorf("参数 %d 太长 (最大 255 字节)", i)
		}
	}

	// 构建二进制数据：[参数数量][参数1长度][参数1][参数2长度][参数2]...
	buf := new(bytes.Buffer)

	// 写入参数数量
	buf.WriteByte(byte(len(params)))

	// 写入每个参数的长度和内容
	for _, param := range params {
		buf.WriteByte(byte(len(param)))
		buf.WriteString(param)
	}

	// 使用Snappy压缩数据
	compressed := snappy.Encode(nil, buf.Bytes())

	// 转为小写hex
	return hex.EncodeToString(compressed), nil
}

// DecodeSSOKeyParams 解码为结构体
func DecodeSSOKeyParams(encodedStr string) (SSOKeyParams, error) {
	params, err := DecodeSSOKeyV2(encodedStr)
	if err != nil {
		return SSOKeyParams{}, err
	}

	result := SSOKeyParams{}

	// 按固定顺序解析参数：AppID, GameID, SSOKey
	// 如果参数数量不足，只解析存在的参数，确保向后兼容
	paramIndex := 0

	if paramIndex < len(params) {
		result.AppID = params[paramIndex]
		paramIndex++
	}

	if paramIndex < len(params) {
		result.GameID = params[paramIndex]
		paramIndex++
	}

	if paramIndex < len(params) {
		result.SSOKey = params[paramIndex]
		paramIndex++
	}

	// 如果还有更多参数，说明是更新版本的数据，忽略额外字段
	// 这确保了向后兼容性

	return result, nil
}

// DecodeSSOKeyV2 解码为字符串数组
func DecodeSSOKeyV2(encodedStr string) ([]string, error) {
	// 1. 从hex解码
	compressed, err := hex.DecodeString(encodedStr)
	if err != nil {
		return nil, fmt.Errorf("hex decode error: %v", err)
	}

	// 2. 解压数据
	decompressed, err := snappy.Decode(nil, compressed)
	if err != nil {
		return nil, fmt.Errorf("decompression error: %v", err)
	}

	// 3. 解析原始数据
	if len(decompressed) < 1 {
		return nil, fmt.Errorf("invalid data format")
	}

	// 检查是否为新的多参数格式
	paramCount := int(decompressed[0])

	// 如果参数数量为1，说明是旧格式（只有appid长度）
	if paramCount == 1 && len(decompressed) >= 2 {
		// 旧格式：[1字节长度][字符串1][字符串2]
		str1Len := int(decompressed[0])
		if len(decompressed) < 1+str1Len {
			return nil, fmt.Errorf("data too short for str1")
		}

		str1 := string(decompressed[1 : 1+str1Len])
		str2 := string(decompressed[1+str1Len:])
		return []string{str1, str2}, nil
	}

	// 新格式：[参数数量][参数1长度][参数1][参数2长度][参数2]...
	if paramCount > 10 {
		return nil, fmt.Errorf("参数数量过多")
	}

	params := make([]string, 0, paramCount)
	offset := 1

	for i := 0; i < paramCount; i++ {
		if offset >= len(decompressed) {
			return nil, fmt.Errorf("数据不完整，缺少参数 %d 的长度信息", i+1)
		}

		paramLen := int(decompressed[offset])
		offset++

		if offset+paramLen > len(decompressed) {
			return nil, fmt.Errorf("数据不完整，缺少参数 %d 的内容", i+1)
		}

		param := string(decompressed[offset : offset+paramLen])
		params = append(params, param)
		offset += paramLen
	}

	return params, nil
}
