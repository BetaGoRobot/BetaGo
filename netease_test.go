package main

import (
	"fmt"
	"log"
	"testing"
)

func TestGetWithCookieParams(t *testing.T) {
	neteaseCtx := &NetEaseContext{}
	err := neteaseCtx.loginNetEase()
	if err != nil {
		log.Fatal(err.Error())
	}
	IDs, err := neteaseCtx.getDailyRecommendID()
	if err != nil {
		log.Fatal(err.Error())
	}
	URLs, err := neteaseCtx.getMusicURLByID(IDs)
	if err != nil {
		log.Fatal(err.Error())
	}
	for _, URL := range URLs {
		fmt.Println(URL.URL)
	}
	neteaseCtx.searchMusicByKeyWord([]string{"Closer"})
}
