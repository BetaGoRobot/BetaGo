package larkutils

import (
	"context"
	"errors"
	"fmt"

	"github.com/BetaGoRobot/BetaGo/utility/otel"
	"github.com/BetaGoRobot/go_utils/reflecting"
	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
	"go.opentelemetry.io/otel/attribute"
)

func TruncString(str string, length int) string {
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
					Uuid(TruncString(msgID+suffix, 50)).
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
	template := GetTemplate(StreamingReasonTemplate)
	cardContent := NewSheetCardContent(
		ctx, template.TemplateID, template.TemplateVersion,
	).
		AddJaegerTraceInfo(span.SpanContext().TraceID().String()).
		AddVariable("content", text)

	resp, err := LarkClient.Im.V1.Message.Reply(
		ctx, larkim.NewReplyMessageReqBuilder().
			MessageId(msgID).
			Body(
				larkim.NewReplyMessageReqBodyBuilder().
					MsgType(larkim.MsgTypeInteractive).
					Content(cardContent.String()).
					Uuid(TruncString(msgID+suffix, 50)).
					ReplyInThread(replyInThread).
					Build(),
			).
			Build(),
	)
	if err != nil {
		return
	}
	if resp.StatusCode != 200 {
		return errors.New(resp.Error())
	}
	RecordReplyMessage2Opensearch(ctx, resp, cardContent.GetVariables()...)
	return
}
