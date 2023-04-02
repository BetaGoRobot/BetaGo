package redis

import (
	"github.com/BetaGoRobot/BetaGo/betagovar"
	"github.com/redis/go-redis/v9"
)

// RedisClient  12
var RedisClient *redis.Client

var RedisConnAddr = "redis-ix-chart.ix-redis:6379"

// GetRedisClient 1
func GetRedisClient() *redis.Client {
	if RedisClient == nil {
		RedisClient = redis.NewClient(&redis.Options{
			Addr:       RedisConnAddr,
			ClientName: betagovar.RobotName,
		})
	}
	return RedisClient
}
