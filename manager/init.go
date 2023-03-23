package manager

import "github.com/BetaGoRobot/BetaGo/betagovar"

func init() {
	AddAllHandler()
}

func AddAllHandler() {
	betagovar.GlobalSession.AddHandler(MessageEventAsyncHandler)
	betagovar.GlobalSession.AddHandler(ClickEventAsyncHandler)
	betagovar.GlobalSession.AddHandler(ChannelJoinedAsyncHandler)
	betagovar.GlobalSession.AddHandler(ChannelLeftAsyncHandler)
}
