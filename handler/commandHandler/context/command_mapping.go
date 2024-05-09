package context

import (
	"github.com/BetaGoRobot/BetaGo/consts"
	"github.com/BetaGoRobot/BetaGo/handler/commandHandler/admin"
	"github.com/BetaGoRobot/BetaGo/handler/commandHandler/cal"
	"github.com/BetaGoRobot/BetaGo/handler/commandHandler/dailyrate"
	"github.com/BetaGoRobot/BetaGo/handler/commandHandler/gpt3"
	"github.com/BetaGoRobot/BetaGo/handler/commandHandler/helper"
	"github.com/BetaGoRobot/BetaGo/handler/commandHandler/hitokoto"
	"github.com/BetaGoRobot/BetaGo/handler/commandHandler/music"
	"github.com/BetaGoRobot/BetaGo/handler/commandHandler/news"
	"github.com/BetaGoRobot/BetaGo/handler/commandHandler/roll"
)

var commandMapping = map[string]CommandContextFunc{
	CommandContextTypeRestart:      admin.RestartHandler,
	CommandContextTypeTryPanic:     helper.TryPanic,
	consts.ShortCommandHelp:        helper.CommandRouter,
	CommandContextTypeHelper:       helper.CommandRouter,
	consts.ShortCommandAddAdmin:    admin.AddAdminHandler,
	CommandContextTypeAddAdmin:     admin.AddAdminHandler,
	consts.ShortCommandRemoveAdmin: admin.RemoveAdminHandler,
	CommandContextTypeRemoveAdmin:  admin.RemoveAdminHandler,
	consts.ShortCommandShowAdmin:   admin.ShowAdminHandler,
	CommandContextTypeShowAdmin:    admin.ShowAdminHandler,
	CommandContextTypeDeleteAll:    admin.DeleteAllMessageHandler,
	CommandContextReconnect:        admin.ReconnectHandler,
	consts.ShortCommandReconnect:   admin.ReconnectHandler,
	CommandContextTypeRoll:         roll.RandRollHandler,
	consts.ShortCommandRoll:        roll.RandRollHandler,
	CommandContextTypeGPT:          gpt3.ClientHandlerStream,
	consts.ShortCommandPing:        helper.PingHandler,
	CommandContextTypePing:         helper.PingHandler,
	consts.ShortCommandHitokoto:    hitokoto.GetHitokotoHandler,
	CommandContextTypeHitokoto:     hitokoto.GetHitokotoHandler,
	consts.ShortCommandMusic:       music.SearchMusicByRobot,
	CommandContextTypeMusic:        music.SearchMusicByRobot,
	CommandContextTypeNews:         news.Handler,
	CommandContextTypeDailyRate:    dailyrate.GetRateHandler,
}

var commandMappingWithGuildID = map[string]CommandContextWithGuildIDFunc{
	CommandContextTypeUser:          helper.GetUserInfoHandler,
	CommandContextTypeCal:           cal.ShowCalHandler,
	CommandContextTypeCalLocal:      cal.ShowCalLocalHandler,
	consts.ShortCommandShowCalLocal: cal.ShowCalLocalHandler,
	consts.ShortCommandShowCal:      cal.ShowCalHandler,
}
