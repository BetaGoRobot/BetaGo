package reaction

import (
	"context"

	"github.com/BetaGoRobot/BetaGo/consts"
	handlerbase "github.com/BetaGoRobot/BetaGo/handler/handler_base"
	"github.com/BetaGoRobot/BetaGo/utility"
	"github.com/BetaGoRobot/BetaGo/utility/database"
	"github.com/BetaGoRobot/BetaGo/utility/larkutils"
	"github.com/BetaGoRobot/BetaGo/utility/otel"
	larkcore "github.com/larksuite/oapi-sdk-go/v3/core"
	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
	"go.opentelemetry.io/otel/attribute"
)

var _ handlerbase.Operator[larkim.P2MessageReactionCreatedV1] = &RecordReactionOperator{}

// RecordReactionOperator Repeat
//
//	@author heyuhengmatt
//	@update 2024-07-17 01:36:07
type RecordReactionOperator struct {
	handlerbase.OperatorBase[larkim.P2MessageReactionCreatedV1]
}

// PreRun Music
//
//	@receiver r *MusicMsgOperator
//	@param ctx context.Context
//	@param event *larkim.P2MessageReactionCreatedV1
//	@return err error
//	@author heyuhengmatt
//	@update 2024-07-17 01:34:09
func (r *RecordReactionOperator) PreRun(ctx context.Context, event *larkim.P2MessageReactionCreatedV1) (err error) {
	ctx, span := otel.LarkRobotOtelTracer.Start(ctx, utility.GetCurrentFunc())
	defer span.End()
	return
}

// Run  Repeat
//
//	@receiver r
//	@param ctx
//	@param event
//	@return err
func (r *RecordReactionOperator) Run(ctx context.Context, event *larkim.P2MessageReactionCreatedV1) (err error) {
	ctx, span := otel.LarkRobotOtelTracer.Start(ctx, utility.GetCurrentFunc())
	span.SetAttributes(attribute.Key("event").String(larkcore.Prettify(event)))
	defer span.End()
	defer span.RecordError(err)
	chatID, err := larkutils.GetChatIDFromMsgID(ctx, *event.Event.MessageId)
	if err != nil {
		return err
	}
	if *event.Event.OperatorType != "user" {
		return nil
	}
	member, err := larkutils.GetUserMemberFromChat(ctx, chatID, *event.Event.UserId.OpenId)
	if err != nil {
		return err
	}
	database.GetDbConnection().Create(&database.InteractionStats{
		OpenID:     *event.Event.UserId.OpenId,
		GuildID:    chatID,
		UserName:   *member.Name,
		ActionType: consts.LarkInteractionAddReaction,
	})

	return
}
