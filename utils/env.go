package utils

import "os"

// 是否为生产环境
func IsProd() bool {
	env := os.Getenv("GFGAME_ENV")
	if env == "prod" {
		return true
	}
	return false
}
