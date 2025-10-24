package health

import (
	"context"
	"fmt"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/healthcheck"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
	"time"
)

func Check(db *gorm.DB, rdb *redis.Client, opts ...HealthCheckOption) fiber.Handler {
	// 默认配置
	config := &healthCheckConfig{
		dbChecker:    defaultDBCheck,
		redisChecker: defaultRedisCheck,
	}

	// 应用选项
	for _, opt := range opts {
		opt(config)
	}
	return healthcheck.New(healthcheck.Config{
		LivenessEndpoint:  "/health/livez",
		ReadinessEndpoint: "/health/readyz",
		// 自定义就绪检查逻辑
		ReadinessProbe: func(c *fiber.Ctx) bool {
			// 检查数据库连接
			if config.dbChecker != nil {
				if err := config.dbChecker(db); err != nil {
					log.Context(c.UserContext()).Errorf("Health Check 数据库连接失败 : %v", err)
					return false
				}
			}

			// 检查Redis连接
			if config.redisChecker != nil && rdb != nil {
				if err := config.redisChecker(rdb); err != nil {
					log.Context(c.UserContext()).Errorf("ealth Check Redis连接失败: %v", err)
					return false
				}
			}
			return true
		},
	})
}

// HealthCheckOption 健康检查选项
type HealthCheckOption func(*healthCheckConfig)

// healthCheckConfig 健康检查配置
type healthCheckConfig struct {
	dbChecker    func(*gorm.DB) error
	redisChecker func(*redis.Client) error
}

// WithDBChecker 设置数据库检查器
func WithDBChecker(checker func(*gorm.DB) error) HealthCheckOption {
	return func(config *healthCheckConfig) {
		config.dbChecker = checker
	}
}

// WithRedisChecker 设置Redis检查器
func WithRedisChecker(checker func(*redis.Client) error) HealthCheckOption {
	return func(config *healthCheckConfig) {
		config.redisChecker = checker
	}
}

// defaultRedisCheck 默认Redis检查
func defaultRedisCheck(redisClient *redis.Client) error {
	if redisClient == nil {
		return fmt.Errorf("redis client not initialized")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	return redisClient.Ping(ctx).Err()
}

// defaultDBCheck 默认数据库检查
func defaultDBCheck(db *gorm.DB) error {
	if db == nil {
		return fiber.NewError(fiber.StatusInternalServerError, "数据库实例未初始化")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	var result int
	if err := db.WithContext(ctx).Raw("SELECT 1").Scan(&result).Error; err != nil {
		return fiber.NewError(fiber.StatusServiceUnavailable, "数据库连接失败: "+err.Error())
	}
	return nil
}
