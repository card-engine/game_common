package player

import (
	"fmt"
	"testing"
)

func TestPackAppidSsokey(t *testing.T) {
	appid := "10001"
	ssokey := "12345678901234567890123456789012"
	encoded, err := EncodedAppidSsokey(appid, ssokey)
	if err != nil {
		t.Fatalf("EncodedAppidSsokey failed: %v", err)
	}
	decodedAppid, decodedSsokey, err := DecodeAppidSsokey(encoded)
	if err != nil {
		t.Fatalf("DecodeAppidSsokey failed: %v", err)
	}
	if decodedAppid != appid {
		t.Fatalf("Decoded appid %v does not match original %v", decodedAppid, appid)
	}
	if decodedSsokey != ssokey {
		t.Fatalf("Decoded ssokey %s does not match original %s", decodedSsokey, ssokey)
	}

	fmt.Println(encoded)
	fmt.Println(decodedAppid)
	fmt.Println(decodedSsokey)
}

func TestSsokeyPackerV2(t *testing.T) {
	// 测试结构体编码解码
	params := SSOKeyParams{
		AppID:  "10001",
		GameID: "334",
		SSOKey: "12345678901234567890123456789012",
	}

	encoded, err := EncodedSSOKeyParams(params)
	if err != nil {
		t.Fatalf("EncodedSSOKeyParams failed: %v", err)
	}

	decoded, err := DecodeSSOKeyParams(encoded)
	if err != nil {
		t.Fatalf("DecodeSSOKeyParams failed: %v", err)
	}

	if decoded.AppID != params.AppID {
		t.Fatalf("Decoded AppID %v does not match original %v", decoded.AppID, params.AppID)
	}
	if decoded.GameID != params.GameID {
		t.Fatalf("Decoded GameID %v does not match original %v", decoded.GameID, params.GameID)
	}
	if decoded.SSOKey != params.SSOKey {
		t.Fatalf("Decoded SSOKey %s does not match original %s", decoded.SSOKey, params.SSOKey)
	}

	fmt.Printf("V2 结构体测试 - 编码: %s\n", encoded)
	fmt.Printf("解码结果 - AppID: %s, GameID: %s, SSOKey: %s\n", decoded.AppID, decoded.GameID, decoded.SSOKey)
}

func TestSsokeyPackerV2ArbitraryParams(t *testing.T) {
	// 测试任意参数编码解码
	params := []string{"10001", "334", "12345678901234567890123456789012"}

	encoded, err := EncodedSSOKeyV2(params...)
	if err != nil {
		t.Fatalf("EncodedSSOKeyV2 failed: %v", err)
	}

	decoded, err := DecodeSSOKeyV2(encoded)
	if err != nil {
		t.Fatalf("DecodeSSOKeyV2 failed: %v", err)
	}

	if len(decoded) != len(params) {
		t.Fatalf("Decoded params count %d does not match original %d", len(decoded), len(params))
	}

	for i, param := range params {
		if decoded[i] != param {
			t.Fatalf("Decoded param %d %s does not match original %s", i, decoded[i], param)
		}
	}

	fmt.Printf("V2 任意参数测试 - 编码: %s\n", encoded)
	fmt.Printf("解码结果: %v\n", decoded)
}

func TestSsokeyPackerV2BackwardCompatibility(t *testing.T) {
	// 测试向后兼容性：旧版本数据用新版本解码
	oldParams := SSOKeyParams{
		AppID:  "10001",
		SSOKey: "12345678901234567890123456789012",
		// GameID 为空
	}

	encoded, err := EncodedSSOKeyParams(oldParams)
	if err != nil {
		t.Fatalf("EncodedSSOKeyParams failed: %v", err)
	}

	decoded, err := DecodeSSOKeyParams(encoded)
	if err != nil {
		t.Fatalf("DecodeSSOKeyParams failed: %v", err)
	}

	if decoded.AppID != oldParams.AppID {
		t.Fatalf("Decoded AppID %v does not match original %v", decoded.AppID, oldParams.AppID)
	}
	if decoded.SSOKey != oldParams.SSOKey {
		t.Fatalf("Decoded SSOKey %s does not match original %s", decoded.SSOKey, oldParams.SSOKey)
	}
	if decoded.GameID != "" {
		t.Fatalf("Decoded GameID should be empty, got %s", decoded.GameID)
	}

	fmt.Printf("V2 向后兼容测试 - 编码: %s\n", encoded)
	fmt.Printf("解码结果 - AppID: %s, GameID: %s, SSOKey: %s\n", decoded.AppID, decoded.GameID, decoded.SSOKey)
}
