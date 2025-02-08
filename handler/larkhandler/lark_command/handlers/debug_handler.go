package handlers

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
)

const (
	getIDText      = "Quoted Msg OpenID is "
	getGroupIDText = "Current ChatID is "
)

// DebugGetIDHandler to be filled
//
//	@param ctx context.Context
//	@param data *larkim.P2MessageReceiveV1
//	@param args ...string
//	@return error
//	@author heyuhengmatt
//	@update 2024-08-06 08:27:33
func DebugGetIDHandler(ctx context.Context, data *larkim.P2MessageReceiveV1, args ...string) error {
	ctx, span := otel.LarkRobotOtelTracer.Start(ctx, utility.GetCurrentFunc())
	span.SetAttributes(attribute.Key("event").String(larkcore.Prettify(data)))
	defer span.End()

	if data.Event.Message.ParentId == nil {
		return errors.New("No parent Msg Quoted")
	}

	err := larkutils.ReplyMsgText(ctx, getIDText+*data.Event.Message.ParentId, *data.Event.Message.MessageId, "_getID", false)
	if err != nil {
		log.ZapLogger.Error("ReplyMessage", zaplog.Error(err), zaplog.String("TraceID", span.SpanContext().TraceID().String()))
		return err
	}
	return nil
}

// DebugGetGroupIDHandler to be filled
//
//	@param ctx context.Context
//	@param data *larkim.P2MessageReceiveV1
//	@param args ...string
//	@return error
//	@author heyuhengmatt
//	@update 2024-08-06 08:27:29
func DebugGetGroupIDHandler(ctx context.Context, data *larkim.P2MessageReceiveV1, args ...string) error {
	ctx, span := otel.LarkRobotOtelTracer.Start(ctx, utility.GetCurrentFunc())
	span.SetAttributes(attribute.Key("event").String(larkcore.Prettify(data)))
	defer span.End()
	chatID := data.Event.Message.ChatId
	if chatID != nil {
		err := larkutils.ReplyMsgText(ctx, getGroupIDText+*chatID, *data.Event.Message.MessageId, "_getGroupID", false)
		if err != nil {
			log.ZapLogger.Error("ReplyMessage", zaplog.Error(err), zaplog.String("TraceID", span.SpanContext().TraceID().String()))
			return err
		}
	}

	return nil
}

// DebugTryPanicHandler to be filled
//
//	@param ctx context.Context
//	@param data *larkim.P2MessageReceiveV1
//	@param args ...string
//	@return error
//	@author heyuhengmatt
//	@update 2024-08-06 08:27:25
func DebugTryPanicHandler(ctx context.Context, data *larkim.P2MessageReceiveV1, args ...string) error {
	ctx, span := otel.LarkRobotOtelTracer.Start(ctx, utility.GetCurrentFunc())
	span.SetAttributes(attribute.Key("event").String(larkcore.Prettify(data)))
	defer span.End()
	panic("try panic!")
}

func getTraceURLMD(traceID string) string {
	return strings.Join([]string{"[Trace-", traceID[:8], "]", "(https://jaeger.kmhomelab.cn/trace/", traceID, ")"}, "")
}

// GetTraceFromMsgID to be filled
//
//	@param ctx context.Context
//	@param msgID string
//	@return []string
//	@return error
//	@author heyuhengmatt
//	@update 2024-08-06 08:27:37
func GetTraceFromMsgID(ctx context.Context, msgID string) ([]string, error) {
	ctx, span := otel.LarkRobotOtelTracer.Start(ctx, utility.GetCurrentFunc())
	defer span.End()

	traceLogs, hitCache := database.FindByCacheFunc(database.MsgTraceLog{MsgID: msgID}, func(d database.MsgTraceLog) string {
		return d.MsgID
	})
	span.SetAttributes(attribute.Bool("MsgTraceLog hitCache", hitCache))
	if len(traceLogs) == 0 {
		return nil, errors.New("No trace log found for the message qouted")
	}
	traceIDs := make([]string, 0)
	for _, traceLog := range traceLogs {
		traceIDs = append(traceIDs, getTraceURLMD(traceLog.TraceID))
	}
	return traceIDs, nil
}

// DebugTraceHandler to be filled
//
//	@param ctx context.Context
//	@param data *larkim.P2MessageReceiveV1
//	@param args ...string
//	@return error
//	@author heyuhengmatt
//	@update 2024-08-06 08:27:23
func DebugTraceHandler(ctx context.Context, data *larkim.P2MessageReceiveV1, args ...string) error {
	ctx, span := otel.LarkRobotOtelTracer.Start(ctx, utility.GetCurrentFunc())
	span.SetAttributes(attribute.Key("event").String(larkcore.Prettify(data)))
	defer span.End()

	if data.Event.Message.ThreadId != nil { // 话题模式，找到所有的traceID
		resp, err := larkutils.LarkClient.Im.Message.List(ctx, larkim.NewListMessageReqBuilder().ContainerIdType("thread").ContainerId(*data.Event.Message.ThreadId).Build())
		if err != nil {
			return err
		}
		traceIDs := make([]string, 0)
		for _, msg := range resp.Data.Items {
			if *msg.Sender.Id == larkutils.BotAppID {
				traceIDsTmp, err := GetTraceFromMsgID(ctx, *msg.MessageId)
				if err != nil {
					return err
				}
				traceIDs = append(traceIDs, traceIDsTmp...)
			}
		}
		traceIDStr := "TraceIDs:\\n" + strings.Join(traceIDs, "\\n")
		err = larkutils.ReplyMsgText(ctx, traceIDStr, *data.Event.Message.MessageId, "_trace", true)
		if err != nil {
			log.ZapLogger.Error("ReplyMessage", zaplog.Error(err), zaplog.String("TraceID", span.SpanContext().TraceID().String()))
			return err
		}
	} else if data.Event.Message.ParentId != nil {
		traceIDs, err := GetTraceFromMsgID(ctx, *data.Event.Message.ParentId)
		if err != nil {
			return err
		}
		traceIDStr := "TraceIDs:\\n" + strings.Join(traceIDs, "\\n")
		err = larkutils.ReplyMsgText(ctx, traceIDStr, *data.Event.Message.MessageId, "_trace", true)
		if err != nil {
			log.ZapLogger.Error("ReplyMessage", zaplog.Error(err), zaplog.String("TraceID", span.SpanContext().TraceID().String()))
			return err
		}
	}
	return nil
}

// DebugRevertHandler DebugTraceHandler to be filled
//
//	@param ctx context.Context
//	@param data *larkim.P2MessageReceiveV1
//	@param args ...string
//	@return error
func DebugRevertHandler(ctx context.Context, data *larkim.P2MessageReceiveV1, args ...string) error {
	ctx, span := otel.LarkRobotOtelTracer.Start(ctx, utility.GetCurrentFunc())
	span.SetAttributes(attribute.Key("event").String(larkcore.Prettify(data)))
	defer span.End()

	if data.Event.Message.ThreadId != nil { // 话题模式，找到所有的traceID
		resp, err := larkutils.LarkClient.Im.Message.List(ctx, larkim.NewListMessageReqBuilder().ContainerIdType("thread").ContainerId(*data.Event.Message.ThreadId).Build())
		if err != nil {
			return err
		}
		for _, msg := range resp.Data.Items {
			if *msg.Sender.Id == larkutils.BotAppID {
				resp, err := larkutils.LarkClient.Im.Message.Delete(ctx, larkim.NewDeleteMessageReqBuilder().MessageId(*msg.MessageId).Build())
				if err != nil {
					return err
				}
				if resp.Code != 0 {
					log.ZapLogger.Error("DeleteMessage", zaplog.String("MessageID", *msg.MessageId), zaplog.Error(errors.New(resp.Error())))
				}
			}
		}
	} else if data.Event.Message.ParentId != nil {
		respMsg := larkutils.GetMsgFullByID(ctx, *data.Event.Message.ParentId)
		msg := respMsg.Data.Items[0]
		if msg == nil {
			return errors.New("No parent message found")
		}
		if msg.Sender.Id == nil || *msg.Sender.Id != larkutils.BotAppID {
			return errors.New("Parent message is not sent by bot")
		}
		resp, err := larkutils.LarkClient.Im.Message.Delete(ctx, larkim.NewDeleteMessageReqBuilder().MessageId(*data.Event.Message.ParentId).Build())
		if err != nil {
			return err
		}
		if resp.Code != 0 {
			return errors.New(resp.Error())
		}
	}
	return nil
}
