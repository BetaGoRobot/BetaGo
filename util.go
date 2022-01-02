package main

import (
	"fmt"
	"net"
	"strings"
)

// 获取机器人部署的当前ip
func GetOutBoundIP() (ip string, err error) {
	conn, err := net.Dial("udp", "101.33.242.146:80")
	if err != nil {
		fmt.Println(err)
		return
	}
	localAddr := conn.LocalAddr().(*net.UDPAddr)
	fmt.Println(localAddr.String())
	ip = strings.Split(localAddr.String(), ":")[0]
	return
}
