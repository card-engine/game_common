package protocol

// ========================================
// cmd = 0
type PreLoginRequest struct {
	Api string `sfs:"api"`
	Cl  string `sfs:"cl"`
}

type PreLoginRespond struct {
	Ct int32  `sfs:"ct"`
	Ms int32  `sfs:"ms"`
	Tk string `sfs:"tk"`
}

// ========================================
// cmd = 1
type LoginPPlatform struct {
	DeviceInfo string `sfs:"deviceInfo"` // JSON 字符串，详细设备信息（浏览器、OS、设备型号、方向等）
	DeviceType string `sfs:"deviceType"` // 简化设备类型，比如 "desktop" / "mobile"
	UserAgent  string `sfs:"userAgent"`  // 浏览器或客户端 UA 字符串
}

// 登录请求中的参数
type LoginP struct {
	Currency     string         `sfs:"currency"`     // 用户选择的货币，例如 "USD"
	Jurisdiction string         `sfs:"jurisdiction"` // 用户所在地区或牌照区域，如 "CW"
	Lang         string         `sfs:"lang"`         // 用户语言，例如 "en"
	Platform     LoginPPlatform `sfs:"platform"`     // 客户端设备信息
	SessionToken string         `sfs:"sessionToken"` // 会话唯一标识，用于验证合法性
	Token        string         `sfs:"token"`        // 用户身份或登录 token
	Version      string         `sfs:"version"`      // 客户端版本号
}

// 登录请求消息
type LoginRequest struct {
	Payload  LoginP `sfs:"p,optional"` // payload 对象，包含客户端环境信息 (可选)
	PassWord string `sfs:"pw"`         // 用户密码，这里可能为空
	UserName string `sfs:"un"`         // 用户名，例如 "27890&&demo"
	ZoneName string `sfs:"zn"`         // 游戏区或房间名，例如 "mines"
}

// 登录响应消息
type LoginRespond struct {
	Id int32         `sfs:"id"` // 用户 ID 或系统分配的唯一标识
	Pi int16         `sfs:"pi"` // 玩家索引 / 内部编号
	Rl []interface{} `sfs:"rl"` // 资源列表或初始化数据 ROOMLIST
	Rs int16         `sfs:"rs"` // 响应状态码，0 = 成功
	Un string        `sfs:"un"` // 用户名
	Zn string        `sfs:"zn"` // 游戏区 / 房间名
}

// ========================================
