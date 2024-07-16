package larkhandler

import (
	"time"

	"github.com/patrickmn/go-cache"
)

var (
	repeatConfigCache   = cache.New(time.Second*30, time.Second*30)
	reactionConfigCache = cache.New(time.Second*30, time.Second*30)
	repeatWordRateCache = cache.New(time.Second*30, time.Second*30)
)
