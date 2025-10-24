package redis

import (
	"context"
	"fmt"
	"os"
	"time"

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

// (使用我们上面修改的版本)
var setOrGetExpireAtScriptV2 = redis.NewScript(`
    local result = redis.call('SET', KEYS[1], ARGV[1], 'EXAT', ARGV[2], 'NX')
    if result then
        -- 返回 {1, new_timestamp}
        return {1, tonumber(ARGV[2])}
    else
        -- 返回 {0, existing_timestamp}
        return {0, redis.call('EXPIRETIME', KEYS[1])}
    end
`)

// SetOrGetExpireAtV2 原子地执行操作，并明确返回操作类型。
//
// 返回值:
// - wasSet (bool): true 表示 key 是新创建的; false 表示 key 已存在。
// - timestamp (int64): 相关的时间戳 (新设置的或已存在的)。
// - error: 如果执行出错。
func SetOrGetExpireAt(ctx context.Context, rdb *redis.Client, key string, value interface{}, expireAt time.Time) (set bool, t time.Time, err error) {
	timestamp := expireAt.Unix()

	result, err := setOrGetExpireAtScriptV2.Run(ctx, rdb, []string{key}, value, timestamp).Result()
	if err != nil {
		if err == redis.Nil {
			return set, t, fmt.Errorf("script execution returned nil: %w", err)
		}
		return set, t, fmt.Errorf("failed to run SetOrGetExpireAt script: %w", err)
	}

	// Lua 脚本返回一个 table, go-redis 将其映射为 []interface{}
	resultSlice, ok := result.([]interface{})
	if !ok {
		return set, t, fmt.Errorf("unexpected script result type: %T (value: %v)", result, result)
	}

	if len(resultSlice) != 2 {
		return set, t, fmt.Errorf("unexpected script result slice length: %d (expected 2)", len(resultSlice))
	}

	// 解析 元素 1 (标志位)
	actionFlag, ok := resultSlice[0].(int64)
	if !ok {
		return set, t, fmt.Errorf("unexpected action flag type: %T", resultSlice[0])
	}

	// 解析 元素 2 (时间戳)
	ts, ok := resultSlice[1].(int64)
	if !ok {
		return set, t, fmt.Errorf("unexpected timestamp type: %T", resultSlice[1])
	}

	// actionFlag == 1 意味着 "新 Set 了"
	wasSet := (actionFlag == 1)

	return wasSet, time.Unix(ts, 0), nil
}
