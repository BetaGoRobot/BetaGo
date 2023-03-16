package manager

import (
	"github.com/BetaGoRobot/BetaGo/betagovar"
	"github.com/lonelyevil/kook"
)

// SendMessageToTestChannel  is a async handler for message event
//
//	@param session
//	@param content
func SendMessageToTestChannel(session *kook.Session, content string) {
	session.MessageCreate(&kook.MessageCreate{
		MessageCreateBase: kook.MessageCreateBase{
			Type:     9,
			TargetID: betagovar.TestChanID,
			Content:  content,
		},
	})
}
