package utils

import (
	"math"
	"math/rand/v2"
)

var currencySymbolMap = map[string]string{
	"USD": "$",   // 美元
	"EUR": "€",   // 欧元
	"GBP": "£",   // 英镑
	"INR": "₹",   // 印度卢比
	"JPY": "¥",   // 日元
	"CNY": "¥",   // 人民币
	"KRW": "₩",   // 韩元
	"RUB": "₽",   // 俄罗斯卢布
	"BRL": "R$",  // 巴西雷亚尔
	"AUD": "A$",  // 澳元
	"CAD": "C$",  // 加元
	"NZD": "NZ$", // 新西兰元
	"SGD": "S$",  // 新加坡元
	"HKD": "HK$", // 港币
	"TWD": "NT$", // 新台币
	"THB": "฿",   // 泰铢
	"PHP": "₱",   // 菲律宾比索
	"IDR": "Rp",  // 印尼卢比
	"MYR": "RM",  // 马来西亚令吉
	"VND": "₫",   // 越南盾

	"CHF": "CHF", // 瑞士法郎
	"SEK": "kr",  // 瑞典克朗
	"NOK": "kr",  // 挪威克朗
	"DKK": "kr",  // 丹麦克朗
	"PLN": "zł",  // 波兰兹罗提
	"CZK": "Kč",  // 捷克克朗
	"HUF": "Ft",  // 匈牙利福林

	"SAR": "﷼",   // 沙特里亚尔
	"AED": "د.إ", // 阿联酋迪拉姆
	"TRY": "₺",   // 土耳其里拉
	"EGP": "£",   // 埃及镑

	"ZAR": "R",   // 南非兰特
	"NGN": "₦",   // 尼日利亚奈拉
	"KES": "KSh", // 肯尼亚先令

	"ARS": "$",  // 阿根廷比索
	"MXN": "$",  // 墨西哥比索
	"CLP": "$",  // 智利比索
	"PEN": "S/", // 秘鲁新索尔
	"COP": "$",  // 哥伦比亚比索
	"PKR": "₨",  // 巴基斯坦卢比
}

func GetCurrencySymbol(currencyCode string) string {
	if symbol, ok := currencySymbolMap[currencyCode]; ok {
		return symbol
	}
	return "" // 如果未找到对应的符号，返回空字符串
}

// 随机获取一个货币代码,如INR，等
func GetRandomCurrencyCode() string {
	keys := []string{
		"USD", // 美元
		"EUR", // 欧元
		"GBP", // 英镑
		"INR", // 卢比
		"JPY", // 日元
		"KRW", // 韩元
		"AUD", // 澳元
		"CAD", // 加元
		"SGD", // 新加坡
		"HKD", // 港币
	}

	return keys[rand.IntN(len(keys))]
}

// 保证机器人下注金额是cashUnit的整数倍,cashUnit保证两位小数
func AdjustBotBetNumStrict(botBetNum float64, cashUnit float64) float64 {
	// 将金额转换为整数（乘以100避免浮点精度问题）
	botBetNumCents := int64(math.Round(botBetNum * 100))
	cashUnitCents := int64(math.Round(cashUnit * 100))

	// 计算倍数
	multiple := botBetNumCents / cashUnitCents

	// 如果倍数小于1，则使用最小倍数1
	if multiple < 1 {
		multiple = 1
	}

	// 计算调整后的金额（以分为单位）
	adjustedBetNumCents := multiple * cashUnitCents

	// 转换回元，保留两位小数
	return float64(adjustedBetNumCents) / 100
}
