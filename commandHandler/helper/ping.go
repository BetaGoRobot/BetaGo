package helper

import (
	"context"

	"github.com/BetaGoRobot/BetaGo/betagovar"
	"github.com/BetaGoRobot/BetaGo/utility"
	"github.com/BetaGoRobot/BetaGo/utility/jaeger_client"
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
	ctx, span := jaeger_client.BetaGoCommandTracer.Start(ctx, utility.GetCurrentFunc())
	span.SetAttributes(attribute.Key("targetID").String(targetID), attribute.Key("quoteID").String(quoteID), attribute.Key("authorID").String(authorID), attribute.Key("args").StringSlice(args))
	defer span.RecordError(err)
	defer span.End()

	betagovar.GlobalSession.MessageCreate(&kook.MessageCreate{
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
