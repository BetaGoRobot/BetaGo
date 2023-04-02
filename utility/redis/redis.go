package redis

import (
	"github.com/BetaGoRobot/BetaGo/betagovar"
	"github.com/redis/go-redis/v9"
)

// RedisClient  12
var RedisClient *redis.Client

// GetRedisClient 1
func GetRedisClient() *redis.Client {
	if RedisClient == nil {
		RedisClient = redis.NewClient(&redis.Options{
			Addr:       "redis-ix-chart.ix-redis:6379",
			ClientName: betagovar.RobotName,
		})
	}
	return RedisClient
}
