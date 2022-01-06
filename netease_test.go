package main

import (
	"fmt"
	"testing"
)

func Test_loginNetEase(t *testing.T) {
	sendGetReq("http://localhost:8080", "login/cellphone", map[string]string{"phone": "18681655914", "password": "heyuheng1.22.3"})
}

func Test_sendGetReq(t *testing.T) {
	sendGetReq("http://localhost:8080", "login/status", nil)
}

func Test_sendGetReq2(t *testing.T) {
	// fmt.Println(sendGetReq("http://localhost:8080", "login/cellphone", map[string]string{"phone": "18681655914", "password": "heyuheng1.22.3"}))
	// // sendGetReq("http://localhost:3333", "login/refresh", nil)
	// fmt.Println(sendGetReq("http://localhost:8080", "artist/sublist", nil))
	// // sendGetReq("http://localhost:3333", "logout", nil)
	// // sendGetReq("http://localhost:8080", "user/account", nil)
	fmt.Println(sendGetWithHeader("https://www.bungie.net/Platform", fmt.Sprintf("/User/GetBungieNetUserById/%s", "24957719"), getHeader{headerName: "X-API-Key", headerValue: "c2c8d26846eb470ca50380ccc105e78d"}, map[string][]string{}))
}
