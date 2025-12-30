package spribe

import (
	"hash/fnv"
	"strconv"
)

func hashTo1_72(s string) int {
	h := fnv.New32a()
	h.Write([]byte(s))
	hashValue := int(h.Sum32())

	// 取绝对值并映射到1-72
	if hashValue < 0 {
		hashValue = -hashValue
	}
	return (hashValue % 72) + 1
}

// 生成头像文件名
func GenerateSpribeAvatarFilename(appid string, playerId string) string {
	num := hashTo1_72(appid + "_" + playerId)
	return "av-" + strconv.Itoa(num) + ".png"
}
