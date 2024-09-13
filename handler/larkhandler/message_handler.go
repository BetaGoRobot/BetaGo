package larkhandler

import (
	"context"
	"strconv"
	"time"

	"github.com/BetaGoRobot/BetaGo/consts"
	"github.com/BetaGoRobot/BetaGo/handler/larkhandler/message"
	"github.com/BetaGoRobot/BetaGo/handler/larkhandler/reaction"
	"github.com/BetaGoRobot/BetaGo/utility"
	"github.com/BetaGoRobot/BetaGo/utility/database"
	"github.com/BetaGoRobot/BetaGo/utility/larkutils"
	"github.com/BetaGoRobot/BetaGo/utility/log"
	"github.com/BetaGoRobot/BetaGo/utility/otel"
	larkcore "github.com/larksuite/oapi-sdk-go/v3/core"
	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
	"go.opentelemetry.io/otel/attribute"
)

func isOutDated(createTime string) bool {
	stamp, err := strconv.ParseInt(createTime, 10, 64)
	if err != nil {
		panic(err)
	}
	return time.Now().Sub(time.UnixMilli(stamp)) > time.Second*10
}

// MessageV2Handler Repeat
//
//	@param ctx
//	@param event
//	@return error
func MessageV2Handler(ctx context.Context, event *larkim.P2MessageReceiveV1) error {
	ctx, span := otel.LarkRobotOtelTracer.Start(ctx, utility.GetCurrentFunc())
	defer larkutils.RecoverMsg(ctx, *event.Event.Message.MessageId)
	span.SetAttributes(attribute.Key("event").String(larkcore.Prettify(event)))
	defer span.End()

	if isOutDated(*event.Event.Message.CreateTime) {
		return nil
	}
	if *event.Event.Sender.SenderId.OpenId == consts.BotOpenID {
		return nil
	}
	go message.Handler.Clean().WithCtx(ctx).WithEvent(event).RunStages()
	go message.Handler.Clean().WithCtx(ctx).WithEvent(event).RunParallelStages()

	chatID, err := larkutils.GetChatIDFromMsgID(ctx, *event.Event.Message.MessageId)
	if err != nil {
		return err
	}
	member, err := larkutils.GetUserMemberFromChat(ctx, chatID, *event.Event.Sender.SenderId.OpenId)
	if err != nil {
		return err
	}
	database.GetDbConnection().Create(&database.InteractionStats{
		OpenID:     *event.Event.Sender.SenderId.OpenId,
		GuildID:    chatID,
		UserName:   *member.Name,
		ActionType: consts.LarkInteractionSendMsg,
	})

	database.GetDbConnection().Create(&database.MessageLog{
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
	})
	log.ZapLogger.Info(larkcore.Prettify(event))
	return nil
}

// MessageReactionHandler Repeat
//
//	@param ctx
//	@param event
//	@return error
func MessageReactionHandler(ctx context.Context, event *larkim.P2MessageReactionCreatedV1) error {
	ctx, span := otel.LarkRobotOtelTracer.Start(ctx, utility.GetCurrentFunc())
	defer larkutils.RecoverMsg(ctx, *event.Event.MessageId)
	defer span.End()

	go reaction.Handler.Clean().WithCtx(ctx).WithEvent(event).RunParallelStages()
	return nil
}
