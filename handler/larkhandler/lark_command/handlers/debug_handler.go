package handlers

import (
	"context"
	"errors"
	"fmt"
	"iter"
	"slices"
	"strings"

	"github.com/BetaGoRobot/BetaGo/consts"
	"github.com/BetaGoRobot/BetaGo/dal/lark"
	handlerbase "github.com/BetaGoRobot/BetaGo/handler/handler_base"
	handlertypes "github.com/BetaGoRobot/BetaGo/handler/handler_types"
	"github.com/BetaGoRobot/BetaGo/utility/chunking"
	"github.com/BetaGoRobot/BetaGo/utility/doubao"
	"github.com/BetaGoRobot/BetaGo/utility/history"
	"github.com/BetaGoRobot/BetaGo/utility/larkutils"
	"github.com/BetaGoRobot/BetaGo/utility/larkutils/larkconsts"
	"github.com/BetaGoRobot/BetaGo/utility/larkutils/larkimg"
	"github.com/BetaGoRobot/BetaGo/utility/larkutils/larkmsgutils"
	"github.com/BetaGoRobot/BetaGo/utility/larkutils/templates"
	"github.com/BetaGoRobot/BetaGo/utility/logs"
	"github.com/BetaGoRobot/BetaGo/utility/message"
	opensearchdal "github.com/BetaGoRobot/BetaGo/utility/opensearch_dal"
	"github.com/BetaGoRobot/BetaGo/utility/otel"
	commonutils "github.com/BetaGoRobot/go_utils/common_utils"
	"github.com/BetaGoRobot/go_utils/reflecting"
	"github.com/bytedance/sonic"
	"github.com/defensestation/osquery"
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
func DebugGetIDHandler(ctx context.Context, data *larkim.P2MessageReceiveV1, metaData *handlerbase.BaseMetaData, args ...string) (err error) {
	ctx, span := otel.LarkRobotOtelTracer.Start(ctx, reflecting.GetCurrentFunc())
	span.SetAttributes(attribute.Key("event").String(larkcore.Prettify(data)))
	defer span.End()
	defer func() { span.RecordError(err) }()

	if data.Event.Message.ParentId == nil {
		return errors.New("No parent Msg Quoted")
	}

	err = larkutils.ReplyCardText(ctx, getIDText+*data.Event.Message.ParentId, *data.Event.Message.MessageId, "_getID", false)
	if err != nil {
		logs.L.Error(ctx, "reply message error", "error", err, "traceID", span.SpanContext().TraceID().String())
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
func DebugGetGroupIDHandler(ctx context.Context, data *larkim.P2MessageReceiveV1, metaData *handlerbase.BaseMetaData, args ...string) (err error) {
	ctx, span := otel.LarkRobotOtelTracer.Start(ctx, reflecting.GetCurrentFunc())
	span.SetAttributes(attribute.Key("event").String(larkcore.Prettify(data)))
	defer span.End()
	defer func() { span.RecordError(err) }()
	chatID := data.Event.Message.ChatId
	if chatID != nil {
		err := larkutils.ReplyCardText(ctx, getGroupIDText+*chatID, *data.Event.Message.MessageId, "_getGroupID", false)
		if err != nil {
			logs.L.Error(ctx, "reply message error", "error", err, "traceID", span.SpanContext().TraceID().String())
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
func DebugTryPanicHandler(ctx context.Context, data *larkim.P2MessageReceiveV1, metaData *handlerbase.BaseMetaData, args ...string) (err error) {
	ctx, span := otel.LarkRobotOtelTracer.Start(ctx, reflecting.GetCurrentFunc())
	span.SetAttributes(attribute.Key("event").String(larkcore.Prettify(data)))
	defer span.End()
	defer func() { span.RecordError(err) }()
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
func DebugTraceHandler(ctx context.Context, data *larkim.P2MessageReceiveV1, metaData *handlerbase.BaseMetaData, args ...string) (err error) {
	ctx, span := otel.LarkRobotOtelTracer.Start(ctx, reflecting.GetCurrentFunc())
	span.SetAttributes(attribute.Key("event").String(larkcore.Prettify(data)))
	defer span.End()
	defer func() { span.RecordError(err) }()
	var (
		m             = map[string]struct{}{}
		traceIDs      = make([]string, 0)
		replyInThread bool
	)
	if data.Event.Message.ThreadId != nil { // 话题模式，找到所有的traceID
		replyInThread = true
		resp, err := lark.LarkClient.Im.Message.List(ctx,
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
	err = larkutils.ReplyCardText(ctx, traceIDStr, *data.Event.Message.MessageId, "_trace", replyInThread)
	if err != nil {
		logs.L.Error(ctx, "reply message error", "error", err, "traceID", span.SpanContext().TraceID().String())
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
func DebugRevertHandler(ctx context.Context, data *larkim.P2MessageReceiveV1, metaData *handlerbase.BaseMetaData, args ...string) (err error) {
	ctx, span := otel.LarkRobotOtelTracer.Start(ctx, reflecting.GetCurrentFunc())
	span.SetAttributes(attribute.Key("event").String(larkcore.Prettify(data)))
	defer span.End()
	defer func() { span.RecordError(err) }()

	if data.Event.Message.ThreadId != nil { // 话题模式，找到所有的traceID
		resp, err := lark.LarkClient.Im.Message.List(ctx, larkim.NewListMessageReqBuilder().ContainerIdType("thread").ContainerId(*data.Event.Message.ThreadId).Build())
		if err != nil {
			return err
		}
		for _, msg := range resp.Data.Items {
			if *msg.Sender.Id == larkconsts.BotAppID {
				resp, err := lark.LarkClient.Im.Message.Delete(ctx, larkim.NewDeleteMessageReqBuilder().MessageId(*msg.MessageId).Build())
				if err != nil {
					return err
				}
				if !resp.Success() {
					logs.L.Error(ctx, "delete message error", "messageID", *msg.MessageId, "error", resp.Error())
				}
			}
		}
	} else if data.Event.Message.ParentId != nil {
		respMsg := larkutils.GetMsgFullByID(ctx, *data.Event.Message.ParentId)
		msg := respMsg.Data.Items[0]
		if msg == nil {
			return errors.New("No parent message found")
		}
		if msg.Sender.Id == nil || *msg.Sender.Id != larkconsts.BotAppID {
			return errors.New("Parent message is not sent by bot")
		}
		resp, err := lark.LarkClient.Im.Message.Delete(ctx, larkim.NewDeleteMessageReqBuilder().MessageId(*data.Event.Message.ParentId).Build())
		if err != nil {
			return err
		}
		if !resp.Success() {
			return errors.New(resp.Error())
		}
	}
	return nil
}

func DebugRepeatHandler(ctx context.Context, data *larkim.P2MessageReceiveV1, metaData *handlerbase.BaseMetaData, args ...string) (err error) {
	ctx, span := otel.LarkRobotOtelTracer.Start(ctx, reflecting.GetCurrentFunc())
	span.SetAttributes(attribute.Key("event").String(larkcore.Prettify(data)))
	defer span.End()
	defer func() { span.RecordError(err) }()

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
		resp, err := lark.LarkClient.Im.V1.Message.Create(ctx, repeatReq)
		if err != nil {
			return err
		}
		if !resp.Success() {
			if strings.Contains(resp.Error(), "invalid image_key") {
				logs.L.Error(ctx, "repeat message error", "error", err, "traceID", span.SpanContext().TraceID().String())
				return nil
			}
			return errors.New(resp.Error())
		}
		larkutils.RecordMessage2Opensearch(ctx, resp)
	}
	return nil
}

func DebugImageHandler(ctx context.Context, data *larkim.P2MessageReceiveV1, metaData *handlerbase.BaseMetaData, args ...string) (err error) {
	ctx, span := otel.LarkRobotOtelTracer.Start(ctx, reflecting.GetCurrentFunc())
	span.SetAttributes(attribute.Key("event").String(larkcore.Prettify(data)))
	defer span.End()
	defer func() { span.RecordError(err) }()
	seq, err := larkimg.GetAllImgURLFromParent(ctx, data)
	if err != nil {
		return err
	}
	if seq == nil {
		return nil
	}
	urls := make([]string, 0)
	for url := range seq {
		// url = strings.ReplaceAll(url, "kmhomelab.cn", "kevinmatt.top")
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
	err = message.SendAndReplyStreamingCard(ctx, data.Event.Message, dataSeq, true)
	if err != nil {
		return err
	}
	return nil
}

func DebugConversationHandler(ctx context.Context, data *larkim.P2MessageReceiveV1, metaData *handlerbase.BaseMetaData, args ...string) (err error) {
	ctx, span := otel.LarkRobotOtelTracer.Start(ctx, reflecting.GetCurrentFunc())
	span.SetAttributes(attribute.Key("event").String(larkcore.Prettify(data)))
	defer span.End()
	defer func() { span.RecordError(err) }()

	msgs, err := larkutils.GetAllParentMsg(ctx, data)
	if err != nil {
		return err
	}

	resp, err := opensearchdal.SearchData(ctx, consts.LarkChunkIndex,
		map[string]any{
			"query": map[string]any{
				"bool": map[string]any{
					"must": map[string]any{
						"terms": map[string]any{
							"msg_ids": commonutils.TransSlice(msgs, func(msg *larkim.Message) string { return *msg.MessageId }),
						},
					},
				},
			},
			"sort": map[string]any{
				"timestamp": map[string]any{
					"order": "desc",
				},
			},
		})
	if err != nil {
		return err
	}
	for _, hit := range resp.Hits.Hits {
		chunkLog := &handlertypes.MessageChunkLogV3{}
		err = sonic.Unmarshal(hit.Source, chunkLog)
		if err != nil {
			return err
		}

		msgList, err := history.New(ctx).Query(
			osquery.Bool().Must(
				osquery.Terms("message_id", commonutils.TransSlice(chunkLog.MsgIDs, func(s string) any { return s })...),
			),
		).
			Source("raw_message", "mentions", "message_str", "create_time", "user_id", "chat_id", "user_name", "message_type").GetAll()
		if err != nil {
			return err
		}
		tpl := templates.GetTemplateV2(templates.ChunkMetaTemplate) // make sure template is loaded
		msgLines := commonutils.TransSlice(msgList, func(msg *handlertypes.MessageIndex) *templates.MsgLine {
			msgTrunc := make([]string, 0)
			for item := range larkmsgutils.Trans2Item(msg.MessageType, msg.RawMessage) {
				switch item.Tag {
				case "image", "sticker":
					msgTrunc = append(msgTrunc, fmt.Sprintf("![something](%s)", item.Content))
				case "text":
					msgTrunc = append(msgTrunc, item.Content)
				}
			}
			return &templates.MsgLine{
				Time:    msg.CreateTime,
				User:    &templates.User{ID: msg.UserID},
				Content: strings.Join(msgTrunc, " "),
			}
		})
		slices.SortFunc(msgLines, func(a, b *templates.MsgLine) int {
			return strings.Compare(a.Time, b.Time)
		})
		metaData := &templates.ChunkMetaData{
			Summary: chunkLog.Summary,

			Intent: chunking.Translate(chunkLog.Intent),
			Participants: Dedup(
				commonutils.TransSlice(msgList, func(m *handlertypes.MessageIndex) *templates.User { return &templates.User{ID: m.UserID} }),
				func(u *templates.User) string { return u.ID },
			),

			Sentiment: chunking.Translate(chunkLog.SentimentAndTone.Sentiment),
			Tones:     commonutils.TransSlice(chunkLog.SentimentAndTone.Tones, func(tone string) *templates.ToneData { return &templates.ToneData{Tone: chunking.Translate(tone)} }),
			Questions: commonutils.TransSlice(chunkLog.InteractionAnalysis.UnresolvedQuestions, func(question string) *templates.Questions { return &templates.Questions{Question: question} }),

			MsgList: msgLines,

			// PlansAndSuggestion: ,
			MainTopicsOrActivities:         commonutils.TransSlice(chunkLog.Entities.MainTopicsOrActivities, templates.ToObjTextArray),
			KeyConceptsAndNouns:            commonutils.TransSlice(chunkLog.Entities.KeyConceptsAndNouns, templates.ToObjTextArray),
			MentionedGroupsOrOrganizations: commonutils.TransSlice(chunkLog.Entities.MentionedGroupsOrOrganizations, templates.ToObjTextArray),
			MentionedPeople:                commonutils.TransSlice(chunkLog.Entities.MentionedPeople, templates.ToObjTextArray),
			LocationsAndVenues:             commonutils.TransSlice(chunkLog.Entities.LocationsAndVenues, templates.ToObjTextArray),
			MediaAndWorks: commonutils.TransSlice(chunkLog.Entities.MediaAndWorks, func(m *handlertypes.MediaAndWork) *templates.MediaAndWork {
				return &templates.MediaAndWork{m.Title, m.Type}
			}),

			Timestamp: chunkLog.Timestamp,
			MsgID:     *data.Event.Message.MessageId,
		}

		tpl.WithData(metaData)
		cardContent := templates.NewCardContentV2(ctx, tpl)
		err = larkutils.ReplyCard(ctx, cardContent, *data.Event.Message.MessageId, "_replyGet", false)
		if err != nil {
			return err
		}
	}

	return err
}

func Map[T any, U any](slice []T, f func(int, T) U) []U {
	result := make([]U, 0, len(slice))
	for idx, v := range slice {
		result = append(result, f(idx, v))
	}
	return result
}

func Dedup[T, K comparable](slice []T, keyFunc func(T) K) []T {
	seen := make(map[K]struct{})
	result := make([]T, 0, len(slice))
	for _, v := range slice {
		key := keyFunc(v)
		if _, ok := seen[key]; !ok {
			seen[key] = struct{}{}
			result = append(result, v)
		}
	}
	return result
}
