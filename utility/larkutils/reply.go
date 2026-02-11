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

func GenUUIDCode(srcKey, specificKey string, length int) string {
	// 重点是分发的时候，是某个消息（srcKey），只会触发一次（specificKey）
	res := srcKey + specificKey
	if len(res) > length {
		res = res[:length]
	}
	return res
}

// ReplyCard  注意：不要传入已经Build过的文本
//
//	@param ctx
//	@param text
//	@param msgID
func ReplyCard(ctx context.Context, cardContent *templates.TemplateCardContent, msgID, suffix string, replyInThread bool) (err error) {
	ctx, span := otel.LarkRobotOtelTracer.Start(ctx, reflecting.GetCurrentFunc())
	defer span.End()
	defer func() { span.RecordError(err) }()

	// 先把卡片发送了，再记录日志和指标，避免指标记录的耗时过程拖慢整个请求
	resp, err := doSendCard(ctx, msgID, suffix, cardContent, replyInThread)
	if err != nil {
		logs.L().Ctx(ctx).Error("doSendCard failed", zap.Error(err))
		return
	}

	span.SetAttributes(attribute.Key("msgID").String(msgID))
	for k, v := range cardContent.Data.TemplateVariable {
		span.SetAttributes(attribute.Key(k).String(fmt.Sprintf("%v", v)))
	}
	logs.L().Ctx(ctx).Info(
		"reply card",
		zap.String("msgID", msgID),
		zap.String("suffix", suffix),
		zap.Bool("replyInThread", replyInThread),
		zap.String("cardContent", cardContent.String()),
	)
	go RecordReplyMessage2Opensearch(ctx, resp, cardContent.GetVariables()...)
	return
}

func doSendCard(ctx context.Context, msgID, suffix string, cardContent *templates.TemplateCardContent, replyInThread bool) (resp *larkim.ReplyMessageResp, err error) {
	ctx, span := otel.LarkRobotOtelTracer.Start(ctx, reflecting.GetCurrentFunc())
	defer span.End()
	resp, err = lark.LarkClient.Im.V1.Message.Reply(
		ctx, larkim.NewReplyMessageReqBuilder().
			MessageId(msgID).
			Body(
				larkim.NewReplyMessageReqBodyBuilder().
					MsgType(larkim.MsgTypeInteractive).
					Content(cardContent.String()).
					Uuid(GenUUIDCode(msgID, suffix, 50)).
					ReplyInThread(replyInThread).
					Build(),
			).
			Build(),
	)
	if err != nil {
		return
	}
	if !resp.Success() {
		return resp, errors.New(resp.Error())
	}
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
	logs.L().Ctx(ctx).Info(
		"reply card text",
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
	logs.L().Ctx(ctx).Info(
		"reply card text",
		zap.String("chatID", chatID),
		zap.String("text", text),
		zap.String("suffix", suffix),
		zap.String("cardContent", cardContent.String()),
	)
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
	logs.L().Ctx(ctx).Info(
		"reply card text",
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
	logs.L().Ctx(ctx).Info(
		"reply card text",
		zap.String("msgID", msgID),
		zap.String("cardContent", cardContent.String()),
	)
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
