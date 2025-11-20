package message

import (
	"context"
	"fmt"
	"strings"

	"github.com/BetaGoRobot/BetaGo/consts"
	handlerbase "github.com/BetaGoRobot/BetaGo/handler/handler_base"
	larkcommand "github.com/BetaGoRobot/BetaGo/handler/larkhandler/lark_command"
	"github.com/BetaGoRobot/BetaGo/utility/larkutils"
	"github.com/BetaGoRobot/BetaGo/utility/logs"
	"github.com/BetaGoRobot/BetaGo/utility/otel"
	"github.com/BetaGoRobot/go_utils/reflecting"
	larkcore "github.com/larksuite/oapi-sdk-go/v3/core"
	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
	"github.com/pkg/errors"
	"go.opentelemetry.io/otel/attribute"
	"go.uber.org/zap"
)

var _ Op = &CommandOperator{}

// CommandOperator Repeat
type CommandOperator struct {
	OpBase
	command string
}

func (r *CommandOperator) Name() string {
	return "CommandOperator"
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
	defer func() { span.RecordError(err) }()
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
	defer func() { span.RecordError(err) }()
	defer span.RecordError(err)
	rawCommand := larkutils.PreGetTextMsg(ctx, event)

	return ExecuteFromRawCommand(ctx, event, meta, rawCommand)
}

func ExecuteFromRawCommand(ctx context.Context, event *larkim.P2MessageReceiveV1, meta *handlerbase.BaseMetaData, rawCommand string) (err error) {
	ctx, span := otel.LarkRobotOtelTracer.Start(ctx, reflecting.GetCurrentFunc())
	span.SetAttributes(attribute.Key("event").String(larkcore.Prettify(event)))
	defer span.End()
	defer func() { span.RecordError(err) }()
	defer span.RecordError(err)

	rawCommand = strings.ReplaceAll(rawCommand, "<b>", " ")
	rawCommand = strings.ReplaceAll(rawCommand, "</b>", " ")
	ctx = context.WithValue(ctx, consts.ContextVarSrcCmd, rawCommand)
	commands := larkutils.GetCommand(ctx, rawCommand)
	if len(commands) > 0 {
		meta.IsCommand = true
		var reactionID string
		reactionID, err = larkutils.AddReaction(ctx, "OnIt", *event.Event.Message.MessageId)
		if err != nil {
			logs.L().Ctx(ctx).Error("Add reaction to msg failed", zap.Error(err))
		} else {
			defer larkutils.RemoveReactionAsync(ctx, reactionID, *event.Event.Message.MessageId)
		}
		err = larkcommand.LarkRootCommand.Execute(ctx, event, meta, commands)
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
				larkutils.ReplyCardText(ctx, text, *event.Event.Message.MessageId, "_OpErr", true)
				logs.L().Ctx(ctx).Error("CommandOperator", zap.Error(err), zap.String("TraceID", span.SpanContext().TraceID().String()))
				return
			}
		}
		if !meta.SkipDone {
			larkutils.AddReactionAsync(ctx, "DONE", *event.Event.Message.MessageId)
		}
	}
	return
}
