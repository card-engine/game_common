package player

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/go-kratos/kratos/v2/log"

	"github.com/card-engine/game_common/models"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

const RedisKeyPlayerInfo = "player:%v-%v"
const RedisKeyPlayerToken = "token:%v"

// 玩家信息在数据库里的过期时间
const RedisPlayInfoExpire = 24 * time.Hour

// 一些玩家相关的工具类
// 生成 32 位唯一字符串
func generate32CharString() string {
	key := uuid.New().String()
	key = strings.ReplaceAll(key, "-", "")
	return key
}

// 获取应用信息
func GetAppInfoByID(db *gorm.DB, appId string) (*models.AppInfo, error) {
	var appInfo models.AppInfo
	result := db.Model(&models.AppInfo{}).Where("app_id = ?", appId).First(&appInfo)
	if result.Error != nil {
		return nil, result.Error
	}
	return &appInfo, nil
}

func GetGameRtp(db *gorm.DB, appId string, brand string, gameID string) (*models.GameRtp, error) {
	var gameRtp models.GameRtp
	result := db.Where("app_id = ? and brand = ? and game_id = ?", appId, brand, gameID).First(&gameRtp)
	if result.Error != nil {
		return nil, result.Error
	}
	return &gameRtp, nil
}

// 返回一下token
func InitPlayer(rdb *redis.Client, db *gorm.DB, appId string, playerId string, ssoKey string, brand string, gameId string) error {
	if appId == "" || playerId == "" {
		return errors.New("appId or playerId is empty")
	}

	appInfo, err := GetAppInfoByID(db, appId)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	// 默认的rtp
	//rtp := appInfo.Rtp
	//
	//if err == nil {
	//	rtp = appInfo.Rtp
	//}
	//
	//// 查看游戏有没有指定rtp
	//gameRtpInfo, err := GetGameRtp(db, appId, brand, gameId)
	//if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
	//	return err
	//}
	//
	//if err == nil {
	//	rtp = gameRtpInfo.Rtp
	//}

	// 检查player是否存在，如果存在获取指定用户的rtp
	var playerInfo models.Player
	err = db.Where("app_id = ? AND account_id = ?", appInfo.AppId, playerId).First(&playerInfo).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// 记录不存在，创建新记录
			err = db.Create(&models.Player{
				AccountId: playerId,
				AppId:     appInfo.AppId,
				NickName:  playerId,
			}).Error
			if err != nil {
				return err
			}
		} else {
			// 其他错误，返回错误信息
			return err
		}
	} else {
		if playerInfo.HasSetRtp {
			//rtp = playerInfo.Rtp
		}
	}

	ctx := context.Background()
	token := generate32CharString()

	err = rdb.HSet(ctx, fmt.Sprintf(RedisKeyPlayerInfo, appId, playerId), []string{
		"rtp", playerInfo.Rtp, // 玩家RTP
		"appId", appId, // 应用ID
		"playerId", playerId, // 玩家ID
		"brand", brand, // 品牌
		"gameId", gameId, // 游戏ID
		"currency", appInfo.Currency, // 货币类型
		"aId", strconv.FormatUint(playerInfo.AId, 10), // 用户ID
	}).Err()

	if err != nil {
		return nil
	}
	// 数据设置过期
	rdb.Expire(ctx, fmt.Sprintf(RedisKeyPlayerInfo, appId, playerId), RedisPlayInfoExpire)

	// token与用户信息做映射关系
	rdb.Set(ctx, fmt.Sprintf(RedisKeyPlayerToken, token), fmt.Sprintf(RedisKeyPlayerInfo, appId, playerId), RedisPlayInfoExpire)

	// 保存 ssoKey 到 token 的索引，10分钟过期

	key := fmt.Sprintf("ssoKey:%v-%v", appId, ssoKey)

	err = rdb.Set(ctx, key, token, RedisPlayInfoExpire).Err()
	if err != nil {
		return err
	}

	return nil
}

// PlayerInfo 定义玩家信息结构体
type PlayerInfo struct {
	RTP      string  `json:"rtp"`
	AppID    string  `json:"appId"`
	PlayerID string  `json:"playerId"`
	Brand    string  `json:"brand"`
	Lang     string  `json:"lang"` //玩家所使用的语言
	GameID   string  `json:"gameId"`
	Currency string  `json:"currency"`
	Balance  float64 `json:"balance"`
	AId      uint64  `json:"aId"`
}

// Deprecated: 通过 token 从 Redis 获取玩家指定信息, 请使用GetPlayerByAppAndPlayerId
func GetPlayerInfoByToken(rdb *redis.Client, token string) (*PlayerInfo, error) {
	ctx := context.Background()

	infoKey, err := rdb.GetEx(ctx, fmt.Sprintf(RedisKeyPlayerToken, token), RedisPlayInfoExpire).Result()
	if err != nil {
		return nil, err
	}

	// 使用 HGetAll 一次性获取所有字段和值
	values, err := rdb.HGetAll(ctx, infoKey).Result()
	if err != nil {
		return nil, fmt.Errorf("获取玩家信息失败: %w", err)
	}
	// 给数据续期
	rdb.Expire(ctx, infoKey, RedisPlayInfoExpire)

	playerInfo := &PlayerInfo{}

	// 尝试从返回的 map 中提取各字段值
	playerInfo.RTP = values["rtp"]
	playerInfo.AppID = values["appId"]
	playerInfo.PlayerID = values["playerId"]
	playerInfo.Brand = values["brand"]
	playerInfo.GameID = values["gameId"]
	playerInfo.Currency = values["currency"]
	aIdStr, ok := values["aId"]
	if ok {
		playerInfo.AId, err = strconv.ParseUint(aIdStr, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("转换 aId 为 uint64 失败: %w", err)
		}
	} else {
		playerInfo.AId = 0
	}

	// 转换 balance 字段为 float64 类型
	balanceStr, ok := values["balance"]
	if ok {
		balance, err := strconv.ParseFloat(balanceStr, 64)
		if err != nil {
			return nil, fmt.Errorf("转换 balance 为 float64 失败: %w", err)
		}
		playerInfo.Balance = balance
	} else {
		// 如果 balance 字段不存在，可根据需求处理，这里设为 0
		playerInfo.Balance = 0
	}

	return playerInfo, nil
}

// Deprecated: token索引的方式废弃了
func CheckTokenExists(rdb *redis.Client, token string) (bool, error) {
	ctx := context.Background()
	result, err := rdb.Exists(ctx, fmt.Sprintf(RedisKeyPlayerToken, token)).Result()
	if err != nil {
		return false, err
	}
	return result > 0, nil
}

// Deprecated: 通过token获取appid, token索引的方式废弃了
func GetAppIDByToken(rdb *redis.Client, token string) (string, error) {
	ctx := context.Background()

	infoKey, err := rdb.GetEx(ctx, fmt.Sprintf(RedisKeyPlayerToken, token), RedisPlayInfoExpire).Result()
	if err != nil {
		return "", err
	}

	// 使用 Redis 的 HGet 方法获取 appId
	appId, err := rdb.HGet(ctx, infoKey, "appId").Result()
	if err != nil {
		if err == redis.Nil {
			return "", errors.New("token not found")
		}
		return "", err
	}

	// 给数据续期
	rdb.Expire(ctx, infoKey, RedisPlayInfoExpire)
	return appId, nil
}

// Deprecated: 通过token获取金币 balance
func GetBalanceByToken(rdb *redis.Client, token string) (float64, error) {
	ctx := context.Background()
	infoKey, err := rdb.GetEx(ctx, fmt.Sprintf(RedisKeyPlayerToken, token), RedisPlayInfoExpire).Result()
	if err != nil {
		return 0, err
	}

	// 使用 Redis 的 HGet 方法获取 balance
	balanceStr, err := rdb.HGet(ctx, infoKey, "balance").Result()
	if err != nil {
		if err == redis.Nil {
			return 0, errors.New("token not found or balance not set")
		}
		return 0, err
	}

	// 将获取到的字符串类型的 balance 转换为 float64 类型
	balance, err := strconv.ParseFloat(balanceStr, 64)
	if err != nil {
		return 0, errors.New("failed to convert balance to float64")
	}
	// 给数据续期
	rdb.Expire(ctx, infoKey, RedisPlayInfoExpire)
	return balance, nil
}

// Deprecated: 获取货币类型, 请使用GetPlayerByAppAndPlayerId
func GetCurrencyByToken(rdb *redis.Client, token string) (string, error) {
	ctx := context.Background()
	infoKey, err := rdb.GetEx(ctx, fmt.Sprintf(RedisKeyPlayerToken, token), RedisPlayInfoExpire).Result()
	if err != nil {
		return "", err
	}
	// 使用 Redis 的 HGet 方法获取 currency
	currency, err := rdb.HGet(ctx, infoKey, "currency").Result()
	if err != nil {
		if err == redis.Nil {
			return "", errors.New("token not found or currency not set")
		}
		return "", err
	}
	// 给数据续期
	rdb.Expire(ctx, infoKey, RedisPlayInfoExpire)
	return currency, nil
}

var currencySymbolMap = map[string]string{
	"INR": "₹",
	// 可以根据需要添加更多货币映射
	"USD": "$",
	"EUR": "€",
	"GBP": "£",
}

// Deprecated: GetCurrencySymbol 已废弃，请使用 utils.GetCurrencySymbol
// GetCurrencySymbol 通过货币字母缩写获取货币符号
func GetCurrencySymbol(currencyCode string) string {
	if symbol, ok := currencySymbolMap[currencyCode]; ok {
		return symbol
	}
	return "" // 如果未找到对应的符号，返回空字符串
}

// Deprecated: 通过token设置金币金额, 请使用UpdateBalance
func SetBalanceByToken(rdb *redis.Client, token string, amount float64) error {
	ctx := context.Background()

	infoKey, err := rdb.GetEx(ctx, fmt.Sprintf(RedisKeyPlayerToken, token), RedisPlayInfoExpire).Result()
	if err != nil {
		return err
	}

	// 将 float64 类型的金额转换为字符串
	amountStr := strconv.FormatFloat(amount, 'f', -1, 64)
	// 使用 Redis 的 HSet 方法设置 balance
	err = rdb.HSet(ctx, infoKey, "balance", amountStr).Err()
	if err != nil {
		return err
	}
	// 给数据续期
	rdb.Expire(ctx, infoKey, RedisPlayInfoExpire)
	return nil
}

// Deprecated: 此函数已弃用
func GetTokenBySSOKey(rdb *redis.Client, appId, ssokey string) (string, error) {
	ctx := context.Background()
	// 使用 Redis 的 Get 方法获取 token
	key := fmt.Sprintf("ssoKey:%v-%v", appId, ssokey)

	token, err := rdb.Get(ctx, key).Result()
	if err == nil {
		return token, nil
	}

	// 如果原始key不存在，尝试查询URL编码后的key
	if errors.Is(err, redis.Nil) {
		encodedSSOKey := url.QueryEscape(ssokey)
		encodedKey := fmt.Sprintf("ssoKey:%v-%v", appId, encodedSSOKey)
		log.Infof("ssokey not found, try to use ssokey %s encoded encodedKey: %s", ssokey, encodedKey)
		token, err = rdb.Get(ctx, encodedKey).Result()
		if err == nil {
			return token, nil
		}
	}

	// 两种方式都未找到
	if errors.Is(err, redis.Nil) {
		return "", errors.New("ssokey not found")
	}
	return "", err
}

// 通过appid以及playerid获取rtp档位
func GetRtpByAppPlayerId(rdb *redis.Client, appId, playerId string) (string, error) {
	ctx := context.Background()

	infoKey := fmt.Sprintf(RedisKeyPlayerInfo, appId, playerId)

	// 使用 Redis 的 HGet 方法获取 currency
	rtp, err := rdb.HGet(ctx, infoKey, "rtp").Result()
	if err != nil {
		if err == redis.Nil {
			return "", errors.New("token not found or currency not set")
		}
		return "", err
	}
	// 给数据续期
	rdb.Expire(ctx, infoKey, RedisPlayInfoExpire)
	return rtp, nil
}

// 通过appid以及playeridg更新rtp档位
func UpdateRtpByAppPlayerId(rdb *redis.Client, appId, playerId string, rtp string) (string, error) {
	ctx := context.Background()

	infoKey := fmt.Sprintf(RedisKeyPlayerInfo, appId, playerId)

	// 使用 Redis 的 HGet 方法获取 currency
	err := rdb.HSet(ctx, infoKey, "rtp", rtp).Err()
	if err != nil {
		if err == redis.Nil {
			return "", errors.New("token not found or currency not set")
		}
		return "", err
	}
	// 给数据续期
	rdb.Expire(ctx, infoKey, RedisPlayInfoExpire)
	return rtp, nil
}

// 通过appid以及playerids批量更新rtp档位
func UpdateRtpByAppPlayerIds(rdb *redis.Client, appId string, playerIds []string, rtp string) error {
	ctx := context.Background()
	// 使用管道批量操作提高性能
	pipe := rdb.TxPipeline()
	for _, playerId := range playerIds {
		infoKey := fmt.Sprintf(RedisKeyPlayerInfo, appId, playerId)
		pipe.HSet(ctx, infoKey, "rtp", rtp)
		pipe.Expire(ctx, infoKey, RedisPlayInfoExpire)
	}
	_, err := pipe.Exec(ctx)
	return err
}

// func GetPlayerId(token string) string {
// 	return ""
// }

// func AddMoney(rdb *redis.Client, token string, amount float64) (float64, error) {
// 	return 0, nil
// }

func FindPlayerByIds(ctx context.Context, db *gorm.DB, appId string, ids []string) ([]*models.Player, error) {
	return gorm.G[*models.Player](db).Where("app_id = ? AND account_id IN ?", appId, ids).Find(ctx)
}

func UpdatePlayerRtp(ctx context.Context, db *gorm.DB, appId string, aids []uint64, rtp string) (rowsAffected int, err error) {
	updates := map[string]interface{}{
		"rtp":      rtp,
		"rtp_time": time.Now(),
	}
	result := db.Model(&models.Player{}).Where("app_id = ? AND a_id IN ?", appId, aids).Updates(updates)
	if result.Error != nil {
		return 0, result.Error
	}
	return int(result.RowsAffected), nil
}

func UpdateGameId(rdb *redis.Client, appId, playerId string, gameBrand, gameId string) error {
	ctx := context.Background()
	infoKey := fmt.Sprintf(RedisKeyPlayerInfo, appId, playerId)
	// 使用 HSet 更新 gameId 字段
	err := rdb.HSet(ctx, infoKey, "gameId", gameId).Err()
	if err != nil {
		return fmt.Errorf("更新 gameId 失败: %w", err)
	}
	// 更新 brand 字段
	err = rdb.HSet(ctx, infoKey, "brand", gameBrand).Err()
	if err != nil {
		return fmt.Errorf("更新 brand 失败: %w", err)
	}
	// 给数据续期
	rdb.Expire(ctx, infoKey, RedisPlayInfoExpire)
	return nil
}

func UpdateBalance(rdb *redis.Client, appId, playerId string, amount float64) error {
	ctx := context.Background()

	infoKey := fmt.Sprintf(RedisKeyPlayerInfo, appId, playerId)

	// 将 float64 类型的金额转换为字符串
	amountStr := strconv.FormatFloat(amount, 'f', -1, 64)
	// 使用 Redis 的 HSet 方法设置 balance
	err := rdb.HSet(ctx, infoKey, "balance", amountStr).Err()
	if err != nil {
		return err
	}
	// 给数据续期
	rdb.Expire(ctx, infoKey, RedisPlayInfoExpire)
	return nil
}

// 通过appid以及playerid获取玩家
func GetPlayerByAppAndPlayerId(rdb *redis.Client, appId, playerId string) (*PlayerInfo, error) {
	ctx := context.Background()
	infoKey := fmt.Sprintf(RedisKeyPlayerInfo, appId, playerId)
	// 使用 HGetAll 一次性获取所有字段和值
	values, err := rdb.HGetAll(ctx, infoKey).Result()
	if err != nil {
		return nil, fmt.Errorf("获取玩家信息失败: %w", err)
	}
	// 给数据续期
	rdb.Expire(ctx, infoKey, RedisPlayInfoExpire)
	playerInfo := &PlayerInfo{}
	// 尝试从返回的 map 中提取各字段值
	playerInfo.RTP = values["rtp"]
	playerInfo.AppID = values["appId"]
	playerInfo.PlayerID = values["playerId"]
	playerInfo.Brand = values["brand"]
	playerInfo.GameID = values["gameId"]
	playerInfo.Currency = values["currency"]
	aIdStr, ok := values["aId"]
	if ok {
		playerInfo.AId, err = strconv.ParseUint(aIdStr, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("转换 aId 为 uint64 失败: %w", err)
		}
	} else {
		playerInfo.AId = 0
	}

	// 转换 balance 字段为 float64 类型
	balanceStr, ok := values["balance"]
	if ok {
		balance, err := strconv.ParseFloat(balanceStr, 64)
		if err != nil {
			return nil, fmt.Errorf("转换 balance 为 float64 失败: %w", err)
		}
		playerInfo.Balance = balance
	} else {
		// 如果 balance 字段不存在，可根据需求处理，这里设为 0
		playerInfo.Balance = 0
	}

	return playerInfo, nil
}
