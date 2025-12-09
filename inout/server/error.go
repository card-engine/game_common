package server

var ErrorStrMap map[string]map[string]string

func init() {
	ErrorStrMap = map[string]map[string]string{
		"en": {"INSUFFICIENT_BALANCE": "Insufficient balance"},
		"zh": {"INSUFFICIENT_BALANCE": "余额不足"},
		"ja": {"INSUFFICIENT_BALANCE": "残高不足"},
		"ko": {"INSUFFICIENT_BALANCE": "잔액 부족"},
		"es": {"INSUFFICIENT_BALANCE": "Saldo insuficiente"},
		"fr": {"INSUFFICIENT_BALANCE": "Solde insuffisant"},
	}
}

// GetErrorMsg 获取错误信息
func GetErrorMsg(code, lang string) string {
	// 检查语言是否存在
	if langMap, exists := ErrorStrMap[lang]; exists {
		if msg, exists := langMap[code]; exists {
			return msg
		}
	}

	// 回退到英文
	if enMsg, exists := ErrorStrMap["en"][code]; exists {
		return enMsg
	}

	return "Unknown error: " + code
}
