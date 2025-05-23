package helper

import (
	"context"

	"github.com/BetaGoRobot/BetaGo/consts"
	"github.com/BetaGoRobot/BetaGo/utility/otel"
	"github.com/BetaGoRobot/go_utils/reflecting"
	"github.com/enescakir/emoji"
	"github.com/lonelyevil/kook"
	"go.opentelemetry.io/otel/attribute"
)

// PingHandler  检查机器人是否运行正常
//
//	@param targetID
//	@param qouteID
//	@param authorID
func PingHandler(ctx context.Context, targetID, quoteID, authorID string, args ...string) (err error) {
	ctx, span := otel.BetaGoOtelTracer.Start(ctx, reflecting.GetCurrentFunc())
	span.SetAttributes(attribute.Key("targetID").String(targetID), attribute.Key("quoteID").String(quoteID), attribute.Key("authorID").String(authorID), attribute.Key("args").StringSlice(args))
	defer span.RecordError(err)
	defer span.End()

	consts.GlobalSession.MessageCreate(&kook.MessageCreate{
		MessageCreateBase: kook.MessageCreateBase{
			Type:     kook.MessageTypeKMarkdown,
			TargetID: targetID,
			Content:  emoji.WavingHand.String() + "pong~",
			Quote:    quoteID,
		},
		TempTargetID: authorID,
	})
	return nil
}
