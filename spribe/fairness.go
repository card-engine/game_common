package spribe

import (
	"crypto/sha1"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"math/rand"
	"strings"
)

// 公平性的计算相关的逻辑，思路就是可以通过roundId做种子

// 你可以使用roundId做种子，这样子，查询也不需要存库，只需要重新计算一下就可以了
// 生成基于哈希的服务器种子
func GenerateServerSeedHashBased(seed int32) (string, error) {

	rng := rand.New(rand.NewSource(int64(seed)))
	randomBytes := make([]byte, 16)
	for i := range randomBytes {
		randomBytes[i] = byte(rng.Intn(256))
	}
	data := randomBytes
	// 计算哈希
	hash := sha256.Sum256(data)

	// Base64 URL 编码
	encoded := base64.URLEncoding.EncodeToString(hash[:])
	// 去掉末尾的 '='
	encoded = strings.TrimRight(encoded, "=")

	return encoded, nil
}

// GenerateHashBasedID 基于哈希生成ID
// 再使用服务器的种子生成ID
func GenerateHashBasedID(seed string) int64 {
	data := []byte(seed)

	hash := sha1.Sum(data)

	// 取前4字节并转换为int32，然后加上基数
	id := int64(binary.BigEndian.Uint32(hash[:4]))
	baseValue := int64(3260000000)

	resultID := baseValue + (id % 1000000) // 限制在合理范围内

	// 转换为字符串返回
	return resultID
}

// generateProfile 根据固定种子和数量，生成 profile 数据
func GenerateProfile(seed int32, count int) []map[string]interface{} {
	rng := rand.New(rand.NewSource(int64(seed)))
	results := make([]map[string]interface{}, 0, count)

	for i := 0; i < count; i++ {
		// profileImage 范围 1~72
		profileImage := fmt.Sprintf("av-%d.png", rng.Intn(72)+1)

		// 随机 seed（Base62 风格）
		encodedSeed := randomString(rng, 20)

		// 昵称从 botNicknames 里取
		nickname := botNicknames[rng.Intn(len(botNicknames))]
		username := fmt.Sprintf("%s_%d", nickname, rng.Intn(99999))

		results = append(results, map[string]interface{}{
			"profileImage": profileImage,
			"seed":         encodedSeed,
			"username":     username,
		})
	}

	return results
}

var botNicknames = []string{
	"Ace", "Bolt", "Comet", "Dash", "Echo", "Fury", "Ghost", "Hawk",
	"Iron", "Jolt", "King", "Lion", "Mage", "Nova", "Orion", "Peak",
	"Queen", "Rage", "Storm", "Titan", "Unit", "Viper", "Wolf", "Xeno",
	"Yeti", "Zest", "Alpha", "Bravo", "Charlie", "Delta", "Eagle",
	"Faker", "Gamer", "Hero", "Icon", "Joker", "Knight", "Legend",
	"Master", "Ninja", "Omega", "Pilot", "Raptor", "Shadow", "T-Rex",
	"Blaze", "Crusher", "Venom", "Hunter", "Phantom", "Samurai", "Reaper",
	"Drake", "Specter", "Wraith", "Phoenix", "Striker", "Sniper", "Cobra",
	"Rogue", "Gladiator", "Berserk", "Monk", "Zephyr", "Rider", "Ranger",
	"Slayer", "Overlord", "Grim", "Titanium", "Cyber", "Punk",
	"Shogun", "Dagger", "Spartan", "Thunder", "Bullet", "Predator",
	"Hurricane", "Avalanche", "Cyclone", "Inferno",
}

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func randomString(rng *rand.Rand, length int) string {
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rng.Intn(len(charset))]
	}
	return string(b)
}
