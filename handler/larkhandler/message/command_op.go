package message

import (
	"context"
	"fmt"

	"github.com/BetaGoRobot/BetaGo/consts"
	handlerbase "github.com/BetaGoRobot/BetaGo/handler/handler_base"
	larkcommand "github.com/BetaGoRobot/BetaGo/handler/larkhandler/lark_command"
	"github.com/BetaGoRobot/BetaGo/utility/larkutils"
	"github.com/BetaGoRobot/BetaGo/utility/log"
	"github.com/BetaGoRobot/BetaGo/utility/otel"
	"github.com/BetaGoRobot/go_utils/reflecting"
	"github.com/kevinmatthe/zaplog"
	larkcore "github.com/larksuite/oapi-sdk-go/v3/core"
	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
	"github.com/pkg/errors"
	"go.opentelemetry.io/otel/attribute"
)

var _ Op = &CommandOperator{}

// CommandOperator Repeat
type CommandOperator struct {
	OpBase
	command string
}

// PreRun Music
//
//	@receiver r *MusicMsgOperator
//	@param ctx context.Context
//	@param event *larkim.P2MessageReceiveV1
//	@return err error
//	@author heyuhengmatt
//	@update 2024-07-17 01:34:09
func (r *CommandOperator) PreRun(ctx context.Context, event *larkim.P2MessageReceiveV1, meta *handlerbase.BaseMetaData) (err error) {
	ctx, span := otel.LarkRobotOtelTracer.Start(ctx, reflecting.GetCurrentFunc())
	defer span.End()
	defer span.RecordError(err)

	if !larkutils.IsCommand(ctx, larkutils.PreGetTextMsg(ctx, event)) {
		return errors.Wrap(consts.ErrStageSkip, "CommandOperator: Not Command")
	}
	return
}

// Run  Repeat
//
//	@receiver r
//	@param ctx
//	@param event
//	@return err
func (r *CommandOperator) Run(ctx context.Context, event *larkim.P2MessageReceiveV1, meta *handlerbase.BaseMetaData) (err error) {
	ctx, span := otel.LarkRobotOtelTracer.Start(ctx, reflecting.GetCurrentFunc())
	span.SetAttributes(attribute.Key("event").String(larkcore.Prettify(event)))
	defer span.End()
	defer span.RecordError(err)

	commands := larkutils.GetCommand(ctx, larkutils.PreGetTextMsg(ctx, event))
	if len(commands) > 0 {
		meta.IsCommand = true
		var reactionID string
		reactionID, err = larkutils.AddReaction(ctx, "OnIt", *event.Event.Message.MessageId)
		if err != nil {
			log.Zlog.Error("Add reaction to msg failed", zaplog.Error(err))
		} else {
			defer larkutils.RemoveReaction(ctx, reactionID, *event.Event.Message.MessageId)
		}
		err = larkcommand.LarkRootCommand.Execute(ctx, event, commands)
		if err != nil {
			span.RecordError(err)
			if errors.Is(err, consts.ErrCommandNotFound) {
				meta.IsCommand = false
				if larkutils.IsMentioned(event.Event.Message.Mentions) {
					larkutils.ReplyCardText(ctx, err.Error(), *event.Event.Message.MessageId, "_OpErr", true)
					return
				}
			} else {
				text := fmt.Sprintf("%v\n[Jaeger Trace](https://jaeger.kmhomelab.cn/trace/%s)", err.Error(), span.SpanContext().TraceID().String())
				larkutils.ReplyCardText(ctx, text, *event.Event.Message.MessageId, "_OpErr", false)
				log.Zlog.Error("CommandOperator", zaplog.Error(err), zaplog.String("TraceID", span.SpanContext().TraceID().String()))
				return
			}
		}
		larkutils.AddReactionAsync(ctx, "DONE", *event.Event.Message.MessageId)
	}

	return nil
}
