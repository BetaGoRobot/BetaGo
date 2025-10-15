package cache

import (
	"log"
	"time"

	"github.com/patrickmn/go-cache"
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

func GetOrExecute[T any](key string, fn func() (T, error)) (val T, err error) {
	if value, found := wrapper.c.Get(key); found {
		log.Printf("✅ Cache HIT for key: %s\n", key)
		return value.(T), nil
	}

	log.Printf("❌ Cache MISS for key: %s. Executing function...\n", key)

	value, err := fn()
	if err != nil {
		return
	}

	wrapper.c.Set(key, value, cache.DefaultExpiration)
	log.Printf("📦 Cache SET for key: %s\n", key)

	return
}
