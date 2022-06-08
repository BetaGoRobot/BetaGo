package helper

import (
	"github.com/BetaGoRobot/BetaGo/betagovar"
	"github.com/enescakir/emoji"
	"github.com/lonelyevil/khl"
)

// PingHandler  检查机器人是否运行正常
//  @param targetID
//  @param qouteID
//  @param authorID
func PingHandler(targetID, quoteID, authorID string) {
	betagovar.GlobalSession.MessageCreate(&khl.MessageCreate{
		MessageCreateBase: khl.MessageCreateBase{
			Type:     khl.MessageTypeKMarkdown,
			TargetID: targetID,
			Content:  emoji.WavingHand.String() + "pong~",
			Quote:    quoteID,
		},
		TempTargetID: authorID,
	})
}
