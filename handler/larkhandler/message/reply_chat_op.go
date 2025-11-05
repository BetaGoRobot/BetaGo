package message

import (
	"context"
	"strings"

	"github.com/BetaGoRobot/BetaGo/consts"
	handlerbase "github.com/BetaGoRobot/BetaGo/handler/handler_base"
	"github.com/BetaGoRobot/BetaGo/handler/larkhandler/lark_command/handlers"
	"github.com/BetaGoRobot/BetaGo/utility/larkutils"
	"github.com/BetaGoRobot/BetaGo/utility/logs"
	"github.com/BetaGoRobot/BetaGo/utility/otel"
	"github.com/BetaGoRobot/go_utils/reflecting"
	larkcore "github.com/larksuite/oapi-sdk-go/v3/core"
	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
	"github.com/pkg/errors"
	"go.opentelemetry.io/otel/attribute"
)

var _ Op = &ReplyChatOperator{}

// ReplyChatOperator Repeat
//
//	@author heyuhengmatt
//	@update 2024-07-17 01:36:07
type ReplyChatOperator struct {
	OpBase
}

// PreRun Music
//
//	@receiver r *MusicMsgOperator
//	@param ctx context.Context
//	@param event *larkim.P2MessageReceiveV1
//	@return err error
//	@author heyuhengmatt
//	@update 2024-07-17 01:34:09
func (r *ReplyChatOperator) PreRun(ctx context.Context, event *larkim.P2MessageReceiveV1, meta *handlerbase.BaseMetaData) (err error) {
	ctx, span := otel.LarkRobotOtelTracer.Start(ctx, reflecting.GetCurrentFunc())
	defer span.End()
	defer func() { span.RecordError(err) }()
	if !larkutils.IsMentioned(event.Event.Message.Mentions) {
		return errors.Wrap(consts.ErrStageSkip, "MusicMsgOperator: Not Mentioned")
	}
	if larkutils.IsCommand(ctx, larkutils.PreGetTextMsg(ctx, event)) {
		return errors.Wrap(consts.ErrStageSkip, "MusicMsgOperator: Is Command")
	}
	return
}

// Run  Repeat
//
//	@receiver r
//	@param ctx
//	@param event
//	@return err
func (r *ReplyChatOperator) Run(ctx context.Context, event *larkim.P2MessageReceiveV1, meta *handlerbase.BaseMetaData) (err error) {
	ctx, span := otel.LarkRobotOtelTracer.Start(ctx, reflecting.GetCurrentFunc())
	span.SetAttributes(attribute.Key("event").String(larkcore.Prettify(event)))
	defer span.End()
	defer func() { span.RecordError(err) }()
	defer span.RecordError(err)

	reactionID, err := larkutils.AddReaction(ctx, "OnIt", *event.Event.Message.MessageId)
	if err != nil {
		logs.L.Error().Ctx(ctx).Err(err).Msg("Add reaction to msg failed")
	} else {
		defer larkutils.RemoveReaction(ctx, reactionID, *event.Event.Message.MessageId)
	}
	defer larkutils.AddReactionAsync(ctx, "DONE", *event.Event.Message.MessageId)

	msg := larkutils.PreGetTextMsg(ctx, event)
	msg = larkutils.TrimAtMsg(ctx, msg)

	return handlers.ChatHandler("chat")(ctx, event, meta, strings.Split(msg, " ")...)
}
