package larkhandler

import (
	"context"
	"strconv"
	"time"

	"github.com/BetaGoRobot/BetaGo/consts"
	"github.com/BetaGoRobot/BetaGo/handler/larkhandler/message"
	"github.com/BetaGoRobot/BetaGo/handler/larkhandler/reaction"
	"github.com/BetaGoRobot/BetaGo/utility/larkutils"
	"github.com/BetaGoRobot/BetaGo/utility/log"
	"github.com/BetaGoRobot/BetaGo/utility/logging"
	"github.com/BetaGoRobot/BetaGo/utility/otel"
	"github.com/BetaGoRobot/go_utils/reflecting"
	larkcore "github.com/larksuite/oapi-sdk-go/v3/core"
	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
	"go.opentelemetry.io/otel/attribute"
)

func isOutDated(createTime string) bool {
	stamp, err := strconv.ParseInt(createTime, 10, 64)
	if err != nil {
		panic(err)
	}
	return time.Now().Sub(time.UnixMilli(stamp)) > time.Second*10
}

// MessageV2Handler Repeat
//
//	@param ctx
//	@param event
//	@return error
func MessageV2Handler(ctx context.Context, event *larkim.P2MessageReceiveV1) (err error) {
	ctx, span := otel.LarkRobotOtelTracer.Start(ctx, reflecting.GetCurrentFunc())
	defer larkutils.RecoverMsg(ctx, *event.Event.Message.MessageId)
	span.SetAttributes(attribute.Key("event").String(larkcore.Prettify(event)))
	defer span.End()
	defer func() { span.RecordError(err) }()

	if isOutDated(*event.Event.Message.CreateTime) {
		return nil
	}
	if *event.Event.Sender.SenderId.OpenId == consts.BotOpenID {
		return nil
	}
	logging.Logger.Info().Ctx(ctx).Str("event", larkcore.Prettify(event)).Msg("Inside the child span for complex handler")
	go message.Handler.Clean().WithCtx(ctx).WithEvent(event).Run()

	log.Zlog.Info(larkcore.Prettify(event))
	return nil
}

// MessageReactionHandler Repeat
//
//	@param ctx
//	@param event
//	@return error
func MessageReactionHandler(ctx context.Context, event *larkim.P2MessageReactionCreatedV1) (err error) {
	ctx, span := otel.LarkRobotOtelTracer.Start(ctx, reflecting.GetCurrentFunc())
	defer larkutils.RecoverMsg(ctx, *event.Event.MessageId)
	defer span.End()
	defer func() { span.RecordError(err) }()

	go reaction.Handler.Clean().WithCtx(ctx).WithEvent(event).Run()
	return nil
}
