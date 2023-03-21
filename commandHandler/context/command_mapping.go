package context

import (
	"github.com/BetaGoRobot/BetaGo/betagovar"
	"github.com/BetaGoRobot/BetaGo/commandHandler/admin"
	"github.com/BetaGoRobot/BetaGo/commandHandler/cal"
	"github.com/BetaGoRobot/BetaGo/commandHandler/dailyrate"
	"github.com/BetaGoRobot/BetaGo/commandHandler/gpt3"
	"github.com/BetaGoRobot/BetaGo/commandHandler/helper"
	"github.com/BetaGoRobot/BetaGo/commandHandler/hitokoto"
	"github.com/BetaGoRobot/BetaGo/commandHandler/music"
	"github.com/BetaGoRobot/BetaGo/commandHandler/news"
	"github.com/BetaGoRobot/BetaGo/commandHandler/roll"
)

var commandMapping = map[string]CommandContextFunc{
	CommandContextTypeTryPanic:        helper.TryPanic,
	betagovar.ShortCommandHelp:        helper.CommandRouter,
	CommandContextTypeHelper:          helper.CommandRouter,
	betagovar.ShortCommandAddAdmin:    admin.AddAdminHandler,
	CommandContextTypeAddAdmin:        admin.AddAdminHandler,
	betagovar.ShortCommandRemoveAdmin: admin.RemoveAdminHandler,
	CommandContextTypeRemoveAdmin:     admin.RemoveAdminHandler,
	betagovar.ShortCommandShowAdmin:   admin.ShowAdminHandler,
	CommandContextTypeShowAdmin:       admin.ShowAdminHandler,
	CommandContextTypeDeleteAll:       admin.DeleteAllMessageHandler,
	CommandContextReconnect:           admin.ReconnectHandler,
	betagovar.ShortCommandReconnect:   admin.ReconnectHandler,
	CommandContextTypeRoll:            roll.RandRollHandler,
	betagovar.ShortCommandRoll:        roll.RandRollHandler,
	CommandContextTypeGPT:             gpt3.ClientHandler,
	betagovar.ShortCommandPing:        helper.PingHandler,
	CommandContextTypePing:            helper.PingHandler,
	betagovar.ShortCommandHitokoto:    hitokoto.GetHitokotoHandler,
	CommandContextTypeHitokoto:        hitokoto.GetHitokotoHandler,
	betagovar.ShortCommandMusic:       music.SearchMusicByRobot,
	CommandContextTypeMusic:           music.SearchMusicByRobot,
	CommandContextTypeNews:            news.Handler,
	CommandContextTypeDailyRate:       dailyrate.GetRateHandler,
}

var commandMappingWithGuildID = map[string]CommandContextWithGuildIDFunc{
	CommandContextTypeUser:             helper.GetUserInfoHandler,
	CommandContextTypeCal:              cal.ShowCalHandler,
	CommandContextTypeCalLocal:         cal.ShowCalLocalHandler,
	betagovar.ShortCommandShowCalLocal: cal.ShowCalLocalHandler,
	betagovar.ShortCommandShowCal:      cal.ShowCalHandler,
}
