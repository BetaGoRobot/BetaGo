package check

import (
	"github.com/BetaGoRobot/BetaGo/betagovar"
	"github.com/BetaGoRobot/BetaGo/manager"
	"github.com/kevinmatthe/zaplog"
)

// CheckEnv  检查环境变量
func CheckEnv() {
	if betagovar.RobotName == "" {
		manager.SendMessageToTestChannel(betagovar.GlobalSession, ">  机器人未配置名称！")
	}
	if betagovar.RobotID == "" {
		zaplog.Logger.Fatal("机器人未配置ID！")
	}
}

func CheckDbStatus() {
}
