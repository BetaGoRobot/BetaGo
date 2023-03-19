package errorsender

import (
	"github.com/BetaGoRobot/BetaGo/betagovar"
	"github.com/BetaGoRobot/BetaGo/utility"
	"github.com/BetaGoRobot/BetaGo/utility/gotify"
	"github.com/enescakir/emoji"
	"github.com/kevinmatthe/zaplog"
	"github.com/lonelyevil/kook"
)

// var  zapLogger = zaplog.New("errorsender")
var (
	ZapLogger   = utility.ZapLogger
	SugerLogger = utility.SugerLogger
)

// SendErrorInfo 发送错误信息
//
//	@param targetID 目标ID
//	@param QuoteID 引用ID
//	@param authorID 作者ID
//	@param err 错误信息
func SendErrorInfo(targetID, QuoteID, authorID string, sourceErr error) {
	cardMessageStr, err := kook.CardMessage{
		&kook.CardMessageCard{
			Theme: "danger",
			Size:  "lg",
			Modules: []interface{}{
				kook.CardMessageHeader{
					Text: kook.CardMessageElementText{
						Content: emoji.Warning.String() + " Command Error",
						Emoji:   true,
					},
				},
				kook.CardMessageSection{
					Text: kook.CardMessageElementKMarkdown{
						Content: sourceErr.Error(),
					},
				},
			},
		},
	}.BuildMessage()
	if err != nil {
		ZapLogger.Error("SendErrorInfo", zaplog.Error(sourceErr))
		return
	}
	betagovar.GlobalSession.MessageCreate(&kook.MessageCreate{
		MessageCreateBase: kook.MessageCreateBase{
			Type:     kook.MessageTypeCard,
			TargetID: targetID,
			Content:  cardMessageStr,
			Quote:    QuoteID,
		},
		// TempTargetID: authorID,
	})
	gotify.SendMessage(emoji.Warning.String()+"CommandError", sourceErr.Error(), 6)
}
