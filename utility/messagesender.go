package utility

import (
	"github.com/BetaGoRobot/BetaGo/betagovar"
	"github.com/enescakir/emoji"
	"github.com/heyuhengmatt/zaplog"
	"github.com/lonelyevil/kook"
)

// SendMessageTemp 发送消息
//
//	@param targetID 目标ID
//	@param QuoteID 引用ID
//	@param authorID 作者ID
//	@param err 错误信息
func SendMessageTemp(targetID, QuoteID, authorID string, message string) {
	cardMessageStr, err := kook.CardMessage{
		&kook.CardMessageCard{
			Theme: "danger",
			Size:  "lg",
			Modules: []interface{}{
				kook.CardMessageHeader{
					Text: kook.CardMessageElementText{
						Content: emoji.Information.String() + "Message:",
						Emoji:   true,
					},
				},
				kook.CardMessageSection{
					Text: kook.CardMessageElementKMarkdown{
						Content: message + "\n" + "注意：这是一条临时消息，仅你可见",
					},
				},
			},
		},
	}.BuildMessage()
	if err != nil {
		ZapLogger.Error("发送消息错误: ", zaplog.Error(err))
		return
	}
	betagovar.GlobalSession.MessageCreate(
		&kook.MessageCreate{
			MessageCreateBase: kook.MessageCreateBase{
				Type:     kook.MessageTypeCard,
				TargetID: targetID,
				Content:  cardMessageStr,
				Quote:    QuoteID,
			},
			TempTargetID: authorID,
		},
	)
}

// SendMessage 发送消息
//
//	@param targetID 目标ID
//	@param QuoteID 引用ID
//	@param authorID 作者ID
//	@param err 错误信息
func SendMessage(targetID, QuoteID, authorID string, message string) {
	cardMessageStr, err := kook.CardMessage{
		&kook.CardMessageCard{
			Theme: "danger",
			Size:  "lg",
			Modules: []interface{}{
				kook.CardMessageHeader{
					Text: kook.CardMessageElementText{
						Content: emoji.Information.String() + "Message:",
						Emoji:   true,
					},
				},
				kook.CardMessageSection{
					Text: kook.CardMessageElementKMarkdown{
						Content: message,
					},
				},
			},
		},
	}.BuildMessage()
	if err != nil {
		ZapLogger.Error("发送消息错误: ", zaplog.Error(err))
		return
	}
	betagovar.GlobalSession.MessageCreate(
		&kook.MessageCreate{
			MessageCreateBase: kook.MessageCreateBase{
				Type:     kook.MessageTypeCard,
				TargetID: targetID,
				Content:  cardMessageStr,
				Quote:    QuoteID,
			},
		},
	)
}

// SendMessageWithTitle 发送消息
//
//	@param targetID 目标ID
//	@param QuoteID 引用ID
//	@param authorID 作者ID
//	@param message
func SendMessageWithTitle(targetID, QuoteID, authorID, message, title string) {
	cardMessageStr, err := kook.CardMessage{
		&kook.CardMessageCard{
			Theme: "danger",
			Size:  "lg",
			Modules: []interface{}{
				kook.CardMessageHeader{
					Text: kook.CardMessageElementText{
						Content: title,
						Emoji:   true,
					},
				},
				kook.CardMessageSection{
					Text: kook.CardMessageElementKMarkdown{
						Content: message,
					},
				},
			},
		},
	}.BuildMessage()
	if err != nil {
		ZapLogger.Error("发送消息错误: ", zaplog.Error(err))
		return
	}
	betagovar.GlobalSession.MessageCreate(
		&kook.MessageCreate{
			MessageCreateBase: kook.MessageCreateBase{
				Type:     kook.MessageTypeCard,
				TargetID: targetID,
				Content:  cardMessageStr,
				Quote:    QuoteID,
			},
		},
	)
}
