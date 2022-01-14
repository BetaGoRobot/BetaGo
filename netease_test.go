package main

import (
	"fmt"
	"os"
	"testing"
	"time"
)

func TestNetEaseContext_getNewCommendMusic(t *testing.T) {
	os.Setenv("NETEASE_PASSWORD", "heyuheng1.22.3")
	os.Setenv("NETEASE_PHONE", "18681655914")
	ctx := &NetEaseContext{}
	ctx.loginNetEase()
	ctx.getNewRecommendMusic()

}

func TestTime(t *testing.T) {
	os.Setenv("NETEASE_PASSWORD", "heyuheng1.22.3")
	os.Setenv("NETEASE_PHONE", "18681655914")
	for {
		time.Sleep(10 * time.Second)
		if string(time.Now().Local().Format("15:04:05")) == "04:54:00" {
			fmt.Println("test")
			break
		}
	}

	fmt.Println(time.Now().Local().Format("15:04:05")) // 13:57:52
}
