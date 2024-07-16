package database

import (
	"reflect"
	"sync"
	"time"

	"github.com/patrickmn/go-cache"
)

// GlobalCache  全局缓存
var GlobalCache = cache.New(time.Second*30, time.Second*30)

var typeKindMap = &sync.Map{}

func getCacheType(t reflect.Type) (name string) {
	if kind, ok := typeKindMap.Load(t); ok {
		return kind.(string)
	}

	if t.Kind() == reflect.Ptr {
		name = t.Elem().Name()
	}
	actual, _ := typeKindMap.LoadOrStore(t, name)
	return actual.(string)
}

// FindByCache 查找缓存
//
//	@param model
//	@return modelList
func FindByCache[T any](model T) (modelList []T, hitCache bool) {
	modelList = make([]T, 0)
	cacheKey := getCacheType(reflect.TypeOf(model))
	if cache, ok := GlobalCache.Get(cacheKey); ok {
		modelList = cache.([]T)
		hitCache = true
		return
	}
	GetDbConnection().Find(&modelList)
	GlobalCache.Set(cacheKey, modelList, cache.DefaultExpiration)
	return
}
