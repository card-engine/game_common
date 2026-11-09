package utils

import (
	"math"
	"math/rand/v2"
	"sort"
	"strings"
)

// CurrencyInfo 平台货币目录条目（ISO 4217 → 符号 / 中文名）。
type CurrencyInfo struct {
	Code   string `json:"code"`
	Symbol string `json:"symbol"`
	Name   string `json:"name,omitempty"` // 展示名（与原 currency_config.name 对齐，中文）
}

// currencyCatalog 权威目录；symbol / name 查找与 ListCurrencies 共用此表。
// Name 优先沿用原 admin currency_config 中文名；其余为常用中文译名。
var currencyCatalog = []CurrencyInfo{
	{Code: "AED", Symbol: "د.إ", Name: "阿联酋迪拉姆"},
	{Code: "AFN", Symbol: "؋", Name: "阿富汗尼"},
	{Code: "ALL", Symbol: "L", Name: "阿尔巴尼亚列克"},
	{Code: "AMD", Symbol: "֏", Name: "亚美尼亚德拉姆"},
	{Code: "ANG", Symbol: "ƒ", Name: "荷属安的列斯盾"},
	{Code: "AOA", Symbol: "Kz", Name: "安哥拉宽扎"},
	{Code: "ARS", Symbol: "$", Name: "阿根廷比索"},
	{Code: "AUD", Symbol: "A$", Name: "澳元"},
	{Code: "AWG", Symbol: "ƒ", Name: "阿鲁巴弗罗林"},
	{Code: "AZN", Symbol: "₼", Name: "阿塞拜疆马纳特"},
	{Code: "BAM", Symbol: "KM", Name: "波黑可兑换马克"},
	{Code: "BBD", Symbol: "$", Name: "巴巴多斯元"},
	{Code: "BDT", Symbol: "৳", Name: "孟加拉塔卡"},
	{Code: "BGN", Symbol: "лв", Name: "保加利亚列弗"},
	{Code: "BHD", Symbol: ".د.ب", Name: "巴林第纳尔"},
	{Code: "BIF", Symbol: "FBu", Name: "布隆迪法郎"},
	{Code: "BMD", Symbol: "$", Name: "百慕大元"},
	{Code: "BND", Symbol: "$", Name: "文莱元"},
	{Code: "BOB", Symbol: "Bs.", Name: "玻利维亚诺"},
	{Code: "BRL", Symbol: "R$", Name: "巴西雷亚尔"},
	{Code: "BSD", Symbol: "$", Name: "巴哈马元"},
	{Code: "BTN", Symbol: "Nu.", Name: "不丹努尔特鲁姆"},
	{Code: "BWP", Symbol: "P", Name: "博茨瓦纳普拉"},
	{Code: "BYN", Symbol: "Br", Name: "白俄罗斯卢布"},
	{Code: "BZD", Symbol: "BZ$", Name: "伯利兹元"},
	{Code: "CAD", Symbol: "C$", Name: "加元"},
	{Code: "CDF", Symbol: "FC", Name: "刚果法郎"},
	{Code: "CHF", Symbol: "CHF", Name: "瑞士法郎"},
	{Code: "CLP", Symbol: "$", Name: "智利比索"},
	{Code: "CNY", Symbol: "¥", Name: "人民币"},
	{Code: "COP", Symbol: "$", Name: "哥伦比亚比索"},
	{Code: "CRC", Symbol: "₡", Name: "哥斯达黎加科朗"},
	{Code: "CUP", Symbol: "$", Name: "古巴比索"},
	{Code: "CVE", Symbol: "$", Name: "佛得角埃斯库多"},
	{Code: "CZK", Symbol: "Kč", Name: "捷克克朗"},
	{Code: "DJF", Symbol: "Fdj", Name: "吉布提法郎"},
	{Code: "DKK", Symbol: "kr", Name: "丹麦克朗"},
	{Code: "DOP", Symbol: "RD$", Name: "多米尼加比索"},
	{Code: "DZD", Symbol: "د.ج", Name: "阿尔及利亚第纳尔"},
	{Code: "EGP", Symbol: "E£", Name: "埃及镑"},
	{Code: "ERN", Symbol: "Nfk", Name: "厄立特里亚纳克法"},
	{Code: "ETB", Symbol: "Br", Name: "埃塞俄比亚比尔"},
	{Code: "EUR", Symbol: "€", Name: "欧元"},
	{Code: "FJD", Symbol: "FJ$", Name: "斐济元"},
	{Code: "FKP", Symbol: "£", Name: "福克兰群岛镑"},
	{Code: "GBP", Symbol: "£", Name: "英镑"},
	{Code: "GEL", Symbol: "₾", Name: "格鲁吉亚拉里"},
	{Code: "GHS", Symbol: "GH₵", Name: "加纳塞地"},
	{Code: "GIP", Symbol: "£", Name: "直布罗陀镑"},
	{Code: "GMD", Symbol: "D", Name: "冈比亚达拉西"},
	{Code: "GNF", Symbol: "FG", Name: "几内亚法郎"},
	{Code: "GTQ", Symbol: "Q", Name: "危地马拉格查尔"},
	{Code: "GYD", Symbol: "$", Name: "圭亚那元"},
	{Code: "HKD", Symbol: "HK$", Name: "港币"},
	{Code: "HNL", Symbol: "L", Name: "洪都拉斯伦皮拉"},
	{Code: "HRK", Symbol: "kn", Name: "克罗地亚库纳"},
	{Code: "HTG", Symbol: "G", Name: "海地古德"},
	{Code: "HUF", Symbol: "Ft", Name: "匈牙利福林"},
	{Code: "IDR", Symbol: "Rp", Name: "印尼卢比"},
	{Code: "ILS", Symbol: "₪", Name: "以色列新谢克尔"},
	{Code: "INR", Symbol: "₹", Name: "印度卢比"},
	{Code: "IQD", Symbol: "ع.د", Name: "伊拉克第纳尔"},
	{Code: "IRR", Symbol: "﷼", Name: "伊朗里亚尔"},
	{Code: "ISK", Symbol: "kr", Name: "冰岛克朗"},
	{Code: "JMD", Symbol: "J$", Name: "牙买加元"},
	{Code: "JOD", Symbol: "JD", Name: "约旦第纳尔"},
	{Code: "JPY", Symbol: "¥", Name: "日元"},
	{Code: "KES", Symbol: "KSh", Name: "肯尼亚先令"},
	{Code: "KGS", Symbol: "с", Name: "吉尔吉斯斯坦索姆"},
	{Code: "KHR", Symbol: "៛", Name: "柬埔寨瑞尔"},
	{Code: "KMF", Symbol: "CF", Name: "科摩罗法郎"},
	{Code: "KPW", Symbol: "₩", Name: "朝鲜圆"},
	{Code: "KRW", Symbol: "₩", Name: "韩元"},
	{Code: "KWD", Symbol: "KD", Name: "科威特第纳尔"},
	{Code: "KYD", Symbol: "CI$", Name: "开曼群岛元"},
	{Code: "KZT", Symbol: "₸", Name: "哈萨克斯坦坚戈"},
	{Code: "LAK", Symbol: "₭", Name: "老挝基普"},
	{Code: "LBP", Symbol: "L£", Name: "黎巴嫩镑"},
	{Code: "LKR", Symbol: "Rs", Name: "斯里兰卡卢比"},
	{Code: "LRD", Symbol: "L$", Name: "利比里亚元"},
	{Code: "LSL", Symbol: "L", Name: "莱索托洛蒂"},
	{Code: "LYD", Symbol: "LD", Name: "利比亚第纳尔"},
	{Code: "MAD", Symbol: "DH", Name: "摩洛哥迪拉姆"},
	{Code: "MDL", Symbol: "L", Name: "摩尔多瓦列伊"},
	{Code: "MGA", Symbol: "Ar", Name: "马达加斯加阿里亚里"},
	{Code: "MKD", Symbol: "ден", Name: "北马其顿第纳尔"},
	{Code: "MMK", Symbol: "K", Name: "缅甸元"},
	{Code: "MNT", Symbol: "₮", Name: "蒙古图格里克"},
	{Code: "MOP", Symbol: "MOP$", Name: "澳门元"},
	{Code: "MRU", Symbol: "UM", Name: "毛里塔尼亚乌吉亚"},
	{Code: "MUR", Symbol: "₨", Name: "毛里求斯卢比"},
	{Code: "MVR", Symbol: "Rf", Name: "马尔代夫拉菲亚"},
	{Code: "MWK", Symbol: "MK", Name: "马拉维克瓦查"},
	{Code: "MXN", Symbol: "$", Name: "墨西哥比索"},
	{Code: "MYR", Symbol: "RM", Name: "马来西亚令吉"},
	{Code: "MZN", Symbol: "MT", Name: "莫桑比克梅蒂卡尔"},
	{Code: "NAD", Symbol: "N$", Name: "纳米比亚元"},
	{Code: "NGN", Symbol: "₦", Name: "尼日利亚奈拉"},
	{Code: "NIO", Symbol: "C$", Name: "尼加拉瓜科多巴"},
	{Code: "NOK", Symbol: "kr", Name: "挪威克朗"},
	{Code: "NPR", Symbol: "Rs", Name: "尼泊尔卢比"},
	{Code: "NZD", Symbol: "NZ$", Name: "新西兰元"},
	{Code: "OMR", Symbol: "﷼", Name: "阿曼里亚尔"},
	{Code: "PAB", Symbol: "B/.", Name: "巴拿马巴波亚"},
	{Code: "PEN", Symbol: "S/", Name: "秘鲁新索尔"},
	{Code: "PGK", Symbol: "K", Name: "巴布亚新几内亚基那"},
	{Code: "PHP", Symbol: "₱", Name: "菲律宾比索"},
	{Code: "PKR", Symbol: "₨", Name: "巴基斯坦卢比"},
	{Code: "PLN", Symbol: "zł", Name: "波兰兹罗提"},
	{Code: "PYG", Symbol: "₲", Name: "巴拉圭瓜拉尼"},
	{Code: "QAR", Symbol: "﷼", Name: "卡塔尔里亚尔"},
	{Code: "RON", Symbol: "L", Name: "罗马尼亚列伊"},
	{Code: "RSD", Symbol: "din", Name: "塞尔维亚第纳尔"},
	{Code: "RUB", Symbol: "₽", Name: "俄罗斯卢布"},
	{Code: "RWF", Symbol: "FRw", Name: "卢旺达法郎"},
	{Code: "SAR", Symbol: "﷼", Name: "沙特里亚尔"},
	{Code: "SBD", Symbol: "SI$", Name: "所罗门群岛元"},
	{Code: "SCR", Symbol: "₨", Name: "塞舌尔卢比"},
	{Code: "SDG", Symbol: "SDG", Name: "苏丹镑"},
	{Code: "SEK", Symbol: "kr", Name: "瑞典克朗"},
	{Code: "SGD", Symbol: "S$", Name: "新加坡元"},
	{Code: "SHP", Symbol: "£", Name: "圣赫勒拿镑"},
	{Code: "SLE", Symbol: "Le", Name: "塞拉利昂利昂"},
	{Code: "SOS", Symbol: "Sh", Name: "索马里先令"},
	{Code: "SRD", Symbol: "$", Name: "苏里南元"},
	{Code: "SSP", Symbol: "SS£", Name: "南苏丹镑"},
	{Code: "STN", Symbol: "Db", Name: "圣多美和普林西比多布拉"},
	{Code: "SYP", Symbol: "£", Name: "叙利亚镑"},
	{Code: "SZL", Symbol: "L", Name: "斯威士兰里兰吉尼"},
	{Code: "THB", Symbol: "฿", Name: "泰铢"},
	{Code: "TJS", Symbol: "SM", Name: "塔吉克斯坦索莫尼"},
	{Code: "TMT", Symbol: "m", Name: "土库曼斯坦马纳特"},
	{Code: "TND", Symbol: "DT", Name: "突尼斯第纳尔"},
	{Code: "TOP", Symbol: "T$", Name: "汤加潘加"},
	{Code: "TRY", Symbol: "₺", Name: "土耳其里拉"},
	{Code: "TTD", Symbol: "TT$", Name: "特立尼达和多巴哥元"},
	{Code: "TWD", Symbol: "NT$", Name: "新台币"},
	{Code: "TZS", Symbol: "TSh", Name: "坦桑尼亚先令"},
	{Code: "UAH", Symbol: "₴", Name: "乌克兰格里夫纳"},
	{Code: "UGX", Symbol: "USh", Name: "乌干达先令"},
	{Code: "USD", Symbol: "$", Name: "美元"},
	{Code: "UYU", Symbol: "$U", Name: "乌拉圭比索"},
	{Code: "UZS", Symbol: "so'm", Name: "乌兹别克斯坦苏姆"},
	{Code: "VES", Symbol: "Bs.S", Name: "委内瑞拉玻利瓦尔"},
	{Code: "VND", Symbol: "₫", Name: "越南盾"},
	{Code: "VUV", Symbol: "VT", Name: "瓦努阿图瓦图"},
	{Code: "WST", Symbol: "WS$", Name: "萨摩亚塔拉"},
	{Code: "XAF", Symbol: "FCFA", Name: "中非法郎"},
	{Code: "XCD", Symbol: "EC$", Name: "东加勒比元"},
	{Code: "XOF", Symbol: "CFA", Name: "西非法郎"},
	{Code: "XPF", Symbol: "₣", Name: "太平洋法郎"},
	{Code: "YER", Symbol: "﷼", Name: "也门里亚尔"},
	{Code: "ZAR", Symbol: "R", Name: "南非兰特"},
	{Code: "ZMW", Symbol: "ZK", Name: "赞比亚克瓦查"},
	{Code: "ZWL", Symbol: "Z$", Name: "津巴布韦元"},
}

var currencySymbolMap map[string]string
var currencyByCode map[string]CurrencyInfo

func init() {
	currencySymbolMap = make(map[string]string, len(currencyCatalog))
	currencyByCode = make(map[string]CurrencyInfo, len(currencyCatalog))
	for _, c := range currencyCatalog {
		currencySymbolMap[c.Code] = c.Symbol
		currencyByCode[c.Code] = c
	}
}

// ListCurrencies 返回平台货币目录（按 code 排序的副本）。
func ListCurrencies() []CurrencyInfo {
	out := make([]CurrencyInfo, len(currencyCatalog))
	copy(out, currencyCatalog)
	sort.Slice(out, func(i, j int) bool { return out[i].Code < out[j].Code })
	return out
}

// LookupCurrency 按 ISO 代码查找目录项；大小写不敏感。
func LookupCurrency(isoCode string) (CurrencyInfo, bool) {
	code := strings.ToUpper(strings.TrimSpace(isoCode))
	if code == "" {
		return CurrencyInfo{}, false
	}
	info, ok := currencyByCode[code]
	return info, ok
}

// HasCurrency 报告代码是否在平台目录中。
func HasCurrency(isoCode string) bool {
	_, ok := LookupCurrency(isoCode)
	return ok
}

// NormalizeCurrency 规范化货币代码；空则 USD。
func NormalizeCurrency(isoCode string) string {
	code := strings.ToUpper(strings.TrimSpace(isoCode))
	if code == "" {
		return "USD"
	}
	return code
}

// GetCurrencySymbol 将货币代码转为展示符号；大小写不敏感。
// 未知代码返回空串（兼容历史调用方）；需要「未知原样返回」时用 CurrencySymbol。
func GetCurrencySymbol(currencyCode string) string {
	code := strings.ToUpper(strings.TrimSpace(currencyCode))
	if code == "" {
		return ""
	}
	if symbol, ok := currencySymbolMap[code]; ok {
		return symbol
	}
	return ""
}

// CurrencySymbol 已知→符号，未知→规范化代码，空→"$"。
func CurrencySymbol(isoCode string) string {
	code := strings.ToUpper(strings.TrimSpace(isoCode))
	if code == "" {
		return "$"
	}
	if sym, ok := currencySymbolMap[code]; ok {
		return sym
	}
	return code
}


// GetRandomCurrencyCode 随机获取一个常见货币代码。
func GetRandomCurrencyCode() string {
	keys := []string{
		"USD", "EUR", "GBP", "INR", "JPY", "KRW", "AUD", "CAD", "SGD", "HKD",
	}
	return keys[rand.IntN(len(keys))]
}

// AdjustBotBetNumStrict 保证机器人下注金额是 cashUnit 的整数倍，cashUnit 保证两位小数。
func AdjustBotBetNumStrict(botBetNum float64, cashUnit float64) float64 {
	botBetNumCents := int64(math.Round(botBetNum * 100))
	cashUnitCents := int64(math.Round(cashUnit * 100))
	multiple := botBetNumCents / cashUnitCents
	if multiple < 1 {
		multiple = 1
	}
	adjustedBetNumCents := multiple * cashUnitCents
	return float64(adjustedBetNumCents) / 100
}
