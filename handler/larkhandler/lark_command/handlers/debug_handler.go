package handlers

import (
	"context"
	"errors"
	"iter"
	"strings"

	"github.com/BetaGoRobot/BetaGo/consts"
	handlertypes "github.com/BetaGoRobot/BetaGo/handler/handler_types"
	"github.com/BetaGoRobot/BetaGo/utility/doubao"
	"github.com/BetaGoRobot/BetaGo/utility/larkutils"
	"github.com/BetaGoRobot/BetaGo/utility/log"
	"github.com/BetaGoRobot/BetaGo/utility/message"
	opensearchdal "github.com/BetaGoRobot/BetaGo/utility/opensearch_dal"
	"github.com/BetaGoRobot/BetaGo/utility/otel"
	"github.com/BetaGoRobot/go_utils/reflecting"
	"github.com/bytedance/sonic"
	"github.com/defensestation/osquery"
	"github.com/kevinmatthe/zaplog"
	larkcore "github.com/larksuite/oapi-sdk-go/v3/core"
	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
	"go.opentelemetry.io/otel/attribute"
)

const (
	getIDText      = "Quoted Msg OpenID is "
	getGroupIDText = "Current ChatID is "
)

type traceItem struct {
	TraceID    string `json:"trace_id"`
	CreateTime string `json:"create_time"`
}

// DebugGetIDHandler to be filled
//
//	@param ctx context.Context
//	@param data *larkim.P2MessageReceiveV1
//	@param args ...string
//	@return error
//	@author heyuhengmatt
//	@update 2024-08-06 08:27:33
func DebugGetIDHandler(ctx context.Context, data *larkim.P2MessageReceiveV1, args ...string) error {
	ctx, span := otel.LarkRobotOtelTracer.Start(ctx, reflecting.GetCurrentFunc())
	span.SetAttributes(attribute.Key("event").String(larkcore.Prettify(data)))
	defer span.End()

	if data.Event.Message.ParentId == nil {
		return errors.New("No parent Msg Quoted")
	}

	err := larkutils.ReplyCardText(ctx, getIDText+*data.Event.Message.ParentId, *data.Event.Message.MessageId, "_getID", false)
	if err != nil {
		log.Zlog.Error("ReplyMessage", zaplog.Error(err), zaplog.String("TraceID", span.SpanContext().TraceID().String()))
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
	ctx, span := otel.LarkRobotOtelTracer.Start(ctx, reflecting.GetCurrentFunc())
	span.SetAttributes(attribute.Key("event").String(larkcore.Prettify(data)))
	defer span.End()
	chatID := data.Event.Message.ChatId
	if chatID != nil {
		err := larkutils.ReplyCardText(ctx, getGroupIDText+*chatID, *data.Event.Message.MessageId, "_getGroupID", false)
		if err != nil {
			log.Zlog.Error("ReplyMessage", zaplog.Error(err), zaplog.String("TraceID", span.SpanContext().TraceID().String()))
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
	ctx, span := otel.LarkRobotOtelTracer.Start(ctx, reflecting.GetCurrentFunc())
	span.SetAttributes(attribute.Key("event").String(larkcore.Prettify(data)))
	defer span.End()
	panic(errors.New("try panic!"))
}

func (t *traceItem) TraceURLMD() string {
	return strings.Join([]string{t.CreateTime, ": [Trace-", t.TraceID[:8], "]", "(https://jaeger.kmhomelab.cn/trace/", t.TraceID, ")"}, "")
}

// GetTraceFromMsgID to be filled
//
//	@param ctx context.Context
//	@param msgID string
//	@return []string
//	@return error
//	@author heyuhengmatt
//	@update 2024-08-06 08:27:37
func GetTraceFromMsgID(ctx context.Context, msgID string) (iter.Seq[*traceItem], error) {
	ctx, span := otel.LarkRobotOtelTracer.Start(ctx, reflecting.GetCurrentFunc())
	defer span.End()

	query := osquery.Search().
		Query(
			osquery.Bool().Must(
				osquery.Term("message_id", msgID),
			).MustNot(
				osquery.MatchPhrase(
					"raw_message_seg", "file _ key",
				),
			),
		).
		SourceIncludes("create_time", "trace_id").
		Sort("create_time", "desc")
	resp, err := opensearchdal.SearchData(
		ctx, consts.LarkMsgIndex, query,
	)
	if err != nil {
		return nil, err
	}
	return func(yield func(*traceItem) bool) {
		for _, hit := range resp.Hits.Hits {
			src := &handlertypes.MessageIndex{}
			err = sonic.Unmarshal(hit.Source, &src)
			if err != nil {
				return
			}
			if src.TraceID != "" {
				if !yield(&traceItem{src.TraceID, src.CreateTime}) {
					return
				}
			}
		}
	}, nil
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
	ctx, span := otel.LarkRobotOtelTracer.Start(ctx, reflecting.GetCurrentFunc())
	span.SetAttributes(attribute.Key("event").String(larkcore.Prettify(data)))
	defer span.End()
	var (
		m             = map[string]struct{}{}
		traceIDs      = make([]string, 0)
		replyInThread bool
	)
	if data.Event.Message.ThreadId != nil { // 话题模式，找到所有的traceID
		replyInThread = true
		resp, err := larkutils.LarkClient.Im.Message.List(ctx,
			larkim.NewListMessageReqBuilder().
				ContainerId(*data.Event.Message.ThreadId).
				ContainerIdType("thread").
				Build(),
		)
		if err != nil {
			return err
		}
		for _, msg := range resp.Data.Items {
			traceIters, err := GetTraceFromMsgID(ctx, *msg.MessageId)
			if err != nil {
				return err
			}
			for item := range traceIters {
				if _, ok := m[item.TraceID]; ok {
					continue
				}
				m[item.TraceID] = struct{}{}
				traceIDs = append(traceIDs, item.TraceURLMD())
			}
		}
	} else if data.Event.Message.ParentId != nil {
		traceIters, err := GetTraceFromMsgID(ctx, *data.Event.Message.ParentId)
		if err != nil {
			return err
		}
		for item := range traceIters {
			if _, ok := m[item.TraceID]; ok {
				continue
			}
			m[item.TraceID] = struct{}{}
			traceIDs = append(traceIDs, item.TraceURLMD())
		}
	}
	if len(traceIDs) == 0 {
		return errors.New("No traceID found")
	}
	traceIDStr := "TraceIDs:\n" + strings.Join(traceIDs, "\n")
	err := larkutils.ReplyCardText(ctx, traceIDStr, *data.Event.Message.MessageId, "_trace", replyInThread)
	if err != nil {
		log.Zlog.Error("ReplyMessage", zaplog.Error(err), zaplog.String("TraceID", span.SpanContext().TraceID().String()))
		return err
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
	ctx, span := otel.LarkRobotOtelTracer.Start(ctx, reflecting.GetCurrentFunc())
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
				if !resp.Success() {
					log.Zlog.Error("DeleteMessage", zaplog.String("MessageID", *msg.MessageId), zaplog.Error(errors.New(resp.Error())))
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
		if !resp.Success() {
			return errors.New(resp.Error())
		}
	}
	return nil
}

func DebugRepeatHandler(ctx context.Context, data *larkim.P2MessageReceiveV1, args ...string) error {
	ctx, span := otel.LarkRobotOtelTracer.Start(ctx, reflecting.GetCurrentFunc())
	span.SetAttributes(attribute.Key("event").String(larkcore.Prettify(data)))
	defer span.End()

	if data.Event.Message.ThreadId != nil {
		return nil
	} else if data.Event.Message.ParentId != nil {
		respMsg := larkutils.GetMsgFullByID(ctx, *data.Event.Message.ParentId)
		msg := respMsg.Data.Items[0]
		if msg == nil {
			return errors.New("No parent message found")
		}
		if msg.Sender.Id == nil {
			return errors.New("Parent message is not sent by bot")
		}
		repeatReq := larkim.NewCreateMessageReqBuilder().
			Body(
				larkim.NewCreateMessageReqBodyBuilder().
					MsgType(*msg.MsgType).
					Content(
						*msg.Body.Content,
					).
					ReceiveId(*msg.ChatId).
					Build(),
			).
			ReceiveIdType(larkim.ReceiveIdTypeChatId).
			Build()
		resp, err := larkutils.LarkClient.Im.V1.Message.Create(ctx, repeatReq)
		if err != nil {
			return err
		}
		if !resp.Success() {
			if strings.Contains(resp.Error(), "invalid image_key") {
				log.Zlog.Error("repeatMessage", zaplog.Error(err), zaplog.String("TraceID", span.SpanContext().TraceID().String()))
				return nil
			}
			return errors.New(resp.Error())
		}
		larkutils.RecordMessage2Opensearch(ctx, resp)
	}
	return nil
}

func DebugImageHandler(ctx context.Context, data *larkim.P2MessageReceiveV1, args ...string) error {
	ctx, span := otel.LarkRobotOtelTracer.Start(ctx, reflecting.GetCurrentFunc())
	span.SetAttributes(attribute.Key("event").String(larkcore.Prettify(data)))
	defer span.End()
	seq, err := larkutils.GetAllImgURLFromParent(ctx, data)
	if err != nil {
		return err
	}
	if seq == nil {
		return nil
	}
	urls := make([]string, 0)
	for url := range seq {
		url = strings.ReplaceAll(url, "kmhomelab.cn", "kevinmatt.top")
		urls = append(urls, url)
	}
	var inputPrompt string
	if _, input := parseArgs(args...); input == "" {
		inputPrompt = "图里都是些什么？"
	} else {
		inputPrompt = input
	}
	dataSeq, err := doubao.SingleChatStreamingPrompt(
		ctx,
		inputPrompt,
		doubao.ARK_VISION_EPID,
		urls...,
	)
	if err != nil {
		return err
	}
	err = message.SendAndUpdateStreamingCard(ctx, data.Event.Message, dataSeq)
	if err != nil {
		return err
	}
	return nil
}
