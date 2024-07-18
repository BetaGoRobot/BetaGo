package database

import (
	"context"
	"reflect"
	"sync"
	"time"

	gocache "github.com/eko/gocache/lib/v4/cache"
	gocache_store "github.com/eko/gocache/store/go_cache/v4"
	"github.com/patrickmn/go-cache"
)

var dbDataCache *gocache.Cache[*sync.Map]

var newMap *sync.Map

func init() {
	gocacheClient := cache.New(time.Second*30, time.Second*30)
	gocacheStore := gocache_store.NewGoCache(gocacheClient)
	dbDataCache = gocache.New[*sync.Map](gocacheStore)
}

// GlobalCache  全局缓存
var GlobalCache = cache.New(time.Second*30, time.Second*30)

var typeKindMap = &sync.Map{}

func getCacheType(t reflect.Type) (name string) {
	if kind, ok := typeKindMap.Load(t); ok {
		return kind.(string)
	}

	if t.Kind() == reflect.Ptr {
		panic("Must not be pointer")
	} else {
		name = t.Name()
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

type cacheGetFunc[T any] func(T) string

// FindByCacheFunc 查找缓存
//
//	@param model T
//	@param f func(T) string
//	@return modelList []T
//	@return hitCache bool
//	@author heyuhengmatt
//	@update 2024-07-17 06:36:14
func FindByCacheFunc[T any](model T, keyFunc func(T) string) (res []T, hitCache bool) {
	res = make([]T, 0)
	modelKey := keyFunc(model)
	cacheKey := getCacheType(reflect.TypeOf(model)) + "_" + modelKey

	if cache, err := dbDataCache.Get(context.Background(), cacheKey); err == nil {
		// 取Cache
		hitCache = true
		if modelKey == "" { // 返回全部
			cache.Range(
				func(key, value any) bool {
					res = append(res, value.([]T)...)
					return true
				},
			)
			return
		}
		if tRes, ok := cache.Load(modelKey); ok {
			res = tRes.([]T)
		}
		return
	}
	// 取筛选条件对应的Key
	GetDbConnection().Find(&res, model)
	cacheValue := &sync.Map{}
	for _, r := range res {
		key := keyFunc(r)
		if v, ok := cacheValue.Load(key); ok {
			cacheValue.Store(key, append(v.([]T), r))
		} else {
			cacheValue.Store(key, []T{r})
		}
	}
	dbDataCache.Set(context.Background(), cacheKey, cacheValue)
	return
}

func FindFuncNoCache[T any](model T) (res []T) {
	res = make([]T, 0)
	// 取筛选条件对应的Key
	GetDbConnection().Find(&res, model)
	return
}
