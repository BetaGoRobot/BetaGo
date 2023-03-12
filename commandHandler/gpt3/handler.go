package gpt3

import (
	"strings"

	"github.com/BetaGoRobot/BetaGo/betagovar"
	"github.com/enescakir/emoji"
	"github.com/lonelyevil/kook"
)

// ClientHandler 随机抽取一个数字
//
// @param targetID 目标ID
// @param quoteID 引用ID
// @param authorID 发送者ID
// @return err 错误信息
func ClientHandler(targetID, quoteID, authorID string, args ...string) (err error) {
	msg := strings.Join(args, " ")
	res, err := CreateChatCompletion(msg)
	if err != nil {
		return
	}
	cardMessageStr, err := kook.CardMessage{
		&kook.CardMessageCard{
			Theme: "info",
			Size:  "lg",
			Modules: []interface{}{
				kook.CardMessageHeader{
					Text: kook.CardMessageElementText{
						Content: string(emoji.BeerMug) + "GPT来帮你",
						Emoji:   false,
					},
				},
				kook.CardMessageSection{
					Text: kook.CardMessageElementKMarkdown{
						Content: res,
					},
				},
			},
		},
	}.BuildMessage()

	_, err = betagovar.GlobalSession.MessageCreate(
		&kook.MessageCreate{
			MessageCreateBase: kook.MessageCreateBase{
				Type:     kook.MessageTypeCard,
				TargetID: targetID,
				Content:  cardMessageStr,
				Quote:    quoteID,
			},
		},
	)
	return
}
