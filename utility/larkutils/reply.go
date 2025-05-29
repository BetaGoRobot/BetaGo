package larkutils

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/BetaGoRobot/BetaGo/utility/otel"
	"github.com/BetaGoRobot/go_utils/reflecting"
	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
	"go.opentelemetry.io/otel/attribute"
)

func GenUUIDStr(str string, length int) string {
	st := strconv.Itoa(int(time.Now().Truncate(time.Minute * 2).Unix()))
	str = st + str
	if len(str) > length {
		return str[:length]
	}
	return str
}

// ReplyCard  注意：不要传入已经Build过的文本
//
//	@param ctx
//	@param text
//	@param msgID
func ReplyCard(ctx context.Context, cardContent *TemplateCardContent, msgID, suffix string, replyInThread bool) (err error) {
	_, span := otel.LarkRobotOtelTracer.Start(ctx, reflecting.GetCurrentFunc())
	span.SetAttributes(attribute.Key("msgID").String(msgID))
	for k, v := range cardContent.Data.TemplateVariable {
		span.SetAttributes(attribute.Key(k).String(fmt.Sprintf("%v", v)))
	}
	defer span.End()
	resp, err := LarkClient.Im.V1.Message.Reply(
		ctx, larkim.NewReplyMessageReqBuilder().
			MessageId(msgID).
			Body(
				larkim.NewReplyMessageReqBodyBuilder().
					MsgType(larkim.MsgTypeInteractive).
					Content(cardContent.String()).
					Uuid(GenUUIDStr(suffix, 50)).
					ReplyInThread(replyInThread).
					Build(),
			).
			Build(),
	)
	if err != nil {
		return
	}
	if !resp.Success() {
		return errors.New(resp.Error())
	}
	RecordReplyMessage2Opensearch(ctx, resp, cardContent.GetVariables()...)
	return
}

// ReplyCardText 123
//
//	@param ctx
//	@param text
//	@param msgID
func ReplyCardText(ctx context.Context, text string, msgID, suffix string, replyInThread bool) (err error) {
	_, span := otel.LarkRobotOtelTracer.Start(ctx, reflecting.GetCurrentFunc())
	span.SetAttributes(attribute.Key("msgID").String(msgID))

	defer span.End()
	cardContent := NewCardContent(
		ctx, NormalCardReplyTemplate,
	).
		AddJaegerTraceInfo(span.SpanContext().TraceID().String()).
		AddVariable("content", text)
	fmt.Println(cardContent.String())
	resp, err := LarkClient.Im.V1.Message.Reply(
		ctx, larkim.NewReplyMessageReqBuilder().
			MessageId(msgID).
			Body(
				larkim.NewReplyMessageReqBodyBuilder().
					MsgType(larkim.MsgTypeInteractive).
					Content(cardContent.String()).
					Uuid(GenUUIDStr(msgID+suffix, 50)).
					ReplyInThread(replyInThread).
					Build(),
			).
			Build(),
	)
	if err != nil {
		return
	}
	if !resp.Success() {
		return errors.New(resp.Error())
	}
	RecordReplyMessage2Opensearch(ctx, resp, cardContent.GetVariables()...)
	return
}

// ReplyCardTextGraph 123
//
//	@param ctx
//	@param text
//	@param msgID
func ReplyCardTextGraph(ctx context.Context, text string, graph any, msgID, suffix string, replyInThread bool) (err error) {
	_, span := otel.LarkRobotOtelTracer.Start(ctx, reflecting.GetCurrentFunc())
	span.SetAttributes(attribute.Key("msgID").String(msgID))

	defer span.End()
	cardContent := NewCardContent(
		ctx, NormalCardGraphReplyTemplate,
	).
		AddJaegerTraceInfo(span.SpanContext().TraceID().String()).
		AddVariable("content", text).
		AddVariable("graph", graph)
	fmt.Println(cardContent.String())
	resp, err := LarkClient.Im.V1.Message.Reply(
		ctx, larkim.NewReplyMessageReqBuilder().
			MessageId(msgID).
			Body(
				larkim.NewReplyMessageReqBodyBuilder().
					MsgType(larkim.MsgTypeInteractive).
					Content(cardContent.String()).
					Uuid(GenUUIDStr(msgID+suffix, 50)).
					ReplyInThread(replyInThread).
					Build(),
			).
			Build(),
	)
	if err != nil {
		return
	}
	if !resp.Success() {
		return errors.New(resp.Error())
	}
	RecordReplyMessage2Opensearch(ctx, resp, cardContent.GetVariables()...)
	return
}
