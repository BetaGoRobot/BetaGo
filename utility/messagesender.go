package utility

import (
	"context"
	"log"

	"github.com/BetaGoRobot/BetaGo/betagovar"
	"github.com/BetaGoRobot/BetaGo/utility/jaeger_client"
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
	oldMsgView, err := betagovar.GlobalSession.MessageView(QuoteID)
	if err != nil {
		log.Println(err.Error())
		oldMsgView = &kook.DetailedChannelMessage{}
	}
	oldMsg := oldMsgView.Content
	cardMessageStr, err := kook.CardMessage{
		&kook.CardMessageCard{
			Theme: "danger",
			Size:  "lg",
			Modules: []interface{}{
				kook.CardMessageHeader{
					Text: kook.CardMessageElementText{
						Content: emoji.Information.String() + " Your Message:",
						Emoji:   true,
					},
				},
				kook.CardMessageDivider{},
				kook.CardMessageSection{
					Text: kook.CardMessageElementKMarkdown{
						Content: oldMsg,
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
		ZapLogger.Error("发送消息错误: ", zaplog.Error(err))
		return
	}
	betagovar.GlobalSession.MessageCreate(
		&kook.MessageCreate{
			MessageCreateBase: kook.MessageCreateBase{
				Type:     kook.MessageTypeCard,
				TargetID: targetID,
				Content:  cardMessageStr,
			},
			TempTargetID: authorID,
		},
	)
	betagovar.GlobalSession.MessageDelete(QuoteID)
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
		ZapLogger.Error("发送消息错误: ", zaplog.Error(err))
		return
	}
	resp, err := betagovar.GlobalSession.MessageCreate(
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
		log.Println(err)
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
	ctx, span := jaeger_client.BetaGoCommandTracer.Start(ctx, GetCurrentFunc())
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
