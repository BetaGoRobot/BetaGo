package message

import (
	"context"

	"github.com/BetaGoRobot/BetaGo/consts"
	handlerbase "github.com/BetaGoRobot/BetaGo/handler/handler_base"
	handlertypes "github.com/BetaGoRobot/BetaGo/handler/handler_types"
	"github.com/BetaGoRobot/BetaGo/utility"
	"github.com/BetaGoRobot/BetaGo/utility/database"
	"github.com/BetaGoRobot/BetaGo/utility/doubao"
	"github.com/BetaGoRobot/BetaGo/utility/larkutils"
	"github.com/BetaGoRobot/BetaGo/utility/log"
	opensearchdal "github.com/BetaGoRobot/BetaGo/utility/opensearch_dal"
	"github.com/BetaGoRobot/BetaGo/utility/otel"
	"github.com/kevinmatthe/zaplog"
	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
)

// Handler  消息处理器
var Handler = &handlerbase.Processor[larkim.P2MessageReceiveV1, handlerbase.BaseMetaData]{}

type (
	OpBase = handlerbase.OperatorBase[larkim.P2MessageReceiveV1, handlerbase.BaseMetaData]
	Op     = handlerbase.Operator[larkim.P2MessageReceiveV1, handlerbase.BaseMetaData]
)

func larkDeferFunc(ctx context.Context, err error, event *larkim.P2MessageReceiveV1) {
	larkutils.SendRecoveredMsg(ctx, err, *event.Event.Message.MessageId)
}

func CollectMessage(ctx context.Context, event *larkim.P2MessageReceiveV1, metaData *handlerbase.BaseMetaData) {
	go func() {
		ctx, span := otel.LarkRobotOtelTracer.Start(ctx, utility.GetCurrentFunc())
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
		database.GetDbConnection().Create(&database.InteractionStats{
			OpenID:     *event.Event.Sender.SenderId.OpenId,
			GuildID:    chatID,
			UserName:   userName,
			ActionType: consts.LarkInteractionSendMsg,
		})
		msgLog := &database.MessageLog{
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
		database.GetDbConnection().Create(msgLog)
		content := larkutils.PreGetTextMsg(ctx, event)
		embedded, usage, err := doubao.EmbeddingText(ctx, content)
		if err != nil {
			log.ZapLogger.Error("EmbeddingText error", zaplog.Error(err))
		}
		err = opensearchdal.InsertData(
			ctx, "lark_msg_index", *event.Event.Message.MessageId,
			&handlertypes.MessageIndex{
				MessageLog:  msgLog,
				ChatName:    larkutils.GetChatName(ctx, chatID),
				RawMessage:  content,
				CreateTime:  utility.EpoMil2DateStr(*event.Event.Message.CreateTime),
				Message:     embedded,
				UserID:      *event.Event.Sender.SenderId.OpenId,
				UserName:    userName,
				TokenUsage:  usage,
				IsCommand:   metaData.IsCommand,
				MainCommand: metaData.MainCommand,
			},
		)
		if err != nil {
			log.ZapLogger.Error("InsertData error", zaplog.Error(err))
		}
		return
	}()
}

func init() {
	Handler = Handler.
		OnPanic(larkDeferFunc).
		WithDefer(CollectMessage).
		AddParallelStages(&RepeatMsgOperator{}).
		AddParallelStages(&ReactMsgOperator{}).
		AddParallelStages(&WordReplyMsgOperator{}).
		AddParallelStages(&MusicMsgOperator{}).
		AddParallelStages(&CommandOperator{})
}
