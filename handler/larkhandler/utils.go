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
	// 获取GuildID(群聊)下的所有启用方法，缓存Key=GuildID
	queryDatas, _ := database.FindByCacheFunc(
		database.FunctionEnabling{GuildID: chatID},
		func(d database.FunctionEnabling) string {
			return d.GuildID
		},
	)
	for _, data := range queryDatas {
		if data.Function == function {
			return true
		}
	}
	return false
}