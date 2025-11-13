package larkutils

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/BetaGoRobot/BetaGo/cts"
	"github.com/BetaGoRobot/BetaGo/dal/lark"
	"github.com/BetaGoRobot/BetaGo/utility/larkutils/templates"
	"github.com/BetaGoRobot/BetaGo/utility/logs"
	"github.com/BetaGoRobot/BetaGo/utility/otel"
	"github.com/BetaGoRobot/BetaGo/utility/vadvisor"
	"github.com/BetaGoRobot/go_utils/reflecting"
	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
	"go.opentelemetry.io/otel/attribute"
	"go.uber.org/zap"
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
func ReplyCard(ctx context.Context, cardContent *templates.TemplateCardContent, msgID, suffix string, replyInThread bool) (err error) {
	_, span := otel.LarkRobotOtelTracer.Start(ctx, reflecting.GetCurrentFunc())
	span.SetAttributes(attribute.Key("msgID").String(msgID))
	for k, v := range cardContent.Data.TemplateVariable {
		span.SetAttributes(attribute.Key(k).String(fmt.Sprintf("%v", v)))
	}
	defer span.End()
	defer func() { span.RecordError(err) }()
	logs.L().Ctx(ctx).Info(
		"reply card",
		zap.String("msgID", msgID),
		zap.String("suffix", suffix),
		zap.Bool("replyInThread", replyInThread),
		zap.String("cardContent", cardContent.String()),
	)

	resp, err := lark.LarkClient.Im.V1.Message.Reply(
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
	go RecordReplyMessage2Opensearch(ctx, resp, cardContent.GetVariables()...)
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
	defer func() { span.RecordError(err) }()
	cardContent := templates.NewCardContent(
		ctx, templates.NormalCardReplyTemplate,
	).
		AddJaegerTraceInfo(span.SpanContext().TraceID().String()).
		AddVariable("content", text)
	fmt.Println(cardContent.String())
	resp, err := lark.LarkClient.Im.V1.Message.Reply(
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
	go RecordReplyMessage2Opensearch(ctx, resp, cardContent.GetVariables()...)
	return
}

// SendCard to be filled
//
//	@param ctx context.Context
//	@param cardContent *templates.TemplateCardContent
//	@param chatID string
//	@param suffix string
//	@return err error
//	@author kevinmatthe
//	@update 2025-06-04 18:02:15
func SendCard(ctx context.Context, cardContent *templates.TemplateCardContent, chatID, suffix string) (err error) {
	_, span := otel.LarkRobotOtelTracer.Start(ctx, reflecting.GetCurrentFunc())
	span.SetAttributes(attribute.Key("chatID").String(chatID))
	span.SetAttributes(attribute.Key("suffix").String(suffix))
	span.SetAttributes(attribute.Key("cardContent").String(cardContent.String()))
	defer span.End()
	defer func() { span.RecordError(err) }()

	resp, err := lark.LarkClient.Im.V1.Message.Create(
		ctx, larkim.NewCreateMessageReqBuilder().ReceiveIdType(larkim.ReceiveIdTypeChatId).
			Body(
				larkim.NewCreateMessageReqBodyBuilder().
					ReceiveId(chatID).
					MsgType(larkim.MsgTypeInteractive).
					Content(cardContent.String()).
					Uuid(GenUUIDStr(chatID+suffix, 50)).
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
	go RecordMessage2Opensearch(ctx, resp, cardContent.GetVariables()...)
	return
}

// SendCardText to be filled
//
//	@param ctx context.Context
//	@param text string
//	@param msgID string
//	@param suffix string
//	@param replyInThread bool
//	@return err error
//	@author kevinmatthe
//	@update 2025-06-04 16:25:42
func SendCardText(ctx context.Context, text string, chatID, suffix string) (err error) {
	_, span := otel.LarkRobotOtelTracer.Start(ctx, reflecting.GetCurrentFunc())
	span.SetAttributes(attribute.Key("chatID").String(chatID))

	defer span.End()
	defer func() { span.RecordError(err) }()
	cardContent := templates.NewCardContent(
		ctx, templates.NormalCardReplyTemplate,
	).
		AddJaegerTraceInfo(span.SpanContext().TraceID().String()).
		AddVariable("content", text)
	fmt.Println(cardContent.String())
	resp, err := lark.LarkClient.Im.V1.Message.Create(
		ctx, larkim.NewCreateMessageReqBuilder().ReceiveIdType(larkim.ReceiveIdTypeChatId).
			Body(
				larkim.NewCreateMessageReqBodyBuilder().
					ReceiveId(chatID).
					MsgType(larkim.MsgTypeInteractive).
					Content(cardContent.String()).
					Uuid(GenUUIDStr(chatID+suffix, 50)).
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
	go RecordMessage2Opensearch(ctx, resp, cardContent.GetVariables()...)
	return
}

// ReplyCardTextGraph 123
//
//	@param ctx
//	@param text
//	@param msgID
func ReplyCardTextGraph[X cts.ValidType, Y cts.Numeric](ctx context.Context, text string, graph *vadvisor.MultiSeriesLineGraph[X, Y], msgID, suffix string, replyInThread bool) (err error) {
	_, span := otel.LarkRobotOtelTracer.Start(ctx, reflecting.GetCurrentFunc())
	span.SetAttributes(attribute.Key("msgID").String(msgID))

	defer span.End()
	defer func() { span.RecordError(err) }()
	cardContent := templates.NewCardContent(
		ctx, templates.NormalCardGraphReplyTemplate,
	).
		AddJaegerTraceInfo(span.SpanContext().TraceID().String()).
		AddVariable("content", text).
		AddVariable("graph", graph)
	fmt.Println(cardContent.String())
	resp, err := lark.LarkClient.Im.V1.Message.Reply(
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
	go RecordReplyMessage2Opensearch(ctx, resp, cardContent.GetVariables()...)
	return
}

// PatchCardTextGraph to be filled
//
//	@param ctx context.Context
//	@param text string
//	@param graph any
//	@param msgID string
//	@param suffix string
//	@param replyInThread bool
//	@return err error
//	@author kevinmatthe
//	@update 2025-06-03 13:29:07
func PatchCardTextGraph(ctx context.Context, text string, graph any, msgID string) (err error) {
	_, span := otel.LarkRobotOtelTracer.Start(ctx, reflecting.GetCurrentFunc())
	span.SetAttributes(attribute.Key("msgID").String(msgID))

	defer span.End()
	defer func() { span.RecordError(err) }()
	cardContent := templates.NewCardContent(
		ctx, templates.NormalCardGraphReplyTemplate,
	).
		AddVariable("content", text).
		AddVariable("graph", graph)
	fmt.Println(cardContent.String())
	resp, err := lark.LarkClient.Im.V1.Message.Patch(
		ctx, larkim.NewPatchMessageReqBuilder().
			MessageId(msgID).
			Body(
				larkim.NewPatchMessageReqBodyBuilder().
					Content(cardContent.String()).
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
	return
}

// PatchCard to be filled PatchCard
//
//	@param ctx context.Context
//	@param cardContent *templates.TemplateCardContent
//	@param msgID string
//	@return err error
//	@author kevinmatthe
//	@update 2025-06-05 13:23:46
func PatchCard(ctx context.Context, cardContent *templates.TemplateCardContent, msgID string) (err error) {
	_, span := otel.LarkRobotOtelTracer.Start(ctx, reflecting.GetCurrentFunc())
	span.SetAttributes(attribute.Key("msgID").String(msgID))
	for k, v := range cardContent.Data.TemplateVariable {
		span.SetAttributes(attribute.Key(k).String(fmt.Sprintf("%v", v)))
	}
	defer span.End()
	defer func() { span.RecordError(err) }()
	resp, err := lark.LarkClient.Im.V1.Message.Patch(
		ctx, larkim.NewPatchMessageReqBuilder().
			MessageId(msgID).
			Body(
				larkim.NewPatchMessageReqBodyBuilder().
					Content(cardContent.String()).
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
	return
}
