package message

import (
	"context"

	"github.com/BetaGoRobot/BetaGo/consts"
	handlerbase "github.com/BetaGoRobot/BetaGo/handler/handler_base"
	"github.com/BetaGoRobot/BetaGo/handler/larkhandler/lark_command/handlers"
	"github.com/BetaGoRobot/BetaGo/utility"
	"github.com/BetaGoRobot/BetaGo/utility/larkutils"
	"github.com/BetaGoRobot/BetaGo/utility/otel"
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
	ctx, span := otel.LarkRobotOtelTracer.Start(ctx, utility.GetCurrentFunc())
	defer span.End()
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
	ctx, span := otel.LarkRobotOtelTracer.Start(ctx, utility.GetCurrentFunc())
	span.SetAttributes(attribute.Key("event").String(larkcore.Prettify(event)))
	defer span.End()
	defer span.RecordError(err)

	msg := larkutils.PreGetTextMsg(ctx, event)
	msg = larkutils.TrimAtMsg(ctx, msg)

	return handlers.ChatHandler(ctx, event, msg)
}
