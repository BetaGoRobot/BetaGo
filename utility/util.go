package utility

import (
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/BetaGoRobot/BetaGo/betagovar"
	"github.com/lonelyevil/khl"
)

//GetOutBoundIP 获取机器人部署的当前ip
func GetOutBoundIP() (ip string, err error) {
	conn, err := net.Dial("udp", "101.132.154.52:80")
	if err != nil {
		fmt.Println(err)
		return
	}
	localAddr := conn.LocalAddr().(*net.UDPAddr)
	fmt.Println(localAddr.String())
	ip = strings.Split(localAddr.String(), ":")[0]
	return
}

//GetCurrentTime 获取当前时间
func GetCurrentTime() (localTime string) {
	timeLocal, _ := time.LoadLocation("Asia/Shanghai")
	time.Local = timeLocal
	localTime = time.Now().Local().Format("2006-01-02 15:04:05")
	return
}

// ForDebug 用于测试
func ForDebug(test ...interface{}) {
	return
}

// IsInSlice 判断机器人是否被at到
//  @param target
//  @param slice
//  @return bool
func IsInSlice(target string, slice []string) bool {
	for i := range slice {
		if slice[i] == target {
			return true
		}
	}
	return false
}

// MustAtoI 将字符串转换为int
//  @param str
//  @return int
func MustAtoI(str string) int {
	i, err := strconv.Atoi(str)
	if err != nil {
		return i
	}
	return i
}

// GetUserInfo 获取用户信息
//  @param userID
//  @param guildID
//  @return userInfo
func GetUserInfo(userID, guildID string) (userInfo *khl.User, err error) {
	userInfo, err = betagovar.GlobalSession.UserView(userID, khl.UserViewWithGuildID(guildID))
	if err != nil {
		return
	}
	return
}
