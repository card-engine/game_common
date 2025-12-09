package utils

import (
	"fmt"
	"github.com/redis/go-redis/v9"
	"testing"
)

func TestSnowflake(t *testing.T) {
	// 初始化Redis客户端
	redisClient := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "z7wmfqLX",
		DB:       0,
	})

	snowflakeUtil, err := NewSnowflakeUtil(redisClient, "snowflake-service")
	if err != nil {
		panic(err)
	}

	for i := 0; i < 10; i++ {
		fmt.Println(snowflakeUtil.GenerateIDInt64())
		fmt.Println(snowflakeUtil.Generate().String())
	}
}
