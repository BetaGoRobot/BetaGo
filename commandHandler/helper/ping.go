package helper

import (
	"github.com/BetaGoRobot/BetaGo/betagovar"
	"github.com/enescakir/emoji"
	"github.com/lonelyevil/khl"
)

// PingHandler  检查机器人是否运行正常
//
//	@param targetID
//	@param qouteID
//	@param authorID
func PingHandler(TargetID, QuoteID, AuthorID string, parameters ...string) error {
	betagovar.GlobalSession.MessageCreate(&khl.MessageCreate{
		MessageCreateBase: khl.MessageCreateBase{
			Type:     khl.MessageTypeKMarkdown,
			TargetID: TargetID,
			Content:  emoji.WavingHand.String() + "pong~",
			Quote:    QuoteID,
		},
		TempTargetID: AuthorID,
	})
	return nil
}
