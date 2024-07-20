package larkcommand

import (
	"context"
	"errors"
	"strings"

	"github.com/BetaGoRobot/BetaGo/utility"
	"github.com/BetaGoRobot/BetaGo/utility/database"
	"github.com/BetaGoRobot/BetaGo/utility/larkutils"
	"github.com/BetaGoRobot/BetaGo/utility/log"
	"github.com/BetaGoRobot/BetaGo/utility/otel"
	"github.com/kevinmatthe/zaplog"
	larkcore "github.com/larksuite/oapi-sdk-go/v3/core"
	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
	"go.opentelemetry.io/otel/attribute"
	"gorm.io/gorm/clause"
)

const (
	getIDText      = "Quoted Msg OpenID is "
	getGroupIDText = "Current ChatID is "
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

	if data.Event.Message.ParentId == nil {
		return errors.New("No parent Msg Quoted")
	}
	req := larkim.NewReplyMessageReqBuilder().Body(
		larkim.NewReplyMessageReqBodyBuilder().Content(larkim.NewTextMsgBuilder().Text(getIDText + *data.Event.Message.ParentId).Build()).MsgType(larkim.MsgTypeText).ReplyInThread(true).Uuid(*data.Event.Message.MessageId + "reply").Build(),
	).MessageId(*data.Event.Message.MessageId).Build()
	resp, err := larkutils.LarkClient.Im.V1.Message.Reply(ctx, req)
	if err != nil {
		log.ZapLogger.Error("ReplyMessage", zaplog.Error(err), zaplog.String("TraceID", span.SpanContext().TraceID().String()))
		return err
	}
	if resp.Code != 0 {
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
		if resp.Code != 0 {
			log.ZapLogger.Error("ReplyMessage", zaplog.String("Error", resp.Error()), zaplog.String("TraceID", span.SpanContext().TraceID().String()))
			return errors.New(resp.Error())
		}
	}

	return nil
}

// getIDHandler get ID Handler
//
//	@param ctx
//	@param data
//	@param args
//	@return error
func tryPanicHandler(ctx context.Context, data *larkim.P2MessageReceiveV1, args ...string) error {
	ctx, span := otel.LarkRobotOtelTracer.Start(ctx, utility.GetCurrentFunc())
	span.SetAttributes(attribute.Key("event").String(larkcore.Prettify(data)))
	defer span.End()
	panic("try panic!")
}

// getIDHandler get ID Handler
//
//	@param ctx
//	@param data
//	@param args
//	@return error
func wordAddHandler(ctx context.Context, data *larkim.P2MessageReceiveV1, args ...string) error {
	ctx, span := otel.LarkRobotOtelTracer.Start(ctx, utility.GetCurrentFunc())
	span.SetAttributes(attribute.Key("event").String(larkcore.Prettify(data)))
	defer span.End()

	if len(args) < 2 {
		return errors.ErrUnsupported
	}
	argMap := parseArgs(args...)
	log.ZapLogger.Info("wordAddHandler", zaplog.Any("args", argMap))

	word, ok := argMap["word"]
	if !ok {
		return errors.New("word is required")
	}
	rate, ok := argMap["rate"]
	if !ok {
		return errors.New("rate is required")
	}

	ChatID := *data.Event.Message.ChatId
	return database.GetDbConnection().Clauses(clause.OnConflict{
		UpdateAll: true,
	}).Create(&database.RepeatWordsRateCustom{
		GuildID: ChatID,
		Word:    word,
		Rate:    utility.MustAtoI(rate),
	}).Error
}

func parseArgs(args ...string) map[string]string {
	resMap := make(map[string]string)
	for _, arg := range args {
		if argKV := strings.Split(arg, "="); len(argKV) == 2 {
			resMap[strings.TrimLeft(argKV[0], "--")] = argKV[1]
		}
	}
	return resMap
}
