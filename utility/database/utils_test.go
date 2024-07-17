package database

import (
	"fmt"
	"testing"
)

func TestFindByCacheFunc(t *testing.T) {
	keyFunc := func(a Administrator) string {
		return a.UserName
	}
	res, hitCache := FindByCacheFunc(Administrator{}, keyFunc)
	fmt.Println(res, hitCache)
	res, hitCache = FindByCacheFunc(Administrator{}, keyFunc)
	fmt.Println(res, hitCache)
	_ = res
}
