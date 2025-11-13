package reaction

import (
	"context"

	"github.com/BetaGoRobot/BetaGo/consts"
	handlerbase "github.com/BetaGoRobot/BetaGo/handler/handler_base"
	"github.com/BetaGoRobot/BetaGo/utility/database"
	"github.com/BetaGoRobot/BetaGo/utility/larkutils"
	"github.com/BetaGoRobot/BetaGo/utility/larkutils/grouputil"
	"github.com/BetaGoRobot/BetaGo/utility/logs"
	"github.com/BetaGoRobot/BetaGo/utility/otel"
	"github.com/BetaGoRobot/go_utils/reflecting"
	larkcore "github.com/larksuite/oapi-sdk-go/v3/core"
	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
	"go.opentelemetry.io/otel/attribute"
	"go.uber.org/zap"
)

var _ Op = &RecordReactionOperator{}

// RecordReactionOperator Repeat
//
//	@author heyuhengmatt
//	@update 2024-07-17 01:36:07
type RecordReactionOperator struct {
	OpBase
}

// PreRun Music
//
//	@receiver r *MusicMsgOperator
//	@param ctx context.Context
//	@param event *larkim.P2MessageReactionCreatedV1
//	@return err error
//	@author heyuhengmatt
//	@update 2024-07-17 01:34:09
func (r *RecordReactionOperator) PreRun(ctx context.Context, event *larkim.P2MessageReactionCreatedV1, meta *handlerbase.BaseMetaData) (err error) {
	ctx, span := otel.LarkRobotOtelTracer.Start(ctx, reflecting.GetCurrentFunc())
	defer span.End()
	defer func() { span.RecordError(err) }()
	return
}

// Run  Repeat
//
//	@receiver r
//	@param ctx
//	@param event
//	@return err
func (r *RecordReactionOperator) Run(ctx context.Context, event *larkim.P2MessageReactionCreatedV1, meta *handlerbase.BaseMetaData) (err error) {
	ctx, span := otel.LarkRobotOtelTracer.Start(ctx, reflecting.GetCurrentFunc())
	span.SetAttributes(attribute.Key("event").String(larkcore.Prettify(event)))
	defer span.End()
	defer func() { span.RecordError(err) }()
	defer span.RecordError(err)
	chatID, err := larkutils.GetChatIDFromMsgID(ctx, *event.Event.MessageId)
	if err != nil {
		return err
	}
	if *event.Event.OperatorType != "user" {
		return nil
	}
	member, err := grouputil.GetUserMemberFromChat(ctx, chatID, *event.Event.UserId.OpenId)
	if err != nil {
		return err
	}
	if member == nil || member.Name == nil {
		logs.L().Ctx(ctx).Error("user not found in chat", zap.String("TraceID", span.SpanContext().TraceID().String()), zap.String("OpenID", *event.Event.UserId.OpenId), zap.String("ChatID", chatID))
		return
	}
	database.GetDbConnection().Create(&database.InteractionStats{
		OpenID:     *event.Event.UserId.OpenId,
		GuildID:    chatID,
		UserName:   *member.Name,
		ActionType: consts.LarkInteractionAddReaction,
	})

	return
}
