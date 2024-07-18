package larkcommand

import (
	"context"
	"errors"

	"github.com/BetaGoRobot/BetaGo/utility"
	"github.com/BetaGoRobot/BetaGo/utility/larkutils"
	"github.com/BetaGoRobot/BetaGo/utility/log"
	"github.com/BetaGoRobot/BetaGo/utility/otel"
	"github.com/kevinmatthe/zaplog"
	larkcore "github.com/larksuite/oapi-sdk-go/v3/core"
	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
	"go.opentelemetry.io/otel/attribute"
)

const (
	getIDText      = "Quoted Msg OpenID is "
	getGroupIDText = "Current GroupID is "
)

// getIDHandler get ID Handler
//
//	@param ctx
//	@param data
//	@param args
//	@return error
func getIDHandler(ctx context.Context, data *larkim.P2MessageReceiveV1, args ...string) error {
	ctx, span := otel.LarkRobotOtelTracer.Start(ctx, utility.GetCurrentFunc())
	span.SetAttributes(attribute.Key("event").String(larkcore.Prettify(data)))
	defer span.End()

	req := larkim.NewReplyMessageReqBuilder().Body(
		larkim.NewReplyMessageReqBodyBuilder().Content(larkim.NewTextMsgBuilder().Text(getIDText + *data.Event.Message.ParentId).Build()).MsgType(larkim.MsgTypeText).ReplyInThread(true).Uuid(*data.Event.Message.MessageId + "reply").Build(),
	).MessageId(*data.Event.Message.MessageId).Build()
	resp, err := larkutils.LarkClient.Im.V1.Message.Reply(ctx, req)
	if err != nil {
		log.ZapLogger.Error("ReplyMessage", zaplog.Error(err), zaplog.String("TraceID", span.SpanContext().TraceID().String()))
		return err
	}
	if resp.Error() != "" {
		log.ZapLogger.Error("ReplyMessage", zaplog.String("Error", resp.Error()), zaplog.String("TraceID", span.SpanContext().TraceID().String()))
		return errors.New(resp.Error())
	}
	return nil
}

// getIDHandler get ID Handler
//
//	@param ctx
//	@param data
//	@param args
//	@return error
func getGroupIDHandler(ctx context.Context, data *larkim.P2MessageReceiveV1, args ...string) error {
	ctx, span := otel.LarkRobotOtelTracer.Start(ctx, utility.GetCurrentFunc())
	span.SetAttributes(attribute.Key("event").String(larkcore.Prettify(data)))
	defer span.End()
	chatID := data.Event.Message.ChatId
	if chatID != nil {
		req := larkim.NewReplyMessageReqBuilder().Body(
			larkim.NewReplyMessageReqBodyBuilder().Content(larkim.NewTextMsgBuilder().Text(getGroupIDText + *chatID).Build()).MsgType(larkim.MsgTypeText).ReplyInThread(true).Uuid(*data.Event.Message.MessageId + "reply").Build(),
		).MessageId(*data.Event.Message.MessageId).Build()
		resp, err := larkutils.LarkClient.Im.V1.Message.Reply(ctx, req)
		if err != nil {
			log.ZapLogger.Error("ReplyMessage", zaplog.Error(err), zaplog.String("TraceID", span.SpanContext().TraceID().String()))
			return err
		}
		if resp.Error() != "" {
			log.ZapLogger.Error("ReplyMessage", zaplog.String("Error", resp.Error()), zaplog.String("TraceID", span.SpanContext().TraceID().String()))
			return errors.New(resp.Error())
		}
	}

	return nil
}
