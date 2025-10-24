package spribe

// 一些错误的信息文本

// 错误代码定义
const (
	CodeInsufficientBalance = 400 // 余额不足
	CodeSystemError         = 500 // 系统错误
	CodeInvalidParameter    = 401 // 参数非法
	CodeNotBettingStage     = 402 // 当前不是下注阶段
)

// 语言类型定义
type Language string

const (
	LangZhCN Language = "zh" // 中文简体
	LangEnUS Language = "en" // 英语美国
	LangThTH Language = "th" // 泰语
	LangViVN Language = "vi" // 越南语
	LangIdID Language = "id" // 印尼语
	LangHiIN Language = "hi" // 印地语
	LangTaIN Language = "ta" // 泰米尔语
	LangMyMM Language = "my" // 缅甸语
	LangJaJP Language = "ja" // 日语
	LangMsMY Language = "ms" // 马来语
	LangKoKR Language = "ko" // 韩语
	LangBnIN Language = "bn" // 孟加拉语
	LangEsAR Language = "es" // 西班牙语阿根廷
	LangPtBR Language = "pt" // 葡萄牙语巴西
	LangItIT Language = "it" // 意大利语
	LangSvSE Language = "sv" // 瑞典语
	LangDeDE Language = "de" // 德语
	LangDaDK Language = "da" // 丹麦语
	LangRoRO Language = "ro" // 罗马尼亚语
	LangNlNL Language = "nl" // 荷兰语
	LangTrTR Language = "tr" // 土耳其语
	LangRuRU Language = "ru" // 俄语
	LangElGR Language = "el" // 希腊语
	LangFrFR Language = "fr" // 法语
)

var errorMessages = map[int]map[Language]string{}

func init() {
	// 余额不足
	errorMessages[CodeInsufficientBalance] = map[Language]string{
		LangZhCN: "余额不足",
		LangEnUS: "Insufficient balance",
		LangThTH: "ยอดเงินไม่เพียงพอ",
		LangViVN: "Số dư không đủ",
		LangIdID: "Saldo tidak cukup",
		LangHiIN: "अपर्याप्त शेष",
		LangTaIN: "போதுமான இருப்பு இல்லை",
		LangMyMM: "လက်ကျန်ငွေ မလုံလောက်ပါ",
		LangJaJP: "残高不足",
		LangMsMY: "Baki tidak mencukupi",
		LangKoKR: "잔액 부족",
		LangBnIN: "অপর্যাপ্ত ব্যালেন্স",
		LangEsAR: "Saldo insuficiente",
		LangPtBR: "Saldo insuficiente",
		LangItIT: "Saldo insufficiente",
		LangSvSE: "Otillräckligt saldo",
		LangDeDE: "Unzureichender Kontostand",
		LangDaDK: "Utilstrækkelig saldo",
		LangRoRO: "Sold insuficient",
		LangNlNL: "Onvoldoende saldo",
		LangTrTR: "Yetersiz bakiye",
		LangRuRU: "Недостаточно средств",
		LangElGR: "Ανεπαρκές υπόλοιπο",
		LangFrFR: "Solde insuffisant",
	}

	// 系统错误
	errorMessages[CodeSystemError] = map[Language]string{
		LangZhCN: "系统错误，请重试",
		LangEnUS: "System error, please try again",
		LangThTH: "ข้อผิดพลาดของระบบ กรุณาลองอีกครั้ง",
		LangViVN: "Lỗi hệ thống, vui lòng thử lại",
		LangIdID: "Kesalahan sistem, silakan coba lagi",
		LangHiIN: "सिस्टम त्रुटि, कृपया पुनः प्रयास करें",
		LangTaIN: "கணினி பிழை, மீண்டும் முயற்சிக்கவும்",
		LangMyMM: "စနစ်ချို့ယွင်းမှု၊ ကျေးဇူးပြု၍ ထပ်မံကြိုးစားပါ",
		LangJaJP: "システムエラー。もう一度お試しください",
		LangMsMY: "Ralat sistem, sila cuba lagi",
		LangKoKR: "시스템 오류, 다시 시도하십시오",
		LangBnIN: "Системная ошибка, попробуйте еще раз",
		LangEsAR: "Error del sistema, por favor intente nuevamente",
		LangPtBR: "Erro do sistema, por favor tente novamente",
		LangItIT: "Errore di sistema, per favore riprova",
		LangSvSE: "Systemfel, försök igen",
		LangDeDE: "Systemfehler, bitte versuchen Sie es erneut",
		LangDaDK: "Systemfejl, prøv venligst igen",
		LangRoRO: "Eroare de sistem, vă rugăm să încercați din nou",
		LangNlNL: "Systeemfout, probeer het opnieuw",
		LangTrTR: "Sistem hatası, lütfen tekrar deneyin",
		LangRuRU: "Системная ошибка, пожалуйста, попробуйте еще раз",
		LangElGR: "Σφάλμα συστήματος, δοκιμάστε ξανά",
		LangFrFR: "Erreur système, veuillez réessayer",
	}

	// 参数非法
	errorMessages[CodeInvalidParameter] = map[Language]string{
		LangZhCN: "参数非法",
		LangEnUS: "Invalid parameter",
		LangThTH: "พารามิเตอร์ไม่ถูกต้อง",
		LangViVN: "Tham số không hợp lệ",
		LangIdID: "Parameter tidak valid",
		LangHiIN: "अमान्य पैरामीटर",
		LangTaIN: "தவறான அளவுரு",
		LangMyMM: "မှားယွင်းသော parameter",
		LangJaJP: "無効なパラメータ",
		LangMsMY: "Parameter tidak sah",
		LangKoKR: "잘못된 매개변수",
		LangBnIN: "অবৈধ প্যারামিটার",
		LangEsAR: "Parámetro inválido",
		LangPtBR: "Parâmetro inválido",
		LangItIT: "Parametro non valido",
		LangSvSE: "Ogiltig parameter",
		LangDeDE: "Ungültiger Parameter",
		LangDaDK: "Ugyldig parameter",
		LangRoRO: "Parametru invalid",
		LangNlNL: "Ongeldige parameter",
		LangTrTR: "Geçersiz parametre",
		LangRuRU: "Неверный параметр",
		LangElGR: "Μη έγκυρη παράμετρος",
		LangFrFR: "Paramètre invalide",
	}

	// 当前不是下注阶段
	errorMessages[CodeNotBettingStage] = map[Language]string{
		LangZhCN: "当前不是下注阶段",
		LangEnUS: "Not in betting stage",
		LangThTH: "ไม่ได้อยู่ในขั้นตอนการเดิมพัน",
		LangViVN: "Không trong giai đoạn đặt cược",
		LangIdID: "Tidak dalam tahap taruhan",
		LangHiIN: "बेटिंग चरण में नहीं",
		LangTaIN: "பந்தய கட்டத்தில் இல்லை",
		LangMyMM: "ဗန်းပြှဲအဆင့်မဟုတ်ပါ",
		LangJaJP: "ベッティング段階ではありません",
		LangMsMY: "Bukan dalam peringkat pertaruhan",
		LangKoKR: "베팅 단계가 아닙니다",
		LangBnIN: "বেটিং পর্যায়ে নেই",
		LangEsAR: "No está en la etapa de apuestas",
		LangPtBR: "Não está na fase de apostas",
		LangItIT: "Non è nella fase di scommessa",
		LangSvSE: "Inte i satsningsstadiet",
		LangDeDE: "Nicht in der Wettphase",
		LangDaDK: "Ikke i indsatsfasen",
		LangRoRO: "Nu este în etapa de pariere",
		LangNlNL: "Niet in de inzetfase",
		LangTrTR: "Bahis aşamasında değil",
		LangRuRU: "Не на стадии ставок",
		LangElGR: "Δεν βρίσκεται στο στάδιο στοιχηματισμού",
		LangFrFR: "Pas dans la phase de pari",
	}
}

func GetErrorMessage(code int, lang Language) string {
	defaultLang := LangEnUS

	if langMap, exists := errorMessages[code]; exists {
		if msg, exists := langMap[lang]; exists {
			return msg
		}
		// 如果指定语言不存在，返回默认语言的错误消息
		if msg, exists := langMap[defaultLang]; exists {
			return msg
		}
	}

	return ""
}
