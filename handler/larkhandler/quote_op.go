package larkhandler

import (
	"context"
	"strings"

	"github.com/BetaGoRobot/BetaGo/utility"
	"github.com/BetaGoRobot/BetaGo/utility/larkutils"
	"github.com/BetaGoRobot/BetaGo/utility/log"
	"github.com/BetaGoRobot/BetaGo/utility/otel"
	"github.com/kevinmatthe/zaplog"
	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
)

var _ LarkMsgOperator = &QuoteMsgOperator{}

type QuoteMsgOperator struct{}

// PreRun  Repeat
//
//	@receiver r
//	@param ctx
//	@param event
//	@return err
func (r *QuoteMsgOperator) PreRun(ctx context.Context, event *larkim.P2MessageReceiveV1) (err error) {
	return
}

// Run  Repeat
//
//	@receiver r
//	@param ctx
//	@param event
//	@return err
func (r *QuoteMsgOperator) Run(ctx context.Context, event *larkim.P2MessageReceiveV1) (err error) {
	ctx, span := otel.LarkRobotOtelTracer.Start(ctx, utility.GetCurrentFunc())
	defer span.End()

	msg := PreGetTextMsg(ctx, event)
	if strings.Contains(msg, "下班") {
		req := larkim.NewReplyMessageReqBuilder().
			Body(
				larkim.NewReplyMessageReqBodyBuilder().
					Content(larkim.NewTextMsgBuilder().Text("这么早你就惦记着下班了?").Build()).
					MsgType(larkim.MsgTypeText).
					ReplyInThread(true).
					Uuid(*event.Event.Message.MessageId).
					Build(),
			).MessageId(*event.Event.Message.MessageId).
			Build()
		_, subSpan := otel.LarkRobotOtelTracer.Start(ctx, utility.GetCurrentFunc())
		resp, err := larkutils.LarkClient.Im.V1.Message.Reply(ctx, req)
		subSpan.End()
		if err != nil {
			log.ZapLogger.Error("ReplyMessage", zaplog.Error(err))
			return err
		}
		log.ZapLogger.Info("ReplyMessage", zaplog.Any("resp", resp))
	}
	return
}

// PostRun  Repeat
//
//	@receiver r
//	@param ctx
//	@param event
//	@return err
func (r *QuoteMsgOperator) PostRun(ctx context.Context, event *larkim.P2MessageReceiveV1) (err error) {
	return
}
