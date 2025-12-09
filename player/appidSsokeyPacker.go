package player

import (
	"bytes"
	"encoding/hex"
	"fmt"

	"github.com/golang/snappy"
)

// ==================有些游戏的在换token过程，只提交ssoKey，需要将token和ssokey打包==============================
func EncodedAppidSsokey(appid, ssokey string) (string, error) {
	// 1. 检查第一个字符串长度
	if len(appid) > 255 {
		return "", fmt.Errorf("first string too long (max 255 bytes)")
	}

	// 2. 构建二进制数据：[1字节长度][字符串1][字符串2]
	buf := new(bytes.Buffer)
	buf.WriteByte(byte(len(appid)))
	buf.WriteString(appid)
	buf.WriteString(ssokey)

	// 3. 压缩数据
	// 3. 使用Snappy压缩数据
	compressed := snappy.Encode(nil, buf.Bytes())

	// 4. 转为小写hex
	return hex.EncodeToString(compressed), nil
}

// 从key中解压出来
func DecodeAppidSsokey(encodedStr string) (string, string, error) {
	// 1. 从hex解码
	compressed, err := hex.DecodeString(encodedStr)
	if err != nil {
		return "", "", fmt.Errorf("hex decode error: %v", err)
	}

	// 2. 解压数据
	decompressed, err := snappy.Decode(nil, compressed)
	if err != nil {
		return "", "", fmt.Errorf("decompression error: %v", err)
	}

	// 3. 解析原始数据
	if len(decompressed) < 1 {
		return "", "", fmt.Errorf("invalid data format")
	}

	// 第一个字节是str1的长度
	str1Len := int(decompressed[0])
	if len(decompressed) < 1+str1Len {
		return "", "", fmt.Errorf("data too short for str1")
	}

	// 分割字符串
	str1 := string(decompressed[1 : 1+str1Len])
	str2 := string(decompressed[1+str1Len:])

	return str1, str2, nil
}

// =========================================================================================================
