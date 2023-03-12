package utility

import (
	"strings"

	"github.com/BetaGoRobot/BetaGo/betagovar"
)

// GetCommandWithParameters 获取命令及参数
//
//	@param rawCommand
//	@return command
//	@return params
func GetCommandWithParameters(rawCommand string) (command string, params []string) {
	var trueContent = strings.TrimSpace(rawCommand)
	if strings.HasPrefix(rawCommand, betagovar.CommandPrefix) {
		trueContent = strings.TrimSpace(strings.Replace(rawCommand, "(met)"+betagovar.RobotID+"(met)", "", 1))
	} else {
		trueContent = strings.Trim(rawCommand, ".")
	}

	// 解析得到不包含At机器人的信息的实际内容
	// 判断是否为空字符串
	if trueContent == "" {
		return
	}

	splittedStr := strings.Split(trueContent, " ")
	if len(splittedStr) == 0 {
		return
	}

	command = strings.ToUpper(splittedStr[0])
	if len(splittedStr) > 1 {
		params = splittedStr[1:]
	}
	return
}
