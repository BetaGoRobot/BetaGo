package redis

import (
	"os"

	"github.com/BetaGoRobot/BetaGo/consts"
	"github.com/redis/go-redis/v9"
)

// RedisClient  12
var RedisClient *redis.Client

var RedisConnAddr = os.Getenv("REDIS_CONN_ADDR")

// GetRedisClient 1
func GetRedisClient() *redis.Client {
	if RedisClient == nil {
		RedisClient = redis.NewClient(&redis.Options{
			Addr:       RedisConnAddr,
			ClientName: consts.RobotName,
		})
	}
	return RedisClient
}
