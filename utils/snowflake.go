package utils

import (
	"context"
	"fmt"

	"github.com/bwmarrin/snowflake"
	"github.com/redis/go-redis/v9"
)

type SnowflakeUtil struct {
	rdb           *redis.Client
	serviceName   string
	snowflakeNode *snowflake.Node
}

func NewSnowflakeUtil(rdb *redis.Client, serviceName string) (*SnowflakeUtil, error) {
	s := &SnowflakeUtil{
		rdb:         rdb,
		serviceName: serviceName,
	}
	err := s.initSnowflakeFromRedis()
	if err != nil {
		return nil, err
	}
	return s, nil
}

// InitSnowflakeFromRedis 从Redis获取Worker ID并初始化Snowflake
func (s *SnowflakeUtil) initSnowflakeFromRedis() error {
	// 获取Worker ID
	workerID, err := s.getWorkerIDFromRedis()
	if err != nil {
		return err
	}
	// 初始化Snowflake节点
	s.snowflakeNode, err = snowflake.NewNode(workerID)
	return err
}

// getWorkerIDFromRedis 从Redis获取唯一的Worker ID
func (s *SnowflakeUtil) getWorkerIDFromRedis() (int64, error) {
	workerIDStr, err := s.rdb.Incr(context.Background(), fmt.Sprintf("snowflake:%s", s.serviceName)).Result()
	if err != nil {
		return 0, err
	}
	// Snowflake节点ID范围是0-1023
	workerID := (workerIDStr - 1) % 1024
	return workerID, nil
}

func (s *SnowflakeUtil) Generate() snowflake.ID {
	return s.snowflakeNode.Generate()
}
func (s *SnowflakeUtil) GenerateId() string {
	return s.Generate().String()
}

func (s *SnowflakeUtil) GenerateIDInt64() int64 {
	return s.Generate().Int64()
}

func (s *SnowflakeUtil) GenerateIDBase64() string {
	return s.Generate().Base64()
}
