package crontask

import (
	"context"

	handlerbase "github.com/BetaGoRobot/BetaGo/handler/handler_base"
	"github.com/BetaGoRobot/BetaGo/utility/otel"
	"github.com/BetaGoRobot/go_utils/reflecting"
	larkcore "github.com/larksuite/oapi-sdk-go/v3/core"
	"go.opentelemetry.io/otel/attribute"
)

var _ Op = &CronTaskRunReactionOperator{}

// RecordReactionOperator Repeat
//
//	@author heyuhengmatt
//	@update 2024-07-17 01:36:07
type CronTaskRunReactionOperator struct {
	OpBase
}

// Run  Repeat
//
//	@receiver r
//	@param ctx
//	@param event
//	@return err
func (r *CronTaskRunReactionOperator) Run(ctx context.Context, event *CronTaskEvent, meta *handlerbase.BaseMetaData) (err error) {
	ctx, span := otel.LarkRobotOtelTracer.Start(ctx, reflecting.GetCurrentFunc())
	span.SetAttributes(attribute.Key("event").String(larkcore.Prettify(event)))
	defer span.End()
	defer func() { span.RecordError(err) }()
	defer span.RecordError(err)
	// chatID, err := larkutils.GetChatIDFromMsgID(ctx, *event.Event.MessageId)
	// if err != nil {
	// 	return err
	// }
	// if *event.Event.OperatorType != "user" {
	// 	return nil
	// }
	// member, err := grouputil.GetUserMemberFromChat(ctx, chatID, *event.Event.UserId.OpenId)
	// if err != nil {
	// 	return err
	// }
	// database.GetDbConnection().Create(&database.InteractionStats{
	// 	OpenID:     *event.Event.UserId.OpenId,
	// 	GuildID:    chatID,
	// 	UserName:   *member.Name,
	// 	ActionType: consts.LarkInteractionAddReaction,
	// })

	return
}
