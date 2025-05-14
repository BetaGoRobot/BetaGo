package cache

import (
	handlerbase "github.com/BetaGoRobot/BetaGo/handler/handler_base"
	"github.com/dgraph-io/ristretto/v2"
)

var MsgIDMetaCache *ristretto.Cache[string, *handlerbase.BaseMetaData]

func init() {
	var err error
	MsgIDMetaCache, err = ristretto.NewCache(
		&ristretto.Config[string, *handlerbase.BaseMetaData]{
			NumCounters: 1e7,     // number of keys to track frequency of (10M).
			MaxCost:     5 << 20, // maximum cost of cache (1GB).
			BufferItems: 64,      // number of keys per Get buffer.
		})
	if err != nil {
		panic(err)
	}
}
