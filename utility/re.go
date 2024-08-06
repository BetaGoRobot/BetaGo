package utility

import (
	"context"
	"time"

	"github.com/BetaGoRobot/BetaGo/utility/log"
	"github.com/dlclark/regexp2"
	gocache "github.com/eko/gocache/lib/v4/cache"
	gocache_store "github.com/eko/gocache/store/go_cache/v4"
	"github.com/kevinmatthe/zaplog"
	"github.com/patrickmn/go-cache"
)

var repatternCache *gocache.Cache[*regexp2.Regexp]

func init() {
	gocacheClient := cache.New(time.Second*30, time.Second*30)
	gocacheStore := gocache_store.NewGoCache(gocacheClient)
	repatternCache = gocache.New[*regexp2.Regexp](gocacheStore)
}

func RegexpMatch(str, pattern string) bool {
	if v, err := repatternCache.Get(context.Background(), pattern); err == nil {
		res, _ := v.MatchString(str)
		return res
	}
	re, err := regexp2.Compile(pattern, 0)
	if err != nil {
		log.ZapLogger.Error("compile regexp error", zaplog.Error(err))
		return false
	}
	repatternCache.Set(context.Background(), pattern, re)
	res, _ := re.MatchString(str)
	return res
}
