package errorsender

import (
	"log"

	"github.com/BetaGoRobot/BetaGo/betagovar"
	"github.com/enescakir/emoji"
	"github.com/lonelyevil/khl"
)

// SendErrorInfo 发送错误信息
//  @param targetID 目标ID
//  @param QuoteID 引用ID
//  @param authorID 作者ID
//  @param err 错误信息
func SendErrorInfo(targetID, QuoteID, authorID string, err error) {
	cardMessageStr, err := khl.CardMessage{
		&khl.CardMessageCard{
			Theme: "danger",
			Size:  "lg",
			Modules: []interface{}{
				khl.CardMessageHeader{
					Text: khl.CardMessageElementText{
						Content: emoji.Warning.String() + " Command Error",
						Emoji:   true,
					},
				},
				khl.CardMessageSection{
					Text: khl.CardMessageElementKMarkdown{
						Content: err.Error(),
					},
				},
			},
		},
	}.BuildMessage()
	if err != nil {
		log.Println(err.Error())
		return
	}
	betagovar.GlobalSession.MessageCreate(&khl.MessageCreate{
		MessageCreateBase: khl.MessageCreateBase{
			Type:     khl.MessageTypeCard,
			TargetID: targetID,
			Content:  cardMessageStr,
			Quote:    QuoteID,
		},
		TempTargetID: authorID,
	})
}
