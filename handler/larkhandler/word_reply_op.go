package larkhandler

import (
	"context"
	"errors"
	"strings"

	"github.com/BetaGoRobot/BetaGo/consts"
	"github.com/BetaGoRobot/BetaGo/utility"
	"github.com/BetaGoRobot/BetaGo/utility/database"
	"github.com/BetaGoRobot/BetaGo/utility/larkutils"
	"github.com/BetaGoRobot/BetaGo/utility/log"
	"github.com/BetaGoRobot/BetaGo/utility/otel"
	"github.com/kevinmatthe/zaplog"
	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
	"go.opentelemetry.io/otel/attribute"
)

var _ LarkMsgOperator = &WordReplyMsgOperator{}

type WordReplyMsgOperator struct{}

// PreRun  Repeat
//
//	@receiver r
//	@param ctx
//	@param event
//	@return err
func (r *WordReplyMsgOperator) PreRun(ctx context.Context, event *larkim.P2MessageReceiveV1) (err error) {
	// 先判断群聊的功能启用情况
	if !checkFunctionEnabling(*event.Event.Message.ChatId, consts.LarkFunctionWordReply) {
		return errors.New("Not enabled")
	}
	return
}

// Run  Repeat
//
//	@receiver r
//	@param ctx
//	@param event
//	@return err
func (r *WordReplyMsgOperator) Run(ctx context.Context, event *larkim.P2MessageReceiveV1) (err error) {
	ctx, span := otel.LarkRobotOtelTracer.Start(ctx, utility.GetCurrentFunc())
	defer span.End()

	msg := larkutils.PreGetTextMsg(ctx, event)
	var replyStr string
	// 检查定制化逻辑
	resData, hitCache := database.FindByCache(&database.QuoteReplyMsgCustom{})
	span.SetAttributes(attribute.Bool("QuoteReplyMsgCustom hitCache", hitCache))
	for _, data := range resData {
		if data.GuildID == *event.Event.Message.ChatId && strings.Contains(msg, data.SubStr) {
			replyStr = data.Reply
		}
	}
	if replyStr == "" {
		// 无定制化逻辑，走通用判断
		data, hitCache := database.FindByCache(&database.QuoteReplyMsg{})
		span.SetAttributes(attribute.Bool("QuoteReplyMsg hitCache", hitCache))
		for _, d := range data {
			if strings.Contains(msg, d.SubStr) {
				replyStr = d.Reply
				break
			}
		}
	}
	if replyStr != "" {
		req := larkim.NewReplyMessageReqBuilder().
			Body(
				larkim.NewReplyMessageReqBodyBuilder().
					Content(larkim.NewTextMsgBuilder().Text(replyStr).Build()).
					MsgType(larkim.MsgTypeText).
					ReplyInThread(false).
					Uuid(*event.Event.Message.MessageId + "reply").
					Build(),
			).MessageId(*event.Event.Message.MessageId).
			Build()
		_, subSpan := otel.LarkRobotOtelTracer.Start(ctx, utility.GetCurrentFunc())
		resp, err := larkutils.LarkClient.Im.V1.Message.Reply(ctx, req)

		subSpan.End()
		if err != nil {
			log.ZapLogger.Error("ReplyMessage", zaplog.Error(err), zaplog.String("TraceID", span.SpanContext().TraceID().String()))
			return err
		}
		log.ZapLogger.Info("ReplyMessage", zaplog.Any("resp", resp), zaplog.String("TraceID", span.SpanContext().TraceID().String()))
	}
	return
}

// PostRun  Repeat
//
//	@receiver r
//	@param ctx
//	@param event
//	@return err
func (r *WordReplyMsgOperator) PostRun(ctx context.Context, event *larkim.P2MessageReceiveV1) (err error) {
	return
}
