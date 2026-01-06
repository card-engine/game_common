package fairness

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"math"
	"math/rand"

	"github.com/card-engine/game_common/utils"
	"github.com/google/uuid"
)

// ===========================
type ClientSeed struct {
	ID         string
	OperatorID string
	UserID     string
	ClientSeed string
	Nickname   string
}

type SeedInfo struct {
	GameUUID         string
	ServerSeed       string
	HashedServerSeed string
	CombinedHash     string
	Decimal          string
}

// ===========================

// 用于有些游戏的roundid为string的情况，将其转换为int64
func RoundIdInt64(s string) int64 {
	h := sha256.New()
	h.Write([]byte(s))
	hashBytes := h.Sum(nil)
	// Convert first 8 bytes to int64 (big-endian)
	return int64(binary.BigEndian.Uint64(hashBytes[:8]))
}

func GenerateClientSeeds(roundId int64, count int) []ClientSeed {
	seed := roundId
	r := rand.New(rand.NewSource(seed)) // 固定种子，保证可复现

	seeds := make([]ClientSeed, count)
	for i := 0; i < count; i++ {
		// 随机生成一些基础数据
		operatorID := uuid.New().String()
		userID := fmt.Sprintf("%d", r.Int63n(2_000_000_000))
		clientSeed := randomHex(r, 16)

		// 用 operator + userID 组合生成唯一 ID
		id := fmt.Sprintf("%s::%s", operatorID, userID)

		nickname := randomNickname(r)

		seeds[i] = ClientSeed{
			ID:         id,
			OperatorID: operatorID,
			UserID:     userID,
			ClientSeed: clientSeed,
			Nickname:   nickname,
		}
	}

	return seeds
}

func GenerateServerSeed(roundID int64, result float64) SeedInfo {
	// Step 1: 用 roundID 派生出 ServerSeed
	key := []byte(fmt.Sprintf("round-key-%d", roundID))
	mac := hmac.New(sha256.New, key)
	mac.Write([]byte("server-seed"))
	serverSeed := mac.Sum(nil)

	// Step 2: HashedServerSeed = sha256(ServerSeed)
	hashed := sha256.Sum256(serverSeed)

	// Step 3: CombinedHash = sha512(ServerSeed + HashedServerSeed)
	h := sha512.New()
	h.Write(serverSeed)
	h.Write(hashed[:])
	combined := h.Sum(nil)

	// Step 4: 生成一个稳定的 UUID（基于 roundID）
	uuidStr := uuid.NewSHA1(uuid.NameSpaceOID, []byte(fmt.Sprintf("%d", roundID))).String()

	// Step 5: 生成确定性随机大数字（科学计数法格式）
	macFloat := hmac.New(sha256.New, key)
	macFloat.Write([]byte(fmt.Sprintf("%f", result)))
	hashBytes := macFloat.Sum(nil)

	// 使用哈希值生成大数字
	num := binary.BigEndian.Uint64(hashBytes[:8])

	// 方法1：生成类似 1.364273877642065e+153 的科学计数法数字
	// 使用哈希的前8字节生成基数（1.0-9.999...范围）
	base := 1.0 + float64(num%9000000000000000)/10000000000000000.0

	// 使用哈希的其他字节生成指数（100-300范围）
	expNum := binary.BigEndian.Uint64(hashBytes[8:16])
	exponent := 100 + int(expNum%201) // 指数范围 100-300

	// 计算最终的科学计数法数字
	scientificNumber := base * math.Pow10(exponent)
	decimal := fmt.Sprintf("%.15e", scientificNumber)

	return SeedInfo{
		GameUUID:         uuidStr,
		ServerSeed:       hex.EncodeToString(serverSeed),
		HashedServerSeed: hex.EncodeToString(hashed[:]),
		CombinedHash:     hex.EncodeToString(combined),
		Decimal:          decimal, // 科学计数法，15位小数精度
	}
}

// 仅生成 ServerSeed 字符串
func GenerateServerSeedStr(roundID int64) string {
	key := []byte(fmt.Sprintf("round-key-%d", roundID))
	mac := hmac.New(sha256.New, key)
	mac.Write([]byte("server-seed"))
	serverSeed := mac.Sum(nil)

	return hex.EncodeToString(serverSeed)
}

func GenerateClientSeedsStr(roundId int64) string {
	seed := roundId
	r := rand.New(rand.NewSource(seed)) // 固定种子，保证可复现
	clientSeed := randomHex(r, 16)
	return clientSeed
}

// ======================================================================
// 随机昵称（可复现）
func randomNickname(r *rand.Rand) string {
	adjectives := utils.BotNicknames
	return adjectives[r.Intn(len(adjectives))]
}

// 生成伪随机 hex
func randomHex(r *rand.Rand, n int) string {
	b := make([]byte, n/2)
	for i := range b {
		b[i] = byte(r.Intn(256))
	}
	return hex.EncodeToString(b)
}
