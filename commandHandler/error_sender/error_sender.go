package errorsender

import (
	"context"

	"github.com/BetaGoRobot/BetaGo/betagovar"
	"github.com/BetaGoRobot/BetaGo/utility"
	"github.com/BetaGoRobot/BetaGo/utility/gotify"
	"github.com/BetaGoRobot/BetaGo/utility/jaeger_client"
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
func SendErrorInfo(targetID, QuoteID, authorID string, sourceErr error, ctx context.Context) {
	ctx, span := jaeger_client.BetaGoCommandTracer.Start(ctx, utility.GetCurrentFunc())
	span.RecordError(sourceErr)
	defer span.End()
	var (
		cardMessageStr string
		err            error
	)
	if sourceErr == betagovar.ErrorOverReq {
		cardMessageStr, err = kook.CardMessage{
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
	} else {
		cardMessageStr, err = kook.CardMessage{
			&kook.CardMessageCard{
				Theme: "danger",
				Size:  "sm",
				Modules: []interface{}{
					kook.CardMessageHeader{
						Text: kook.CardMessageElementText{
							Content: emoji.Warning.String() + " Command Error: 指令错误",
							Emoji:   true,
						},
					},
					&kook.CardMessageDivider{},
					kook.CardMessageSection{
						Mode: kook.CardMessageSectionModeRight,
						Text: kook.CardMessageElementKMarkdown{
							Content: "请联系开发者并提供此ID\n\nTraceID: `" +
								span.SpanContext().TraceID().String() + "`\n",
						},
						Accessory: kook.CardMessageElementButton{
							Theme: kook.CardThemeWarning,
							Value: "https://jaeger.kevinmatt.top/trace/" + span.SpanContext().TraceID().String(),
							Click: "link",
							Text:  "链路追踪",
						},
					},
				},
			},
		}.BuildMessage()
	}

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
	gotify.SendMessage(emoji.Warning.String()+"CommandError", sourceErr.Error()+"\n"+"[追踪链接](https://jaeger.kevinmatt.top/trace/"+span.SpanContext().TraceID().String()+")", 6)
}
