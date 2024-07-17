package database

import (
	"fmt"
	"strconv"
	"testing"

	"gorm.io/gorm"
)

func TestFindByCacheFunc(t *testing.T) {
	keyFunc := func(a Administrator) string {
		return strconv.Itoa(int(a.Model.ID))
	}
	res, hitCache := FindByCacheFunc(Administrator{Model: gorm.Model{ID: 1}}, keyFunc)
	fmt.Println(res, hitCache)
	res, hitCache = FindByCacheFunc(Administrator{Model: gorm.Model{ID: 1}}, keyFunc)
	fmt.Println(res, hitCache)
	_ = res
}

func BenchmarkFindByCacheFunc(b *testing.B) {
	keyFunc := func(a Administrator) string {
		return strconv.Itoa(int(a.Model.ID))
	}
	b.ReportAllocs()
	b.ResetTimer()
	b.Run("FindByCacheFunc", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			FindByCacheFunc(Administrator{}, keyFunc)
		}
	})
	b.Run("FindFuncNoCache", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			FindFuncNoCache(Administrator{})
		}
	})
}
