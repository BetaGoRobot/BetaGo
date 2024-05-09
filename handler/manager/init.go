package manager

import "github.com/BetaGoRobot/BetaGo/consts"

func init() {
	AddAllHandler()
}

func AddAllHandler() {
	consts.GlobalSession.AddHandler(MessageEventAsyncHandler)
	consts.GlobalSession.AddHandler(ClickEventAsyncHandler)
	consts.GlobalSession.AddHandler(ChannelJoinedAsyncHandler)
	consts.GlobalSession.AddHandler(ChannelLeftAsyncHandler)
}
