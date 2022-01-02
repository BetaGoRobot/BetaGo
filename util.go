package main

import (
	"fmt"
	"net"
	"strings"
	"time"
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

func GetCurrentTime() (localTime time.Time) {
	timeStr := time.Now().Format("2006-01-02 15:04:05")
	location, _ := time.LoadLocation("local")
	localTime, _ = time.ParseInLocation("2006-01-02 15:04:05", timeStr, location)
	return
}
