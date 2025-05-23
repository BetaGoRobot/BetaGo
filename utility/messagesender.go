package utility

import (
	"context"

	"github.com/BetaGoRobot/BetaGo/consts"
	"github.com/BetaGoRobot/BetaGo/utility/log"
	"github.com/BetaGoRobot/BetaGo/utility/otel"
	"github.com/enescakir/emoji"
	"github.com/kevinmatthe/zaplog"
	"github.com/lonelyevil/kook"
)

// SendMessageTempAndDelete 1
//
//	@param targetID
//	@param QuoteID
//	@param authorID
//	@param message
func SendMessageTempAndDelete(targetID, QuoteID, authorID, newMsg string) {
	cardMessageStr, err := BuildCardMessage(
		"danger",
		"lg",
		"重启更新",
		QuoteID, nil, kook.CardMessageSection{
			Text: kook.CardMessageElementKMarkdown{
				Content: newMsg,
			},
		},
		kook.CardMessageDivider{},
		kook.CardMessageSection{
			Text: kook.CardMessageElementKMarkdown{
				Content: "> 注意：这是一条临时消息，仅你可见",
			},
		})
	if err != nil {
		log.Zlog.Error("发送消息错误: ", zaplog.Error(err))
		return
	}
	_, err = consts.GlobalSession.MessageCreate(
		&kook.MessageCreate{
			MessageCreateBase: kook.MessageCreateBase{
				Type:     kook.MessageTypeCard,
				TargetID: targetID,
				Content:  cardMessageStr,
			},
			TempTargetID: authorID,
		},
	)
	consts.GlobalSession.MessageDelete(QuoteID)
}

// SendMessageTemp 发送消息
//
//	@param targetID 目标ID
//	@param QuoteID 引用ID
//	@param authorID 作者ID
//	@param err 错误信息
func SendMessageTemp(targetID, QuoteID, authorID, newMsg string) {
	cardMessageStr, err := kook.CardMessage{
		&kook.CardMessageCard{
			Theme: "danger",
			Size:  "lg",
			Modules: []interface{}{
				kook.CardMessageHeader{
					Text: kook.CardMessageElementText{
						Content: emoji.Information.String() + " Message:",
						Emoji:   true,
					},
				},
				kook.CardMessageDivider{},
				kook.CardMessageSection{
					Text: kook.CardMessageElementKMarkdown{
						Content: newMsg,
					},
				},
				kook.CardMessageDivider{},
				kook.CardMessageSection{
					Text: kook.CardMessageElementKMarkdown{
						Content: "> 注意：这是一条临时消息，仅你可见",
					},
				},
			},
		},
	}.BuildMessage()
	if err != nil {
		log.Zlog.Error("发送消息错误: ", zaplog.Error(err))
		return
	}
	consts.GlobalSession.MessageCreate(
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
		log.Zlog.Error("发送消息错误: ", zaplog.Error(err))
		return
	}
	consts.GlobalSession.MessageCreate(
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

func SendMessageWithTitle(targetID, QuoteID, authorID, message, title string, ctx context.Context) (msgID string) {
	cardMessageStr, err := kook.CardMessage{
		&kook.CardMessageCard{
			Theme: kook.CardThemeSecondary,
			Size:  "lg",
			Modules: []interface{}{
				kook.CardMessageHeader{
					Text: kook.CardMessageElementText{
						Content: title,
						Emoji:   true,
					},
				},
				kook.CardMessageDivider{},
				kook.CardMessageSection{
					Text: kook.CardMessageElementKMarkdown{
						Content: message,
					},
				},
			},
		},
	}.BuildMessage()
	if err != nil {
		log.Zlog.Error("发送消息错误: ", zaplog.Error(err))
		return
	}
	resp, err := consts.GlobalSession.MessageCreate(
		&kook.MessageCreate{
			MessageCreateBase: kook.MessageCreateBase{
				Type:     kook.MessageTypeCard,
				TargetID: targetID,
				Content:  cardMessageStr,
				Quote:    QuoteID,
			},
		},
	)
	if err != nil {
		log.Zlog.Error("Send msg failed", zaplog.Error(err))
	}
	return resp.MsgID
}

// SendErrorMessageWithTitle 发送消息
//
//	@param targetID 目标ID
//	@param QuoteID 引用ID
//	@param authorID 作者ID
//	@param message
func SendErrorMessageWithTitle(targetID, QuoteID, authorID, message, title string, ctx context.Context) {
	ctx, span := otel.BetaGoOtelTracer.Start(ctx, GetCurrentFunc())
	defer span.End()

	cardMessageStr, err := kook.CardMessage{
		&kook.CardMessageCard{
			Theme: "danger",
			Size:  "sm",
			Modules: []interface{}{
				kook.CardMessageHeader{
					Text: kook.CardMessageElementText{
						Content: title,
						Emoji:   true,
					},
				},
				&kook.CardMessageDivider{},
				GenerateTraceButtonSection(span.SpanContext().TraceID().String()),
			},
		},
	}.BuildMessage()
	if err != nil {
		log.Zlog.Error("发送消息错误: ", zaplog.Error(err))
		return
	}
	consts.GlobalSession.MessageCreate(
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
