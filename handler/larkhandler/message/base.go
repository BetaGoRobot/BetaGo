package message

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/BetaGoRobot/BetaGo/consts"
	handlerbase "github.com/BetaGoRobot/BetaGo/handler/handler_base"
	handlertypes "github.com/BetaGoRobot/BetaGo/handler/handler_types"
	"github.com/BetaGoRobot/BetaGo/utility"
	"github.com/BetaGoRobot/BetaGo/utility/chunking"
	"github.com/BetaGoRobot/BetaGo/utility/doubao"
	"github.com/BetaGoRobot/BetaGo/utility/larkutils"
	"github.com/BetaGoRobot/BetaGo/utility/larkutils/larkmsgutils"
	"github.com/BetaGoRobot/BetaGo/utility/log"
	opensearchdal "github.com/BetaGoRobot/BetaGo/utility/opensearch_dal"
	"github.com/BetaGoRobot/BetaGo/utility/otel"
	"github.com/BetaGoRobot/go_utils/reflecting"
	"github.com/bytedance/sonic"
	"github.com/kevinmatthe/zaplog"
	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
	"github.com/yanyiwu/gojieba"
)

// Handler  消息处理器
var Handler = &handlerbase.Processor[larkim.P2MessageReceiveV1, handlerbase.BaseMetaData]{}

type LarkMessageEvent struct {
	*larkim.P2MessageReceiveV1
}

func (m *LarkMessageEvent) GroupID() (res string) {
	return *m.Event.Message.ChatId
}

func (m *LarkMessageEvent) MsgID() (res string) {
	return *m.Event.Message.MessageId
}

func (m *LarkMessageEvent) TimeStamp() (res int64) {
	t, err := strconv.ParseInt(*m.Event.Message.CreateTime, 10, 64)
	if err != nil {
		zaplog.Logger.Error("getTimestampFunc error", zaplog.Error(err))
		return time.Now().UnixMilli()
	}
	return t
}

func (m *LarkMessageEvent) BuildLine() (res string) {
	mentions := m.Event.Message.Mentions

	tmpList := make([]string, 0)
	for msgItem := range larkmsgutils.
		GetContentItemsSeq(
			&larkim.EventMessage{
				Content:     m.Event.Message.Content,
				MessageType: m.Event.Message.MessageType,
			},
		) {
		switch msgItem.Tag {
		case "at", "text":
			if msgItem.Tag == "text" {
				m := map[string]string{}
				if err := sonic.UnmarshalString(msgItem.Content, &m); err == nil {
					msgItem.Content = m["text"]
				}
			}
			if len(mentions) > 0 {
				for _, mention := range mentions {
					if mention.Key != nil {
						if *mention.Name == "不太正经的网易云音乐机器人" {
							*mention.Name = "你"
						}
						msgItem.Content = strings.ReplaceAll(msgItem.Content, *mention.Key, fmt.Sprintf("@%s", *mention.Name))
					}
				}
			}
			fallthrough
		default:
			content := strings.ReplaceAll(msgItem.Content, "\n", "<换行>")
			if strings.TrimSpace(content) != "" {
				tmpList = append(tmpList, content)
			}
		}
	}
	member, err := larkutils.GetUserMemberFromChat(context.Background(), *m.Event.Message.ChatId, *m.Event.Sender.SenderId.OpenId)
	if err != nil {
		log.Zlog.Error("got error openID", zaplog.String("openID", *m.Event.Sender.SenderId.OpenId))
	}
	userName := ""
	if member == nil {
		userName = "NULL"
	} else {
		userName = *member.Name
	}

	createTime := time.UnixMilli(m.TimeStamp()).Local().Format(time.DateTime)
	return fmt.Sprintf("[%s] <%s>: %s", createTime, userName, strings.Join(tmpList, ";"))
}

type (
	OpBase = handlerbase.OperatorBase[larkim.P2MessageReceiveV1, handlerbase.BaseMetaData]
	Op     = handlerbase.Operator[larkim.P2MessageReceiveV1, handlerbase.BaseMetaData]
)

func larkDeferFunc(ctx context.Context, err error, event *larkim.P2MessageReceiveV1, metaData *handlerbase.BaseMetaData) {
	larkutils.SendRecoveredMsg(ctx, err, *event.Event.Message.MessageId)
}

func CollectMessage(ctx context.Context, event *larkim.P2MessageReceiveV1, metaData *handlerbase.BaseMetaData) {
	go func() {
		ctx, span := otel.LarkRobotOtelTracer.Start(ctx, reflecting.GetCurrentFunc())
		defer span.End()

		chatID, err := larkutils.GetChatIDFromMsgID(ctx, *event.Event.Message.MessageId)
		if err != nil {
			return
		}
		member, err := larkutils.GetUserMemberFromChat(ctx, chatID, *event.Event.Sender.SenderId.OpenId)
		if err != nil {
			return
		}
		userName := ""
		if member == nil {
			userName = "NULL"
		} else {
			userName = *member.Name
		}
		msgLog := &handlertypes.MessageLog{
			MessageID:   utility.AddressORNil(event.Event.Message.MessageId),
			RootID:      utility.AddressORNil(event.Event.Message.RootId),
			ParentID:    utility.AddressORNil(event.Event.Message.ParentId),
			ChatID:      utility.AddressORNil(event.Event.Message.ChatId),
			ThreadID:    utility.AddressORNil(event.Event.Message.ThreadId),
			ChatType:    utility.AddressORNil(event.Event.Message.ChatType),
			MessageType: utility.AddressORNil(event.Event.Message.MessageType),
			UserAgent:   utility.AddressORNil(event.Event.Message.UserAgent),
			Mentions:    utility.MustMashal(event.Event.Message.Mentions),
			RawBody:     utility.MustMashal(event),
			Content:     utility.AddressORNil(event.Event.Message.Content),
			TraceID:     span.SpanContext().TraceID().String(),
		}

		content := larkutils.PreGetTextMsg(ctx, event)
		embedded, usage, err := doubao.EmbeddingText(ctx, content)
		if err != nil {
			log.Zlog.Error("EmbeddingText error", zaplog.Error(err))
		}
		jieba := gojieba.NewJieba()
		defer jieba.Free()
		ws := jieba.Cut(content, true)

		err = opensearchdal.InsertData(
			ctx, consts.LarkMsgIndex, *event.Event.Message.MessageId,
			&handlertypes.MessageIndex{
				MessageLog:      msgLog,
				ChatName:        larkutils.GetChatName(ctx, chatID),
				RawMessage:      content,
				RawMessageJieba: strings.Join(ws, " "),
				CreateTime:      utility.EpoMil2DateStr(*event.Event.Message.CreateTime),
				Message:         embedded,
				UserID:          *event.Event.Sender.SenderId.OpenId,
				UserName:        userName,
				TokenUsage:      usage,
				IsCommand:       metaData.IsCommand,
				MainCommand:     metaData.MainCommand,
			},
		)
		if err != nil {
			log.Zlog.Error("InsertData error", zaplog.Error(err))
		}
	}()
}

func init() {
	m := chunking.NewManagement[*LarkMessageEvent]()
	m.StartBackgroundCleaner(context.Background())
	Handler = Handler.
		OnPanic(larkDeferFunc).
		WithDefer(CollectMessage).
		WithDefer(func(ctx context.Context, event *larkim.P2MessageReceiveV1, _ *handlerbase.BaseMetaData) {
			m.SubmitMessage(ctx, &LarkMessageEvent{event})
		}).
		AddParallelStages(&RecordMsgOperator{}).
		AddParallelStages(&RepeatMsgOperator{}).
		AddParallelStages(&ReactMsgOperator{}).
		AddParallelStages(&WordReplyMsgOperator{}).
		AddParallelStages(&ReplyChatOperator{}).
		AddParallelStages(&CommandOperator{}).
		AddParallelStages(&ChatMsgOperator{})
}
