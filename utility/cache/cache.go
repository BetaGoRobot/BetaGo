package cache

import (
	"context"
	"time"

	"github.com/BetaGoRobot/BetaGo/utility/logs"
	"github.com/BetaGoRobot/go_utils/reflecting"
	"github.com/patrickmn/go-cache"
	"go.uber.org/zap"
)

var wrapper = NewCacheWrapper(30*time.Minute, 30*time.Minute)

// CacheWrapper 结构体和 New/GetOrExecute 方法保持不变
type CacheWrapper struct {
	c *cache.Cache
}

func NewCacheWrapper(defaultExpiration, cleanupInterval time.Duration) *CacheWrapper {
	return &CacheWrapper{
		c: cache.New(defaultExpiration, cleanupInterval),
	}
}

func GetOrExecute[T any](ctx context.Context, key string, fn func() (T, error)) (value T, err error) {
	fName := reflecting.GetFunctionName(fn)
	if value, found := wrapper.c.Get(fName + ":" + key); found {
		logs.L().Ctx(ctx).Info("[✅ Cache HIT] Executing function to get cache value,", zap.String("key", key), zap.String("function", fName))
		return value.(T), nil
	}

	logs.L().Ctx(ctx).Warn("[❌ Cache MISS] Executing function to get cache value,", zap.String("key", key), zap.String("function", fName))
	value, err = fn()
	if err != nil {
		return
	}

	wrapper.c.Set(key, value, cache.DefaultExpiration)
	logs.L().Ctx(ctx).Info("📦 Cache SET", zap.String("key", key))

	return
}
