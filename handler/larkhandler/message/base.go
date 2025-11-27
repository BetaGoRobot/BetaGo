package message

import (
	"context"
	"strings"
	"time"

	"github.com/BetaGoRobot/BetaGo/consts"
	handlerbase "github.com/BetaGoRobot/BetaGo/handler/handler_base"
	handlertypes "github.com/BetaGoRobot/BetaGo/handler/handler_types"
	"github.com/BetaGoRobot/BetaGo/utility"
	"github.com/BetaGoRobot/BetaGo/utility/ark"
	"github.com/BetaGoRobot/BetaGo/utility/larkutils"
	"github.com/BetaGoRobot/BetaGo/utility/larkutils/grouputil"
	"github.com/BetaGoRobot/BetaGo/utility/larkutils/larkchunking"
	"github.com/BetaGoRobot/BetaGo/utility/logs"
	opensearchdal "github.com/BetaGoRobot/BetaGo/utility/opensearch_dal"
	"github.com/BetaGoRobot/BetaGo/utility/otel"
	"github.com/BetaGoRobot/BetaGo/utility/retriver"
	"github.com/BetaGoRobot/go_utils/reflecting"
	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
	"github.com/tmc/langchaingo/schema"
	"github.com/yanyiwu/gojieba"
	"go.uber.org/zap"
)

// Handler  消息处理器
var Handler = &handlerbase.Processor[larkim.P2MessageReceiveV1, handlerbase.BaseMetaData]{}

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
		member, err := grouputil.GetUserMemberFromChat(ctx, chatID, *event.Event.Sender.SenderId.OpenId)
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
		embedded, usage, err := ark.EmbeddingText(ctx, content)
		if err != nil {
			logs.L().Ctx(ctx).Error("EmbeddingText error", zap.Error(err))
		}
		jieba := gojieba.NewJieba()
		defer jieba.Free()
		for _, mention := range event.Event.Message.Mentions {
			jieba.AddWord("@" + *mention.Name)
		}
		ws := jieba.Cut(content, true)
		wts := jieba.Tag(content)
		wsTags := []*handlertypes.WordWithTag{}
		for _, tag := range wts {
			sp := strings.Split(tag, "/")
			if sp[0] = strings.TrimSpace(sp[0]); sp[0] == "" {
				continue
			}
			wsTags = append(wsTags, &handlertypes.WordWithTag{Word: sp[0], Tag: sp[1]})
		}
		err = opensearchdal.InsertData(
			ctx, consts.LarkMsgIndex, *event.Event.Message.MessageId,
			&handlertypes.MessageIndex{
				MessageLog:           msgLog,
				ChatName:             larkutils.GetChatName(ctx, chatID),
				RawMessage:           content,
				RawMessageJieba:      strings.Join(ws, " "),
				RawMessageJiebaArray: ws,
				RawMessageJiebaTag:   wsTags,
				CreateTime:           utility.Epo2DateZoneMil(utility.MustInt(*event.Event.Message.CreateTime), time.UTC, time.DateTime),
				CreateTimeV2:         utility.Epo2DateZoneMil(utility.MustInt(*event.Event.Message.CreateTime), utility.UTCPlus8Loc(), time.RFC3339),
				Message:              embedded,
				UserID:               *event.Event.Sender.SenderId.OpenId,
				UserName:             userName,
				TokenUsage:           usage,
				IsCommand:            metaData.IsCommand,
				MainCommand:          metaData.MainCommand,
			},
		)
		if err != nil {
			logs.L().Ctx(ctx).Error("InsertData error", zap.Error(err))
		}
		err = retriver.Cli().AddDocuments(ctx, utility.AddressORNil(event.Event.Message.ChatId),
			[]schema.Document{{
				PageContent: content,
				Metadata: map[string]any{
					"chat_id":     utility.AddressORNil(event.Event.Message.ChatId),
					"user_id":     utility.AddressORNil(event.Event.Sender.SenderId.OpenId),
					"msg_id":      utility.AddressORNil(event.Event.Message.MessageId),
					"create_time": utility.EpoMil2DateStr(*event.Event.Message.CreateTime),
					"user_name":   userName,
				},
			}})
		if err != nil {
			logs.L().Ctx(ctx).Error("AddDocuments error", zap.Error(err))
		}
	}()
}

func init() {
	Handler = Handler.
		OnPanic(larkDeferFunc).
		WithDefer(CollectMessage).
		WithDefer(func(ctx context.Context, event *larkim.P2MessageReceiveV1, meta *handlerbase.BaseMetaData) {
			if !meta.IsCommand { // 过滤Command
				larkchunking.M.SubmitMessage(ctx, &larkchunking.LarkMessageEvent{P2MessageReceiveV1: event})
			}
		}).
		AddParallelStages(&RecordMsgOperator{}).
		AddParallelStages(&RepeatMsgOperator{}).
		AddParallelStages(&ReactMsgOperator{}).
		AddParallelStages(&WordReplyMsgOperator{}).
		AddParallelStages(&ReplyChatOperator{}).
		AddParallelStages(&CommandOperator{}).
		AddParallelStages(&ChatMsgOperator{})
}
