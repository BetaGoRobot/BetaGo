package check

import (
	"github.com/BetaGoRobot/BetaGo/consts"
	"github.com/BetaGoRobot/BetaGo/handler/manager"
	"github.com/BetaGoRobot/BetaGo/utility/logs"
)

// CheckEnv  检查环境变量
func CheckEnv() {
	if consts.RobotName == "" {
		manager.SendMessageToTestChannel(consts.GlobalSession, ">  机器人未配置名称！")
	}
	if consts.RobotID == "" {
		logs.L.Fatal().Msg("机器人未配置ID！")
	}
}

func CheckDbStatus() {
}
