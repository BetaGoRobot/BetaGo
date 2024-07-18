package larkhandler

import (
	"github.com/BetaGoRobot/BetaGo/consts"
	"github.com/BetaGoRobot/BetaGo/utility"
	"github.com/BetaGoRobot/BetaGo/utility/database"
)

func getRandomEmoji() string {
	return utility.SampleSlice(emojiTypeList)
}

func checkFunctionEnabling(chatID string, function consts.LarkFunctionEnum) bool {
	queryDatas, _ := database.FindByCacheFunc(
		database.FunctionEnabling{GuildID: chatID, Function: function},
		func(d database.FunctionEnabling) string {
			return d.GuildID + string(d.Function)
		},
	)
	for _, data := range queryDatas {
		if data.Function == function {
			return true
		}
	}
	return false
}
