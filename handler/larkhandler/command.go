package larkhandler

import (
	"context"

	commandcli "github.com/BetaGoRobot/BetaGo/handler/larkhandler/command_cli"
	"github.com/BetaGoRobot/BetaGo/utility"
	"github.com/BetaGoRobot/BetaGo/utility/larkutils"
	"github.com/BetaGoRobot/BetaGo/utility/otel"
	larkcore "github.com/larksuite/oapi-sdk-go/v3/core"
	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
	"github.com/pkg/errors"
	"go.opentelemetry.io/otel/attribute"
)

var _ LarkMsgOperator = &CommandOperator{}

// CommandOperator Repeat
type CommandOperator struct {
	LarkMsgOperatorBase
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
func (r *CommandOperator) PreRun(ctx context.Context, event *larkim.P2MessageReceiveV1) (err error) {
	ctx, span := otel.LarkRobotOtelTracer.Start(ctx, utility.GetCurrentFunc())
	defer span.End()
	if !larkutils.IsMentioned(event.Event.Message.Mentions) {
		return errors.Wrap(ErrStageSkip, "CommandOperator: Not Mentioned")
	}
	if event.Event.Message.ParentId == nil {
		return errors.Wrap(ErrStageSkip, "CommandOperator: No ParentId")
	}

	// Get Command

	return
}

// Run  Repeat
//
//	@receiver r
//	@param ctx
//	@param event
//	@return err
func (r *CommandOperator) Run(ctx context.Context, event *larkim.P2MessageReceiveV1) (err error) {
	ctx, span := otel.LarkRobotOtelTracer.Start(ctx, utility.GetCurrentFunc())
	span.SetAttributes(attribute.Key("event").String(larkcore.Prettify(event)))
	defer span.End()
	_ = commandcli.Re
	// parentChat := larkutils.GetMsgByID(ctx, *event.Event.Message.ParentId)

	return nil
}
