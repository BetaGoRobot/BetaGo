package gpt3

import (
	"context"
	"strings"

	"github.com/BetaGoRobot/BetaGo/betagovar"
	"github.com/BetaGoRobot/BetaGo/utility"
	"github.com/BetaGoRobot/BetaGo/utility/jaeger_client"
	"github.com/enescakir/emoji"
	"github.com/lonelyevil/kook"
	"go.opentelemetry.io/otel/attribute"
)

// ClientHandler 随机抽取一个数字
//
// @param targetID 目标ID
// @param quoteID 引用ID
// @param authorID 发送者ID
// @return err 错误信息
func ClientHandler(ctx context.Context, targetID, quoteID, authorID string, args ...string) (err error) {
	ctx, span := jaeger_client.BetaGoCommandTracer.Start(ctx, utility.GetCurrentFunc())
	span.SetAttributes(attribute.Key("targetID").String(targetID), attribute.Key("quoteID").String(quoteID), attribute.Key("authorID").String(authorID), attribute.Key("args").StringSlice(args))
	defer span.RecordError(err)
	defer span.End()

	msg := strings.Join(args, " ")
	res, err := CreateChatCompletion(ctx, msg, authorID)
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
						Content: emoji.Robot.String() + "GPT来帮你",
						Emoji:   false,
					},
				},
				kook.CardMessageSection{
					Text: kook.CardMessageElementKMarkdown{
						Content: res,
					},
				},
				&kook.CardMessageDivider{},
				&kook.CardMessageSection{
					Mode: kook.CardMessageSectionModeRight,
					Text: &kook.CardMessageElementKMarkdown{
						Content: "TraceID: `" + span.SpanContext().TraceID().String() + "`",
					},
					Accessory: kook.CardMessageElementButton{
						Theme: kook.CardThemeSuccess,
						Value: "https://jaeger.kevinmatt.top/trace/" + span.SpanContext().TraceID().String(),
						Click: "link",
						Text:  "链路追踪",
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
