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
	queryDatas, _ := database.FindByCache(&database.FunctionEnabling{})
	for _, data := range queryDatas {
		if data.GuildID == chatID && data.Function == function {
			return true
		}
	}
	return false
}
