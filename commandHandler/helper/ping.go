package helper

import (
	"github.com/BetaGoRobot/BetaGo/betagovar"
	"github.com/enescakir/emoji"
	"github.com/lonelyevil/kook"
)

// PingHandler  检查机器人是否运行正常
//
//	@param targetID
//	@param qouteID
//	@param authorID
func PingHandler(TargetID, QuoteID, AuthorID string, parameters ...string) error {
	betagovar.GlobalSession.MessageCreate(&kook.MessageCreate{
		MessageCreateBase: kook.MessageCreateBase{
			Type:     kook.MessageTypeKMarkdown,
			TargetID: TargetID,
			Content:  emoji.WavingHand.String() + "pong~",
			Quote:    QuoteID,
		},
		TempTargetID: AuthorID,
	})
	return nil
}
