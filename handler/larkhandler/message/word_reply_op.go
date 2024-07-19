package message

import (
	"context"
	"strings"

	"github.com/BetaGoRobot/BetaGo/consts"
	"github.com/BetaGoRobot/BetaGo/handler/larkhandler/base"
	"github.com/BetaGoRobot/BetaGo/utility"
	"github.com/BetaGoRobot/BetaGo/utility/database"
	"github.com/BetaGoRobot/BetaGo/utility/larkutils"
	"github.com/BetaGoRobot/BetaGo/utility/log"
	"github.com/BetaGoRobot/BetaGo/utility/otel"
	"github.com/kevinmatthe/zaplog"
	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
	"github.com/pkg/errors"
	"go.opentelemetry.io/otel/attribute"
)

var _ base.Operator[larkim.P2MessageReceiveV1] = &WordReplyMsgOperator{}

// WordReplyMsgOperator  Repeat
//
//	@author heyuhengmatt
//	@update 2024-07-17 01:35:11
type WordReplyMsgOperator struct {
	base.OperatorBase[larkim.P2MessageReceiveV1]
}

// PreRun Repeat
//
//	@receiver r *WordReplyMsgOperator
//	@param ctx context.Context
//	@param event *larkim.P2MessageReceiveV1
//	@return err error
//	@author heyuhengmatt
//	@update 2024-07-17 01:35:17
func (r *WordReplyMsgOperator) PreRun(ctx context.Context, event *larkim.P2MessageReceiveV1) (err error) {
	ctx, span := otel.LarkRobotOtelTracer.Start(ctx, utility.GetCurrentFunc())
	defer span.End()
	defer span.RecordError(err)

	// 先判断群聊的功能启用情况
	if !larkutils.CheckFunctionEnabling(*event.Event.Message.ChatId, consts.LarkFunctionWordReply) {
		return errors.Wrap(consts.ErrStageSkip, "WordReplyMsgOperator: Not enabled")
	}

	if larkutils.IsCommand(ctx, larkutils.PreGetTextMsg(ctx, event)) {
		return errors.Wrap(consts.ErrStageSkip, "WordReplyMsgOperator: Is Command")
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
	defer span.RecordError(err)

	msg := larkutils.PreGetTextMsg(ctx, event)
	var replyStr string
	// 检查定制化逻辑, Key为GuildID, 拿到GUI了dID下的所有SubStr配置
	customConfig, hitCache := database.FindByCacheFunc(database.QuoteReplyMsgCustom{GuildID: *event.Event.Message.ChatId},
		func(d database.QuoteReplyMsgCustom) string {
			return d.GuildID
		},
	)
	span.SetAttributes(attribute.Bool("QuoteReplyMsgCustom hitCache", hitCache))
	for _, data := range customConfig {
		if strings.Contains(msg, data.SubStr) {
			replyStr = data.Reply
		}
	}

	if replyStr == "" {
		// 无定制化逻辑，走通用判断
		data, hitCache := database.FindByCacheFunc(
			database.QuoteReplyMsg{},
			func(d database.QuoteReplyMsg) string {
				return d.SubStr
			},
		)
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
